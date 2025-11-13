package main

import (
	"context"
	"net"
	"os"

	"apps-hosting.com/logging"
	"apps-hosting.com/messaging"

	"build/database"
	gitclient "build/git_client"
	"build/grpc_server"
	"build/kaniko"
	"build/nats_service"
	"build/proto/build_service_pb"
	"build/repositories"
	"build/runtime"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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
			semconv.ServiceName("build-service"),
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

	logger := logging.NewServiceLogger(logging.ServiceBuild)

	logger.LogInfo("Build Service running")

	logger.LogInfo("Connecting to the database.")
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
	natsConnection, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}

	jetStream, err := natsConnection.JetStream()
	if err != nil {
		panic(err)
	}

	_, err = jetStream.AddStream(&nats.StreamConfig{
		Name: "BUILD_EVENTS",
		Subjects: []string{
			messaging.BuildCompleted,
			messaging.BuildFailed,
		},
	})
	if err != nil {
		panic(err)
	}

	kanikoBuilder := kaniko.NewKanikoBuilder(clientset, logger)
	gitRepoManager := gitclient.NewGitRepoManager()
	runtimeBuilder := runtime.NewRuntimeBuilder(logger)
	natsHandler := nats_service.NewNatsHandler(jetStream, kanikoBuilder, gitRepoManager, runtimeBuilder, buildRepository, logger)

	natsService := nats_service.NewNatsClient(jetStream, natsHandler, logger)

	natsService.SubscribeToEvents()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcBuildServiceServer := grpc_server.NewGRPCBuildServiceServer(buildRepository)
	build_service_pb.RegisterBuildServiceServer(grpcServer, grpcBuildServiceServer)

	// Start server
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8083"
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
