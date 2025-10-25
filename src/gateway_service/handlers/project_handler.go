package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/project_service_pb"
	"gateway/utils"
	"net/http"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectHandler struct {
	ProjectServiceClient project_service_pb.ProjectServiceClient
	Logger               logging.ServiceLogger
}

func NewProjectHandler(projectServiceClient project_service_pb.ProjectServiceClient, logger logging.ServiceLogger) ProjectHandler {
	return ProjectHandler{
		ProjectServiceClient: projectServiceClient,
		Logger:               logger,
	}
}

func (handler *ProjectHandler) OwnershipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		projectId := params["project_id"]
		userId := r.URL.Query().Get("user_id")

		handler.Logger.LogErrorF("projectId=%s, userId=%s", projectId, userId)

		_, err := handler.ProjectServiceClient.GetUserProjectById(context.Background(), &project_service_pb.GetUserProjectByIdRequest{
			ProjectId: projectId,
			UserId:    userId,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)

			if status.Code() == codes.NotFound {
				messaging.WriteError(w, http.StatusUnauthorized, "you don't own this project")
				return
			}

			messaging.WriteError(w, http.StatusInternalServerError, status.Message())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (handler *ProjectHandler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	createProjectRequest := project_service_pb.CreateProjectRequest{}

	err := json.NewDecoder(r.Body).Decode(&createProjectRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, "failed at decoding json request")
		return
	}

	createProjectRequest.UserId = userId
	createAppResponse, err := handler.ProjectServiceClient.CreateProject(context.Background(), &createProjectRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "Project Created Successfully", createAppResponse.Project)
}

func (handler *ProjectHandler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	_, err := handler.ProjectServiceClient.DeleteUserProject(context.Background(), &project_service_pb.DeleteUserProjectRequest{
		ProjectId: projectId,
		UserId:    userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "App Deleted Successfully", nil)
}

func (handler *ProjectHandler) GetUserProjectByIdHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	handler.Logger.LogErrorF("projectId=%s, userId=%s", projectId, userId)

	getUserProjectByIdResponse, err := handler.ProjectServiceClient.GetUserProjectById(context.Background(), &project_service_pb.GetUserProjectByIdRequest{
		ProjectId: projectId,
		UserId:    userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "Project Fetched Successfully", getUserProjectByIdResponse.Project)
}

func (handler *ProjectHandler) GetUserProjectsHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	handler.Logger.LogInfoF("userId=%s", userId)
	getUserProjectsResponse, err := handler.ProjectServiceClient.GetUserProjects(context.Background(), &project_service_pb.GetUserProjectsRequest{UserId: userId})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	if getUserProjectsResponse.Projects == nil {
		messaging.WriteSuccess(w, "Projects Fetched Successfully", []*project_service_pb.Project{})
		return
	}

	messaging.WriteSuccess(w, "Projects Fetched Successfully", getUserProjectsResponse.Projects)
}
