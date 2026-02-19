// Package handlers provides HTTP handlers for the API endpoints.
package handlers

import (
	"database/sql"
	"net/http"

	"devproxy/internal/database"
	"devproxy/internal/models"
	"devproxy/internal/services"

	"github.com/gin-gonic/gin"
)

// GetRoutes returns all routes.
func GetRoutes(c *gin.Context) {
	routes, err := database.GetAllRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routes)
}

// GetRoute returns a single route by ID.
func GetRoute(c *gin.Context) {
	route, err := database.GetRouteByID(c.Param("id"))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, route)
}

// CreateRoute creates a new route.
func CreateRoute(c *gin.Context) {
	var r models.Route
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.CreateRoute(&r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	services.GenerateCaddyfile()
	c.JSON(http.StatusCreated, r)
}

// UpdateRoute updates an existing route.
func UpdateRoute(c *gin.Context) {
	var r models.Route
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.UpdateRoute(c.Param("id"), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	services.GenerateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route updated"})
}

// DeleteRoute deletes a route.
func DeleteRoute(c *gin.Context) {
	if err := database.DeleteRoute(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	services.GenerateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}

// ToggleRoute toggles the enabled state of a route.
func ToggleRoute(c *gin.Context) {
	if err := database.ToggleRoute(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	services.GenerateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route toggled"})
}
