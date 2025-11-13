package handlers

import (
	"gateway/proto/build_service_pb"
	"net/http"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app_id", appId))

	getBuildsResponse, err := handler.BuildServiceClient.GetBuilds(r.Context(), &build_service_pb.GetBuildsRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.Int("app_builds_count", len(getBuildsResponse.Builds)))

	if getBuildsResponse.Builds == nil {
		messaging.WriteSuccess(w, "Builds Fetched Successfully", []*build_service_pb.Build{})
		return
	}

	messaging.WriteSuccess(w, "Builds Fetched Successfully", getBuildsResponse.Builds)
}
