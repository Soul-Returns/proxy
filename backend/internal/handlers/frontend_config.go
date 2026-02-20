package handlers

import (
	"devproxy/internal/config"
	"encoding/json"
	"net/http"
)

// FrontendConfigResponse holds configuration data for the frontend
type FrontendConfigResponse struct {
	AgentURL string `json:"agentUrl"`
	Domain   string `json:"domain"`
	IsRemote bool   `json:"isRemote"`
}

// GetFrontendConfig returns frontend configuration
func GetFrontendConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()

	response := FrontendConfigResponse{
		AgentURL: config.GetAgentURL(),
		Domain:   cfg.Domain,
		IsRemote: cfg.IsRemote,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
