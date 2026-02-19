//go:build windows

package autostart

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const regKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const regValue = "DevProxyAgent"

// Enable adds the agent to Windows autostart via Registry.
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	return key.SetStringValue(regValue, fmt.Sprintf(`"%s"`, exe))
}

// Disable removes the agent from Windows autostart.
func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	err = key.DeleteValue(regValue)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsEnabled checks if autostart is currently enabled.
func IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(regValue)
	return err == nil
}
