package main

import (
	"context"
	"gateway/handlers"
	"net/http"
	"os"

	"apps-hosting.com/logging"

	grpcclients "gateway/grpc_clients"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
			semconv.ServiceName("gateway-service"),
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

	logger := logging.NewServiceLogger(logging.ServiceGateway)

	grpcClients, err := grpcclients.NewGrpcClients()
	if err != nil {
		logger.LogError(err.Error())
		return
	}

	projectHandler := handlers.NewProjectHandler(grpcClients.ProjectServiceClient, logger)
	appHandler := handlers.NewAppHandler(grpcClients.AppServiceClient, logger)
	userHandler := handlers.NewUserHandler(grpcClients.UserServiceClient, logger)
	buildHandler := handlers.NewBuildHandler(grpcClients.BuildServiceClient, logger)
	deployHandler := handlers.NewDeployHandler(grpcClients.DeployServiceClient, logger)
	logHandler := handlers.NewLogHandler(grpcClients.LogServiceClient, logger)
	gatewayHandler := handlers.NewGatewayHandler(logger)

	router := mux.NewRouter()
	router.Use(gatewayHandler.LoggingMiddleware)

	// API Endpoints
	router.Handle("/health", otelhttp.NewHandler(
		http.HandlerFunc(gatewayHandler.HealthCheckHandler),
		"HealthCheckHandler",
	)).Methods("GET")

	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.Handle("/auth", otelhttp.NewHandler(http.HandlerFunc(userHandler.AuthHandler), "AuthHandler")).Methods("GET")
	userRouter.Handle("/sign_in", otelhttp.NewHandler(http.HandlerFunc(userHandler.SignInHandler), "SignInHandler")).Methods("POST")
	userRouter.Handle("/sign_up", otelhttp.NewHandler(http.HandlerFunc(userHandler.SignUpHandler), "SignUpHandler")).Methods("POST")

	userGithubRouter := userRouter.PathPrefix("/github").Subrouter()
	userGithubRouter.Use(userHandler.AuthMiddleware)
	userGithubRouter.Handle("/callback", otelhttp.NewHandler(http.HandlerFunc(userHandler.GithubCallbackHandler), "GithubCallbackHandler")).Methods("GET")
	userGithubRouter.Handle("/repositories", otelhttp.NewHandler(http.HandlerFunc(userHandler.GetGithubRepositoriesHandler), "GetGithubRepositoriesHandler")).Methods("GET")

	router.Handle("/projects", userHandler.AuthMiddleware(
		otelhttp.NewHandler(http.HandlerFunc(projectHandler.GetUserProjectsHandler), "GetUserProjectsHandler")),
	).Methods("GET")

	projectRouter := router.PathPrefix("/projects").Subrouter()
	projectRouter.Use(userHandler.AuthMiddleware)
	projectRouter.Handle("/create", otelhttp.NewHandler(http.HandlerFunc(projectHandler.CreateProjectHandler), "CreateProjectHandler")).Methods("POST")

	projectScoped := projectRouter.PathPrefix("/{project_id}").Subrouter()
	projectScoped.Use(projectHandler.OwnershipMiddleware)

	projectScoped.Handle("", otelhttp.NewHandler(http.HandlerFunc(projectHandler.GetUserProjectByIdHandler), "GetUserProjectByIdHandler")).Methods("GET")
	projectScoped.Handle("/", otelhttp.NewHandler(http.HandlerFunc(projectHandler.GetUserProjectByIdHandler), "GetUserProjectByIdHandlerSlash")).Methods("GET")
	projectScoped.Handle("/delete", otelhttp.NewHandler(http.HandlerFunc(projectHandler.DeleteProjectHandler), "DeleteProjectHandler")).Methods("DELETE")
	projectScoped.Handle("/apps", otelhttp.NewHandler(http.HandlerFunc(appHandler.GetAppsHandler), "GetAppsHandler")).Methods("GET")
	projectScoped.Handle("/apps/create", otelhttp.NewHandler(http.HandlerFunc(appHandler.CreateAppHandler), "CreateAppHandler")).Methods("POST")

	appScoped := projectScoped.PathPrefix("/apps/{app_id}").Subrouter()
	appScoped.Use(appHandler.OwnershipMiddleware)
	appScoped.Handle("", otelhttp.NewHandler(http.HandlerFunc(appHandler.GetAppHandler), "GetAppHandler")).Methods("GET")
	appScoped.Handle("/", otelhttp.NewHandler(http.HandlerFunc(appHandler.GetAppHandler), "GetAppHandlerSlash")).Methods("GET")
	appScoped.Handle("/update", otelhttp.NewHandler(http.HandlerFunc(appHandler.UpdateAppHandler), "UpdateAppHandler")).Methods("PATCH")
	appScoped.Handle("/delete", otelhttp.NewHandler(http.HandlerFunc(appHandler.DeleteAppHandler), "DeleteAppHandler")).Methods("DELETE")
	appScoped.Handle("/environment_variables", otelhttp.NewHandler(http.HandlerFunc(appHandler.GetEnvironmentVariablesHandler), "GetEnvironmentVariablesHandler")).Methods("GET")
	appScoped.Handle("/environment_variables/create", otelhttp.NewHandler(http.HandlerFunc(appHandler.CreateEnvironmentVariablesHandler), "CreateEnvironmentVariablesHandler")).Methods("POST")
	appScoped.Handle("/environment_variables/update", otelhttp.NewHandler(http.HandlerFunc(appHandler.UpdateEnvironmentVariablesHandler), "UpdateEnvironmentVariablesHandler")).Methods("PATCH")
	appScoped.Handle("/environment_variables/delete", otelhttp.NewHandler(http.HandlerFunc(appHandler.DeleteEnvironmentVariablesHandler), "DeleteEnvironmentVariablesHandler")).Methods("DELETE")
	appScoped.Handle("/builds", otelhttp.NewHandler(http.HandlerFunc(buildHandler.GetBuildsHandler), "GetBuildsHandler")).Methods("GET")
	appScoped.Handle("/deployments", otelhttp.NewHandler(http.HandlerFunc(deployHandler.GetDeploymentsHandler), "GetDeploymentsHandler")).Methods("GET")
	appScoped.Handle("/logs", otelhttp.NewHandler(http.HandlerFunc(logHandler.QueryLogsHandler), "QueryLogsHandler")).Methods("GET")

	// Start server
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	// CORS middleware
	// FIXME: should not allow any domain
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(router)

	logger.LogInfo("Http Server running on port " + PORT)

	http.ListenAndServe(":"+PORT, corsHandler)
}
