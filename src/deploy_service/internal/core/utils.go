package core

import (
	"apps-hosting.com/deployservice/internal/models"
	"apps-hosting.com/deployservice/proto/deploy_service_pb"
)

func DeploymentToProto(deployment *models.Deployment) *deploy_service_pb.Deployment {
	return &deploy_service_pb.Deployment{
		Id:        deployment.Id,
		BuildId:   deployment.BuildId,
		AppId:     deployment.AppId,
		Status:    string(deployment.Status),
		CreatedAt: deployment.CreatedAt.String(),
	}
}

func DeploymentListToProto(deployments []models.Deployment) []*deploy_service_pb.Deployment {
	_deployments := make([]*deploy_service_pb.Deployment, 0, len(deployments))
	for _, deployment := range deployments {
		_deployments = append(_deployments, DeploymentToProto(&deployment))
	}
	return _deployments
}

func ExtractDeploymentIDs(deployments []models.Deployment) []string {
	deploymentIDs := make([]string, 0, len(deployments))
	for _, deployment := range deployments {
		deploymentIDs = append(deploymentIDs, deployment.Id)
	}
	return deploymentIDs
}
