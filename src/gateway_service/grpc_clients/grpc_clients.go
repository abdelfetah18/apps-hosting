package grpcclients

import (
	"gateway/proto/app_service_pb"
	"gateway/proto/build_service_pb"
	"gateway/proto/deploy_service_pb"
	"gateway/proto/log_service_pb"
	"gateway/proto/project_service_pb"
	"gateway/proto/user_service_pb"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	UserServiceClient    user_service_pb.UserServiceClient
	AppServiceClient     app_service_pb.AppServiceClient
	BuildServiceClient   build_service_pb.BuildServiceClient
	DeployServiceClient  deploy_service_pb.DeployServiceClient
	LogServiceClient     log_service_pb.LogServiceClient
	ProjectServiceClient project_service_pb.ProjectServiceClient
}

func NewGrpcClients() (*GrpcClients, error) {
	userServiceClient, err := grpc.NewClient(
		os.Getenv("USER_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	appServiceClient, err := grpc.NewClient(
		os.Getenv("APP_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	buildServiceClient, err := grpc.NewClient(
		os.Getenv("BUILD_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	deployServiceClient, err := grpc.NewClient(
		os.Getenv("DEPLOY_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	logServiceClient, err := grpc.NewClient(
		os.Getenv("LOG_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	projectServiceClient, err := grpc.NewClient(
		os.Getenv("PROJECT_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	return &GrpcClients{
		UserServiceClient:    user_service_pb.NewUserServiceClient(userServiceClient),
		AppServiceClient:     app_service_pb.NewAppServiceClient(appServiceClient),
		BuildServiceClient:   build_service_pb.NewBuildServiceClient(buildServiceClient),
		DeployServiceClient:  deploy_service_pb.NewDeployServiceClient(deployServiceClient),
		LogServiceClient:     log_service_pb.NewLogServiceClient(logServiceClient),
		ProjectServiceClient: project_service_pb.NewProjectServiceClient(projectServiceClient),
	}, nil
}
