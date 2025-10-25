package main

import (
	"net"
	"net/http"
	"os"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"build/database"
	gitclient "build/git_client"
	"build/grpc_server"
	"build/handlers"
	"build/kaniko"
	"build/nats_service"
	"build/proto/build_service_pb"
	"build/repositories"
	"build/runtime"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
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

	go StartRestServer(logger)

	grpcServer := grpc.NewServer()
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

func StartRestServer(logger logging.ServiceLogger) {
	router := mux.NewRouter()

	buildHandler := handlers.NewBuildHandler()

	// API endpoints
	router.HandleFunc("/health", buildHandler.HealthCheckHandler)
	router.HandleFunc("/repos/{repo_id}", buildHandler.DownloadSourceCode).Methods("GET")

	PORT := "8081"

	logger.LogInfo("Http Server running on port " + PORT)
	err := http.ListenAndServe(":"+PORT, router)
	if err != nil {
		logger.LogError(err.Error())
		return
	}
}
