package utils

import (
	"encoding/json"
	"net/http"
)

type ErrResponse struct {
	Message string `json:"message"`
}

func SendErrorMessage(w http.ResponseWriter, statusCode int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrResponse{
		Message: message,
	}

	return json.NewEncoder(w).Encode(response)
}

type EnvelopResponse[T any] struct {
	Data T `json:"data"`
}

func SendResponse[T any](w http.ResponseWriter, data T, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(EnvelopResponse[T]{
		Data: data,
	})
}
