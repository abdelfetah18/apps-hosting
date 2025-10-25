package utils

import (
	"net/http"
)

func HealthCheck(serviceURL string) bool {
	resp, err := http.Get(serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
