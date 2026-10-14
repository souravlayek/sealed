package handlers

import (
	"encoding/json"
	"net/http"
)

// HomeHandler handles the home route
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Welcome to sealed-server!",
		"status":  "running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthHandler handles the health check route
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "sealed-server",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
