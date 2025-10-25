package main

import (
	"deploy/database"
	grpcclients "deploy/grpc_clients"
	"deploy/grpc_server"
	"deploy/nats_service"
	"deploy/proto/deploy_service_pb"
	"deploy/repositories"
	"net"
	"os"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	logger := logging.NewServiceLogger(logging.ServiceDeploy)

	logger.LogInfo("Deploy Service running")

	logger.LogInfo("Connecting to the database.")
	database := database.NewDatabase()

	deploymentRepository := repositories.NewDeploymentRepository(database, logger)
	_, err := deploymentRepository.CreateDeploymentsTable()
	if err != nil {
		panic(err)
	}

	grpcClients, err := grpcclients.NewGrpcClients()
	if err != nil {
		panic(err)
	}

	natsURL := os.Getenv("NATS_URL")
	natsConnection, err := nats.Connect(natsURL)
	if err != nil {
		logger.LogError(err.Error())
	}

	jetStream, err := natsConnection.JetStream()
	if err != nil {
		panic(err)
	}

	_, err = jetStream.AddStream(&nats.StreamConfig{
		Name: "DEPLOY_EVENTS",
		Subjects: []string{
			messaging.DeployCompleted,
			messaging.DeployFailed,
		},
	})
	if err != nil {
		panic(err)
	}

	natsHandler := nats_service.NewNatsHandler(jetStream, *grpcClients, deploymentRepository, logger)

	natsClient := nats_service.NewNatsClient(jetStream, natsHandler)

	natsClient.SubscribeToEvents()

	grpcServer := grpc.NewServer()
	grpcDeployServiceServer := grpc_server.NewGRPCDeployServiceServer(deploymentRepository)
	deploy_service_pb.RegisterDeployServiceServer(grpcServer, grpcDeployServiceServer)

	// Start server
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
