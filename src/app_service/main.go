package main

import (
	"app/database"
	"app/grpc_server"
	"app/nats"
	"app/proto/app_service_pb"
	"app/repositories"
	"net"
	"os"

	"apps-hosting.com/logging"

	"google.golang.org/grpc"
)

func main() {
	logger := logging.NewServiceLogger(logging.ServiceApp)

	logger.LogInfo("App Service running")

	logger.LogInfo("Connecting to the repository.Database.")
	database := database.NewDatabase()

	appRepository := repositories.NewAppRepository(database, logger)
	environmentVariablesRepository := repositories.NewEnvironmentVariablesRepository(database, logger)

	_, err := appRepository.CreateAppsTable()
	if err != nil {
		panic(err)
	}

	_, err = environmentVariablesRepository.CreateEnvironmentVariablesTable()
	if err != nil {
		panic(err)
	}

	natsService, err := nats.NewNatsService(logger)
	if err != nil {
		panic(err)
	}

	natsService.SubscribeToEvents()

	grpcServer := grpc.NewServer()
	grpcAppServiceServer := grpc_server.NewGRPCAppServiceServer(appRepository, environmentVariablesRepository, *natsService, logger)
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
