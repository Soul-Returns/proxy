package handlers

import (
	"net/http"

	"devproxy/internal/github"
	"devproxy/internal/version"
	"github.com/gin-gonic/gin"
)

// GetBackendVersion returns the current backend version
func GetBackendVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.GetVersion(),
	})
}

// CheckBackendUpdates checks for backend updates directly from GitHub
func CheckBackendUpdates(c *gin.Context) {
	var req struct {
		Channel string `json:"channel"`
	}
	c.ShouldBindJSON(&req)

	// Default to release channel
	if req.Channel == "" {
		req.Channel = "release"
	}

	currentVersion := version.GetVersion()
	updateInfo, err := github.CheckForUpdates(currentVersion, req.Channel)
	if err != nil {
		// Still return the info (contains error details)
		c.JSON(http.StatusOK, updateInfo)
		return
	}

	c.JSON(http.StatusOK, updateInfo)
}
