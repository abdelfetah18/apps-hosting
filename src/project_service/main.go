package main

import (
	"context"
	"net"
	"os"
	"project/database"
	"project/grpc_server"
	"project/proto/project_service_pb"
	"project/repositories"

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
			semconv.ServiceName("project-service"),
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

	logger := logging.NewServiceLogger(logging.ServiceProject)

	database := database.NewDatabase()
	projectRepository := repositories.NewProjectRepository(database, logger)

	err := projectRepository.CreateProjectsTable()
	if err != nil {
		panic(err)
	}

	natsURL := os.Getenv("NATS_URL")
	eventBus, err := messaging.NewEventBus(
		natsURL,
		events_pb.StreamName_PROJECT_STREAM,
		[]events_pb.EventName{
			events_pb.EventName_PROJECT_DELETED,
		},
	)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcProjectServiceServer := grpc_server.NewGRPCProjectServiceServer(&projectRepository, eventBus, logger)
	project_service_pb.RegisterProjectServiceServer(grpcServer, grpcProjectServiceServer)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8087"
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
