package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the operator's configuration for the captive portal.
type Config struct {
	Iface          string `toml:"iface"`
	PortalPort     int    `toml:"portal_port"`
	SessionTimeout int    `toml:"session_timeout"`
	IdleTimeout    int    `toml:"idle_timeout"`
	DBPath         string `toml:"db_path"`
	RedirectURL    string `toml:"redirect_url"`
	WanIface       string `toml:"wan_iface"`

	DownloadSpeed string `toml:"download_speed"`
	UploadSpeed   string `toml:"upload_speed"`

	AdminUser string `toml:"admin_user"`
	AdminPass string `toml:"admin_pass"`
}

// Load reads the configuration from a TOML file and overrides it with CLI flags.
func Load(path string, args []string) (*Config, error) {
	cfg := &Config{
		Iface:          "wlan0",
		PortalPort:     8080,
		SessionTimeout: 3600,
		IdleTimeout:    600,
		DBPath:         "captive.db",
		RedirectURL:    "http://192.168.1.1:8080/portal",
		WanIface:       "eth0",
		DownloadSpeed:  "0",
		UploadSpeed:    "0",
		AdminUser:      "admin",
		AdminPass:      "catsplash",
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, fmt.Errorf("failed to decode config file: %w", err)
			}
			if !filepath.IsAbs(cfg.DBPath) {
				cfg.DBPath = filepath.Join(filepath.Dir(path), cfg.DBPath)
			}
		}
	}

	// CLI flags override config file
	fs := flag.NewFlagSet("catsplash", flag.ContinueOnError)
	fs.StringVar(&cfg.Iface, "iface", cfg.Iface, "Network interface to monitor")
	fs.StringVar(&cfg.WanIface, "wan", cfg.WanIface, "External/WAN network interface for internet access")
	fs.IntVar(&cfg.PortalPort, "port", cfg.PortalPort, "Port for the portal server")
	fs.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Path to the SQLite database")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return cfg, nil
}
