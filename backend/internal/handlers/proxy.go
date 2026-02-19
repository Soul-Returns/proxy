package handlers

import (
	"net/http"

	"devproxy/internal/services"

	"github.com/gin-gonic/gin"
)

// ReloadCaddy regenerates Caddyfile and reloads Caddy.
func ReloadCaddy(c *gin.Context) {
	message, warning := services.ReloadCaddy()

	response := gin.H{"message": message}
	if warning != "" {
		response["warning"] = warning
	}

	c.JSON(http.StatusOK, response)
}

// GetAppliedState returns the last applied route configuration.
func GetAppliedState(c *gin.Context) {
	c.JSON(http.StatusOK, services.GetAppliedState())
}
