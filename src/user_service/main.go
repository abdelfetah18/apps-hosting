package main

import (
	"net"
	"os"
	"user/database"
	"user/grpc_server"
	"user/repositories"

	"apps-hosting.com/logging"

	user_service_pb "user/proto/user_service_pb"

	"google.golang.org/grpc"
)

func main() {
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

	grpcServer := grpc.NewServer()
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
