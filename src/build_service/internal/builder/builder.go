package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"apps-hosting.com/buildservice/internal/buildexecutor"
	"apps-hosting.com/buildservice/internal/models"
	"apps-hosting.com/buildservice/internal/repomanager"
	"apps-hosting.com/buildservice/internal/storage"
	"apps-hosting.com/buildservice/proto/user_service_pb"
	"apps-hosting.com/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var runtimeDockerFilesPaths = map[string]string{
	"NodeJS": "assets/runtime/NodeJS.Dockerfile",
}

type Builder struct {
	gitRepoManager    repomanager.GitRepoManager
	buildExecutor     buildexecutor.BuildExecutor
	userServiceClient user_service_pb.UserServiceClient

	serviceLogger logging.ServiceLogger
	userAppLogger logging.UserAppLogger
}

func NewBuilder(
	gitRepoManager repomanager.GitRepoManager,
	buildExecutor buildexecutor.BuildExecutor,
	userServiceClient user_service_pb.UserServiceClient,
	serviceLogger logging.ServiceLogger,
	userAppLogger logging.UserAppLogger,
) *Builder {
	return &Builder{
		gitRepoManager:    gitRepoManager,
		buildExecutor:     buildExecutor,
		userServiceClient: userServiceClient,
		serviceLogger:     serviceLogger,
		userAppLogger:     userAppLogger,
	}
}

func (b *Builder) CloneGitRepository(ctx context.Context, userId string, cloneUrl string, isPrivate bool) (*repomanager.GitRepo, error) {
	span := trace.SpanFromContext(ctx)

	getGithubUserAccessTokenResponse, err := b.userServiceClient.
		GetGithubUserAccessToken(
			ctx,
			&user_service_pb.GetGithubUserAccessTokenRequest{
				UserId: userId,
			},
		)
	if err != nil {
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	b.serviceLogger.LogInfo(fmt.Sprintf("Cloning github repository '%s'...", cloneUrl))
	gitRepo, err := b.gitRepoManager.Clone(
		cloneUrl,
		isPrivate,
		getGithubUserAccessTokenResponse.GithubUserAccessToken,
		b.userAppLogger,
	)
	if err != nil {
		b.userAppLogger.LogError(err.Error())
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	return gitRepo, nil
}

func (b *Builder) PrepareSourceCode(ctx context.Context, runtime, gitRepositoryFilename, gitRepositoryPath string) error {
	span := trace.SpanFromContext(ctx)

	// Copy Docker Image
	b.serviceLogger.LogInfo(fmt.Sprintf("Copy Dockerfile for the target runtime '%s' to repository path '%s'...", runtime, gitRepositoryPath))
	_, err := b.AddDockerfile(gitRepositoryPath, runtime)
	if err != nil {
		b.userAppLogger.LogError(err.Error())
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	// Create Tar Archive
	b.serviceLogger.LogInfo("Compressing the repository path to a tar archive...")
	err = CompressTarGZ(gitRepositoryPath, gitRepositoryFilename)
	if err != nil {
		b.serviceLogger.LogError("Failed to create tar archive")
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	// Upload Tar Archive to Minio Storage
	b.serviceLogger.LogInfo("Uploading the tar archive to minio storage...")
	minioStorage := storage.NewMinioStorage(false)
	err = minioStorage.PutFile(gitRepositoryFilename, gitRepositoryFilename)
	if err != nil {
		b.serviceLogger.LogError("Failed to upload tar archive to minio storage")
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	return nil
}

func (b *Builder) BuildAndPushDockerImage(ctx context.Context, appId, appName, repositoryFileName string) (*string, error) {
	span := trace.SpanFromContext(ctx)
	registryURL := os.Getenv("REGISTRY_URL")
	imageURL := registryURL + buildexecutor.ToImageName(appName)
	srcContext := fmt.Sprintf("s3://apps-source/%s", repositoryFileName)

	b.serviceLogger.LogInfoF("Running kaniko build job for image '%s'...", imageURL)
	err := b.buildExecutor.Execute(srcContext, imageURL, appId, appName)
	if err != nil {
		b.serviceLogger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	return &imageURL, nil
}

func (b *Builder) AddDockerfile(repoPath string, runtime string) (string, error) {
	src, exists := runtimeDockerFilesPaths[runtime]
	if !exists {
		return "", fmt.Errorf("unsupported runtime: %s", runtime)
	}

	dest := filepath.Join(repoPath, "Dockerfile")

	err := CopyFile(src, dest)
	if err != nil {
		return "", err
	}

	return dest, nil
}

func (b *Builder) StartBuilding(ctx context.Context, userId, appId, appName, appRuntime, cloneURL string, isPrivate bool) (*models.Build, error) {
	repository, err := b.CloneGitRepository(ctx, userId, cloneURL, isPrivate)
	if err != nil {
		return nil, err
	}

	repositoryFileName := fmt.Sprintf("%s.tar.gz", repository.Id)
	err = b.PrepareSourceCode(ctx, appRuntime, repositoryFileName, repository.Path)
	if err != nil {
		return nil, err
	}

	imageUrl, err := b.BuildAndPushDockerImage(ctx, appId, appName, repositoryFileName)
	if err != nil {
		return nil, err
	}

	return &models.Build{
		Status:     models.BuildStatusSuccessed,
		ImageURL:   *imageUrl,
		CommitHash: repository.LastCommitHash,
	}, nil
}
