package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	// Domain is the base domain for the application (e.g., "proxy.soulreturns.com" or "localhost:8090")
	Domain string
	// AgentPort is the port the agent runs on (default: 9099)
	AgentPort string
	// IsRemote indicates if this is a remote deployment (not localhost)
	IsRemote bool
}

var current Config

// Init initializes the configuration from environment variables
func Init() {
	domain := getEnv("DOMAIN", "localhost:8090")
	agentPort := getEnv("AGENT_PORT", "9099")

	// Determine if this is a remote deployment
	isRemote := domain != "localhost:8090" && domain != "localhost" && domain != "127.0.0.1"

	current = Config{
		Domain:    domain,
		AgentPort: agentPort,
		IsRemote:  isRemote,
	}
}

// Get returns the current configuration
func Get() Config {
	return current
}

// GetAgentURL returns the full URL to connect to the agent
func GetAgentURL() string {
	cfg := Get()
	if cfg.IsRemote {
		// For remote deployments, agent runs on same domain but different port
		return "http://" + cfg.Domain + ":" + cfg.AgentPort
	}
	// For local deployments, agent is on localhost
	return "http://localhost:" + cfg.AgentPort
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
