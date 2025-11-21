package handlers

import (
	"encoding/json"
	"gateway/proto/app_service_pb"
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

type AppHandler struct {
	AppServiceClient app_service_pb.AppServiceClient
	Logger           logging.ServiceLogger
}

func NewAppHandler(appServiceClient app_service_pb.AppServiceClient, logger logging.ServiceLogger) AppHandler {
	return AppHandler{
		AppServiceClient: appServiceClient,
		Logger:           logger,
	}
}

func (handler *AppHandler) OwnershipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		params := mux.Vars(r)

		projectId := params["project_id"]
		appId := params["app_id"]

		span.SetAttributes(
			attribute.String("project.id", projectId),
			attribute.String("app.id", appId),
		)

		_, err := handler.AppServiceClient.GetApp(r.Context(), &app_service_pb.GetAppRequest{
			AppId:     appId,
			ProjectId: projectId,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)

			span.SetAttributes(attribute.String("error", err.Error()))

			if status.Code() == codes.NotFound {
				messaging.WriteError(w, http.StatusUnauthorized, "you don't own this app")
				return
			}

			messaging.WriteError(w, http.StatusInternalServerError, status.Message())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (handler *AppHandler) CreateAppHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())

	params := mux.Vars(r)
	projectId := params["project_id"]
	userId := r.URL.Query().Get("user_id")

	span.SetAttributes(
		attribute.String("user.id", userId),
		attribute.String("project.id", projectId),
	)

	createAppRequest := app_service_pb.CreateAppRequest{}
	err := json.NewDecoder(r.Body).Decode(&createAppRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, "failed at decoding json request")
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	createAppRequest.ProjectId = projectId
	createAppRequest.UserId = userId
	createAppResponse, err := handler.AppServiceClient.CreateApp(r.Context(), &createAppRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("app.id", createAppResponse.App.Id))

	messaging.WriteSuccess(w, "App Created Successfully", createAppResponse.App)
}

func (handler *AppHandler) UpdateAppHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())

	params := mux.Vars(r)
	appId := params["app_id"]
	projectId := params["project_id"]

	span.SetAttributes(
		attribute.String("project.id", projectId),
		attribute.String("app.id", appId),
	)

	updateAppRequest := app_service_pb.UpdateAppRequest{}

	err := json.NewDecoder(r.Body).Decode(&updateAppRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	updateAppRequest.ProjectId = projectId
	updateAppRequest.AppId = appId

	updateAppResponse, err := handler.AppServiceClient.UpdateApp(r.Context(), &updateAppRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())

		span.SetAttributes(attribute.String("error", err.Error()))

		return
	}

	messaging.WriteSuccess(w, "App Updated Successfully", updateAppResponse.App)
}

func (handler *AppHandler) DeleteAppHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	projectId := params["project_id"]

	span.SetAttributes(
		attribute.String("project.id", projectId),
		attribute.String("app.id", appId),
	)

	_, err := handler.AppServiceClient.DeleteApp(r.Context(), &app_service_pb.DeleteAppRequest{
		ProjectId: projectId,
		AppId:     appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "App Deleted Successfully", nil)
}

func (handler *AppHandler) GetAppHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	projectId := params["project_id"]

	span.SetAttributes(
		attribute.String("project.id", projectId),
		attribute.String("app.id", appId),
	)

	getAppResponse, err := handler.AppServiceClient.GetApp(r.Context(), &app_service_pb.GetAppRequest{
		ProjectId: projectId,
		AppId:     appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "App Fetched Successfully", getAppResponse.App)
}

func (handler *AppHandler) GetAppsHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())

	params := mux.Vars(r)
	projectId := params["project_id"]

	span.SetAttributes(attribute.String("project.id", projectId))

	getAppsResponse, err := handler.AppServiceClient.GetApps(r.Context(), &app_service_pb.GetAppsRequest{ProjectId: projectId})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.Int("apps.count", len(getAppsResponse.Apps)))

	if getAppsResponse.Apps == nil {
		messaging.WriteSuccess(w, "Apps Fetched Successfully", []*app_service_pb.App{})
		return
	}

	messaging.WriteSuccess(w, "Apps Fetched Successfully", getAppsResponse.Apps)
}

func (handler *AppHandler) GetEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app.id", appId))

	getEnvironmentVariablesResponse, err := handler.AppServiceClient.GetEnvironmentVariables(r.Context(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	if getEnvironmentVariablesResponse.EnvironmentVariable == nil {
		messaging.WriteSuccess(w, "Environment Variables Fetched Successfully", []*app_service_pb.EnvironmentVariables{})
		return
	}

	span.SetAttributes(attribute.String("environment_variable.id", getEnvironmentVariablesResponse.EnvironmentVariable.Id))

	messaging.WriteSuccess(w, "Environment Variables Fetched Successfully", getEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) CreateEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app.id", appId))

	createEnvironmentVariablesRequest := app_service_pb.CreateEnvironmentVariablesRequest{}

	err := json.NewDecoder(r.Body).Decode(&createEnvironmentVariablesRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	createEnvironmentVariablesRequest.AppId = appId
	createEnvironmentVariablesResponse, err := handler.AppServiceClient.CreateEnvironmentVariables(r.Context(), &createEnvironmentVariablesRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Created Successfully", createEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) UpdateEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app.id", appId))

	updateEnvironmentVariablesRequest := app_service_pb.UpdateEnvironmentVariablesRequest{}

	err := json.NewDecoder(r.Body).Decode(&updateEnvironmentVariablesRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	updateEnvironmentVariablesRequest.AppId = appId
	updateEnvironmentVariablesResponse, err := handler.AppServiceClient.UpdateEnvironmentVariables(r.Context(), &updateEnvironmentVariablesRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Updated Successfully", updateEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) DeleteEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app.id", appId))

	_, err := handler.AppServiceClient.DeleteEnvironmentVariables(r.Context(), &app_service_pb.DeleteEnvironmentVariablesRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Deleted Successfully", nil)
}
