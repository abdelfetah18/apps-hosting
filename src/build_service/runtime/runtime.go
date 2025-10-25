package runtime

import (
	"fmt"
	"path/filepath"

	"apps-hosting.com/logging"

	"build/utils"
)

type Runtime = string

const (
	RuntimeNodeJS Runtime = "NodeJS"
)

var DockerfileMap = map[Runtime]string{
	RuntimeNodeJS: "assets/runtime/NodeJS.Dockerfile",
}

type RuntimeBuilder struct {
	Logger logging.ServiceLogger
}

func NewRuntimeBuilder(logger logging.ServiceLogger) RuntimeBuilder {
	return RuntimeBuilder{
		Logger: logger,
	}
}

func (builder *RuntimeBuilder) CopyDockerfile(repoPath string, runtime Runtime) (string, error) {
	src, exists := DockerfileMap[runtime]
	if !exists {
		return "", fmt.Errorf("unsupported runtime: %s", runtime)
	}

	dest := filepath.Join(repoPath, "Dockerfile")

	err := utils.CopyFile(src, dest)
	if err != nil {
		return "", err
	}

	builder.Logger.LogInfo(fmt.Sprintf("Copied %s Dockerfile to %s", runtime, dest))

	return dest, nil
}
