package main

import (
	"gateway/handlers"
	"net/http"
	"os"

	"apps-hosting.com/logging"

	"github.com/gorilla/mux"

	grpcclients "gateway/grpc_clients"

	gorillaHandlers "github.com/gorilla/handlers"
)

func main() {
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
	router.HandleFunc("/health", gatewayHandler.HealthCheckHandler).Methods("GET")

	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.HandleFunc("/auth", userHandler.AuthHandler).Methods("GET")
	userRouter.HandleFunc("/sign_in", userHandler.SignInHandler).Methods("POST")
	userRouter.HandleFunc("/sign_up", userHandler.SignUpHandler).Methods("POST")

	userGithubRouter := userRouter.PathPrefix("/github").Subrouter()
	userGithubRouter.Use(userHandler.AuthMiddleware)
	userGithubRouter.HandleFunc("/callback", userHandler.GithubCallbackHandler).Methods("GET")
	userGithubRouter.HandleFunc("/repositories", userHandler.GetGithubRepositoriesHandler).Methods("GET")

	router.HandleFunc("/projects", userHandler.AuthMiddleware(http.HandlerFunc(projectHandler.GetUserProjectsHandler)).ServeHTTP).Methods("GET")
	projectRouter := router.PathPrefix("/projects").Subrouter()
	projectRouter.Use(userHandler.AuthMiddleware)

	projectRouter.HandleFunc("/create", projectHandler.CreateProjectHandler).Methods("POST")

	projectScoped := projectRouter.PathPrefix("/{project_id}").Subrouter()
	projectScoped.Use(projectHandler.OwnershipMiddleware)

	projectScoped.HandleFunc("", projectHandler.GetUserProjectByIdHandler).Methods("GET")
	projectScoped.HandleFunc("/", projectHandler.GetUserProjectByIdHandler).Methods("GET")

	projectScoped.HandleFunc("/delete", projectHandler.DeleteProjectHandler).Methods("DELETE")
	projectScoped.HandleFunc("/apps", appHandler.GetAppsHandler).Methods("GET")
	projectScoped.HandleFunc("/apps/create", appHandler.CreateAppHandler).Methods("POST")

	appScoped := projectScoped.PathPrefix("/apps/{app_id}").Subrouter()
	appScoped.Use(appHandler.OwnershipMiddleware)

	appScoped.HandleFunc("", appHandler.GetAppHandler).Methods("GET")
	appScoped.HandleFunc("/", appHandler.GetAppHandler).Methods("GET")

	appScoped.HandleFunc("/update", appHandler.UpdateAppHandler).Methods("PATCH")
	appScoped.HandleFunc("/delete", appHandler.DeleteAppHandler).Methods("DELETE")
	appScoped.HandleFunc("/environment_variables", appHandler.GetEnvironmentVariablesHandler).Methods("GET")
	appScoped.HandleFunc("/environment_variables/create", appHandler.CreateEnvironmentVariablesHandler).Methods("POST")
	appScoped.HandleFunc("/environment_variables/update", appHandler.UpdateEnvironmentVariablesHandler).Methods("PATCH")
	appScoped.HandleFunc("/environment_variables/delete", appHandler.DeleteEnvironmentVariablesHandler).Methods("DELETE")
	appScoped.HandleFunc("/builds", buildHandler.GetBuildsHandler).Methods("GET")
	appScoped.HandleFunc("/deployments", deployHandler.GetDeploymentsHandler).Methods("GET")
	appScoped.HandleFunc("/logs", logHandler.QueryLogsHandler).Methods("GET")

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
