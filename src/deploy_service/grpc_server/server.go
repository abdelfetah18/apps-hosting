package grpc_server

import (
	"context"
	"deploy/proto/deploy_service_pb"
	"deploy/repositories"

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
	deployments, err := server.DeploymentRepository.GetDeployments(getDeploymentsRequest.AppId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	var _deployments []*deploy_service_pb.Deployment
	for _, deployment := range deployments {
		_deployment := deploy_service_pb.Deployment{
			Id:        deployment.Id,
			BuildId:   deployment.BuildId,
			AppId:     deployment.AppId,
			Status:    string(deployment.Status),
			CreatedAt: deployment.CreatedAt.String(),
		}
		_deployments = append(_deployments, &_deployment)
	}

	return &deploy_service_pb.GetDeploymentsResponse{
		Deployments: _deployments,
	}, nil
}
