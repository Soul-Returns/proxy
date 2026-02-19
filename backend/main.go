package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed all:frontend/dist
var staticFiles embed.FS

type Route struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Target    string    `json:"target"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HealthStatus struct {
	RouteID      int64  `json:"route_id"`
	Domain       string `json:"domain"`
	Target       string `json:"target"`
	Healthy      bool   `json:"healthy"`
	LastCheck    string `json:"last_check"`
	Error        string `json:"error,omitempty"`
	ErrorType    string `json:"error_type,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseTime int64  `json:"response_time_ms,omitempty"`
	DNSResolved  bool   `json:"dns_resolved"`
	ResolvedIP   string `json:"resolved_ip,omitempty"`
	Tip          string `json:"tip,omitempty"`
}

var (
	db             *sql.DB
	dbPath         string
	caddyfilePath  string
	caddyAPI       string
	healthCache    = make(map[int64]*HealthStatus)
	healthCacheMux sync.RWMutex
)

func main() {
	dbPath = getEnv("DB_PATH", "/app/data/devproxy.db")
	caddyfilePath = getEnv("CADDYFILE_PATH", "/app/data/Caddyfile")
	caddyAPI = getEnv("CADDY_API", "http://caddy:2019")

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	initDB()
	generateCaddyfile()

	// Start health checker
	go healthChecker()

	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		api.GET("/routes", getRoutes)
		api.GET("/routes/:id", getRoute)
		api.POST("/routes", createRoute)
		api.PUT("/routes/:id", updateRoute)
		api.DELETE("/routes/:id", deleteRoute)
		api.POST("/routes/:id/toggle", toggleRoute)
		api.GET("/health", getHealthStatus)
		api.POST("/reload", reloadCaddy)
		api.GET("/export", exportConfig)
		api.POST("/import", importConfig)
	}

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to load embedded static files: %v", err)
	}

	// Serve assets directory
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets"+c.Param("filepath"), http.FS(staticFS))
	})

	// Serve favicon
	r.GET("/favicon.svg", func(c *gin.Context) {
		c.FileFromFS("favicon.svg", http.FS(staticFS))
	})

	// Serve index.html for root and SPA fallback
	serveIndex := func(c *gin.Context) {
		indexFile, err := staticFS.Open("index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index.html not found")
			return
		}
		defer indexFile.Close()
		data, _ := fs.ReadFile(staticFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
	r.GET("/", serveIndex)
	r.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		serveIndex(c)
	})

	log.Println("DevProxy API running on :8080")
	r.Run(":8080")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func initDB() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS routes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			domain TEXT NOT NULL UNIQUE,
			target TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func getRoutes(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, domain, target, enabled, created_at, updated_at FROM routes ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var r Route
		var enabled int
		err := rows.Scan(&r.ID, &r.Name, &r.Domain, &r.Target, &enabled, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		r.Enabled = enabled == 1
		routes = append(routes, r)
	}

	if routes == nil {
		routes = []Route{}
	}
	c.JSON(http.StatusOK, routes)
}

func getRoute(c *gin.Context) {
	id := c.Param("id")
	var r Route
	var enabled int
	err := db.QueryRow("SELECT id, name, domain, target, enabled, created_at, updated_at FROM routes WHERE id = ?", id).
		Scan(&r.ID, &r.Name, &r.Domain, &r.Target, &enabled, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	r.Enabled = enabled == 1
	c.JSON(http.StatusOK, r)
}

func createRoute(c *gin.Context) {
	var r Route
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := 0
	if r.Enabled {
		enabled = 1
	}

	result, err := db.Exec("INSERT INTO routes (name, domain, target, enabled) VALUES (?, ?, ?, ?)",
		r.Name, r.Domain, r.Target, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r.ID, _ = result.LastInsertId()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	generateCaddyfile()
	c.JSON(http.StatusCreated, r)
}

func updateRoute(c *gin.Context) {
	id := c.Param("id")
	var r Route
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := 0
	if r.Enabled {
		enabled = 1
	}

	_, err := db.Exec("UPDATE routes SET name = ?, domain = ?, target = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		r.Name, r.Domain, r.Target, enabled, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	generateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route updated"})
}

func deleteRoute(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM routes WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	generateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}

func toggleRoute(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("UPDATE routes SET enabled = NOT enabled, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	generateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": "Route toggled"})
}

func generateCaddyfile() {
	rows, err := db.Query("SELECT id, domain, target FROM routes WHERE enabled = 1")
	if err != nil {
		log.Printf("Error querying routes: %v", err)
		return
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("# DevProxy Caddyfile - Auto-generated\n")
	sb.WriteString("{\n")
	sb.WriteString("    admin :2019\n")
	sb.WriteString("    auto_https off\n")
	sb.WriteString("    http_port 80\n")
	sb.WriteString("}\n\n")

	for rows.Next() {
		var id int64
		var domain, target string
		if err := rows.Scan(&id, &domain, &target); err != nil {
			continue
		}
		// Use http:// prefix since auto_https is off
		sb.WriteString(fmt.Sprintf("http://%s {\n", domain))
		sb.WriteString(fmt.Sprintf("    reverse_proxy %s\n", target))
		sb.WriteString("}\n\n")
	}

	err = os.WriteFile(caddyfilePath, []byte(sb.String()), 0644)
	if err != nil {
		log.Printf("Error writing Caddyfile: %v", err)
	}
}

func reloadCaddy(c *gin.Context) {
	generateCaddyfile()

	// Read the generated Caddyfile
	caddyfileContent, err := os.ReadFile(caddyfilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read Caddyfile", "error": err.Error()})
		return
	}

	// Send to Caddy's /load endpoint with Caddyfile adapter
	req, err := http.NewRequest("POST", caddyAPI+"/load", strings.NewReader(string(caddyfileContent)))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Caddyfile regenerated", "warning": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "text/caddyfile")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Caddyfile regenerated, Caddy reload failed", "warning": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusOK, gin.H{"message": "Caddyfile regenerated, Caddy returned error", "warning": string(body)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy reloaded successfully"})
}

func getHealthStatus(c *gin.Context) {
	healthCacheMux.RLock()
	defer healthCacheMux.RUnlock()

	statuses := make([]*HealthStatus, 0, len(healthCache))
	for _, status := range healthCache {
		statuses = append(statuses, status)
	}

	c.JSON(http.StatusOK, statuses)
}

func healthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	checkHealth()
	for range ticker.C {
		checkHealth()
	}
}

func checkHealth() {
	rows, err := db.Query("SELECT id, domain, target FROM routes WHERE enabled = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var domain, target string
		if err := rows.Scan(&id, &domain, &target); err != nil {
			continue
		}

		status := &HealthStatus{
			RouteID:   id,
			Domain:    domain,
			Target:    target,
			LastCheck: time.Now().Format(time.RFC3339),
		}

		// Parse target to get hostname for DNS check
		targetURL := target
		if !strings.HasPrefix(targetURL, "http") {
			targetURL = "http://" + targetURL
		}

		// Extract hostname from target
		hostname := target
		if colonIdx := strings.Index(hostname, ":"); colonIdx > 0 {
			hostname = hostname[:colonIdx]
		}

		// DNS resolution check
		addrs, dnsErr := net.LookupHost(hostname)
		if dnsErr != nil {
			status.DNSResolved = false
			status.Healthy = false
			status.ErrorType = "dns_failure"
			status.Error = fmt.Sprintf("DNS lookup failed for '%s': %v", hostname, dnsErr)
			status.Tip = "Container may not be running or not connected to dev-proxy network. Check: docker network inspect dev-proxy"
			healthCacheMux.Lock()
			healthCache[id] = status
			healthCacheMux.Unlock()
			continue
		}
		status.DNSResolved = true
		if len(addrs) > 0 {
			status.ResolvedIP = addrs[0]
		}

		// HTTP connection check
		client := &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		}

		start := time.Now()
		resp, err := client.Get(targetURL)
		status.ResponseTime = time.Since(start).Milliseconds()

		if err != nil {
			status.Healthy = false
			status.Error = err.Error()

			// Categorize the error
			errStr := err.Error()
			switch {
			case strings.Contains(errStr, "connection refused"):
				status.ErrorType = "connection_refused"
				status.Tip = "Container is reachable but not accepting connections on this port. Check if the web server is running inside the container."
			case strings.Contains(errStr, "timeout"):
				status.ErrorType = "timeout"
				status.Tip = "Connection timed out. Container may be overloaded or the port might be blocked."
			case strings.Contains(errStr, "no such host"):
				status.ErrorType = "dns_failure"
				status.Tip = "Container name not found. Ensure container is on the dev-proxy network."
			case strings.Contains(errStr, "connection reset"):
				status.ErrorType = "connection_reset"
				status.Tip = "Connection was reset by the server. The application may have crashed or rejected the connection."
			default:
				status.ErrorType = "connection_error"
				status.Tip = "Unable to connect. Verify the target container and port are correct."
			}
		} else {
			defer resp.Body.Close()
			status.StatusCode = resp.StatusCode

			if resp.StatusCode >= 500 {
				status.Healthy = false
				status.ErrorType = "server_error"
				status.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
				status.Tip = "Server returned an error. Check application logs inside the container."
			} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				// 4xx is often OK for health checks (e.g., 404 on root, 401 auth required)
				status.Healthy = true
				status.StatusCode = resp.StatusCode
			} else {
				status.Healthy = true
			}
		}

		healthCacheMux.Lock()
		healthCache[id] = status
		healthCacheMux.Unlock()
	}
}

func exportConfig(c *gin.Context) {
	rows, err := db.Query("SELECT name, domain, target, enabled FROM routes ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var routes []map[string]interface{}
	for rows.Next() {
		var name, domain, target string
		var enabled int
		if err := rows.Scan(&name, &domain, &target, &enabled); err != nil {
			continue
		}
		routes = append(routes, map[string]interface{}{
			"name":    name,
			"domain":  domain,
			"target":  target,
			"enabled": enabled == 1,
		})
	}

	c.Header("Content-Disposition", "attachment; filename=devproxy-config.json")
	c.JSON(http.StatusOK, routes)
}

func importConfig(c *gin.Context) {
	var routes []struct {
		Name    string `json:"name"`
		Domain  string `json:"domain"`
		Target  string `json:"target"`
		Enabled bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&routes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imported := 0
	for _, r := range routes {
		enabled := 0
		if r.Enabled {
			enabled = 1
		}
		_, err := db.Exec("INSERT OR REPLACE INTO routes (name, domain, target, enabled) VALUES (?, ?, ?, ?)",
			r.Name, r.Domain, r.Target, enabled)
		if err == nil {
			imported++
		}
	}

	generateCaddyfile()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d routes", imported)})
}

// Unused but needed for compilation
var _ = strconv.Itoa
var _ = json.Marshal
