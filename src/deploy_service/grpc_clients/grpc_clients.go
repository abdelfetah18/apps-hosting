package grpcclients

import (
	"deploy/proto/app_service_pb"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	AppServiceClient app_service_pb.AppServiceClient
}

func NewGrpcClients() (*GrpcClients, error) {
	appServiceClient, err := grpc.NewClient(
		os.Getenv("APP_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		return nil, err
	}

	return &GrpcClients{
		AppServiceClient: app_service_pb.NewAppServiceClient(appServiceClient),
	}, nil
}
