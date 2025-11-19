package core

import (
	"apps-hosting.com/buildservice/internal/models"
	"apps-hosting.com/buildservice/proto/build_service_pb"
)

func BuildToProto(build *models.Build) *build_service_pb.Build {
	return &build_service_pb.Build{
		Id:         build.Id,
		AppId:      build.AppId,
		Status:     string(build.Status),
		ImageUrl:   build.ImageURL,
		CommitHash: build.CommitHash,
		CreatedAt:  build.CreatedAt.String(),
	}
}

func BuildListToProto(builds []models.Build) []*build_service_pb.Build {
	_builds := make([]*build_service_pb.Build, 0, len(builds))
	for _, build := range builds {
		_builds = append(_builds, BuildToProto(&build))
	}
	return _builds
}

func ExtractBuildIDs(builds []models.Build) []string {
	buildIDs := make([]string, 0, len(builds))
	for _, build := range builds {
		buildIDs = append(buildIDs, build.Id)
	}
	return buildIDs
}
