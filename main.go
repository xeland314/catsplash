package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/server"
	"github.com/xeland314/catsplash/state"
)

func main() {
	// 1. Load configuration
	cfgPath := "config.toml"
	if _, err := os.Stat("/opt/catsplash/config.toml"); err == nil {
		cfgPath = "/opt/catsplash/config.toml"
	}
	cfg, err := config.Load(cfgPath, os.Args[1:])
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
	fw := firewall.New(cfg.Iface, cfg.WanIface, nil)
	fw.DownloadSpeed = cfg.DownloadSpeed
	fw.UploadSpeed = cfg.UploadSpeed
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
	if u, err := url.Parse(cfg.RedirectURL); err == nil && u.Hostname() != "" {
		portalIP = u.Hostname()
	}
	if err := fw.SetupRedirect(portalIP, cfg.PortalPort); err != nil {
		log.Printf("Warning: failed to setup redirect to %s: %v", portalIP, err) // lopdp:ignore — portalIP is config, not user PII
	}

	// 5. Start Session Reaper
	reaper := state.NewReaper(db, cfg.SessionTimeout, cfg.IdleTimeout, func(mac, ip string) error {
		log.Printf("Expiring session for %s (%s)", state.MaskMAC(mac), state.MaskIP(ip))
		return fw.BlockClient(mac, ip)
	})
	go reaper.Start(10 * time.Second)

	// 5.5 Start Traffic Monitor
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			stats, err := fw.QueryTrafficCounters()
			if err != nil {
				log.Printf("Error querying traffic counters: %v", err)
				continue
			}

			// List authenticated clients to check quotas
			clients, err := db.ListAuthenticated()
			if err != nil {
				log.Printf("Error listing authenticated clients: %v", err)
				continue
			}

			for _, c := range clients {
				stat, found := stats[strings.ToLower(c.MAC)]
				if !found {
					continue
				}

				// Update database
			if err := db.UpdateTraffic(c.MAC, stat.BytesIn, stat.BytesOut); err != nil {
				log.Printf("Error updating traffic for %s: %v", state.MaskMAC(c.MAC), err)
			}

			// Check data limit
			totalBytes := stat.BytesIn + stat.BytesOut
			if c.MaxBytes > 0 && totalBytes >= c.MaxBytes {
				log.Printf("Client %s (%s) exceeded data quota (%d >= %d bytes). Expiring session...", state.MaskMAC(c.MAC), state.MaskIP(c.IP), totalBytes, c.MaxBytes)
					if err := fw.BlockClient(c.MAC, c.IP); err != nil {
						log.Printf("Error blocking client: %v", err)
					}
					db.Deauthenticate(c.MAC)
				}
			}
		}
	}()

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
