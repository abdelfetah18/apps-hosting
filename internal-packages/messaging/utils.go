package messaging

import (
	"encoding/json"
	"net/http"
)

type HttpResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(HttpResponse[any]{
		Status:  "error",
		Message: message,
	})
}

func WriteSuccess(w http.ResponseWriter, message string, data any) {
	response := HttpResponse[any]{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
