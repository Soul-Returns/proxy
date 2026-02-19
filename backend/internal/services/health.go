package services

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"devproxy/internal/database"
	"devproxy/internal/models"
)

var (
	healthCache    = make(map[int64]*models.HealthStatus)
	healthCacheMux sync.RWMutex
)

// StartHealthChecker starts the background health check loop.
func StartHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	checkHealth()
	for range ticker.C {
		checkHealth()
	}
}

// GetHealthStatuses returns all cached health statuses.
func GetHealthStatuses() []*models.HealthStatus {
	healthCacheMux.RLock()
	defer healthCacheMux.RUnlock()

	statuses := make([]*models.HealthStatus, 0, len(healthCache))
	for _, status := range healthCache {
		statuses = append(statuses, status)
	}
	return statuses
}

func checkHealth() {
	routes, err := database.GetEnabledRoutes()
	if err != nil {
		return
	}

	for _, route := range routes {
		status := checkRouteHealth(route)
		healthCacheMux.Lock()
		healthCache[route.ID] = status
		healthCacheMux.Unlock()
	}
}

func checkRouteHealth(route models.Route) *models.HealthStatus {
	status := &models.HealthStatus{
		RouteID:   route.ID,
		Domain:    route.Domain,
		Target:    route.Target,
		LastCheck: time.Now().Format(time.RFC3339),
	}

	// Prepare target URL
	targetURL := route.Target
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "http://" + targetURL
	}

	// Extract hostname
	hostname := route.Target
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
		return status
	}

	status.DNSResolved = true
	if len(addrs) > 0 {
		status.ResolvedIP = addrs[0]
	}

	// HTTP connection check
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Get(targetURL)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Healthy = false
		status.Error = err.Error()
		categorizeError(status, err.Error())
		return status
	}
	defer resp.Body.Close()

	status.StatusCode = resp.StatusCode

	if resp.StatusCode >= 500 {
		status.Healthy = false
		status.ErrorType = "server_error"
		status.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		status.Tip = "Server returned an error. Check application logs inside the container."
	} else {
		// 2xx, 3xx, and 4xx are considered healthy (4xx often expected for health checks)
		status.Healthy = true
	}

	return status
}

func categorizeError(status *models.HealthStatus, errStr string) {
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
}
