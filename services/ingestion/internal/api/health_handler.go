package api

import (
	"net/http"
)

// HealthHandler returns the health check handler.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "ingestion",
		})
	}
}
