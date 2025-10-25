package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func ToK8sLabelValue(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
}

func ToK8sDeploymentName(appName string) string {
	return ToK8sLabelValue(appName) + "-deployment"
}

func ToK8sContainerName(appName string) string {
	return ToK8sLabelValue(appName) + "-container"
}

func ToK8sServiceName(appName string) string {
	return ToK8sLabelValue(appName) + "-service"
}

func ToK8sIngressName(appName string) string {
	return ToK8sLabelValue(appName) + "-ingress"
}

func ConvertJSONToEnvVars(jsonStr string) ([]v1.EnvVar, error) {
	var envMap map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &envMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	envVars := make([]v1.EnvVar, 0, len(envMap))
	for k, v := range envMap {
		envVars = append(envVars, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return envVars, nil
}
