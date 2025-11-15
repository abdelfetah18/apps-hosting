package nats_service

import (
	"context"
	grpcclients "deploy/grpc_clients"
	"deploy/kubernetes"
	"deploy/proto/app_service_pb"
	"deploy/repositories"
	"deploy/utils"

	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"

	v1Core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

const (
	NameSpace = "default"
)

type NatsHandler struct {
	EventBus             messaging.EventBus
	GrpcClients          grpcclients.GrpcClients
	DeploymentRepository repositories.DeploymentRepository
	Logger               logging.ServiceLogger
}

func NewNatsHandler(
	eventBus messaging.EventBus,
	grpcClients grpcclients.GrpcClients,
	deploymentRepository repositories.DeploymentRepository,
	logger logging.ServiceLogger,
) NatsHandler {
	return NatsHandler{
		EventBus:             eventBus,
		GrpcClients:          grpcClients,
		DeploymentRepository: deploymentRepository,
		Logger:               logger,
	}
}

func (handler *NatsHandler) HandleBuildCompletedEvent(ctx context.Context, message *events_pb.Message) {
	handler.Logger.LogInfo("Handle 'build.completed' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetBuildCompletedData()
	if data == nil {
		handler.Logger.LogError("Invalid build completed message")
		span.SetAttributes(attribute.String("error", "Invalid build completed message"))
		return
	}

	span.SetAttributes(
		attribute.String("app_id", data.AppId),
		attribute.String("app_name", data.AppName),
		attribute.String("build_id", data.BuildId),
		attribute.String("domain_name", data.DomainName),
		attribute.String("image_url", data.ImageUrl),
	)

	handleDeploymentFailure := func(deploymentId string, err error) {
		handler.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		handler.DeploymentRepository.UpdateDeploymentById(
			ctx,
			deploymentId,
			repositories.UpdateDeploymentParams{Status: repositories.DeploymentStatusFailed},
		)
	}

	// 1. Create Deployment
	handler.Logger.LogInfo("Creating deployment entity...")
	deployment, err := handler.DeploymentRepository.CreateDeployment(ctx, data.BuildId, data.AppId, repositories.CreateDeploymentParams{
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
	handler.Logger.LogInfo("Loading cluster config...")
	config, err := rest.InClusterConfig()
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	handler.Logger.LogInfo("Creating kubernetes client...")
	kubernetesClient := kubernetes.NewKubernetesClient(config)

	// Get Environment Varaibels
	handler.Logger.LogInfo("Get environemnt variables for the target app...")
	getEnvironmentVariablesResponse, err := handler.GrpcClients.AppServiceClient.GetEnvironmentVariables(context.Background(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: data.AppId,
	})
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	if getEnvironmentVariablesResponse != nil {
		span.SetAttributes(
			attribute.String("environment_variable_id", getEnvironmentVariablesResponse.EnvironmentVariable.Id),
			attribute.String("environment_variable_value", getEnvironmentVariablesResponse.EnvironmentVariable.Value),
			attribute.String("environment_variable_created_at", getEnvironmentVariablesResponse.EnvironmentVariable.CreateAt),
		)
	}

	envVars := []v1Core.EnvVar{}
	if getEnvironmentVariablesResponse != nil {
		envVars, err = utils.ConvertJSONToEnvVars(getEnvironmentVariablesResponse.EnvironmentVariable.Value)
		if err != nil {
			handleDeploymentFailure(deployment.Id, err)
			return
		}
	}

	// FIXME: Maybe env vars should come from a config instead of passing them to deployment
	envVars = append(envVars, v1Core.EnvVar{Name: "NODE_ENV", Value: "production"})

	// Deploy the image to Kubernetes
	handler.Logger.LogInfo("Generating kubernetes deployment object...")
	deploymentObject := kubernetes.GenerateDeploymentObject(data.AppId, data.AppName, data.ImageUrl, envVars)
	err = kubernetesClient.DeployImage(NameSpace, data.ImageUrl, deploymentObject)
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	// Expose Service
	handler.Logger.LogInfo("Generating kubernetes service object...")
	serviceObject := kubernetes.GenerateServiceObject(NameSpace, data.AppId, data.AppName)
	_, err = kubernetesClient.CreateServiceForDeployment(NameSpace, serviceObject)
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	// Create Ingress Resource
	handler.Logger.LogInfo("Generating kubernetes ingress object...")
	ingressObject := kubernetes.GenerateIngressObject(NameSpace, data.AppId, data.AppName, data.DomainName, serviceObject.Name)
	err = kubernetesClient.CreateIngress(NameSpace, ingressObject.Name, ingressObject)
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	// Update Deployment
	handler.Logger.LogInfo("Updating deployment entity...")
	handler.DeploymentRepository.UpdateDeploymentById(ctx, deployment.Id, repositories.UpdateDeploymentParams{
		Status: repositories.DeploymentStatusSuccessed,
	})

	handler.Logger.LogInfo("Publishing 'deploy.completed' event...")
	handler.EventBus.Publish(ctx, events_pb.EventName_DEPLOY_COMPLETED, &events_pb.EventData{
		Value: &events_pb.EventData_DeployCompletedData{
			DeployCompletedData: &events_pb.DeployCompletedData{
				AppName:  data.AppName,
				DeployId: deployment.Id,
			},
		},
	})
}

func (handler *NatsHandler) HandleAppDeletedEvent(ctx context.Context, message *events_pb.Message) {
	handler.Logger.LogInfo("Handle 'app.deleted' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetAppDeletedData()
	if data == nil {
		handler.Logger.LogError("Invalid app deleted message")
		span.SetAttributes(attribute.String("error", "Invalid app deleted message"))
		return
	}

	span.SetAttributes(
		attribute.String("app_id", data.AppId),
		attribute.String("app_name", data.AppName),
	)

	handler.Logger.LogInfoF("Deleting deployments related to app with id '%s'", data.AppId)
	err := handler.DeploymentRepository.DeleteDeployments(ctx, data.AppId)
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

	handler.Logger.LogInfo("Deleting kubernetes deployments related to the app...")
	err = kubernetesClient.DeleteDeployment(NameSpace, data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes deployment: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	handler.Logger.LogInfo("Deleting kubernetes services related to the app...")
	err = kubernetesClient.DeleteServiceForDeployment(NameSpace, data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes service: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	handler.Logger.LogInfo("Deleting kubernetes ingress related to the app...")
	err = kubernetesClient.DeleteIngress(NameSpace, data.AppName)
	if err != nil {
		handler.Logger.LogErrorF("Failed to delete kubernetes ingress: %v\n", err)
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
