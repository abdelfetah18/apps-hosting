package main

import (
	"log_service/grpc_server"
	"log_service/kubernetes"
	"log_service/proto/log_service_pb"
	"net"
	"os"

	"apps-hosting.com/logging"

	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
)

func main() {
	logger := logging.NewServiceLogger(logging.ServiceLog)

	config, _ := rest.InClusterConfig()
	kubernetesClient := kubernetes.NewKubernetesClient(config)

	grpcServer := grpc.NewServer()
	grpcLogServiceServer := grpc_server.NewGRPCLogServiceServer(kubernetesClient, logger)
	log_service_pb.RegisterLogServiceServer(grpcServer, grpcLogServiceServer)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8086"
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
