package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devproxy/internal/version"
	"github.com/gin-gonic/gin"
)

const agentDir = "/app/agent"

// DownloadAgent serves the agent binary for the specified OS.
func DownloadAgent(c *gin.Context) {
	osName := c.Param("os")
	agentVersion := getAgentVersionFromFiles()

	var filename, contentName string
	switch osName {
	case "windows":
		filename = "devproxy-agent.exe"
		contentName = fmt.Sprintf("devproxy-agent-v%s.exe", agentVersion)
	case "linux":
		filename = "devproxy-agent"
		contentName = fmt.Sprintf("devproxy-agent-v%s", agentVersion)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OS. Use 'windows' or 'linux'."})
		return
	}

	filePath := filepath.Join(agentDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent binary not found. Rebuild the container."})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+contentName)
	c.File(filePath)
}

// GetAgentInfo returns information about available agent downloads.
func GetAgentInfo(c *gin.Context) {
	platforms := []gin.H{}
	agentVersion := getAgentVersionFromFiles()

	// Check for Windows agent
	windowsFile := filepath.Join(agentDir, "devproxy-agent.exe")
	if info, err := os.Stat(windowsFile); err == nil {
		platforms = append(platforms, gin.H{
			"os":       "windows",
			"filename": fmt.Sprintf("devproxy-agent-v%s.exe", agentVersion),
			"size":     info.Size(),
			"url":      "/api/agent/download/windows",
			"version":  agentVersion,
		})
	}

	// Check for Linux agent
	linuxFile := filepath.Join(agentDir, "devproxy-agent")
	if info, err := os.Stat(linuxFile); err == nil {
		platforms = append(platforms, gin.H{
			"os":       "linux",
			"filename": fmt.Sprintf("devproxy-agent-v%s", agentVersion),
			"size":     info.Size(),
			"url":      "/api/agent/download/linux",
			"version":  agentVersion,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"available":       len(platforms) > 0,
		"platforms":       platforms,
		"agent_version":   agentVersion,
		"backend_version": version.GetVersion(),
	})
}

// GetAgentVersion queries the running agent for its version
func GetAgentVersion(c *gin.Context) {
	version, err := queryAgent("http://localhost:9099/api/version")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Agent not running or not accessible",
			"details": err.Error(),
		})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(version, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse agent response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckAgentUpdates triggers an update check on the running agent
func CheckAgentUpdates(c *gin.Context) {
	// Parse optional channel parameter
	var req struct {
		Channel string `json:"channel"`
	}
	// Use ShouldBindJSON to avoid errors on empty body
	c.ShouldBindJSON(&req)

	// Create request body for agent
	var body []byte
	if req.Channel != "" {
		body, _ = json.Marshal(map[string]string{"channel": req.Channel})
	}

	updateInfo, err := queryAgentWithBody("http://localhost:9099/api/updates/check", body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Agent not running or not accessible",
			"details": err.Error(),
		})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(updateInfo, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse agent response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// queryAgent makes an HTTP request to the agent's local API
func queryAgent(url string) ([]byte, error) {
	return queryAgentWithBody(url, nil)
}

// queryAgentWithBody makes an HTTP request to the agent's local API with optional body
func queryAgentWithBody(url string, body []byte) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	var req *http.Request
	var err error

	// POST for update check, GET for version
	if url == "http://localhost:9099/api/updates/check" {
		if body != nil {
			req, err = http.NewRequest("POST", url, strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest("POST", url, nil)
		}
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// getAgentVersionFromFiles reads the agent version from VERSION file
func getAgentVersionFromFiles() string {
	data, err := os.ReadFile("/app/VERSION")
	if err != nil {
		return "unknown"
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "AGENT_VERSION=") {
			return strings.TrimPrefix(line, "AGENT_VERSION=")
		}
	}
	return "unknown"
}
