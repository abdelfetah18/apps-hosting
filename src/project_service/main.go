package main

import (
	"net"
	"os"
	"project/database"
	"project/grpc_server"
	"project/proto/project_service_pb"
	"project/repositories"

	"apps-hosting.com/logging"

	"google.golang.org/grpc"
)

func main() {
	logger := logging.NewServiceLogger(logging.ServiceProject)

	database := database.NewDatabase()
	projectRepository := repositories.NewProjectRepository(database, logger)

	err := projectRepository.CreateProjectsTable()
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	grpcProjectServiceServer := grpc_server.NewGRPCProjectServiceServer(&projectRepository, logger)
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
