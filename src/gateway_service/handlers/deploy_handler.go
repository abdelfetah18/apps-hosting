package handlers

import (
	"net/http"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"apps-hosting.com/logging"

	"gateway/proto/deploy_service_pb"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/status"
)

type DeployHandler struct {
	DeployServiceClient deploy_service_pb.DeployServiceClient
	Logger              logging.ServiceLogger
}

func NewDeployHandler(deployServiceClient deploy_service_pb.DeployServiceClient, logger logging.ServiceLogger) DeployHandler {
	return DeployHandler{
		DeployServiceClient: deployServiceClient,
		Logger:              logger,
	}
}

func (handler *DeployHandler) GetDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	params := mux.Vars(r)

	appId := params["app_id"]
	span.SetAttributes(attribute.String("app_id", appId))

	getDeploymentsResponse, err := handler.DeployServiceClient.GetDeployments(r.Context(), &deploy_service_pb.GetDeploymentsRequest{
		AppId: appId,
	})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.Int("app_deployments_count", len(getDeploymentsResponse.Deployments)))

	if getDeploymentsResponse.Deployments == nil {
		messaging.WriteSuccess(w, "Deployments Fetched Successfully", []*deploy_service_pb.Deployment{})
		return
	}

	messaging.WriteSuccess(w, "Deployments Fetched Successfully", getDeploymentsResponse.Deployments)
}
