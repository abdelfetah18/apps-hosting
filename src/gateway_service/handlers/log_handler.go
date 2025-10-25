package handlers

import (
	"context"
	"gateway/proto/log_service_pb"
	"net/http"

	"apps-hosting.com/messaging"

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
	params := mux.Vars(r)
	appId := params["app_id"]

	query := r.URL.Query()
	userId := query.Get("user_id")

	QueryLogsResponse, err := handler.LogServiceClient.QueryLogs(context.Background(), &log_service_pb.QueryLogsRequest{
		UserId: userId,
		AppId:  appId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "Builds Fetched Successfully", QueryLogsResponse.Logs)
}
