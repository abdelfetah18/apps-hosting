package nats_service

import (
	"context"
	"fmt"
	"os"

	gitclient "build/git_client"
	"build/kaniko"
	"build/repositories"
	"build/runtime"
	"build/storage"
	"build/utils"

	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"
)

var registryURL = os.Getenv("REGISTRY_URL")

type NatsHandler struct {
	Logger          logging.ServiceLogger
	EventBus        messaging.EventBus
	KanikoBuilder   kaniko.KanikoBuilder
	GitRepoManager  gitclient.GitRepoManager
	RuntimeBuilder  runtime.RuntimeBuilder
	BuildRepository repositories.BuildRepository
}

func NewNatsHandler(
	eventBus messaging.EventBus,
	kanikoBuilder kaniko.KanikoBuilder,
	gitRepoManager gitclient.GitRepoManager,
	runtimeBuilder runtime.RuntimeBuilder,
	buildRepository repositories.BuildRepository,
	logger logging.ServiceLogger,
) NatsHandler {
	return NatsHandler{
		Logger:          logger,
		EventBus:        eventBus,
		KanikoBuilder:   kanikoBuilder,
		GitRepoManager:  gitRepoManager,
		RuntimeBuilder:  runtimeBuilder,
		BuildRepository: buildRepository,
	}
}

func (handler *NatsHandler) HandleAppCreatedEvent(ctx context.Context, message *events_pb.Message) {
	handler.Logger.LogInfo("Handle 'app.created' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetAppCreatedData()
	if data == nil {
		handler.Logger.LogError("Invalid app.created event message")
		span.SetAttributes(attribute.String("error", "Invalid app.created event message"))
		return
	}

	span.SetAttributes(
		attribute.String("user_id", data.UserId),
		attribute.String("app_id", data.AppId),
		attribute.String("app_name", data.AppName),
		attribute.String("domain_name", data.DomainName),
		attribute.String("runtime", data.Runtime),
		attribute.String("repo_url", data.RepoUrl),
		attribute.String("build_cmd", data.BuildCmd),
		attribute.String("start_cmd", data.StartCmd),
	)

	handleBuildFailure := func(buildId string, err error) {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		handler.EventBus.Publish(ctx, events_pb.EventName_BUILD_FAILED, &events_pb.EventData{
			Value: &events_pb.EventData_BuildFailedData{
				BuildFailedData: &events_pb.BuildFailedData{
					AppId:   data.AppId,
					BuildId: buildId,
					AppName: data.AppName,
					Reason:  err.Error(),
				},
			},
		})

		handler.BuildRepository.UpdateBuildById(ctx,
			data.AppId,
			buildId,
			repositories.UpdateBuildParams{Status: repositories.BuildStatusFailed})
	}

	userAppLogger := logging.NewUserAppLogger(data.AppId, data.UserId, logging.StageBuild)

	// 1. Create Build Entity
	handler.Logger.LogInfo("Creating build entity...")
	build, err := handler.BuildRepository.CreateBuild(ctx, data.AppId, repositories.CreateBuildParams{Status: repositories.BuildStatusPending})
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// 2. Clone Github Repository
	handler.Logger.LogInfo(fmt.Sprintf("Cloning github repository '%s'...", data.RepoUrl))
	gitRepo, err := handler.GitRepoManager.Clone(data.RepoUrl, userAppLogger)
	if err != nil {
		userAppLogger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 3. Copy Docker Image
	handler.Logger.LogInfo(fmt.Sprintf("Copy Dockerfile for the target runtime '%s' to repository path '%s'...", data.Runtime, gitRepo.Path))
	_, err = handler.RuntimeBuilder.CopyDockerfile(gitRepo.Path, data.Runtime)
	if err != nil {
		userAppLogger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 4. Create Tar Archive
	handler.Logger.LogInfo("Compressing the repository path to a tar archive...")
	err = utils.CompressTarGZ(gitRepo.Path, gitRepo.Id+".tar.gz")
	if err != nil {
		handler.Logger.LogError("Failed to create tar archive")
		handleBuildFailure(build.Id, err)
		return
	}

	// 5. Upload Tar Archive to Minio Storage
	handler.Logger.LogInfo("Uploading the tar archive to minio storage...")
	minioStorage := storage.NewMinioStorage(false)
	err = minioStorage.PutFile(gitRepo.Id+".tar.gz", gitRepo.Id+".tar.gz")
	if err != nil {
		handler.Logger.LogError("Failed to upload tar archive to minio storage")
		handleBuildFailure(build.Id, err)
		return
	}

	// 6. Build & Push Docker image
	imageURL := registryURL + utils.ToImageName(data.AppName)
	handler.Logger.LogInfoF("Running kaniko build job for image '%s'...", imageURL)
	err = handler.KanikoBuilder.RunKanikoBuild(gitRepo.Id+".tar.gz", data.AppId, data.AppName, imageURL)
	if err != nil {
		handler.Logger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 7. Update build status
	handler.Logger.LogInfo("Updating build entity status to success...")
	build, err = handler.BuildRepository.UpdateBuildById(ctx, data.AppId, build.Id, repositories.UpdateBuildParams{
		Status:     repositories.BuildStatusSuccessed,
		ImageURL:   imageURL,
		CommitHash: gitRepo.LastCommitHash,
	})
	if err != nil {
		handler.Logger.LogError("Failed to update build status.")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(
		attribute.String("build_id", build.Id),
		attribute.String("build_image_url", imageURL),
		attribute.String("build_commit_hash", gitRepo.LastCommitHash),
		attribute.String("build_created_at", build.CreatedAt.String()),
	)

	handler.Logger.LogInfo("Publishing 'build.completed' event...")
	err = handler.EventBus.Publish(ctx, events_pb.EventName_BUILD_COMPLETED, &events_pb.EventData{
		Value: &events_pb.EventData_BuildCompletedData{
			BuildCompletedData: &events_pb.BuildCompletedData{
				ImageUrl:   imageURL,
				AppName:    data.AppName,
				AppId:      data.AppId,
				BuildId:    build.Id,
				DomainName: data.DomainName,
			},
		},
	})
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}

func (handler *NatsHandler) HandleAppDeletedEvent(ctx context.Context, message *events_pb.Message) {
	handler.Logger.LogInfo("Handle 'app.deleted' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetAppDeletedData()
	if data == nil {
		handler.Logger.LogError("Invalid app deleted message")
		span.SetAttributes(attribute.String("error", "Invalid app deleted message"))
		return
	}

	span.SetAttributes(
		attribute.String("app_id", data.AppId),
		attribute.String("app_name", data.AppName),
	)

	handler.Logger.LogInfoF("Deleting all builds entities related to app '%s'", data.AppId)
	err := handler.BuildRepository.DeleteBuilds(ctx, data.AppId)
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	handler.Logger.LogInfoF("Deleting all builds jobs related to app '%s'", data.AppName)
	err = handler.KanikoBuilder.DeleteJobs(data.AppName)
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
