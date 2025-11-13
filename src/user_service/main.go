package main

import (
	"context"
	"net"
	"os"
	"user/database"
	"user/grpc_server"
	"user/repositories"

	"apps-hosting.com/logging"

	user_service_pb "user/proto/user_service_pb"

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
			semconv.ServiceName("user-service"),
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

	logger := logging.NewServiceLogger(logging.ServiceUser)

	logger.LogInfo("User Service running")

	logger.LogInfo("Connecting to the Database.")
	database := database.NewDatabase()

	userRepository := repositories.NewUserRepository(database, logger)
	userSessionRepository := repositories.NewUserSessionRepository(database, logger)

	_, err := userRepository.CreateUsersTable()
	if err != nil {
		panic(err)
	}

	_, err = userSessionRepository.CreateUserSessionsTable()
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcUserServiceServer := grpc_server.NewGRPCUserServiceServer(userRepository, userSessionRepository, logger)
	user_service_pb.RegisterUserServiceServer(grpcServer, grpcUserServiceServer)

	// Start server
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8081"
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
