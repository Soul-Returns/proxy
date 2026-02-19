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
)

//go:embed static/index.html
var staticFiles embed.FS

// Start launches the config GUI web server.
func Start(port int) {
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

	// API endpoints
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/sync", handleSync)
	mux.HandleFunc("/api/pause", handlePause)
	mux.HandleFunc("/api/entries", handleEntries)
	mux.HandleFunc("/api/backups", handleBackups)
	mux.HandleFunc("/api/restore", handleRestore)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("Agent config GUI available at http://%s", addr)

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

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
