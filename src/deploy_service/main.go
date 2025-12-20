package main

import (
	"context"
	"net"
	"os"

	"apps-hosting.com/deployservice/internal/core"
	"apps-hosting.com/deployservice/internal/database"
	"apps-hosting.com/deployservice/internal/eventshandlers"
	"apps-hosting.com/deployservice/internal/repositories"
	"apps-hosting.com/deployservice/internal/tracer"
	"apps-hosting.com/deployservice/proto/app_service_pb"
	"apps-hosting.com/deployservice/proto/deploy_service_pb"
	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	serviceName := os.Getenv("SERVICE_NAME")

	shutdown := tracer.SetupTracer(ctx, serviceName)
	defer shutdown(ctx)

	logger := logging.NewServiceLogger(logging.ServiceDeploy)

	logger.LogInfo("Deploy Service running")

	logger.LogInfo("Connecting to the database.")
	database := database.NewDatabase()

	deploymentRepository := repositories.NewDeploymentRepository(database, logger)
	_, err := deploymentRepository.CreateDeploymentsTable()
	if err != nil {
		panic(err)
	}

	_appServiceClient, err := grpc.NewClient(
		os.Getenv("APP_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}))
	if err != nil {
		panic(err)
	}
	appServiceClient := app_service_pb.NewAppServiceClient(_appServiceClient)

	natsURL := os.Getenv("NATS_URL")
	eventBus, err := messaging.NewEventBus(
		serviceName,
		natsURL,
		events_pb.StreamName_DEPLOY_STREAM,
		[]events_pb.EventName{
			events_pb.EventName_DEPLOY_COMPLETED,
			events_pb.EventName_DEPLOY_FAILED,
		},
	)
	if err != nil {
		panic(err)
	}

	eventsHandlers := eventshandlers.NewEventsHandlers(*eventBus, appServiceClient, deploymentRepository, logger)

	err = eventBus.Subscribe(events_pb.EventName_BUILD_COMPLETED, eventsHandlers.HandleBuildCompletedEvent)
	if err != nil {
		logger.LogErrorF("failed to subscribe to '%s': %v", events_pb.EventName_name[int32(events_pb.EventName_BUILD_COMPLETED)], err)
	}
	err = eventBus.Subscribe(events_pb.EventName_APP_DELETED, eventsHandlers.HandleAppDeletedEvent)
	if err != nil {
		logger.LogErrorF("failed to subscribe to '%s': %v", events_pb.EventName_name[int32(events_pb.EventName_APP_DELETED)], err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcDeployServiceServer := core.NewGRPCDeployServiceServer(deploymentRepository)
	deploy_service_pb.RegisterDeployServiceServer(grpcServer, grpcDeployServiceServer)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	lis, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		logger.LogErrorF("failed to listen: %v", err)
	}

	logger.LogErrorF("GRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		logger.LogErrorF("failed to serve: %v", err)
		return
	}
}
