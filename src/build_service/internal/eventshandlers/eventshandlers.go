package eventshandlers

import (
	"context"

	"apps-hosting.com/buildservice/internal/builder"
	"apps-hosting.com/buildservice/internal/buildexecutor"
	"apps-hosting.com/buildservice/internal/models"
	"apps-hosting.com/buildservice/internal/repomanager"
	"apps-hosting.com/buildservice/internal/repositories"
	"apps-hosting.com/buildservice/proto/user_service_pb"

	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"
)

type EventsHandlers struct {
	eventBus          messaging.EventBus
	buildExecutor     buildexecutor.BuildExecutor
	gitRepoManager    repomanager.GitRepoManager
	buildRepository   repositories.BuildRepository
	userServiceClient user_service_pb.UserServiceClient
	logger            logging.ServiceLogger
}

func NewEventsHandlers(
	eventBus messaging.EventBus,
	buildExecutor buildexecutor.BuildExecutor,
	gitRepoManager repomanager.GitRepoManager,
	buildRepository repositories.BuildRepository,
	userServiceClient user_service_pb.UserServiceClient,
	logger logging.ServiceLogger,
) EventsHandlers {
	return EventsHandlers{
		eventBus:          eventBus,
		buildExecutor:     buildExecutor,
		gitRepoManager:    gitRepoManager,
		buildRepository:   buildRepository,
		userServiceClient: userServiceClient,
		logger:            logger,
	}
}

func (h *EventsHandlers) HandleAppCreatedEvent(ctx context.Context, message *events_pb.Message) {
	h.logger.LogInfo("Handle 'app.created' event")
	span := trace.SpanFromContext(ctx)

	data := message.Data.GetAppCreatedData()
	if data == nil {
		h.logger.LogError("Invalid app.created event message")
		span.SetAttributes(attribute.String("error", "Invalid app.created event message"))
		return
	}

	span.SetAttributes(
		attribute.String("app.id", data.App.Id),
		attribute.String("project.id", data.App.ProjectId),
		attribute.String("git_repository.id", data.GitRepository.Id),
	)

	// Create Build Entity
	h.logger.LogInfo("Creating build entity...")
	build, err := h.buildRepository.CreateBuild(
		ctx,
		data.App.Id,
		repositories.CreateBuildParams{
			Status: models.BuildStatusPending,
		},
	)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(
		attribute.String("build.id", build.Id),
	)

	userAppLogger := logging.NewUserAppLogger(data.App.Id, data.UserId, logging.StageBuild)
	builder := builder.NewBuilder(
		h.gitRepoManager,
		h.buildExecutor,
		h.userServiceClient,
		h.logger,
		userAppLogger,
	)

	buildResult, err := builder.StartBuilding(
		ctx,
		data.UserId,
		data.App.Id,
		data.App.Name,
		data.App.Runtime,
		data.GitRepository.CloneUrl,
		data.GitRepository.IsPrivate,
	)

	if err != nil {
		h.eventBus.Publish(ctx, events_pb.EventName_BUILD_FAILED, &events_pb.EventData{
			Value: &events_pb.EventData_BuildFailedData{
				BuildFailedData: &events_pb.BuildFailedData{
					AppId:   data.App.Id,
					BuildId: build.Id,
					AppName: data.App.Name,
					Reason:  err.Error(),
				},
			},
		})

		h.buildRepository.UpdateBuildById(
			ctx,
			data.App.Id,
			build.Id,
			repositories.UpdateBuildParams{Status: models.BuildStatusFailed},
		)
		return
	}

	build, err = h.buildRepository.UpdateBuildById(ctx, data.App.Id, build.Id, repositories.UpdateBuildParams{
		Status:     buildResult.Status,
		ImageURL:   buildResult.ImageURL,
		CommitHash: buildResult.CommitHash,
	})
	if err != nil {
		h.logger.LogError("Failed to update build status.")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	h.logger.LogInfo("Publishing 'build.completed' event...")
	err = h.eventBus.Publish(ctx, events_pb.EventName_BUILD_COMPLETED, &events_pb.EventData{
		Value: &events_pb.EventData_BuildCompletedData{
			BuildCompletedData: &events_pb.BuildCompletedData{
				ImageUrl:   buildResult.ImageURL,
				AppName:    data.App.Name,
				AppId:      data.App.Id,
				BuildId:    build.Id,
				DomainName: data.App.DomainName,
			},
		},
	})
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
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

	span.SetAttributes(
		attribute.String("app_id", data.AppId),
		attribute.String("app_name", data.AppName),
	)

	h.logger.LogInfoF("Deleting all builds entities related to app '%s'", data.AppId)
	err := h.buildRepository.DeleteBuilds(ctx, data.AppId)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	// FIXME: find a safe way to do this.
	err = h.buildExecutor.(*buildexecutor.KanikoExecutor).DeleteJobs(data.AppName)
	if err != nil {
		h.logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
}
