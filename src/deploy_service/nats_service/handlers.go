package nats_service

import (
	"context"
	grpcclients "deploy/grpc_clients"
	"deploy/kubernetes"
	"deploy/proto/app_service_pb"
	"deploy/repositories"
	"deploy/utils"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/nats-io/nats.go"
	v1Core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

const (
	NameSpace = "default"
)

type NatsHandler struct {
	JetStream            nats.JetStream
	GrpcClients          grpcclients.GrpcClients
	DeploymentRepository repositories.DeploymentRepository
	Logger               logging.ServiceLogger
}

func NewNatsHandler(
	jetStream nats.JetStream,
	grpcClients grpcclients.GrpcClients,
	deploymentRepository repositories.DeploymentRepository,
	logger logging.ServiceLogger,
) NatsHandler {
	return NatsHandler{
		JetStream:            jetStream,
		GrpcClients:          grpcClients,
		DeploymentRepository: deploymentRepository,
		Logger:               logger,
	}
}

func (handler *NatsHandler) HandleBuildCompletedEvent(message messaging.Message[messaging.BuildCompletedData]) {
	handler.Logger.LogInfo("HandleBuildCompletedEvent")

	// 1. Create Deployment
	deployment, err := handler.DeploymentRepository.CreateDeployment(message.Data.BuildId, message.Data.AppId, repositories.CreateDeploymentParams{
		Status: repositories.DeploymentStatusPending,
	})

	if err != nil {
		handler.Logger.LogError(err.Error())
		return
	}

	// Initialize Kubernetes client using utils
	// config, err := kubernetes.GetKubernetesConfigFromEnv()
	config, err := rest.InClusterConfig()
	if err != nil {
		handler.Logger.LogError(err.Error())
	}

	kubernetesClient := kubernetes.NewKubernetesClient(config)

	// Get Environment Varaibels
	getEnvironmentVariablesResponse, err := handler.GrpcClients.AppServiceClient.GetEnvironmentVariables(context.Background(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: message.Data.AppId,
	})
	if err != nil {
		handler.Logger.LogError(err.Error())
	}

	envVars := []v1Core.EnvVar{}
	if getEnvironmentVariablesResponse != nil {
		envVars, err = utils.ConvertJSONToEnvVars(getEnvironmentVariablesResponse.EnvironmentVariable.Value)
		if err != nil {
			handler.Logger.LogError(err.Error())
		}
	}

	// FIXME: Maybe env vars should come from a config instead of passing them to deployment
	envVars = append(envVars, v1Core.EnvVar{Name: "NODE_ENV", Value: "production"})

	// Deploy the image to Kubernetes
	deploymentObject := kubernetes.GenerateDeploymentObject(message.Data.AppId, message.Data.AppName, message.Data.ImageURL, envVars)
	err = kubernetesClient.DeployImage(NameSpace, message.Data.ImageURL, deploymentObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to deploy image: %v\n", err)
		return
	}

	// Expose Service
	serviceObject := kubernetes.GenerateServiceObject(NameSpace, message.Data.AppId, message.Data.AppName)
	_, err = kubernetesClient.CreateServiceForDeployment(NameSpace, serviceObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to create kubernetes service: %v\n", err)
		return
	}

	// Create Ingress Resource
	ingressObject := kubernetes.GenerateIngressObject(NameSpace, message.Data.AppId, message.Data.AppName, message.Data.DomainName, serviceObject.Name)
	err = kubernetesClient.CreateIngress(NameSpace, ingressObject.Name, ingressObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to create kubernetes ingress: %v\n", err)
		return
	}

	// Update Deployment
	handler.DeploymentRepository.UpdateDeploymentById(deployment.Id, repositories.UpdateDeploymentParams{
		Status: repositories.DeploymentStatusSuccessed,
	})

	messaging.PublishMessage(handler.JetStream, messaging.NewMessage(messaging.DeployCompleted, messaging.DeployCompletedData{
		AppName:  message.Data.AppName,
		DeployId: deployment.Id,
	}))
}

func (handler *NatsHandler) HandleAppDeletedEvent(message messaging.Message[messaging.AppDeletedData]) {
	err := handler.DeploymentRepository.DeleteDeployments(message.Data.AppId)
	if err != nil {
		handler.Logger.LogError(err.Error())
		return
	}

	// Initialize Kubernetes client using utils
	// config, err := kubernetes.GetKubernetesConfigFromEnv()
	config, err := rest.InClusterConfig()
	if err != nil {
		handler.Logger.LogError(err.Error())
	}

	kubernetesClient := kubernetes.NewKubernetesClient(config)

	err = kubernetesClient.DeleteDeployment(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes deployment: %v\n", err)
		return
	}

	err = kubernetesClient.DeleteServiceForDeployment(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes service: %v\n", err)
		return
	}

	err = kubernetesClient.DeleteIngress(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes ingress: %v\n", err)
		return
	}
}
