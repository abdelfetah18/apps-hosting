package messaging

import (
	"encoding/json"
	"io"
	"log"
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

// [ Tests ]=========================

func SendGETRequest(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func SendPOSTRequest(url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func RedirectResponse(w http.ResponseWriter, resp *http.Response) {
	w.WriteHeader(resp.StatusCode)

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Println(err)
		WriteError(w, http.StatusInternalServerError, "Internal Server Error")
	}
}
