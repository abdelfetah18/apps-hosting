package grpc_server

import (
	"context"
	"project/proto/project_service_pb"
	"project/repositories"

	"apps-hosting.com/logging"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCProjectServiceServer struct {
	project_service_pb.UnimplementedProjectServiceServer

	ProjectRepository repositories.ProjectRepositoryInterface
	Logger            logging.ServiceLogger
}

func NewGRPCProjectServiceServer(projectRepository repositories.ProjectRepositoryInterface, logger logging.ServiceLogger) *GRPCProjectServiceServer {
	return &GRPCProjectServiceServer{
		ProjectRepository: projectRepository,
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
	project, err := server.ProjectRepository.CreateProject(createProjectRequest.UserId, repositories.CreateProjectParams{
		Name: createProjectRequest.Name,
	})

	if err == repositories.ErrProjectNameInUse {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	project, err := server.ProjectRepository.GetProjectById(getUserProjectByIdRequest.UserId, getUserProjectByIdRequest.ProjectId)

	if err == repositories.ErrProjectNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
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
	projects, err := server.ProjectRepository.GetProjects(getUserProjectsRequest.UserId)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	_projects := []*project_service_pb.Project{}
	for _, project := range projects {
		_project := project_service_pb.Project{
			Id:        project.Id,
			Name:      project.Name,
			UserId:    project.UserId,
			CreatedAt: project.CreatedAt.String(),
		}
		_projects = append(_projects, &_project)
	}

	return &project_service_pb.GetUserProjectsResponse{
		Projects: _projects,
	}, nil
}

func (server *GRPCProjectServiceServer) DeleteUserProject(ctx context.Context, deleteUserProjectRequest *project_service_pb.DeleteUserProjectRequest) (*project_service_pb.DeleteUserProjectResponse, error) {
	err := server.ProjectRepository.DeleteProjectById(deleteUserProjectRequest.UserId, deleteUserProjectRequest.ProjectId)

	if err == repositories.ErrProjectNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &project_service_pb.DeleteUserProjectResponse{}, nil
}
