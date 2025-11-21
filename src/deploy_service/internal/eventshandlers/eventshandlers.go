package eventshandlers

import (
	"context"

	"apps-hosting.com/deployservice/internal/deployer"
	"apps-hosting.com/deployservice/internal/models"
	"apps-hosting.com/deployservice/internal/repositories"
	"apps-hosting.com/deployservice/proto/app_service_pb"
	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1Core "k8s.io/api/core/v1"
)

const (
	NameSpace = "default"
)

type EventsHandlers struct {
	eventBus             messaging.EventBus
	appServiceClient     app_service_pb.AppServiceClient
	deploymentRepository repositories.DeploymentRepository
	logger               logging.ServiceLogger
}

func NewEventsHandlers(
	eventBus messaging.EventBus,
	appServiceClient app_service_pb.AppServiceClient,
	deploymentRepository repositories.DeploymentRepository,
	logger logging.ServiceLogger,
) EventsHandlers {
	return EventsHandlers{
		eventBus:             eventBus,
		appServiceClient:     appServiceClient,
		deploymentRepository: deploymentRepository,
		logger:               logger,
	}
}

func (h *EventsHandlers) HandleBuildCompletedEvent(ctx context.Context, message *events_pb.Message) {
	h.logger.LogInfo("Handle 'build.completed' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetBuildCompletedData()
	if data == nil {
		h.logger.LogError("Invalid build completed message")
		span.SetAttributes(attribute.String("error", "Invalid build completed message"))
		return
	}

	span.SetAttributes(
		attribute.String("app.id", data.AppId),
		attribute.String("build.id", data.BuildId),
	)

	handleDeploymentFailure := func(deploymentId string, err error) {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		h.deploymentRepository.UpdateDeploymentById(
			ctx,
			deploymentId,
			repositories.UpdateDeploymentParams{Status: models.DeploymentStatusFailed},
		)
	}

	// 1. Create Deployment
	h.logger.LogInfo("Creating deployment entity...")
	deployment, err := h.deploymentRepository.CreateDeployment(ctx, data.BuildId, data.AppId, repositories.CreateDeploymentParams{
		Status: models.DeploymentStatusPending,
	})
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("deployment.id", deployment.Id))

	// func GetKubernetesConfigFromEnv() (*rest.Config, error) {
	// 	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to create kube config: %v", err)
	// 	}
	// 	return config, nil
	// }

	// Initialize Kubernetes client using utils
	h.logger.LogInfo("Loading cluster config...")
	config, err := rest.InClusterConfig()
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	h.logger.LogInfo("Creating kubernetes client...")
	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	// Get Environment Varaibels
	h.logger.LogInfo("Get environemnt variables for the target app...")
	getEnvironmentVariablesResponse, err := h.appServiceClient.GetEnvironmentVariables(context.Background(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: data.AppId,
	})
	if err != nil {
		handleDeploymentFailure(deployment.Id, err)
		return
	}

	if getEnvironmentVariablesResponse != nil {
		span.SetAttributes(attribute.String("environment_variable.id", getEnvironmentVariablesResponse.EnvironmentVariable.Id))
	}

	envVars := []v1Core.EnvVar{}
	if getEnvironmentVariablesResponse != nil {
		envVars, err = ConvertJSONToEnvVars(getEnvironmentVariablesResponse.EnvironmentVariable.Value)
		if err != nil {
			handleDeploymentFailure(deployment.Id, err)
			return
		}
	}

	// FIXME: Maybe env vars should come from a config instead of passing them to deployment
	envVars = append(envVars, v1Core.EnvVar{Name: "NODE_ENV", Value: "production"})

	deployer := deployer.NewDeployer(kubernetesClient)

	err = deployer.Deploy(data.AppId, data.AppName, data.DomainName, data.ImageUrl, envVars)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		h.logger.LogError(err.Error())
		h.deploymentRepository.UpdateDeploymentById(
			ctx,
			deployment.Id,
			repositories.UpdateDeploymentParams{Status: models.DeploymentStatusFailed},
		)
		h.eventBus.Publish(ctx, events_pb.EventName_DEPLOY_FAILED, &events_pb.EventData{
			Value: &events_pb.EventData_DeployFailedData{
				DeployFailedData: &events_pb.DeployFailedData{
					AppId:        data.AppId,
					AppName:      data.AppId,
					BuildId:      data.BuildId,
					DeploymentId: deployment.Id,
					Reason:       err.Error(),
				},
			},
		})
		return
	}

	h.deploymentRepository.UpdateDeploymentById(ctx, deployment.Id, repositories.UpdateDeploymentParams{
		Status: models.DeploymentStatusSuccessed,
	})

	h.logger.LogInfo("Publishing 'deploy.completed' event...")
	h.eventBus.Publish(ctx, events_pb.EventName_DEPLOY_COMPLETED, &events_pb.EventData{
		Value: &events_pb.EventData_DeployCompletedData{
			DeployCompletedData: &events_pb.DeployCompletedData{
				AppName:  data.AppName,
				DeployId: deployment.Id,
			},
		},
	})
}

func (h *EventsHandlers) HandleAppDeletedEvent(ctx context.Context, message *events_pb.Message) {
	h.logger.LogInfo("Handle 'app.deleted' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetAppDeletedData()
	if data == nil {
		h.logger.LogError("Invalid app deleted message")
		span.SetAttributes(attribute.String("error", "Invalid app deleted message"))
		return
	}

	span.SetAttributes(attribute.String("app_id", data.AppId))

	h.logger.LogInfoF("Deleting deployments related to app with id '%s'", data.AppId)
	err := h.deploymentRepository.DeleteDeployments(ctx, data.AppId)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// Initialize Kubernetes client using utils
	// config, err := kubernetes.GetKubernetesConfigFromEnv()
	config, err := rest.InClusterConfig()
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	h.logger.LogInfo("Creating kubernetes client...")
	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	deployer := deployer.NewDeployer(kubernetesClient)
	err = deployer.Destroy(data.AppName)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
