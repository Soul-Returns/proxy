package handlers

import (
	"net/http"

	"devproxy/internal/version"
	"github.com/gin-gonic/gin"
)

// GetBackendVersion returns the current backend version
func GetBackendVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version.GetVersion(),
	})
}
