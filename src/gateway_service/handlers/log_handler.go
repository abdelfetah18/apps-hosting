package handlers

import (
	"gateway/proto/log_service_pb"
	"net/http"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/status"
)

type LogHandler struct {
	LogServiceClient log_service_pb.LogServiceClient
	Logger           logging.ServiceLogger
}

func NewLogHandler(logServiceClient log_service_pb.LogServiceClient, logger logging.ServiceLogger) LogHandler {
	return LogHandler{
		LogServiceClient: logServiceClient,
		Logger:           logger,
	}
}

func (handler *LogHandler) QueryLogsHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	query := r.URL.Query()
	userId := query.Get("user_id")

	span.SetAttributes(
		attribute.String("app.id", appId),
		attribute.String("user.id", userId),
	)

	QueryLogsResponse, err := handler.LogServiceClient.QueryLogs(r.Context(), &log_service_pb.QueryLogsRequest{
		UserId: userId,
		AppId:  appId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	messaging.WriteSuccess(w, "Builds Fetched Successfully", QueryLogsResponse.Logs)
}
