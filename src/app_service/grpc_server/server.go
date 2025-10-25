package grpc_server

import (
	"app/nats"
	"app/proto/app_service_pb"
	"app/repositories"
	"app/utils"
	"context"
	"net/url"
	"slices"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAppServiceServer struct {
	app_service_pb.UnimplementedAppServiceServer

	AppRepository                  repositories.AppRepository
	EnvironmentVariablesRepository repositories.EnvironmentVariablesRepository
	NatsService                    nats.NatsService
	Logger                         logging.ServiceLogger
}

func NewGRPCAppServiceServer(
	appRepository repositories.AppRepository,
	environmentVariablesRepository repositories.EnvironmentVariablesRepository,
	natsService nats.NatsService,
	logger logging.ServiceLogger,
) *GRPCAppServiceServer {
	return &GRPCAppServiceServer{
		AppRepository:                  appRepository,
		EnvironmentVariablesRepository: environmentVariablesRepository,
		NatsService:                    natsService,
		Logger:                         logger,
	}
}

func (server *GRPCAppServiceServer) Health(ctx context.Context, _ *app_service_pb.HealthRequest) (*app_service_pb.HealthResponse, error) {
	return &app_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCAppServiceServer) CreateApp(context context.Context, createAppRequest *app_service_pb.CreateAppRequest) (*app_service_pb.CreateAppResponse, error) {
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

	createdApp, err := server.AppRepository.CreateApp(createAppRequest.ProjectId, repositories.CreateAppParams{
		Name:       createAppRequest.Name,
		Runtime:    createAppRequest.Runtime,
		RepoURL:    createAppRequest.RepoUrl,
		StartCMD:   createAppRequest.StartCmd,
		BuildCMD:   createAppRequest.BuildCmd,
		DomainName: utils.GetDomainName(createAppRequest.Name),
	})

	if err == repositories.ErrDomainNameInUse {
		appName := createAppRequest.Name + "-" + uuid.NewString()
		createdApp, err = server.AppRepository.CreateApp(createAppRequest.ProjectId, repositories.CreateAppParams{
			Name:       createAppRequest.Name,
			Runtime:    createAppRequest.Runtime,
			RepoURL:    createAppRequest.RepoUrl,
			StartCMD:   createAppRequest.StartCmd,
			BuildCMD:   createAppRequest.BuildCmd,
			DomainName: utils.GetDomainName(appName),
		})
	}

	if err == repositories.ErrAppNameInUse {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		server.Logger.LogError(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	server.Logger.LogInfo("App created successfully")

	if createAppRequest.EnvironmentVariables != nil {
		_, err = server.EnvironmentVariablesRepository.CreateEnvironmentVariables(createdApp.Id, repositories.CreateEnvironmentVariableParams{
			Value: *createAppRequest.EnvironmentVariables,
		})

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		server.Logger.LogInfo("Environment variables created successfully")
	}

	server.Logger.LogInfo("Send AppCreated Event")
	_, err = messaging.PublishMessage(server.NatsService.JetStream,
		messaging.NewMessage(
			messaging.AppCreated,
			messaging.AppCreatedData{
				AppId:      createdApp.Id,
				AppName:    createdApp.Name,
				Runtime:    createdApp.Runtime,
				RepoURL:    createdApp.RepoURL,
				StartCMD:   createdApp.StartCMD,
				BuildCMD:   createdApp.BuildCMD,
				UserId:     createdApp.ProjectId,
				DomainName: createdApp.DomainName,
			}))
	if err != nil {
		// FIXME: Handle this case
		server.Logger.LogError(err.Error())
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

func (server *GRPCAppServiceServer) GetApp(context context.Context, getAppRequest *app_service_pb.GetAppRequest) (*app_service_pb.GetAppResponse, error) {
	app, err := server.AppRepository.GetAppById(getAppRequest.ProjectId, getAppRequest.AppId)

	if err == repositories.ErrAppNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
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

func (server *GRPCAppServiceServer) GetApps(context context.Context, getAppsRequest *app_service_pb.GetAppsRequest) (*app_service_pb.GetAppsResponse, error) {
	apps, err := server.AppRepository.GetApps(getAppsRequest.ProjectId)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	_apps := []*app_service_pb.App{}
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
	}

	return &app_service_pb.GetAppsResponse{
		Apps: _apps,
	}, nil
}

func (server *GRPCAppServiceServer) UpdateApp(context context.Context, updateAppRequest *app_service_pb.UpdateAppRequest) (*app_service_pb.UpdateAppResponse, error) {
	if len(*updateAppRequest.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
	}

	if !slices.Contains(repositories.Runtimes, *updateAppRequest.Runtime) {
		return nil, status.Error(codes.InvalidArgument, "Unsupported runtime")
	}

	repoURL, err := url.Parse(*updateAppRequest.RepoUrl)
	if err != nil || repoURL.Hostname() != "github.com" {
		return nil, status.Error(codes.InvalidArgument, "Invalid GitHub URL")
	}

	updatedApp, err := server.AppRepository.UpdateApp(
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
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
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

func (server *GRPCAppServiceServer) DeleteApp(context context.Context, deleteAppRequest *app_service_pb.DeleteAppRequest) (*app_service_pb.DeleteAppResponse, error) {
	app, err := server.AppRepository.GetAppById(deleteAppRequest.ProjectId, deleteAppRequest.AppId)

	if err == repositories.ErrAppNotFound {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = server.AppRepository.DeleteAppById(deleteAppRequest.ProjectId, deleteAppRequest.AppId)
	if err == repositories.ErrAppNotFound {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = messaging.PublishMessage(server.NatsService.JetStream,
		messaging.NewMessage(
			messaging.AppDeleted,
			messaging.AppDeletedData{
				AppId:   deleteAppRequest.AppId,
				AppName: app.Name, // FIXME: should only pass id
			}))
	if err != nil {
		// FIXME: handle this case
		server.Logger.LogError(err.Error())
	}

	return &app_service_pb.DeleteAppResponse{}, nil
}

func (server *GRPCAppServiceServer) GetEnvironmentVariables(context context.Context, getEnvironmentVariablesRequest *app_service_pb.GetEnvironmentVariablesRequest) (*app_service_pb.GetEnvironmentVariablesResponse, error) {
	environmentVariables, err := server.EnvironmentVariablesRepository.GetEnvironmentVariable(getEnvironmentVariablesRequest.AppId)

	if err == repositories.ErrEnvVarNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &app_service_pb.GetEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) CreateEnvironmentVariables(context context.Context, createEnvironmentVariablesRequest *app_service_pb.CreateEnvironmentVariablesRequest) (*app_service_pb.CreateEnvironmentVariablesResponse, error) {
	environmentVariables, err := server.EnvironmentVariablesRepository.CreateEnvironmentVariables(createEnvironmentVariablesRequest.AppId, repositories.CreateEnvironmentVariableParams{
		Value: createEnvironmentVariablesRequest.Value,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &app_service_pb.CreateEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) UpdateEnvironmentVariables(context context.Context, updateEnvironmentVariablesRequest *app_service_pb.UpdateEnvironmentVariablesRequest) (*app_service_pb.UpdateEnvironmentVariablesResponse, error) {
	_, err := server.EnvironmentVariablesRepository.GetEnvironmentVariable(updateEnvironmentVariablesRequest.AppId)
	if err == repositories.ErrEnvVarNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		environmentVariables, err := server.EnvironmentVariablesRepository.CreateEnvironmentVariables(updateEnvironmentVariablesRequest.AppId, repositories.CreateEnvironmentVariableParams{
			Value: updateEnvironmentVariablesRequest.Value,
		})

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &app_service_pb.UpdateEnvironmentVariablesResponse{
			EnvironmentVariable: &app_service_pb.EnvironmentVariables{
				Id:       environmentVariables.Id,
				Value:    environmentVariables.Value,
				CreateAt: environmentVariables.CreatedAt.String(),
			},
		}, nil
	}

	environmentVariables, err := server.EnvironmentVariablesRepository.UpdateEnvironmentVariables(updateEnvironmentVariablesRequest.AppId, repositories.UpdateEnvironmentVariableParams{
		Value: updateEnvironmentVariablesRequest.Value,
	})

	if err == repositories.ErrEnvVarNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &app_service_pb.UpdateEnvironmentVariablesResponse{
		EnvironmentVariable: &app_service_pb.EnvironmentVariables{
			Id:       environmentVariables.Id,
			Value:    environmentVariables.Value,
			CreateAt: environmentVariables.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCAppServiceServer) DeleteEnvironmentVariables(context context.Context, deleteEnvironmentVariablesRequest *app_service_pb.DeleteEnvironmentVariablesRequest) (*app_service_pb.DeleteEnvironmentVariablesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented")
}
