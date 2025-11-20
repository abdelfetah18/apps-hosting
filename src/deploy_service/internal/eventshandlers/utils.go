package eventshandlers

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

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
