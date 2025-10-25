package handlers

import (
	"context"
	"gateway/proto/build_service_pb"
	"net/http"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/status"
)

type BuildHandler struct {
	BuildServiceClient build_service_pb.BuildServiceClient
	Logger             logging.ServiceLogger
}

func NewBuildHandler(buildServiceClient build_service_pb.BuildServiceClient, logger logging.ServiceLogger) BuildHandler {
	return BuildHandler{
		BuildServiceClient: buildServiceClient,
		Logger:             logger,
	}
}

func (handler *BuildHandler) GetBuildsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appId := params["app_id"]

	getBuildsResponse, err := handler.BuildServiceClient.GetBuilds(context.Background(), &build_service_pb.GetBuildsRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	if getBuildsResponse.Builds == nil {
		messaging.WriteSuccess(w, "Builds Fetched Successfully", []*build_service_pb.Build{})
		return
	}

	messaging.WriteSuccess(w, "Builds Fetched Successfully", getBuildsResponse.Builds)
}
