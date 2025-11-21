package eventshandlers

import (
	"app/repositories"
	"context"

	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"
)

type EventsHandlers struct {
	eventBus                       messaging.EventBus
	appRepository                  repositories.AppRepository
	environmentVariablesRepository repositories.EnvironmentVariablesRepository
	gitRepositoryRepository        repositories.GitRepositoryRepository
	logger                         logging.ServiceLogger
}

func NewEventsHandlers(
	eventBus messaging.EventBus,
	appRepository repositories.AppRepository,
	environmentVariablesRepository repositories.EnvironmentVariablesRepository,
	gitRepositoryRepository repositories.GitRepositoryRepository,
	logger logging.ServiceLogger,
) EventsHandlers {
	return EventsHandlers{
		eventBus:                       eventBus,
		appRepository:                  appRepository,
		environmentVariablesRepository: environmentVariablesRepository,
		gitRepositoryRepository:        gitRepositoryRepository,
		logger:                         logger,
	}
}

func (h *EventsHandlers) HandleProjectDeletedEvent(ctx context.Context, message *events_pb.Message) {
	h.logger.LogInfo("Handle 'project.deleted' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetProjectDeletedData()
	if data == nil {
		h.logger.LogError("Invalid project deleted message")
		span.SetAttributes(attribute.String("error", "Invalid project deleted message"))
		return
	}

	span.SetAttributes(attribute.String("project.id", data.ProjectId))

	apps, err := h.appRepository.GetApps(ctx, data.ProjectId)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	appIds := make([]string, len(apps))
	for i, app := range apps {
		appIds[i] = app.Id
	}

	if err := h.environmentVariablesRepository.DeleteEnvironmentVariablesByAppIds(ctx, appIds); err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	if err := h.gitRepositoryRepository.DeleteGitRepositoriessByAppIds(ctx, appIds); err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	if err := h.appRepository.DeleteAppsByProjectId(ctx, data.ProjectId); err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	for _, app := range apps {
		h.eventBus.Publish(ctx, events_pb.EventName_APP_DELETED, &events_pb.EventData{
			Value: &events_pb.EventData_AppDeletedData{
				AppDeletedData: &events_pb.AppDeletedEventData{
					AppId:   app.Id,
					AppName: app.Name,
				},
			},
		})
	}
}
