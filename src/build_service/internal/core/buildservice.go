package core

import (
	"context"

	"apps-hosting.com/buildservice/internal/repositories"
	"apps-hosting.com/buildservice/proto/build_service_pb"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BuildServiceServer struct {
	build_service_pb.UnimplementedBuildServiceServer

	buildRepository repositories.BuildRepository
}

func NewBuildServiceServer(buildRepository repositories.BuildRepository) *BuildServiceServer {
	return &BuildServiceServer{
		buildRepository: buildRepository,
	}
}

func (s *BuildServiceServer) Health(ctx context.Context, _ *build_service_pb.HealthRequest) (*build_service_pb.HealthResponse, error) {
	return &build_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (s *BuildServiceServer) GetBuilds(ctx context.Context, getBuildsRequest *build_service_pb.GetBuildsRequest) (*build_service_pb.GetBuildsResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("app_id", getBuildsRequest.AppId))

	builds, err := s.buildRepository.GetBuilds(ctx, getBuildsRequest.AppId)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(
		attribute.StringSlice("builds_ids", ExtractBuildIDs(builds)),
		attribute.Int("builds_count", len(builds)),
	)

	return &build_service_pb.GetBuildsResponse{
		Builds: BuildListToProto(builds),
	}, nil
}
