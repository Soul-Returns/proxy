package handlers

import (
	"net/http"

	"devproxy/internal/services"

	"github.com/gin-gonic/gin"
)

// GetHealthStatus returns all health statuses.
func GetHealthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, services.GetHealthStatuses())
}
