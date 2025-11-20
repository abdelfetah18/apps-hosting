package core

import (
	"context"

	"apps-hosting.com/deployservice/internal/repositories"
	"apps-hosting.com/deployservice/proto/deploy_service_pb"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCDeployServiceServer struct {
	deploy_service_pb.UnimplementedDeployServiceServer

	deploymentRepository repositories.DeploymentRepository
}

func NewGRPCDeployServiceServer(deploymentRepository repositories.DeploymentRepository) *GRPCDeployServiceServer {
	return &GRPCDeployServiceServer{
		deploymentRepository: deploymentRepository,
	}
}

func (server *GRPCDeployServiceServer) Health(ctx context.Context, _ *deploy_service_pb.HealthRequest) (*deploy_service_pb.HealthResponse, error) {
	return &deploy_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCDeployServiceServer) GetDeployments(ctx context.Context, getDeploymentsRequest *deploy_service_pb.GetDeploymentsRequest) (*deploy_service_pb.GetDeploymentsResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", getDeploymentsRequest.AppId),
	)

	deployments, err := server.deploymentRepository.GetDeployments(ctx, getDeploymentsRequest.AppId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_deployments := DeploymentListToProto(deployments)
	deployments_ids := ExtractDeploymentIDs(deployments)

	span.SetAttributes(
		attribute.StringSlice("deployments_ids", deployments_ids),
		attribute.Int("deployments_count", len(deployments)),
	)

	return &deploy_service_pb.GetDeploymentsResponse{
		Deployments: _deployments,
	}, nil
}
