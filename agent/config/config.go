package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Config holds all agent configuration.
type Config struct {
	APIURL              string `json:"api_url"`
	SyncIntervalSeconds int    `json:"sync_interval_seconds"`
	Autostart           bool   `json:"autostart"`
	RunInBackground     bool   `json:"run_in_background"`
	MaxBackups          int    `json:"max_backups"`
	GUIPort             int    `json:"gui_port"`
	GUIBindAddr         string `json:"gui_bind_addr"`  // Bind address for GUI server (default: "127.0.0.1", use "0.0.0.0" for remote access)
	UpdateChannel       string `json:"update_channel"` // "release" or "pre-release"
}

var (
	current Config
	mu      sync.RWMutex
	cfgPath string
)

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		APIURL:              "http://localhost:8090",
		SyncIntervalSeconds: 5,
		Autostart:           false,
		RunInBackground:     true,
		MaxBackups:          20,
		GUIPort:             9099,
		GUIBindAddr:         "127.0.0.1", // Default to localhost for security
		UpdateChannel:       "release",
	}
}

// DefaultConfigDir returns the platform-specific config directory.
func DefaultConfigDir() string {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appdata, "DevProxy")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devproxy")
}

// Init loads or creates the config file in the given directory.
func Init(configDir string) error {
	if configDir == "" {
		configDir = DefaultConfigDir()
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	cfgPath = filepath.Join(configDir, "config.json")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		current = DefaultConfig()
		return Save()
	}

	return Load()
}

// Load reads config from disk.
func Load() error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	current = cfg
	return nil
}

// Save writes config to disk.
func Save() error {
	mu.RLock()
	data, err := json.MarshalIndent(current, "", "  ")
	mu.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

// Get returns a copy of the current config.
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Update replaces the current config and saves to disk.
func Update(cfg Config) error {
	mu.Lock()
	current = cfg
	mu.Unlock()
	return Save()
}

// ConfigDir returns the directory containing the config file.
func ConfigDir() string {
	return filepath.Dir(cfgPath)
}
