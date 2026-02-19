//go:build !windows

package tray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// Run is a no-op on non-Windows platforms. It blocks until quit is closed.
func Run(port int, quit chan struct{}) {
	<-quit
}

// Quit is a no-op on non-Windows platforms.
func Quit() {}

// Available returns false on non-Windows platforms.
func Available() bool {
	return false
}

// OpenConfigURL opens the config GUI in a browser.
func OpenConfigURL(port int) {
	url := fmt.Sprintf("http://localhost:%d", port)
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	}
}

func init() {
	log.Println("System tray support: not available (non-Windows)")
}
