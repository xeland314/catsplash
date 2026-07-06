package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Catsctl 🐱 - Control CLI for Catsplash\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  catsctl [options] <command> [arguments]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  status                Show system statistics\n")
		fmt.Fprintf(os.Stderr, "  list                  List all clients in the database\n")
		fmt.Fprintf(os.Stderr, "  auth <mac> <ip>       Manually authorize a client\n")
		fmt.Fprintf(os.Stderr, "  kick <mac>            Manually disconnect a client\n")
		fmt.Fprintf(os.Stderr, "  extend <mac> <mins>   Extend a client's session by N minutes\n")
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]

	// Find configuration file
	cfgFile := *configPath
	if cfgFile == "" {
		if _, err := os.Stat("config.toml"); err == nil {
			cfgFile = "config.toml"
		} else if _, err := os.Stat("/opt/catsplash/config.toml"); err == nil {
			cfgFile = "/opt/catsplash/config.toml"
		}
	}

	// Load configuration
	cfg, err := config.Load(cfgFile, nil)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Open database
	db, err := state.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Error opening database (%s): %v", cfg.DBPath, err)
	}
	defer db.Close()

	// Initialize firewall instance (required for auth/kick)
	fw := firewall.New(cfg.Iface, cfg.WanIface, nil)

	switch command {
	case "status":
		showStatus(db, cfg)
	case "list":
		listClients(db, cfg.SessionTimeout, cfg.IdleTimeout)
	case "auth":
		if len(args) < 3 {
			log.Fatalf("Usage: catsctl auth <mac> <ip>")
		}
		authClient(db, fw, args[1], args[2])
	case "kick":
		if len(args) < 2 {
			log.Fatalf("Usage: catsctl kick <mac>")
		}
		kickClient(db, fw, args[1])
	case "extend":
		if len(args) < 3 {
			log.Fatalf("Usage: catsctl extend <mac> <minutes>")
		}
		var minutes int
		if _, err := fmt.Sscanf(args[2], "%d", &minutes); err != nil || minutes <= 0 {
			log.Fatalf("Invalid minutes value: %s", args[2])
		}
		extendClient(db, args[1], minutes)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func showStatus(db *state.DB, cfg *config.Config) {
	clients, err := db.ListAll()
	if err != nil {
		log.Fatalf("Error listing clients: %v", err)
	}

	active := 0
	pending := 0
	for _, c := range clients {
		if c.State == state.StateAuthenticated {
			active++
		} else {
			pending++
		}
	}

	fmt.Printf("Catsplash Status 🐱\n")
	fmt.Printf("===================\n")
	fmt.Printf("Interface AP:      %s\n", cfg.Iface)
	fmt.Printf("Interface WAN:     %s\n", cfg.WanIface)
	fmt.Printf("Portal Port:       %d\n", cfg.PortalPort)
	fmt.Printf("DB Path:           %s\n", cfg.DBPath)
	fmt.Printf("Session Timeout:   %d seconds\n", cfg.SessionTimeout)
	fmt.Printf("Idle Timeout:      %d seconds\n", cfg.IdleTimeout)
	fmt.Printf("Total Clients:     %d\n", len(clients))
	fmt.Printf("Active Sessions:   %d\n", active)
	fmt.Printf("Pending Clients:   %d\n", pending)
}

func listClients(db *state.DB, sessionTimeout, idleTimeout int) {
	clients, err := db.ListAll()
	if err != nil {
		log.Fatalf("Error listing clients: %v", err)
	}

	if len(clients) == 0 {
		fmt.Println("No clients found in the database.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "STATE\tMAC ADDRESS\tIP ADDRESS\tCONNECTED AT\tEXPIRES IN")
	fmt.Fprintln(w, "-----\t-----------\t----------\t------------\t----------")

	now := time.Now().Unix()
	for _, c := range clients {
		connectedStr := "N/A"
		expiresStr := "N/A"

		if c.State == state.StateAuthenticated {
			connectedStr = time.Unix(c.ConnectedAt, 0).Format("2006-01-02 15:04:05")

			// Calculate remaining time
			remSession := int64(sessionTimeout) - (now - c.ConnectedAt)
			remIdle := int64(idleTimeout) - (now - c.LastSeen)

			rem := remSession
			if idleTimeout > 0 && (remIdle < rem || sessionTimeout <= 0) {
				rem = remIdle
			}

			if rem <= 0 {
				expiresStr = "Expired (Reaping)"
			} else {
				dur := time.Duration(rem) * time.Second
				expiresStr = dur.Round(time.Second).String()
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.State, c.MAC, c.IP, connectedStr, expiresStr)
	}
	w.Flush()
}

func authClient(db *state.DB, fw *firewall.Firewall, mac, ip string) {
	// First update state in DB
	if err := db.UpsertClient(mac, ip); err != nil {
		log.Fatalf("Error upserting client: %v", err)
	}
	if err := db.Authenticate(mac, ip); err != nil {
		log.Fatalf("Error authenticating client in database: %v", err)
	}

	// Apply firewall rules
	if err := fw.AllowClient(mac, ip); err != nil {
		log.Fatalf("Error applying firewall rule: %v", err)
	}

	fmt.Printf("Manually authorized client %s (%s) successfully.\n", mac, ip)
}

func kickClient(db *state.DB, fw *firewall.Firewall, mac string) {
	client, err := db.GetClient(mac)
	if err != nil {
		log.Fatalf("Error querying client: %v", err)
	}
	if client == nil {
		log.Fatalf("Client with MAC %s not found in database.", mac)
	}

	// Deauthenticate in DB
	if err := db.Deauthenticate(mac); err != nil {
		log.Fatalf("Error deauthenticating client in database: %v", err)
	}

	// Block in firewall
	if err := fw.BlockClient(mac, client.IP); err != nil {
		log.Fatalf("Error removing firewall rule: %v", err)
	}

	fmt.Printf("Manually kicked client %s (%s) successfully.\n", mac, client.IP)
}

func extendClient(db *state.DB, mac string, minutes int) {
	client, err := db.GetClient(mac)
	if err != nil {
		log.Fatalf("Error querying client: %v", err)
	}
	if client == nil {
		log.Fatalf("Client with MAC %s not found in database.", mac)
	}
	if client.State != state.StateAuthenticated {
		log.Fatalf("Client %s is not authenticated.", mac)
	}

	// Add minutes to connected_at
	newConnAt := client.ConnectedAt + int64(minutes*60)
	_, err = db.Conn.Exec("UPDATE clients SET connected_at = ? WHERE mac = ?", newConnAt, mac)
	if err != nil {
		log.Fatalf("Error updating connection time in database: %v", err)
	}

	fmt.Printf("Extended session of client %s by %d minutes successfully.\n", mac, minutes)
}
