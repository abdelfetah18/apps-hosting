package grpc_server

import (
	"app/proto/app_service_pb"
	"app/repositories"
	"app/utils"
	"context"
	"net/url"
	"slices"

	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAppServiceServer struct {
	app_service_pb.UnimplementedAppServiceServer

	AppRepository                  repositories.AppRepository
	EnvironmentVariablesRepository repositories.EnvironmentVariablesRepository
	EventBus                       messaging.EventBus
	Logger                         logging.ServiceLogger
}

func NewGRPCAppServiceServer(
	appRepository repositories.AppRepository,
	environmentVariablesRepository repositories.EnvironmentVariablesRepository,
	eventBus messaging.EventBus,
	logger logging.ServiceLogger,
) *GRPCAppServiceServer {
	return &GRPCAppServiceServer{
		AppRepository:                  appRepository,
		EnvironmentVariablesRepository: environmentVariablesRepository,
		EventBus:                       eventBus,
		Logger:                         logger,
	}
}

func (server *GRPCAppServiceServer) Health(ctx context.Context, _ *app_service_pb.HealthRequest) (*app_service_pb.HealthResponse, error) {
	return &app_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCAppServiceServer) CreateApp(ctx context.Context, createAppRequest *app_service_pb.CreateAppRequest) (*app_service_pb.CreateAppResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("project_id", createAppRequest.ProjectId),
		attribute.String("app_name", createAppRequest.Name),
		attribute.String("app_runtime", createAppRequest.Runtime),
		attribute.String("app_repo_url", createAppRequest.RepoUrl),
		attribute.String("app_build_cmd", createAppRequest.BuildCmd),
		attribute.String("app_start_cmd", createAppRequest.StartCmd),
	)

	if createAppRequest.EnvironmentVariables != nil {
		span.SetAttributes(attribute.String("app_environment_variables", *createAppRequest.EnvironmentVariables))
	}

	if len(createAppRequest.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name is required")
	}

	if !slices.Contains(repositories.Runtimes, createAppRequest.Runtime) {
		return nil, status.Error(codes.InvalidArgument, "Unsupported runtime")
	}

	repoURL, err := url.Parse(createAppRequest.RepoUrl)
	if err != nil || repoURL.Hostname() != "github.com" {
		return nil, status.Error(codes.InvalidArgument, "Invalid GitHub URL")
	}

	createdApp, err := server.AppRepository.CreateApp(ctx, createAppRequest.ProjectId, repositories.CreateAppParams{
		Name:       createAppRequest.Name,
		Runtime:    createAppRequest.Runtime,
		RepoURL:    createAppRequest.RepoUrl,
		StartCMD:   createAppRequest.StartCmd,
		BuildCMD:   createAppRequest.BuildCmd,
		DomainName: utils.GetDomainName(createAppRequest.Name),
	})

	if err == repositories.ErrDomainNameInUse {
		appName := createAppRequest.Name + "-" + uuid.NewString()
		createdApp, err = server.AppRepository.CreateApp(ctx, createAppRequest.ProjectId, repositories.CreateAppParams{
			Name:       createAppRequest.Name,
			Runtime:    createAppRequest.Runtime,
			RepoURL:    createAppRequest.RepoUrl,
			StartCMD:   createAppRequest.StartCmd,
			BuildCMD:   createAppRequest.BuildCmd,
			DomainName: utils.GetDomainName(appName),
		})
	}

	if err == repositories.ErrAppNameInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	server.Logger.LogInfo("App created successfully")

	if createAppRequest.EnvironmentVariables != nil {
		_, err = server.EnvironmentVariablesRepository.CreateEnvironmentVariables(ctx, createdApp.Id, repositories.CreateEnvironmentVariableParams{
			Value: *createAppRequest.EnvironmentVariables,
		})

		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		server.Logger.LogInfo("Environment variables created successfully")
	}

	span.SetAttributes(
		attribute.String("app_id", createdApp.Id),
		attribute.String("app_domain_name", createdApp.DomainName),
		attribute.String("app_created_at", createdApp.CreatedAt.String()),
	)

	server.Logger.LogInfo("Send AppCreated Event")
	err = server.EventBus.Publish(ctx, events_pb.EventName_APP_CREATED, &events_pb.EventData{
		Value: &events_pb.EventData_AppCreatedData{
			AppCreatedData: &events_pb.AppCreatedEventData{
				AppId:      createdApp.Id,
				AppName:    createdApp.Name,
				Runtime:    createdApp.Runtime,
				RepoUrl:    createdApp.RepoURL,
				StartCmd:   createdApp.StartCMD,
				BuildCmd:   createdApp.BuildCMD,
				UserId:     createdApp.ProjectId,
				DomainName: createdApp.DomainName,
			},
		},
	})
	if err != nil {
		// FIXME: Handle this case
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	return &app_service_pb.CreateAppResponse{
		App: &app_service_pb.App{
			Id:         createdApp.Id,
			Name:       createdApp.Name,
			DomainName: createdApp.DomainName,
			Runtime:    createdApp.Runtime,
			RepoUrl:    createdApp.RepoURL,
			BuildCmd:   createdApp.BuildCMD,
			StartCmd:   createdApp.StartCMD,
			CreatedAt:  createdApp.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) GetApp(ctx context.Context, getAppRequest *app_service_pb.GetAppRequest) (*app_service_pb.GetAppResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", getAppRequest.AppId),
		attribute.String("project_id", getAppRequest.ProjectId),
	)

	app, err := server.AppRepository.GetAppById(ctx, getAppRequest.ProjectId, getAppRequest.AppId)

	span.SetAttributes(
		attribute.String("app_name", app.Name),
		attribute.String("app_runtime", app.Runtime),
		attribute.String("app_repo_url", app.RepoURL),
		attribute.String("app_domain_name", app.DomainName),
		attribute.String("app_build_cmd", app.BuildCMD),
		attribute.String("app_start_cmd", app.StartCMD),
		attribute.String("app_created_at", app.CreatedAt.String()),
	)

	if err == repositories.ErrAppNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &app_service_pb.GetAppResponse{
		App: &app_service_pb.App{
			Id:         app.Id,
			Name:       app.Name,
			DomainName: app.DomainName,
			Runtime:    app.Runtime,
			RepoUrl:    app.RepoURL,
			BuildCmd:   app.BuildCMD,
			StartCmd:   app.StartCMD,
			CreatedAt:  app.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) GetApps(ctx context.Context, getAppsRequest *app_service_pb.GetAppsRequest) (*app_service_pb.GetAppsResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("project_id", getAppsRequest.ProjectId))

	apps, err := server.AppRepository.GetApps(ctx, getAppsRequest.ProjectId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	_apps := []*app_service_pb.App{}
	apps_ids := []string{}
	for _, app := range apps {
		_app := app_service_pb.App{
			Id:         app.Id,
			Name:       app.Name,
			DomainName: app.DomainName,
			Runtime:    app.Runtime,
			RepoUrl:    app.RepoURL,
			BuildCmd:   app.BuildCMD,
			StartCmd:   app.StartCMD,
			CreatedAt:  app.CreatedAt.String(),
		}
		_apps = append(_apps, &_app)
		apps_ids = append(apps_ids, app.Id)
	}

	span.SetAttributes(attribute.StringSlice("apps_ids", apps_ids))

	return &app_service_pb.GetAppsResponse{
		Apps: _apps,
	}, nil
}

func (server *GRPCAppServiceServer) UpdateApp(ctx context.Context, updateAppRequest *app_service_pb.UpdateAppRequest) (*app_service_pb.UpdateAppResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("project_id", updateAppRequest.ProjectId),
		attribute.String("app_id", updateAppRequest.AppId),
		attribute.String("app_name", utils.SafeString(updateAppRequest.Name)),
		attribute.String("app_runtime", utils.SafeString(updateAppRequest.Runtime)),
		attribute.String("app_repo_url", utils.SafeString(updateAppRequest.RepoUrl)),
		attribute.String("app_build_cmd", utils.SafeString(updateAppRequest.BuildCmd)),
		attribute.String("app_start_cmd", utils.SafeString(updateAppRequest.StartCmd)),
	)

	if len(*updateAppRequest.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
	}

	if !slices.Contains(repositories.Runtimes, *updateAppRequest.Runtime) {
		return nil, status.Error(codes.InvalidArgument, "Unsupported runtime")
	}

	repoURL, err := url.Parse(*updateAppRequest.RepoUrl)
	if err != nil || repoURL.Hostname() != "github.com" {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, "Invalid GitHub URL")
	}

	updatedApp, err := server.AppRepository.UpdateApp(
		ctx,
		updateAppRequest.ProjectId,
		updateAppRequest.AppId,
		repositories.UpdateAppParams{
			Name:       *updateAppRequest.Name,
			Runtime:    *updateAppRequest.Runtime,
			RepoURL:    *updateAppRequest.RepoUrl,
			StartCMD:   *updateAppRequest.StartCmd,
			BuildCMD:   *updateAppRequest.BuildCmd,
			DomainName: utils.GetDomainName(*updateAppRequest.Name),
		})

	if err == repositories.ErrDomainNameInUse {
		appName := *updateAppRequest.Name + "-" + uuid.NewString()
		updatedApp, err = server.AppRepository.UpdateApp(
			ctx,
			updateAppRequest.ProjectId,
			updateAppRequest.AppId,
			repositories.UpdateAppParams{
				Name:       *updateAppRequest.Name,
				Runtime:    *updateAppRequest.Runtime,
				RepoURL:    *updateAppRequest.RepoUrl,
				StartCMD:   *updateAppRequest.StartCmd,
				BuildCMD:   *updateAppRequest.BuildCmd,
				DomainName: utils.GetDomainName(appName),
			})
	}

	if err == repositories.ErrAppNameInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &app_service_pb.UpdateAppResponse{
		App: &app_service_pb.App{
			Id:         updatedApp.Id,
			Name:       updatedApp.Name,
			DomainName: updatedApp.DomainName,
			Runtime:    updatedApp.Runtime,
			RepoUrl:    updatedApp.RepoURL,
			BuildCmd:   updatedApp.BuildCMD,
			StartCmd:   updatedApp.StartCMD,
			CreatedAt:  updatedApp.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) DeleteApp(ctx context.Context, deleteAppRequest *app_service_pb.DeleteAppRequest) (*app_service_pb.DeleteAppResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("project_id", deleteAppRequest.ProjectId),
		attribute.String("app_id", deleteAppRequest.AppId),
	)

	app, err := server.AppRepository.GetAppById(ctx, deleteAppRequest.ProjectId, deleteAppRequest.AppId)

	if err == repositories.ErrAppNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = server.AppRepository.DeleteAppById(ctx, deleteAppRequest.ProjectId, deleteAppRequest.AppId)
	if err == repositories.ErrAppNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = server.EventBus.Publish(ctx, events_pb.EventName_APP_DELETED, &events_pb.EventData{
		Value: &events_pb.EventData_AppDeletedData{
			AppDeletedData: &events_pb.AppDeletedEventData{
				AppId:   deleteAppRequest.AppId,
				AppName: app.Name,
			},
		},
	})
	if err != nil {
		// FIXME: handle this case
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	return &app_service_pb.DeleteAppResponse{}, nil
}

func (server *GRPCAppServiceServer) GetEnvironmentVariables(ctx context.Context, getEnvironmentVariablesRequest *app_service_pb.GetEnvironmentVariablesRequest) (*app_service_pb.GetEnvironmentVariablesResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("app_id", getEnvironmentVariablesRequest.AppId))

	environmentVariables, err := server.EnvironmentVariablesRepository.GetEnvironmentVariable(ctx, getEnvironmentVariablesRequest.AppId)

	if err == repositories.ErrEnvVarNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(
		attribute.String("app_environment_variables_id", environmentVariables.Id),
		attribute.String("app_environment_variables_value", environmentVariables.Value),
	)

	return &app_service_pb.GetEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) CreateEnvironmentVariables(ctx context.Context, createEnvironmentVariablesRequest *app_service_pb.CreateEnvironmentVariablesRequest) (*app_service_pb.CreateEnvironmentVariablesResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", createEnvironmentVariablesRequest.AppId),
		attribute.String("environment_variables_value", createEnvironmentVariablesRequest.Value),
	)

	environmentVariables, err := server.EnvironmentVariablesRepository.CreateEnvironmentVariables(ctx, createEnvironmentVariablesRequest.AppId, repositories.CreateEnvironmentVariableParams{
		Value: createEnvironmentVariablesRequest.Value,
	})

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(attribute.String("environment_variables_id", environmentVariables.Id))

	return &app_service_pb.CreateEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) UpdateEnvironmentVariables(ctx context.Context, updateEnvironmentVariablesRequest *app_service_pb.UpdateEnvironmentVariablesRequest) (*app_service_pb.UpdateEnvironmentVariablesResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("app_id", updateEnvironmentVariablesRequest.AppId),
		attribute.String("environment_variables_value", updateEnvironmentVariablesRequest.Value),
	)

	_, err := server.EnvironmentVariablesRepository.GetEnvironmentVariable(ctx, updateEnvironmentVariablesRequest.AppId)
	if err == repositories.ErrEnvVarNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		environmentVariables, err := server.EnvironmentVariablesRepository.CreateEnvironmentVariables(ctx, updateEnvironmentVariablesRequest.AppId, repositories.CreateEnvironmentVariableParams{
			Value: updateEnvironmentVariablesRequest.Value,
		})

		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		span.SetAttributes(attribute.String("environment_variables_id", environmentVariables.Id))

		return &app_service_pb.UpdateEnvironmentVariablesResponse{
			EnvironmentVariable: &app_service_pb.EnvironmentVariables{
				Id:       environmentVariables.Id,
				Value:    environmentVariables.Value,
				CreateAt: environmentVariables.CreatedAt.String(),
			},
		}, nil
	}

	environmentVariables, err := server.EnvironmentVariablesRepository.UpdateEnvironmentVariables(ctx, updateEnvironmentVariablesRequest.AppId, repositories.UpdateEnvironmentVariableParams{
		Value: updateEnvironmentVariablesRequest.Value,
	})

	if err == repositories.ErrEnvVarNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(attribute.String("environment_variables_id", environmentVariables.Id))

	return &app_service_pb.UpdateEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) DeleteEnvironmentVariables(ctx context.Context, deleteEnvironmentVariablesRequest *app_service_pb.DeleteEnvironmentVariablesRequest) (*app_service_pb.DeleteEnvironmentVariablesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}
