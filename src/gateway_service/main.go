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
	router.Handle("/health", http.HandlerFunc(gatewayHandler.HealthCheckHandler)).Methods("GET")

	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.Handle("/auth", http.HandlerFunc(userHandler.AuthHandler)).Methods("GET")
	userRouter.Handle("/sign_in", http.HandlerFunc(userHandler.SignInHandler)).Methods("POST")
	userRouter.Handle("/sign_up", http.HandlerFunc(userHandler.SignUpHandler)).Methods("POST")

	userGithubRouter := userRouter.PathPrefix("/github").Subrouter()
	userGithubRouter.Use(userHandler.AuthMiddleware)
	userGithubRouter.Handle("/callback", http.HandlerFunc(userHandler.GithubCallbackHandler)).Methods("GET")
	userGithubRouter.Handle("/repositories", http.HandlerFunc(userHandler.GetGithubRepositoriesHandler)).Methods("GET")

	router.Handle("/projects", userHandler.AuthMiddleware(http.HandlerFunc(projectHandler.GetUserProjectsHandler))).Methods("GET")

	projectRouter := router.PathPrefix("/projects").Subrouter()
	projectRouter.Use(userHandler.AuthMiddleware)
	projectRouter.Handle("/create", http.HandlerFunc(projectHandler.CreateProjectHandler)).Methods("POST")

	projectScoped := projectRouter.PathPrefix("/{project_id}").Subrouter()
	projectScoped.Use(projectHandler.OwnershipMiddleware)

	projectScoped.Handle("", http.HandlerFunc(projectHandler.GetUserProjectByIdHandler)).Methods("GET")
	projectScoped.Handle("/", http.HandlerFunc(projectHandler.GetUserProjectByIdHandler)).Methods("GET")
	projectScoped.Handle("/update", http.HandlerFunc(projectHandler.UpdateProjectHandler)).Methods("PATCH")
	projectScoped.Handle("/delete", http.HandlerFunc(projectHandler.DeleteProjectHandler)).Methods("DELETE")
	projectScoped.Handle("/apps", http.HandlerFunc(appHandler.GetAppsHandler)).Methods("GET")
	projectScoped.Handle("/apps/create", http.HandlerFunc(appHandler.CreateAppHandler)).Methods("POST")

	appScoped := projectScoped.PathPrefix("/apps/{app_id}").Subrouter()
	appScoped.Use(appHandler.OwnershipMiddleware)
	appScoped.Handle("", http.HandlerFunc(appHandler.GetAppHandler)).Methods("GET")
	appScoped.Handle("/", http.HandlerFunc(appHandler.GetAppHandler)).Methods("GET")
	appScoped.Handle("/update", http.HandlerFunc(appHandler.UpdateAppHandler)).Methods("PATCH")
	appScoped.Handle("/delete", http.HandlerFunc(appHandler.DeleteAppHandler)).Methods("DELETE")
	appScoped.Handle("/environment_variables", http.HandlerFunc(appHandler.GetEnvironmentVariablesHandler)).Methods("GET")
	appScoped.Handle("/environment_variables/create", http.HandlerFunc(appHandler.CreateEnvironmentVariablesHandler)).Methods("POST")
	appScoped.Handle("/environment_variables/update", http.HandlerFunc(appHandler.UpdateEnvironmentVariablesHandler)).Methods("PATCH")
	appScoped.Handle("/environment_variables/delete", http.HandlerFunc(appHandler.DeleteEnvironmentVariablesHandler)).Methods("DELETE")
	appScoped.Handle("/builds", http.HandlerFunc(buildHandler.GetBuildsHandler)).Methods("GET")
	appScoped.Handle("/deployments", http.HandlerFunc(deployHandler.GetDeploymentsHandler)).Methods("GET")
	appScoped.Handle("/logs", http.HandlerFunc(logHandler.QueryLogsHandler)).Methods("GET")

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
