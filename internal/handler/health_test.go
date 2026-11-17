package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Health("node-test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("incorrect status: excepted %d, received %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("incorrect status field: %q", resp.Status)
	}

	if resp.NodeID != "node-test" {
		t.Errorf("incorrect node_id: %q", resp.NodeID)
	}
}
