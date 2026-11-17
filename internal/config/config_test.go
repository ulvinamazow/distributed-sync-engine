package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("NODE_ID", "")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.NodeID != "node-1" {
		t.Errorf("incorrect default NodeID: %q", cfg.NodeID)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("incorrect default HTTPAddr: %q", cfg.HTTPAddr)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("NODE_ID", "node-42")
	t.Setenv("HTTP_ADDR", ":9999")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.NodeID != "node-42" {
		t.Errorf("NodeID not readed .env: %s", cfg.NodeID)
	}

	if cfg.HTTPAddr != ":9999" {
		t.Errorf("HTTP_ADDR not readed .env: %q", cfg.HTTPAddr)
	}
}
