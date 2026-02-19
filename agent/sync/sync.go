package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"devproxy-agent/config"
	"devproxy-agent/hosts"
)

// Route matches the DevProxy API route structure.
type Route struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

// Status represents the current sync state.
type Status struct {
	Connected     bool      `json:"connected"`
	LastSync      time.Time `json:"last_sync"`
	LastError     string    `json:"last_error,omitempty"`
	RouteCount    int       `json:"route_count"`
	Paused        bool      `json:"paused"`
	HasPermission bool      `json:"has_permission"`
}

var (
	status      Status
	statusMu    sync.RWMutex
	stopCh      chan struct{}
	pauseMu     sync.Mutex
	syncNowCh   chan struct{}
	client      = &http.Client{Timeout: 5 * time.Second}
	lastEntries string // hash of last written entries to avoid redundant writes
)

// GetStatus returns the current sync status.
func GetStatus() Status {
	statusMu.RLock()
	defer statusMu.RUnlock()
	return status
}

// Start begins the sync loop.
func Start() {
	stopCh = make(chan struct{})
	syncNowCh = make(chan struct{}, 1)

	// Check permissions on start
	if err := hosts.CheckPermissions(); err != nil {
		statusMu.Lock()
		status.HasPermission = false
		status.LastError = err.Error()
		statusMu.Unlock()
		log.Printf("Warning: %s", err)
	} else {
		statusMu.Lock()
		status.HasPermission = true
		statusMu.Unlock()
	}

	go syncLoop()
}

// Stop stops the sync loop.
func Stop() {
	if stopCh != nil {
		close(stopCh)
	}
}

// Pause pauses the sync loop.
func Pause() {
	statusMu.Lock()
	status.Paused = true
	statusMu.Unlock()
}

// Resume resumes the sync loop.
func Resume() {
	statusMu.Lock()
	status.Paused = false
	statusMu.Unlock()
}

// TogglePause toggles the pause state.
func TogglePause() {
	statusMu.Lock()
	status.Paused = !status.Paused
	statusMu.Unlock()
}

// SyncNow triggers an immediate sync.
func SyncNow() {
	select {
	case syncNowCh <- struct{}{}:
	default:
	}
}

func syncLoop() {
	// Initial sync
	doSync()

	for {
		cfg := config.Get()
		interval := time.Duration(cfg.SyncIntervalSeconds) * time.Second
		if interval < time.Second {
			interval = time.Second
		}

		select {
		case <-stopCh:
			return
		case <-syncNowCh:
			doSync()
		case <-time.After(interval):
			doSync()
		}
	}
}

func doSync() {
	statusMu.RLock()
	paused := status.Paused
	statusMu.RUnlock()

	if paused {
		return
	}

	cfg := config.Get()
	routes, err := fetchRoutes(cfg.APIURL)
	if err != nil {
		statusMu.Lock()
		status.Connected = false
		status.LastError = err.Error()
		statusMu.Unlock()
		return
	}

	// Build desired entries from enabled routes
	var entries []string
	for _, r := range routes {
		if r.Enabled && r.Domain != "" {
			entries = append(entries, fmt.Sprintf("127.0.0.1 %s", r.Domain))
		}
	}
	sort.Strings(entries)

	// Check if anything changed
	entriesKey := strings.Join(entries, "\n")
	if entriesKey == lastEntries {
		statusMu.Lock()
		status.Connected = true
		status.LastSync = time.Now()
		status.LastError = ""
		status.RouteCount = len(entries)
		statusMu.Unlock()
		return
	}

	// Write to hosts file
	backupDir := fmt.Sprintf("%s/backups", config.ConfigDir())
	if err := hosts.UpdateEntries(entries, backupDir); err != nil {
		statusMu.Lock()
		status.Connected = true
		status.LastError = fmt.Sprintf("hosts update failed: %s", err)
		status.HasPermission = false
		statusMu.Unlock()
		log.Printf("Failed to update hosts file: %s", err)
		return
	}

	// Prune old backups
	hosts.PruneBackups(backupDir, cfg.MaxBackups)

	lastEntries = entriesKey
	statusMu.Lock()
	status.Connected = true
	status.LastSync = time.Now()
	status.LastError = ""
	status.RouteCount = len(entries)
	status.HasPermission = true
	statusMu.Unlock()

	log.Printf("Synced %d entries to hosts file", len(entries))
}

func fetchRoutes(apiURL string) ([]Route, error) {
	resp, err := client.Get(apiURL + "/api/routes")
	if err != nil {
		return nil, fmt.Errorf("connect to DevProxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DevProxy returned status %d", resp.StatusCode)
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return routes, nil
}
