package grpc_server

import (
	"context"
	"project/proto/project_service_pb"
	"project/repositories"

	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCProjectServiceServer struct {
	project_service_pb.UnimplementedProjectServiceServer

	ProjectRepository repositories.ProjectRepositoryInterface
	EventBus          *messaging.EventBus
	Logger            logging.ServiceLogger
}

func NewGRPCProjectServiceServer(projectRepository repositories.ProjectRepositoryInterface, eventBus *messaging.EventBus, logger logging.ServiceLogger) *GRPCProjectServiceServer {
	return &GRPCProjectServiceServer{
		ProjectRepository: projectRepository,
		EventBus:          eventBus,
		Logger:            logger,
	}
}

func (server *GRPCProjectServiceServer) Health(ctx context.Context, _ *project_service_pb.HealthRequest) (*project_service_pb.HealthResponse, error) {
	return &project_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCProjectServiceServer) CreateProject(ctx context.Context, createProjectRequest *project_service_pb.CreateProjectRequest) (*project_service_pb.CreateProjectResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("user.id", createProjectRequest.UserId))

	project, err := server.ProjectRepository.CreateProject(ctx, createProjectRequest.UserId, repositories.CreateProjectParams{
		Name: createProjectRequest.Name,
	})

	if err == repositories.ErrProjectNameInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(attribute.String("project.id", project.Id))

	return &project_service_pb.CreateProjectResponse{
		Project: &project_service_pb.Project{
			Id:        project.Id,
			Name:      project.Name,
			UserId:    project.UserId,
			CreatedAt: project.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCProjectServiceServer) GetUserProjectById(ctx context.Context, getUserProjectByIdRequest *project_service_pb.GetUserProjectByIdRequest) (*project_service_pb.GetUserProjectByIdResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("user.id", getUserProjectByIdRequest.UserId),
		attribute.String("project.id", getUserProjectByIdRequest.ProjectId),
	)

	project, err := server.ProjectRepository.GetProjectById(ctx, getUserProjectByIdRequest.UserId, getUserProjectByIdRequest.ProjectId)

	if err == repositories.ErrProjectNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &project_service_pb.GetUserProjectByIdResponse{
		Project: &project_service_pb.Project{
			Id:        project.Id,
			Name:      project.Name,
			UserId:    project.UserId,
			CreatedAt: project.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCProjectServiceServer) GetUserProjects(ctx context.Context, getUserProjectsRequest *project_service_pb.GetUserProjectsRequest) (*project_service_pb.GetUserProjectsResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("user.id", getUserProjectsRequest.UserId))

	projects, err := server.ProjectRepository.GetProjects(ctx, getUserProjectsRequest.UserId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	_projects := []*project_service_pb.Project{}
	projects_ids := []string{}
	for _, project := range projects {
		_project := project_service_pb.Project{
			Id:        project.Id,
			Name:      project.Name,
			UserId:    project.UserId,
			CreatedAt: project.CreatedAt.String(),
		}
		_projects = append(_projects, &_project)
		projects_ids = append(projects_ids, _project.Id)
	}

	span.SetAttributes(
		attribute.StringSlice("projects.ids", projects_ids),
		attribute.Int("projects.count", len(_projects)),
	)

	return &project_service_pb.GetUserProjectsResponse{
		Projects: _projects,
	}, nil
}

func (server *GRPCProjectServiceServer) DeleteUserProject(ctx context.Context, deleteUserProjectRequest *project_service_pb.DeleteUserProjectRequest) (*project_service_pb.DeleteUserProjectResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("user.id", deleteUserProjectRequest.UserId),
		attribute.String("project.id", deleteUserProjectRequest.ProjectId),
	)

	err := server.ProjectRepository.DeleteProjectById(ctx, deleteUserProjectRequest.UserId, deleteUserProjectRequest.ProjectId)

	if err == repositories.ErrProjectNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = server.EventBus.Publish(ctx, events_pb.EventName_PROJECT_DELETED, &events_pb.EventData{
		Value: &events_pb.EventData_ProjectDeletedData{
			ProjectDeletedData: &events_pb.ProjectDeletedEventData{
				ProjectId: deleteUserProjectRequest.ProjectId,
			},
		},
	})
	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	return &project_service_pb.DeleteUserProjectResponse{}, nil
}

func (server *GRPCProjectServiceServer) UpdateProject(ctx context.Context, updateProjectRequest *project_service_pb.UpdateProjectRequest) (*project_service_pb.UpdateProjectResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("project.id", updateProjectRequest.ProjectId),
	)

	project, err := server.ProjectRepository.UpdateProjectById(ctx, updateProjectRequest.ProjectId, repositories.UpdateProjectParams{Name: updateProjectRequest.Name})
	if err == repositories.ErrProjectNameInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &project_service_pb.UpdateProjectResponse{
		Project: &project_service_pb.Project{
			Id:        project.Id,
			Name:      project.Name,
			UserId:    project.UserId,
			CreatedAt: project.CreatedAt.String(),
		},
	}, nil
}
