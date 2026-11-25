package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config stores the entire runtime configuration of a node
type Config struct {
	NodeID          string
	HTTPAddr        string
	DatabaseURL     string
	PeerURLs        []string
	ShutdownTimeout time.Duration
}

// Load reads the configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error while loading environment variables")
	}

	cfg := &Config{
		NodeID:          getEnv("NODE_ID", "node-1"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		PeerURLs:        getEnvList("PEER_URLS"),
		ShutdownTimeout: 10 * time.Second,
	}

	if cfg.NodeID == "" {
		return nil, fmt.Errorf("NODE_ID can not be empty")
	}
	return cfg, nil
}

// getEnv returns the value if the environment variable exists, otherwise, it returns the default.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvList(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}

	var res []string

	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			res = append(res, p)
		}
	}
	return res
}
