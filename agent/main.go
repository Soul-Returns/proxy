// DevProxy Agent - Automatic hosts file management for DevProxy.
//
// This agent runs on the host machine (outside Docker) and automatically
// syncs DevProxy routes to the system hosts file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"devproxy-agent/config"
	"devproxy-agent/gui"
	agentsync "devproxy-agent/sync"
	"devproxy-agent/tray"
	"devproxy-agent/version"
)

func main() {
	configDir := flag.String("config-dir", "", "Config directory (default: platform-specific)")
	apiURL := flag.String("api-url", "", "DevProxy API URL (overrides config)")
	noTray := flag.Bool("no-tray", false, "Disable system tray icon")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("devproxy-agent %s\n", version.GetVersion())
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("DevProxy Agent %s starting...", version.GetVersion())

	// Initialize config
	if err := config.Init(*configDir); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	log.Printf("Config directory: %s", config.ConfigDir())

	// Override API URL if specified
	if *apiURL != "" {
		cfg := config.Get()
		cfg.APIURL = *apiURL
		config.Update(cfg)
	}

	cfg := config.Get()

	// Start the config GUI web server
	gui.Start(cfg.GUIPort)

	// Start the sync engine
	agentsync.Start()

	// System tray or signal-based shutdown
	quitCh := make(chan struct{})

	if !*noTray && tray.Available() {
		// On Windows, run the system tray (blocks on message loop)
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			tray.Quit()
		}()
		tray.Run(cfg.GUIPort, quitCh)
	} else {
		// On Linux or with --no-tray, wait for signal
		log.Println("Running without system tray. Use Ctrl+C to stop.")
		log.Printf("Config GUI: http://localhost:%d", cfg.GUIPort)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}

	log.Println("Shutting down...")
	agentsync.Stop()
}
