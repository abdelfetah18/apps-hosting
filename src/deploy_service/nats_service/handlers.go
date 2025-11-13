package nats_service

import (
	"context"
	grpcclients "deploy/grpc_clients"
	"deploy/kubernetes"
	"deploy/proto/app_service_pb"
	"deploy/repositories"
	"deploy/utils"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	ctx := message.Context
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", message.Data.AppId),
		attribute.String("app_name", message.Data.AppName),
		attribute.String("build_id", message.Data.BuildId),
		attribute.String("domain_name", message.Data.DomainName),
		attribute.String("image_url", message.Data.ImageURL),
		attribute.Int("duration", message.Data.Duration),
	)

	handler.Logger.LogInfo("HandleBuildCompletedEvent")

	// 1. Create Deployment
	deployment, err := handler.DeploymentRepository.CreateDeployment(ctx, message.Data.BuildId, message.Data.AppId, repositories.CreateDeploymentParams{
		Status: repositories.DeploymentStatusPending,
	})

	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(
		attribute.String("deployment_id", deployment.Id),
		attribute.String("deployment_created_at", deployment.CreatedAt.String()),
	)

	// Initialize Kubernetes client using utils
	// config, err := kubernetes.GetKubernetesConfigFromEnv()
	config, err := rest.InClusterConfig()
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	kubernetesClient := kubernetes.NewKubernetesClient(config)

	// Get Environment Varaibels
	getEnvironmentVariablesResponse, err := handler.GrpcClients.AppServiceClient.GetEnvironmentVariables(context.Background(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: message.Data.AppId,
	})
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	span.SetAttributes(
		attribute.String("environment_variable_id", getEnvironmentVariablesResponse.EnvironmentVariable.Id),
		attribute.String("environment_variable_value", getEnvironmentVariablesResponse.EnvironmentVariable.Value),
		attribute.String("environment_variable_created_at", getEnvironmentVariablesResponse.EnvironmentVariable.CreateAt),
	)

	envVars := []v1Core.EnvVar{}
	if getEnvironmentVariablesResponse != nil {
		envVars, err = utils.ConvertJSONToEnvVars(getEnvironmentVariablesResponse.EnvironmentVariable.Value)
		if err != nil {
			handler.Logger.LogError(err.Error())
			span.SetAttributes(attribute.String("error", err.Error()))
		}
	}

	// FIXME: Maybe env vars should come from a config instead of passing them to deployment
	envVars = append(envVars, v1Core.EnvVar{Name: "NODE_ENV", Value: "production"})

	// Deploy the image to Kubernetes
	deploymentObject := kubernetes.GenerateDeploymentObject(message.Data.AppId, message.Data.AppName, message.Data.ImageURL, envVars)
	err = kubernetesClient.DeployImage(NameSpace, message.Data.ImageURL, deploymentObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to deploy image: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// Expose Service
	serviceObject := kubernetes.GenerateServiceObject(NameSpace, message.Data.AppId, message.Data.AppName)
	_, err = kubernetesClient.CreateServiceForDeployment(NameSpace, serviceObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to create kubernetes service: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// Create Ingress Resource
	ingressObject := kubernetes.GenerateIngressObject(NameSpace, message.Data.AppId, message.Data.AppName, message.Data.DomainName, serviceObject.Name)
	err = kubernetesClient.CreateIngress(NameSpace, ingressObject.Name, ingressObject)
	if err != nil {
		handler.Logger.LogErrorF("Failed to create kubernetes ingress: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// Update Deployment
	handler.DeploymentRepository.UpdateDeploymentById(ctx, deployment.Id, repositories.UpdateDeploymentParams{
		Status: repositories.DeploymentStatusSuccessed,
	})

	messaging.PublishMessage(ctx, handler.JetStream, messaging.NewMessage(messaging.DeployCompleted, messaging.DeployCompletedData{
		AppName:  message.Data.AppName,
		DeployId: deployment.Id,
	}))
}

func (handler *NatsHandler) HandleAppDeletedEvent(message messaging.Message[messaging.AppDeletedData]) {
	ctx := message.Context
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", message.Data.AppId),
		attribute.String("app_name", message.Data.AppName),
	)

	err := handler.DeploymentRepository.DeleteDeployments(ctx, message.Data.AppId)
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// Initialize Kubernetes client using utils
	// config, err := kubernetes.GetKubernetesConfigFromEnv()
	config, err := rest.InClusterConfig()
	if err != nil {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	kubernetesClient := kubernetes.NewKubernetesClient(config)
	err = kubernetesClient.DeleteDeployment(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes deployment: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	err = kubernetesClient.DeleteServiceForDeployment(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes service: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	err = kubernetesClient.DeleteIngress(NameSpace, message.Data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes ingress: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
