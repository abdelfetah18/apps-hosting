package main

import (
	"app/database"
	"app/grpc_server"
	"app/proto/app_service_pb"
	"app/repositories"
	"context"
	"net"
	"os"

	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"
	"apps-hosting.com/messaging/proto/events_pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"google.golang.org/grpc"
)

// setupTracer initializes OpenTelemetry tracing.
func setupTracer(ctx context.Context) func(context.Context) error {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("app-service"),
		)),
	)
	otel.SetTracerProvider(tp)

	// Set up propagator.
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	return tp.Shutdown
}

func main() {
	ctx := context.Background()
	shutdown := setupTracer(ctx)
	defer shutdown(ctx)

	logger := logging.NewServiceLogger(logging.ServiceApp)

	logger.LogInfo("App Service running")

	logger.LogInfo("Connecting to the app_service Database.")
	database := database.NewDatabase()

	appRepository := repositories.NewAppRepository(database, logger)
	environmentVariablesRepository := repositories.NewEnvironmentVariablesRepository(database, logger)
	gitRepositoryRepository := repositories.NewGitRepositoryRepository(database, logger)

	_, err := appRepository.CreateAppsTable()
	if err != nil {
		panic(err)
	}

	_, err = environmentVariablesRepository.CreateEnvironmentVariablesTable()
	if err != nil {
		panic(err)
	}

	_, err = gitRepositoryRepository.CreateGitRepositoryRepositoryTable()
	if err != nil {
		panic(err)
	}

	natsURL := os.Getenv("NATS_URL")
	eventBus, err := messaging.NewEventBus(
		natsURL,
		events_pb.StreamName_APP_STREAM,
		[]events_pb.EventName{
			events_pb.EventName_APP_CREATED,
			events_pb.EventName_APP_DELETED,
		},
	)

	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcAppServiceServer := grpc_server.NewGRPCAppServiceServer(appRepository, environmentVariablesRepository, gitRepositoryRepository, *eventBus, logger)
	app_service_pb.RegisterAppServiceServer(grpcServer, grpcAppServiceServer)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8084"
	}

	lis, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		logger.LogErrorF("failed to listen: %v", err)
	}

	logger.LogErrorF("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		logger.LogErrorF("failed to serve: %v", err)
		return
	}
}
