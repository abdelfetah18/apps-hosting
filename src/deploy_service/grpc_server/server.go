package grpc_server

import (
	"context"
	"deploy/proto/deploy_service_pb"
	"deploy/repositories"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCDeployServiceServer struct {
	deploy_service_pb.UnimplementedDeployServiceServer

	DeploymentRepository repositories.DeploymentRepository
}

func NewGRPCDeployServiceServer(deploymentRepository repositories.DeploymentRepository) *GRPCDeployServiceServer {
	return &GRPCDeployServiceServer{
		DeploymentRepository: deploymentRepository,
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

	deployments, err := server.DeploymentRepository.GetDeployments(ctx, getDeploymentsRequest.AppId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	var _deployments []*deploy_service_pb.Deployment
	deployments_ids := []string{}
	for _, deployment := range deployments {
		_deployment := deploy_service_pb.Deployment{
			Id:        deployment.Id,
			BuildId:   deployment.BuildId,
			AppId:     deployment.AppId,
			Status:    string(deployment.Status),
			CreatedAt: deployment.CreatedAt.String(),
		}
		_deployments = append(_deployments, &_deployment)
		deployments_ids = append(deployments_ids, deployment.Id)
	}

	span.SetAttributes(
		attribute.StringSlice("deployments_ids", deployments_ids),
		attribute.Int("deployments_count", len(deployments)),
	)

	return &deploy_service_pb.GetDeploymentsResponse{
		Deployments: _deployments,
	}, nil
}
