package grpc_server

import (
	"build/proto/build_service_pb"
	"build/repositories"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("app_id", getBuildsRequest.AppId))

	builds, err := server.BuildRepository.GetBuilds(ctx, getBuildsRequest.AppId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: create a util function
	var _builds []*build_service_pb.Build
	builds_ids := []string{}
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
		builds_ids = append(builds_ids, build.Id)
	}

	span.SetAttributes(
		attribute.StringSlice("builds_ids", builds_ids),
		attribute.Int("builds_count", len(builds)),
	)

	return &build_service_pb.GetBuildsResponse{
		Builds: _builds,
	}, nil
}
