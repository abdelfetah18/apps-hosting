package nats_service

import (
	"fmt"
	"os"

	gitclient "build/git_client"
	"build/kaniko"
	"build/repositories"
	"build/runtime"
	"build/utils"

	"apps-hosting.com/messaging"

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
	handler.Logger.LogInfo("HandleAppCreatedEvent")
	userAppLogger := logging.NewUserAppLogger(message.Data.AppId, message.Data.UserId, logging.StageBuild)

	handleBuildFailure := func(buildId string, err error) {
		messaging.PublishMessage(handler.JetStream,
			messaging.NewMessage(
				messaging.BuildFailed,
				messaging.BuildFailedData{
					AppId:   message.Data.AppId,
					AppName: message.Data.AppName,
					BuildId: buildId,
					Reason:  err.Error(),
				}))

		handler.BuildRepository.UpdateBuildById(
			message.Data.AppId,
			buildId,
			repositories.UpdateBuildParams{Status: repositories.BuildStatusFailed})
	}

	// 1. Create Build Entity
	handler.Logger.LogInfo("Create Build model")
	build, err := handler.BuildRepository.CreateBuild(message.Data.AppId, repositories.CreateBuildParams{Status: repositories.BuildStatusPending})
	if err != nil {
		handler.Logger.LogError(err.Error())
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
	err = utils.CompressTarGZ(gitRepo.Path, "/shared/repos/"+gitRepo.Id+".tar.gz")
	if err != nil {
		handler.Logger.LogError("Failed to create tar archive")
		handleBuildFailure(build.Id, err)
		return
	}

	// 5. Build & Push Docker image
	imageURL := registryURL + utils.ToImageName(message.Data.AppName)
	handler.Logger.LogInfoF("RunKanikoBuild: repoPath=%s, imageURL=%s\n", gitRepo.Id+".tar.gz", imageURL)
	handler.Logger.LogInfoF("URL=http://build-service:8081/repos/%s", gitRepo.Id+".tar.gz")
	err = handler.KanikoBuilder.RunKanikoBuild(gitRepo.Id+".tar.gz", message.Data.AppId, message.Data.AppName, imageURL)
	if err != nil {
		handler.Logger.LogError(err.Error())
		handleBuildFailure(build.Id, err)
		return
	}

	// 6. Update build status
	build, err = handler.BuildRepository.UpdateBuildById(message.Data.AppId, build.Id, repositories.UpdateBuildParams{
		Status:     repositories.BuildStatusSuccessed,
		ImageURL:   imageURL,
		CommitHash: gitRepo.LastCommitHash,
	})
	if err != nil {
		handler.Logger.LogError("Failed to update build status")
		return
	}

	_, err = messaging.PublishMessage(handler.JetStream,
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
		return
	}
}

func (handler *NatsHandler) HandleAppDeletedEvent(message messaging.Message[messaging.AppDeletedData]) {
	handler.Logger.LogInfo("HandleAppDeletedEvent")
	err := handler.BuildRepository.DeleteBuilds(message.Data.AppId)
	if err != nil {
		handler.Logger.LogError(err.Error())
		return
	}

	err = handler.KanikoBuilder.DeleteJobs(message.Data.AppName)
	if err != nil {
		handler.Logger.LogError(err.Error())
		return
	}
}
