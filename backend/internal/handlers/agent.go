package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const agentDir = "/app/agent"

// DownloadAgent serves the agent binary for the specified OS.
func DownloadAgent(c *gin.Context) {
	osName := c.Param("os")

	var filename, contentName string
	switch osName {
	case "windows":
		filename = "devproxy-agent.exe"
		contentName = "devproxy-agent.exe"
	case "linux":
		filename = "devproxy-agent"
		contentName = "devproxy-agent"
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

	if _, err := os.Stat(filepath.Join(agentDir, "devproxy-agent.exe")); err == nil {
		info, _ := os.Stat(filepath.Join(agentDir, "devproxy-agent.exe"))
		platforms = append(platforms, gin.H{
			"os":       "windows",
			"filename": "devproxy-agent.exe",
			"size":     info.Size(),
			"url":      "/api/agent/download/windows",
		})
	}

	if _, err := os.Stat(filepath.Join(agentDir, "devproxy-agent")); err == nil {
		info, _ := os.Stat(filepath.Join(agentDir, "devproxy-agent"))
		platforms = append(platforms, gin.H{
			"os":       "linux",
			"filename": "devproxy-agent",
			"size":     info.Size(),
			"url":      "/api/agent/download/linux",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"available": len(platforms) > 0,
		"platforms": platforms,
	})
}
