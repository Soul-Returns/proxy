package handlers

import (
	"fmt"
	"net/http"

	"devproxy/internal/database"
	"devproxy/internal/models"
	"devproxy/internal/services"

	"github.com/gin-gonic/gin"
)

// ExportConfig exports all routes as JSON.
func ExportConfig(c *gin.Context) {
	routes, err := database.GetExportRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=devproxy-config.json")
	c.JSON(http.StatusOK, routes)
}

// ImportConfig imports routes from JSON.
func ImportConfig(c *gin.Context) {
	var routes []models.ImportRoute
	if err := c.ShouldBindJSON(&routes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imported, _ := database.ImportRoutes(routes)
	services.GenerateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d routes", imported)})
}
