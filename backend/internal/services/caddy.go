// Package services provides background services and business logic.
package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"devproxy/internal/database"
	"devproxy/internal/models"
)

var (
	caddyfilePath   string
	caddyAPI        string
	appliedState    []models.AppliedRoute
	appliedStateMux sync.RWMutex
)

// InitCaddy initializes the Caddy service with configuration paths.
func InitCaddy(caddyfile, api string) {
	caddyfilePath = caddyfile
	caddyAPI = api
}

// GenerateCaddyfile generates the Caddyfile from enabled routes.
func GenerateCaddyfile() error {
	routes, err := database.GetEnabledRoutes()
	if err != nil {
		log.Printf("Error querying routes: %v", err)
		return err
	}

	var sb strings.Builder
	sb.WriteString("# DevProxy Caddyfile - Auto-generated\n")
	sb.WriteString("{\n")
	sb.WriteString("    admin :2019\n")
	sb.WriteString("    auto_https off\n")
	sb.WriteString("    http_port 80\n")
	sb.WriteString("}\n\n")

	for _, r := range routes {
		sb.WriteString(fmt.Sprintf("http://%s {\n", r.Domain))
		sb.WriteString(fmt.Sprintf("    reverse_proxy %s\n", r.Target))
		sb.WriteString("}\n\n")
	}

	// Ensure directory exists
	dir := filepath.Dir(caddyfilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return err
	}

	if err := os.WriteFile(caddyfilePath, []byte(sb.String()), 0644); err != nil {
		log.Printf("Error writing Caddyfile: %v", err)
		return err
	}

	log.Println("Caddyfile generated successfully")
	return nil
}

// ReloadCaddy regenerates the Caddyfile and reloads Caddy.
// Returns a message and optional warning.
func ReloadCaddy() (message string, warning string) {
	if err := GenerateCaddyfile(); err != nil {
		return "Failed to generate Caddyfile", err.Error()
	}

	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return "Failed to read Caddyfile", err.Error()
	}

	req, err := http.NewRequest("POST", caddyAPI+"/load", strings.NewReader(string(content)))
	if err != nil {
		return "Caddyfile regenerated", err.Error()
	}
	req.Header.Set("Content-Type", "text/caddyfile")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "Caddyfile regenerated, Caddy reload failed", err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "Caddyfile regenerated, Caddy returned error", string(body)
	}

	SaveAppliedState()
	return "Proxy reloaded successfully", ""
}

// SaveAppliedState captures the current route configuration as the applied state.
func SaveAppliedState() {
	routes, err := database.GetAppliedRoutes()
	if err != nil {
		log.Printf("Error querying routes for applied state: %v", err)
		return
	}

	appliedStateMux.Lock()
	appliedState = routes
	appliedStateMux.Unlock()
}

// GetAppliedState returns the last applied route configuration.
func GetAppliedState() []models.AppliedRoute {
	appliedStateMux.RLock()
	defer appliedStateMux.RUnlock()

	if appliedState == nil {
		return []models.AppliedRoute{}
	}
	return appliedState
}
