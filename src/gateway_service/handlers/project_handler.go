package handlers

import (
	"encoding/json"
	"gateway/proto/project_service_pb"
	"gateway/utils"
	"net/http"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
		span := trace.SpanFromContext(r.Context())

		params := mux.Vars(r)
		projectId := params["project_id"]
		userId := r.URL.Query().Get("user_id")

		span.SetAttributes(attribute.String("user_id", userId))
		span.SetAttributes(attribute.String("project_id", projectId))

		handler.Logger.LogErrorF("projectId=%s, userId=%s", projectId, userId)

		_, err := handler.ProjectServiceClient.GetUserProjectById(r.Context(), &project_service_pb.GetUserProjectByIdRequest{
			ProjectId: projectId,
			UserId:    userId,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)

			span.SetAttributes(attribute.String("error", err.Error()))

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
	span := trace.SpanFromContext(r.Context())

	userId := r.URL.Query().Get("user_id")
	span.SetAttributes(attribute.String("user_id", userId))
	createProjectRequest := project_service_pb.CreateProjectRequest{}

	err := json.NewDecoder(r.Body).Decode(&createProjectRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, "failed at decoding json request")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("project_name", createProjectRequest.Name))

	createProjectRequest.UserId = userId
	createAppResponse, err := handler.ProjectServiceClient.CreateProject(r.Context(), &createProjectRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Project Created Successfully", createAppResponse.Project)
}

func (handler *ProjectHandler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	span.SetAttributes(attribute.String("user_id", userId))
	span.SetAttributes(attribute.String("project_id", projectId))

	_, err := handler.ProjectServiceClient.DeleteUserProject(r.Context(), &project_service_pb.DeleteUserProjectRequest{
		ProjectId: projectId,
		UserId:    userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "App Deleted Successfully", nil)
}

func (handler *ProjectHandler) GetUserProjectByIdHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	span.SetAttributes(attribute.String("user_Id", userId))
	span.SetAttributes(attribute.String("project_id", projectId))

	handler.Logger.LogErrorF("projectId=%s, userId=%s", projectId, userId)

	getUserProjectByIdResponse, err := handler.ProjectServiceClient.GetUserProjectById(r.Context(), &project_service_pb.GetUserProjectByIdRequest{
		ProjectId: projectId,
		UserId:    userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("project_name", getUserProjectByIdResponse.Project.Name))

	messaging.WriteSuccess(w, "Project Fetched Successfully", getUserProjectByIdResponse.Project)
}

func (handler *ProjectHandler) GetUserProjectsHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())

	userId := r.URL.Query().Get("user_id")
	span.SetAttributes(attribute.String("user_id", userId))

	handler.Logger.LogInfoF("userId=%s", userId)
	getUserProjectsResponse, err := handler.ProjectServiceClient.GetUserProjects(r.Context(), &project_service_pb.GetUserProjectsRequest{UserId: userId})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}
	span.SetAttributes(attribute.Int("user_projects_count", len(getUserProjectsResponse.Projects)))

	if getUserProjectsResponse.Projects == nil {
		messaging.WriteSuccess(w, "Projects Fetched Successfully", []*project_service_pb.Project{})
		return
	}

	messaging.WriteSuccess(w, "Projects Fetched Successfully", getUserProjectsResponse.Projects)
}

func (handler *ProjectHandler) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	span.SetAttributes(
		attribute.String("user.id", userId),
		attribute.String("project.id", projectId),
	)

	updateProjectRequest := project_service_pb.UpdateProjectRequest{}
	err := json.NewDecoder(r.Body).Decode(&updateProjectRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, "failed at decoding json request")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	updateProjectRequest.ProjectId = projectId
	updateAppResponse, err := handler.ProjectServiceClient.UpdateProject(r.Context(), &updateProjectRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Project Updated Successfully", updateAppResponse.Project)
}
