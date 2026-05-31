package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/server"
	"github.com/xeland314/catsplash/state"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load("config.toml", os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Open database
	db, err := state.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Initialize firewall
	fw := firewall.New(cfg.Iface, nil)
	if err := fw.Init(); err != nil {
		log.Fatalf("Failed to initialize firewall: %v", err)
	}
	// Ensure cleanup on exit
	defer func() {
		log.Println("Cleaning up firewall rules...")
		fw.Teardown()
	}()

	// 4. Setup redirect (DNAT)
	portalIP := "192.168.10.1" 
	if err := fw.SetupRedirect(portalIP, cfg.PortalPort); err != nil {
		log.Printf("Warning: failed to setup redirect: %v", err)
	}

	// 5. Start Session Reaper
	reaper := state.NewReaper(db, cfg.SessionTimeout, cfg.IdleTimeout, func(mac, ip string) error {
		log.Printf("Expiring session for %s (%s)", mac, ip)
		return fw.BlockClient(mac, ip)
	})
	go reaper.Start(10 * time.Second)

	// 6. Start Web Server
	srv := server.New(cfg, db, fw)
	go func() {
		log.Printf("Starting portal server on :%d", cfg.PortalPort)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 7. Handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Catsplash is running. Press Ctrl+C to stop.")
	<-sigChan
	log.Println("Shutting down...")
}
