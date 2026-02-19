// Package database handles SQLite database operations for DevProxy.
package database

import (
	"database/sql"
	"log"
	"time"

	"devproxy/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database connection pool.
var DB *sql.DB

// Init opens the database connection and creates tables if needed.
func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`
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
		return err
	}

	log.Println("Database initialized")
	return nil
}

// Close closes the database connection.
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// GetAllRoutes retrieves all routes ordered by name.
func GetAllRoutes() ([]models.Route, error) {
	rows, err := DB.Query("SELECT id, name, domain, target, enabled, created_at, updated_at FROM routes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var r models.Route
		var enabled int
		if err := rows.Scan(&r.ID, &r.Name, &r.Domain, &r.Target, &enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Enabled = enabled == 1
		routes = append(routes, r)
	}

	if routes == nil {
		routes = []models.Route{}
	}
	return routes, nil
}

// GetRouteByID retrieves a single route by ID.
func GetRouteByID(id string) (*models.Route, error) {
	var r models.Route
	var enabled int
	err := DB.QueryRow("SELECT id, name, domain, target, enabled, created_at, updated_at FROM routes WHERE id = ?", id).
		Scan(&r.ID, &r.Name, &r.Domain, &r.Target, &enabled, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.Enabled = enabled == 1
	return &r, nil
}

// CreateRoute inserts a new route and returns the created route.
func CreateRoute(r *models.Route) error {
	enabled := 0
	if r.Enabled {
		enabled = 1
	}

	result, err := DB.Exec("INSERT INTO routes (name, domain, target, enabled) VALUES (?, ?, ?, ?)",
		r.Name, r.Domain, r.Target, enabled)
	if err != nil {
		return err
	}

	r.ID, _ = result.LastInsertId()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateRoute updates an existing route.
func UpdateRoute(id string, r *models.Route) error {
	enabled := 0
	if r.Enabled {
		enabled = 1
	}

	_, err := DB.Exec("UPDATE routes SET name = ?, domain = ?, target = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		r.Name, r.Domain, r.Target, enabled, id)
	return err
}

// DeleteRoute removes a route by ID.
func DeleteRoute(id string) error {
	_, err := DB.Exec("DELETE FROM routes WHERE id = ?", id)
	return err
}

// ToggleRoute toggles the enabled state of a route.
func ToggleRoute(id string) error {
	_, err := DB.Exec("UPDATE routes SET enabled = NOT enabled, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// GetEnabledRoutes retrieves all enabled routes.
func GetEnabledRoutes() ([]models.Route, error) {
	rows, err := DB.Query("SELECT id, domain, target FROM routes WHERE enabled = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var r models.Route
		if err := rows.Scan(&r.ID, &r.Domain, &r.Target); err != nil {
			continue
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// GetAppliedRoutes retrieves all routes for the applied state comparison.
func GetAppliedRoutes() ([]models.AppliedRoute, error) {
	rows, err := DB.Query("SELECT id, name, domain, target, enabled FROM routes ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.AppliedRoute
	for rows.Next() {
		var r models.AppliedRoute
		var enabled int
		if err := rows.Scan(&r.ID, &r.Name, &r.Domain, &r.Target, &enabled); err != nil {
			continue
		}
		r.Enabled = enabled == 1
		routes = append(routes, r)
	}
	return routes, nil
}

// GetExportRoutes retrieves routes for export.
func GetExportRoutes() ([]map[string]interface{}, error) {
	rows, err := DB.Query("SELECT name, domain, target, enabled FROM routes ORDER BY name")
	if err != nil {
		return nil, err
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
	return routes, nil
}

// ImportRoutes imports routes from a list, using INSERT OR REPLACE.
func ImportRoutes(routes []models.ImportRoute) (int, error) {
	imported := 0
	for _, r := range routes {
		enabled := 0
		if r.Enabled {
			enabled = 1
		}
		_, err := DB.Exec("INSERT OR REPLACE INTO routes (name, domain, target, enabled) VALUES (?, ?, ?, ?)",
			r.Name, r.Domain, r.Target, enabled)
		if err == nil {
			imported++
		}
	}
	return imported, nil
}
