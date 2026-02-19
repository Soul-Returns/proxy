// DevProxy - A reverse proxy manager for Docker Compose projects.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"devproxy/internal/database"
	"devproxy/internal/handlers"
	"devproxy/internal/services"

	"github.com/gin-gonic/gin"
)

//go:embed all:frontend/dist
var staticFiles embed.FS

func main() {
	// Configuration
	dbPath := getEnv("DB_PATH", "/app/data/devproxy.db")
	caddyfilePath := getEnv("CADDYFILE_PATH", "/app/data/Caddyfile")
	caddyAPI := getEnv("CADDY_API", "http://caddy:2019")

	// Initialize database
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Caddy service
	services.InitCaddy(caddyfilePath, caddyAPI)
	services.GenerateCaddyfile()
	services.SaveAppliedState()

	// Start background health checker
	go services.StartHealthChecker()

	// Setup router
	r := gin.Default()
	setupRoutes(r)
	setupStaticFiles(r)

	log.Println("DevProxy API running on :8080")
	r.Run(":8080")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Route management
		api.GET("/routes", handlers.GetRoutes)
		api.GET("/routes/:id", handlers.GetRoute)
		api.POST("/routes", handlers.CreateRoute)
		api.PUT("/routes/:id", handlers.UpdateRoute)
		api.DELETE("/routes/:id", handlers.DeleteRoute)
		api.POST("/routes/:id/toggle", handlers.ToggleRoute)

		// Health & Status
		api.GET("/health", handlers.GetHealthStatus)
		api.GET("/applied-state", handlers.GetAppliedState)

		// Proxy control
		api.POST("/reload", handlers.ReloadCaddy)

		// Config import/export
		api.GET("/export", handlers.ExportConfig)
		api.POST("/import", handlers.ImportConfig)

		// Host Agent
		api.GET("/agent/info", handlers.GetAgentInfo)
		api.GET("/agent/download/:os", handlers.DownloadAgent)
		api.GET("/agent/version", handlers.GetAgentVersion)
		api.POST("/agent/updates/check", handlers.CheckAgentUpdates)

		// Version
		api.GET("/version", handlers.GetBackendVersion)
	}
}

func setupStaticFiles(r *gin.Engine) {
	staticFS, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to load embedded static files: %v", err)
	}

	// Serve assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets"+c.Param("filepath"), http.FS(staticFS))
	})

	// Serve favicon
	r.GET("/favicon.svg", func(c *gin.Context) {
		c.FileFromFS("favicon.svg", http.FS(staticFS))
	})

	// SPA fallback
	serveIndex := func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index.html not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}

	r.GET("/", serveIndex)
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		serveIndex(c)
	})
}
