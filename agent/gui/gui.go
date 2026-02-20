package gui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"devproxy-agent/autostart"
	"devproxy-agent/config"
	"devproxy-agent/hosts"
	agentsync "devproxy-agent/sync"
	"devproxy-agent/version"
)

//go:embed static/index.html
var staticFiles embed.FS

// Start launches the config GUI web server.
func Start(port int, bindAddr string) {
	if bindAddr == "" {
		bindAddr = "127.0.0.1" // Default to localhost
	}
	mux := http.NewServeMux()

	// Serve the embedded HTML UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API endpoints (with CORS for DevProxy WebUI)
	mux.HandleFunc("/api/status", cors(handleStatus))
	mux.HandleFunc("/api/config", cors(handleConfig))
	mux.HandleFunc("/api/sync", cors(handleSync))
	mux.HandleFunc("/api/pause", cors(handlePause))
	mux.HandleFunc("/api/entries", cors(handleEntries))
	mux.HandleFunc("/api/backups", cors(handleBackups))
	mux.HandleFunc("/api/restore", cors(handleRestore))
	mux.HandleFunc("/api/version", cors(handleVersion))
	mux.HandleFunc("/api/updates/check", cors(handleUpdateCheck))

	addr := fmt.Sprintf("%s:%d", bindAddr, port)
	log.Printf("Agent config GUI available at http://%s", addr)
	if bindAddr == "0.0.0.0" {
		log.Printf("WARNING: GUI listening on all interfaces (0.0.0.0) - accessible from network")
	}

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("GUI server error: %s", err)
		}
	}()
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, agentsync.GetStatus())
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, config.Get())
	case http.MethodPut:
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Handle autostart toggle
		oldCfg := config.Get()
		if cfg.Autostart != oldCfg.Autostart {
			if cfg.Autostart {
				if err := autostart.Enable(); err != nil {
					writeError(w, http.StatusInternalServerError, fmt.Sprintf("enable autostart: %s", err))
					return
				}
			} else {
				if err := autostart.Disable(); err != nil {
					writeError(w, http.StatusInternalServerError, fmt.Sprintf("disable autostart: %s", err))
					return
				}
			}
		}

		if err := config.Update(cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, map[string]string{"message": "saved"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	agentsync.SyncNow()
	writeJSON(w, map[string]string{"message": "sync triggered"})
}

func handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	agentsync.TogglePause()
	writeJSON(w, map[string]string{"message": "toggled"})
}

func handleEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := hosts.GetManagedEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entries == nil {
		entries = []string{}
	}
	writeJSON(w, map[string]interface{}{
		"entries": entries,
		"path":    hosts.GetHostsFilePath(),
	})
}

func handleBackups(w http.ResponseWriter, r *http.Request) {
	backupDir := filepath.Join(config.ConfigDir(), "backups")
	backups, err := hosts.ListBackups(backupDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if backups == nil {
		backups = []hosts.BackupInfo{}
	}
	writeJSON(w, map[string]interface{}{
		"backups":    backups,
		"backup_dir": backupDir,
	})
}

func handleRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	backupDir := filepath.Join(config.ConfigDir(), "backups")
	if err := hosts.RestoreFromBackup(req.Path, backupDir); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, map[string]string{"message": "restored"})
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"version": version.GetVersion(),
	})
}

func handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body for channel override
	var req struct {
		Channel string `json:"channel"`
	}

	// Try to decode body, but don't fail if empty
	json.NewDecoder(r.Body).Decode(&req)

	// Determine channel
	var channel version.UpdateChannel
	if req.Channel != "" {
		channel = version.UpdateChannel(req.Channel)
	} else {
		cfg := config.Get()
		channel = version.UpdateChannel(cfg.UpdateChannel)
	}

	// Validate channel
	if channel != version.ChannelRelease && channel != version.ChannelPreRelease {
		channel = version.ChannelRelease
	}

	updateInfo, err := version.CheckForUpdates(channel)
	if err != nil {
		// Still return the info even on error (it contains error details)
		writeJSON(w, updateInfo)
		return
	}

	writeJSON(w, updateInfo)
}
