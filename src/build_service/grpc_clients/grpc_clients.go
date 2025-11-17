package grpcclients

import (
	"build/proto/user_service_pb"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	UserServiceClient user_service_pb.UserServiceClient
}

func NewGrpcClients() (*GrpcClients, error) {
	userServiceClient, err := grpc.NewClient(
		os.Getenv("USER_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &GrpcClients{
		UserServiceClient: user_service_pb.NewUserServiceClient(userServiceClient),
	}, nil
}
