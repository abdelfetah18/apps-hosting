package nats_service

import (
	"fmt"
	"os"

	gitclient "build/git_client"
	"build/kaniko"
	"build/repositories"
	"build/runtime"
	"build/storage"
	"build/utils"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"

	"github.com/nats-io/nats.go"
)

var registryURL = os.Getenv("REGISTRY_URL")

type NatsHandler struct {
	Logger          logging.ServiceLogger
	JetStream       nats.JetStream
	KanikoBuilder   kaniko.KanikoBuilder
	GitRepoManager  gitclient.GitRepoManager
	RuntimeBuilder  runtime.RuntimeBuilder
	BuildRepository repositories.BuildRepository
}

func NewNatsHandler(
	jetStream nats.JetStream,
	kanikoBuilder kaniko.KanikoBuilder,
	gitRepoManager gitclient.GitRepoManager,
	runtimeBuilder runtime.RuntimeBuilder,
	buildRepository repositories.BuildRepository,
	logger logging.ServiceLogger,
) NatsHandler {
	return NatsHandler{
		Logger:          logger,
		JetStream:       jetStream,
		KanikoBuilder:   kanikoBuilder,
		GitRepoManager:  gitRepoManager,
		RuntimeBuilder:  runtimeBuilder,
		BuildRepository: buildRepository,
	}
}

func (handler *NatsHandler) HandleAppCreatedEvent(message messaging.Message[messaging.AppCreatedData]) {
	ctx := message.Context
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("user_id", message.Data.UserId),
		attribute.String("app_id", message.Data.AppId),
		attribute.String("app_name", message.Data.AppName),
		attribute.String("domain_name", message.Data.DomainName),
		attribute.String("runtime", message.Data.Runtime),
		attribute.String("repo_url", message.Data.RepoURL),
		attribute.String("build_cmd", message.Data.BuildCMD),
		attribute.String("start_cmd", message.Data.StartCMD),
	)

	handler.Logger.LogInfo("HandleAppCreatedEvent")

	userAppLogger := logging.NewUserAppLogger(message.Data.AppId, message.Data.UserId, logging.StageBuild)

	handleBuildFailure := func(buildId string, err error) {
		span.SetAttributes(attribute.String("error", err.Error()))
		messaging.PublishMessage(ctx, handler.JetStream,
			messaging.NewMessage(
				messaging.BuildFailed,
				messaging.BuildFailedData{
					AppId:   message.Data.AppId,
					AppName: message.Data.AppName,
					BuildId: buildId,
					Reason:  err.Error(),
				}))

		handler.BuildRepository.UpdateBuildById(ctx,
			message.Data.AppId,
			buildId,
			repositories.UpdateBuildParams{Status: repositories.BuildStatusFailed})
	}

	// 1. Create Build Entity
	handler.Logger.LogInfo("Create Build model")
	build, err := handler.BuildRepository.CreateBuild(ctx, message.Data.AppId, repositories.CreateBuildParams{Status: repositories.BuildStatusPending})
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// 2. Clone Github Repository
	handler.Logger.LogInfo(fmt.Sprintf("CloneRepo: repoURL=%s", message.Data.RepoURL))
	gitRepo, err := handler.GitRepoManager.Clone(message.Data.RepoURL, userAppLogger)
	if err != nil {
		userAppLogger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 3. Copy Docker Image
	handler.Logger.LogInfo(fmt.Sprintf("CopyDockerfile: repoPath=%s, runtime=%s\n", gitRepo.Path, message.Data.Runtime))
	_, err = handler.RuntimeBuilder.CopyDockerfile(gitRepo.Path, message.Data.Runtime)
	if err != nil {
		userAppLogger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 4. Create Tar Archive
	err = utils.CompressTarGZ(gitRepo.Path, gitRepo.Id+".tar.gz")
	if err != nil {
		handler.Logger.LogError("Failed to create tar archive")
		handleBuildFailure(build.Id, err)
		return
	}

	// 5. Upload Tar Archive to Minio Storage
	minioStorage := storage.NewMinioStorage(false)
	err = minioStorage.PutFile(gitRepo.Id+".tar.gz", gitRepo.Id+".tar.gz")
	if err != nil {
		handler.Logger.LogError("Failed to upload tar archive to minio bucket")
		handleBuildFailure(build.Id, err)
		return
	}

	// 6. Build & Push Docker image
	imageURL := registryURL + utils.ToImageName(message.Data.AppName)
	handler.Logger.LogInfoF("RunKanikoBuild: repoPath=%s, imageURL=%s\n", gitRepo.Id+".tar.gz", imageURL)
	handler.Logger.LogInfoF("URL=http://build-service:8081/repos/%s", gitRepo.Id+".tar.gz")
	err = handler.KanikoBuilder.RunKanikoBuild(gitRepo.Id+".tar.gz", message.Data.AppId, message.Data.AppName, imageURL)
	if err != nil {
		handler.Logger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 7. Update build status
	build, err = handler.BuildRepository.UpdateBuildById(ctx, message.Data.AppId, build.Id, repositories.UpdateBuildParams{
		Status:     repositories.BuildStatusSuccessed,
		ImageURL:   imageURL,
		CommitHash: gitRepo.LastCommitHash,
	})
	if err != nil {
		handler.Logger.LogError("Failed to update build status")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(
		attribute.String("build_id", build.Id),
		attribute.String("build_image_url", imageURL),
		attribute.String("build_commit_hash", gitRepo.LastCommitHash),
		attribute.String("build_created_at", build.CreatedAt.String()),
	)

	_, err = messaging.PublishMessage(ctx, handler.JetStream,
		messaging.NewMessage(
			messaging.BuildCompleted,
			messaging.BuildCompletedData{
				ImageURL:   imageURL,
				AppName:    message.Data.AppName,
				AppId:      message.Data.AppId,
				BuildId:    build.Id,
				DomainName: message.Data.DomainName,
			}))
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}

func (handler *NatsHandler) HandleAppDeletedEvent(message messaging.Message[messaging.AppDeletedData]) {
	ctx := message.Context
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", message.Data.AppId),
		attribute.String("app_name", message.Data.AppName),
	)

	handler.Logger.LogInfo("HandleAppDeletedEvent")
	err := handler.BuildRepository.DeleteBuilds(ctx, message.Data.AppId)
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	err = handler.KanikoBuilder.DeleteJobs(message.Data.AppName)
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
