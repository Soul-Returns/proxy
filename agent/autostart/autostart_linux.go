//go:build linux

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

const desktopEntry = `[Desktop Entry]
Type=Application
Name=DevProxy Agent
Comment=Automatic hosts file management for DevProxy
Exec=%s
Hidden=false
NoDisplay=true
X-GNOME-Autostart-enabled=true
`

func autostartDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "autostart")
}

func desktopFilePath() string {
	return filepath.Join(autostartDir(), "devproxy-agent.desktop")
}

// Enable adds the agent to XDG autostart.
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	if err := os.MkdirAll(autostartDir(), 0755); err != nil {
		return fmt.Errorf("create autostart dir: %w", err)
	}

	content := fmt.Sprintf(desktopEntry, exe)
	return os.WriteFile(desktopFilePath(), []byte(content), 0644)
}

// Disable removes the agent from XDG autostart.
func Disable() error {
	err := os.Remove(desktopFilePath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsEnabled checks if autostart is currently enabled.
func IsEnabled() bool {
	_, err := os.Stat(desktopFilePath())
	return err == nil
}
