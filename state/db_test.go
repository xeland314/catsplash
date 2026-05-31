package state

import (
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	dbPath := "test_captive.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Check if table exists
	var name string
	err = db.Conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='clients'").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query table existence: %v", err)
	}
	if name != "clients" {
		t.Errorf("expected table clients, got %s", name)
	}
}
