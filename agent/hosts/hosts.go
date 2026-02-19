package hosts

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MarkerStart = "# DevProxy managed entries - START"
	MarkerEnd   = "# DevProxy managed entries - END"
)

var mu sync.Mutex

// GetHostsFilePath returns the platform-specific hosts file path.
func GetHostsFilePath() string {
	if runtime.GOOS == "windows" {
		sysRoot := os.Getenv("SystemRoot")
		if sysRoot == "" {
			sysRoot = `C:\Windows`
		}
		return filepath.Join(sysRoot, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// GetManagedEntries reads the currently managed entries from the hosts file.
func GetManagedEntries() ([]string, error) {
	mu.Lock()
	defer mu.Unlock()

	lines, err := readLines()
	if err != nil {
		return nil, err
	}

	var managed []string
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == MarkerStart {
			inSection = true
			continue
		}
		if trimmed == MarkerEnd {
			break
		}
		if inSection && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			managed = append(managed, trimmed)
		}
	}
	return managed, nil
}

// UpdateEntries replaces the managed section with the given entries.
// Each entry should be in "IP domain" format (e.g., "127.0.0.1 myapp.test").
// A backup is created before writing.
func UpdateEntries(entries []string, backupDir string) error {
	mu.Lock()
	defer mu.Unlock()

	lines, err := readLines()
	if err != nil {
		return fmt.Errorf("read hosts file: %w", err)
	}

	// Create backup before modifying
	if backupDir != "" {
		if err := createBackup(backupDir, lines); err != nil {
			return fmt.Errorf("create backup: %w", err)
		}
	}

	// Remove existing managed section
	var newLines []string
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == MarkerStart {
			inSection = true
			continue
		}
		if trimmed == MarkerEnd {
			inSection = false
			continue
		}
		if !inSection {
			newLines = append(newLines, line)
		}
	}

	// Remove trailing blank lines
	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}

	// Add managed section if there are entries
	if len(entries) > 0 {
		newLines = append(newLines, "")
		newLines = append(newLines, MarkerStart)
		for _, entry := range entries {
			if strings.TrimSpace(entry) != "" {
				newLines = append(newLines, entry)
			}
		}
		newLines = append(newLines, MarkerEnd)
	}

	return writeLines(newLines)
}

// CheckPermissions verifies write access to the hosts file.
func CheckPermissions() error {
	hostsPath := GetHostsFilePath()
	f, err := os.OpenFile(hostsPath, os.O_RDWR, 0644)
	if err != nil {
		if os.IsPermission(err) {
			if runtime.GOOS == "windows" {
				return fmt.Errorf("permission denied: run the agent as Administrator")
			}
			return fmt.Errorf("permission denied: run the agent with sudo")
		}
		return err
	}
	f.Close()
	return nil
}

// ListBackups returns available backup files sorted newest first.
func ListBackups(backupDir string) ([]BackupInfo, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "hosts_") || !strings.HasSuffix(entry.Name(), ".bak") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(backupDir, entry.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})
	return backups, nil
}

// RestoreFromBackup restores the hosts file from a backup.
func RestoreFromBackup(backupPath, backupDir string) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	// Backup current state before restoring
	currentLines, readErr := readLines()
	if readErr == nil && backupDir != "" {
		createBackup(backupDir, currentLines)
	}

	hostsPath := GetHostsFilePath()
	return os.WriteFile(hostsPath, data, 0644)
}

// PruneBackups keeps only the most recent N backups.
func PruneBackups(backupDir string, keep int) error {
	backups, err := ListBackups(backupDir)
	if err != nil || len(backups) <= keep {
		return err
	}

	for _, b := range backups[keep:] {
		os.Remove(b.Path)
	}
	return nil
}

// BackupInfo represents a hosts file backup.
type BackupInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

func readLines() ([]string, error) {
	f, err := os.Open(GetHostsFilePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string) error {
	hostsPath := GetHostsFilePath()
	content := strings.Join(lines, "\n") + "\n"

	// Write to temp file first
	tmpPath := hostsPath + ".devproxy.tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, hostsPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func createBackup(backupDir string, lines []string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("hosts_%s.bak", timestamp))
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(backupPath, []byte(content), 0644)
}
