package handler

import (
	"encoding/json"
	"net/http"
)

// HealthResponse - the endpoint's JSON response
type HealthResponse struct {
	Status string `json:"status"`
	NodeID string `json:"node_id"`
}

// returns an http.HandlerFunc for /health endpoint
func Health(nodeID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status: "ok",
			NodeID: nodeID,
		})
	}
}
