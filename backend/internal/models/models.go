// Package models defines the data structures used throughout DevProxy.
package models

import "time"

// Route represents a proxy route configuration.
type Route struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Target    string    `json:"target"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AppliedRoute represents a route that has been applied to Caddy.
// Used to track configuration changes.
type AppliedRoute struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

// HealthStatus represents the health check result for a route.
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

// ImportRoute is used for importing routes from JSON.
type ImportRoute struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}
