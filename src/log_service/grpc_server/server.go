package grpc_server

import (
	"context"
	"log_service/kubernetes"
	"log_service/proto/log_service_pb"

	"apps-hosting.com/logging"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCLogServiceServer struct {
	log_service_pb.UnimplementedLogServiceServer

	KubernetesClient kubernetes.KubernetesClient
	Logger           logging.ServiceLogger
}

func NewGRPCLogServiceServer(kubernetesClient kubernetes.KubernetesClient, logger logging.ServiceLogger) *GRPCLogServiceServer {
	return &GRPCLogServiceServer{
		KubernetesClient: kubernetesClient,
		Logger:           logger,
	}
}

func (server *GRPCLogServiceServer) Health(ctx context.Context, _ *log_service_pb.HealthRequest) (*log_service_pb.HealthResponse, error) {
	return &log_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCLogServiceServer) QueryLogs(ctx context.Context, queryLogsRequest *log_service_pb.QueryLogsRequest) (*log_service_pb.QueryLogsResponse, error) {
	logs, err := server.KubernetesClient.ReadPodLogs(queryLogsRequest.AppId)
	if err != nil {
		server.Logger.LogError(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &log_service_pb.QueryLogsResponse{
		Logs: logs,
	}, nil
}
