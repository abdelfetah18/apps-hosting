package grpc_server

import (
	"build/proto/build_service_pb"
	"build/repositories"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCBuildServiceServer struct {
	build_service_pb.UnimplementedBuildServiceServer
	BuildRepository repositories.BuildRepository
}

func NewGRPCBuildServiceServer(buildRepository repositories.BuildRepository) *GRPCBuildServiceServer {
	return &GRPCBuildServiceServer{
		BuildRepository: buildRepository,
	}
}

func (server *GRPCBuildServiceServer) Health(ctx context.Context, _ *build_service_pb.HealthRequest) (*build_service_pb.HealthResponse, error) {
	return &build_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCBuildServiceServer) GetBuilds(ctx context.Context, getBuildsRequest *build_service_pb.GetBuildsRequest) (*build_service_pb.GetBuildsResponse, error) {
	builds, err := server.BuildRepository.GetBuilds(getBuildsRequest.AppId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	var _builds []*build_service_pb.Build
	for _, build := range builds {
		_build := build_service_pb.Build{
			Id:         build.Id,
			AppId:      build.AppId,
			Status:     string(build.Status),
			ImageUrl:   build.ImageURL,
			CommitHash: build.CommitHash,
			CreatedAt:  build.CreatedAt.String(),
		}
		_builds = append(_builds, &_build)
	}

	return &build_service_pb.GetBuildsResponse{
		Builds: _builds,
	}, nil
}
