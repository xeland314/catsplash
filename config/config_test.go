package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
iface = "eth0"
portal_port = 9090
session_timeout = 7200
idle_timeout = 1200
db_path = "test.db"
redirect_url = "http://localhost:9090/portal"
download_speed = "5mbit"
upload_speed = "2mbit"
admin_user = "root"
admin_pass = "secret"
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name(), []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Iface != "eth0" {
		t.Errorf("expected iface eth0, got %s", cfg.Iface)
	}
	if cfg.PortalPort != 9090 {
		t.Errorf("expected portal_port 9090, got %d", cfg.PortalPort)
	}
	if cfg.SessionTimeout != 7200 {
		t.Errorf("expected session_timeout 7200, got %d", cfg.SessionTimeout)
	}
	expectedDBPath := filepath.Join(filepath.Dir(tmpFile.Name()), "test.db")
	if cfg.DBPath != expectedDBPath {
		t.Errorf("expected db_path %s, got %s", expectedDBPath, cfg.DBPath)
	}
}

func TestLoadResolvesRelativeDBPathFromConfigDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	if err := os.WriteFile(cfgPath, []byte(`db_path = "state.db"`), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(cfgPath, []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DBPath != dir+"/state.db" {
		t.Fatalf("expected DB path %s, got %s", dir+"/state.db", cfg.DBPath)
	}
}

func TestLoadDefault(t *testing.T) {
	cfg, err := Load("", []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Iface != "wlan0" {
		t.Errorf("expected default iface wlan0, got %s", cfg.Iface)
	}
	if cfg.PortalPort != 8080 {
		t.Errorf("expected default portal_port 8080, got %d", cfg.PortalPort)
	}
}

func TestLoadOverride(t *testing.T) {
	cfg, err := Load("", []string{"-iface", "eth1", "-port", "7070"})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Iface != "eth1" {
		t.Errorf("expected overriden iface eth1, got %s", cfg.Iface)
	}
	if cfg.PortalPort != 7070 {
		t.Errorf("expected overriden port 7070, got %d", cfg.PortalPort)
	}
}

func TestAdminPassIsHashedAfterLoad(t *testing.T) {
	cfg, err := Load("", []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AdminPass == "" {
		t.Fatal("expected AdminPass to be set")
	}
	if cfg.AdminPass == "catsplash" {
		t.Error("AdminPass must not be stored as plaintext")
	}
	if !strings.HasPrefix(cfg.AdminPass, "$2") {
		t.Errorf("AdminPass should be a bcrypt hash (starts with $2), got: %s", cfg.AdminPass[:20])
	}
}

func TestAdminPassAlreadyHashedIsNotRehashed(t *testing.T) {
	original, err := bcrypt.GenerateFromPassword([]byte("mypassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "config*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
admin_user = "admin"
admin_pass = "` + string(original) + `"
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name(), []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AdminPass != string(original) {
		t.Error("pre-hashed password should not be re-hashed")
	}
}

func TestAdminPassPlaintextIsHashed(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
admin_user = "admin"
admin_pass = "catsplash"
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name(), []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AdminPass == "catsplash" {
		t.Error("plaintext password should have been hashed")
	}

	err = bcrypt.CompareHashAndPassword([]byte(cfg.AdminPass), []byte("catsplash"))
	if err != nil {
		t.Errorf("hashed password should verify against original plaintext: %v", err)
	}
}

func TestEmptyAdminPassIsNotHashed(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
admin_user = "admin"
admin_pass = ""
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name(), []string{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AdminPass != "" {
		t.Error("empty password should remain empty")
	}
}
