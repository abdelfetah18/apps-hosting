package main

import (
	"context"
	"net"
	"os"

	"apps-hosting.com/buildservice/internal/buildexecutor"
	"apps-hosting.com/buildservice/internal/core"
	"apps-hosting.com/buildservice/internal/database"
	"apps-hosting.com/buildservice/internal/eventshandlers"
	"apps-hosting.com/buildservice/internal/repomanager"
	"apps-hosting.com/buildservice/internal/repositories"
	"apps-hosting.com/buildservice/internal/tracer"
	"apps-hosting.com/buildservice/proto/build_service_pb"
	"apps-hosting.com/buildservice/proto/user_service_pb"
	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	logger := logging.NewServiceLogger(logging.ServiceBuild)
	logger.LogInfoF("Service '%s' is starting", os.Getenv("SERVICE_NAME"))

	ctx := context.Background()
	shutdown := tracer.SetupTracer(ctx, serviceName)
	defer shutdown(ctx)

	logger.LogInfo("Connecting to the database")
	database := database.NewDatabase()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	buildRepository := repositories.NewBuildRepository(database, logger)

	_, err = buildRepository.CreateBuildsTable()
	if err != nil {
		logger.LogError(err.Error())
		return
	}

	natsURL := os.Getenv("NATS_URL")
	eventBus, err := messaging.NewEventBus(
		serviceName,
		natsURL,
		events_pb.StreamName_BUILD_STREAM,
		[]events_pb.EventName{
			events_pb.EventName_BUILD_COMPLETED,
			events_pb.EventName_BUILD_FAILED,
		},
	)
	if err != nil {
		panic(err)
	}

	_userServiceClient, err := grpc.NewClient(
		os.Getenv("USER_SERVICE"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.LogError(err.Error())
		return
	}

	userServiceClient := user_service_pb.NewUserServiceClient(_userServiceClient)

	kanikoExecutor := buildexecutor.NewKanikoExecutor(clientset, logger)
	gitRepoManager := repomanager.NewGitRepoManager()

	eventsHandlers := eventshandlers.NewEventsHandlers(
		*eventBus,
		&kanikoExecutor,
		gitRepoManager,
		buildRepository,
		userServiceClient,
		logger,
	)

	eventBus.Subscribe(events_pb.EventName_APP_CREATED, eventsHandlers.HandleAppCreatedEvent)
	eventBus.Subscribe(events_pb.EventName_APP_DELETED, eventsHandlers.HandleAppDeletedEvent)

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	buildServiceServer := core.NewBuildServiceServer(buildRepository)
	build_service_pb.RegisterBuildServiceServer(grpcServer, buildServiceServer)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	lis, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		logger.LogErrorF("failed to listen: %v", err)
	}

	logger.LogErrorF("GRPC server is listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		logger.LogErrorF("failed to serve: %v", err)
		return
	}
}
