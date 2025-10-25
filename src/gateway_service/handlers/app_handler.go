package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/app_service_pb"
	"gateway/utils"
	"net/http"

	"apps-hosting.com/messaging"

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
		params := mux.Vars(r)

		projectId := params["project_id"]
		appId := params["app_id"]

		_, err := handler.AppServiceClient.GetApp(context.Background(), &app_service_pb.GetAppRequest{
			AppId:     appId,
			ProjectId: projectId,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)

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
	params := mux.Vars(r)
	projectId := params["project_id"]
	createAppRequest := app_service_pb.CreateAppRequest{}

	err := json.NewDecoder(r.Body).Decode(&createAppRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, "failed at decoding json request")
		return
	}

	createAppRequest.ProjectId = projectId
	createAppResponse, err := handler.AppServiceClient.CreateApp(context.Background(), &createAppRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "App Created Successfully", createAppResponse.App)
}

func (handler *AppHandler) UpdateAppHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["app_id"]
	projectId := params["project_id"]

	updateAppRequest := app_service_pb.UpdateAppRequest{}

	err := json.NewDecoder(r.Body).Decode(&updateAppRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	updateAppRequest.ProjectId = projectId
	updateAppRequest.AppId = appId

	updateAppResponse, err := handler.AppServiceClient.UpdateApp(context.Background(), &updateAppRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "App Updated Successfully", updateAppResponse.App)
}

func (handler *AppHandler) DeleteAppHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	appId := params["app_id"]
	projectId := params["project_id"]

	_, err := handler.AppServiceClient.DeleteApp(context.Background(), &app_service_pb.DeleteAppRequest{
		ProjectId: projectId,
		AppId:     appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "App Deleted Successfully", nil)
}

func (handler *AppHandler) GetAppHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	appId := params["app_id"]
	projectId := params["project_id"]

	getAppResponse, err := handler.AppServiceClient.GetApp(context.Background(), &app_service_pb.GetAppRequest{
		ProjectId: projectId,
		AppId:     appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "App Fetched Successfully", getAppResponse.App)
}

func (handler *AppHandler) GetAppsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	projectId := params["project_id"]

	getAppsResponse, err := handler.AppServiceClient.GetApps(context.Background(), &app_service_pb.GetAppsRequest{ProjectId: projectId})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	if getAppsResponse.Apps == nil {
		messaging.WriteSuccess(w, "Apps Fetched Successfully", []*app_service_pb.App{})
		return
	}

	messaging.WriteSuccess(w, "Apps Fetched Successfully", getAppsResponse.Apps)
}

func (handler *AppHandler) GetEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["app_id"]

	getEnvironmentVariablesResponse, err := handler.AppServiceClient.GetEnvironmentVariables(context.Background(), &app_service_pb.GetEnvironmentVariablesRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	if getEnvironmentVariablesResponse.EnvironmentVariable == nil {
		messaging.WriteSuccess(w, "Environment Variables Fetched Successfully", []*app_service_pb.EnvironmentVariables{})
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Fetched Successfully", getEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) CreateEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["app_id"]

	createEnvironmentVariablesRequest := app_service_pb.CreateEnvironmentVariablesRequest{}

	err := json.NewDecoder(r.Body).Decode(&createEnvironmentVariablesRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	createEnvironmentVariablesRequest.AppId = appId
	createEnvironmentVariablesResponse, err := handler.AppServiceClient.CreateEnvironmentVariables(context.Background(), &createEnvironmentVariablesRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Created Successfully", createEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) UpdateEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["app_id"]

	updateEnvironmentVariablesRequest := app_service_pb.UpdateEnvironmentVariablesRequest{}

	err := json.NewDecoder(r.Body).Decode(&updateEnvironmentVariablesRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	updateEnvironmentVariablesRequest.AppId = appId
	updateEnvironmentVariablesResponse, err := handler.AppServiceClient.UpdateEnvironmentVariables(context.Background(), &updateEnvironmentVariablesRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Updated Successfully", updateEnvironmentVariablesResponse.EnvironmentVariable)
}

func (handler *AppHandler) DeleteEnvironmentVariablesHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	appId := params["app_id"]

	_, err := handler.AppServiceClient.DeleteEnvironmentVariables(context.Background(), &app_service_pb.DeleteEnvironmentVariablesRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, utils.GrpcCodeToHttpStatusCode(status.Code()), status.Message())
		return
	}

	messaging.WriteSuccess(w, "Environment Variables Deleted Successfully", nil)
}
