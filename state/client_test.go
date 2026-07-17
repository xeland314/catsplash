package state

import (
	"os"
	"testing"
)

func TestClientOperations(t *testing.T) {
	dbPath := "test_ops.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	mac := "00:11:22:33:44:55"
	ip := "192.168.1.10"

	// Upsert
	if err := db.UpsertClient(mac, ip, true); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Get
	c, err := db.GetClient(mac)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if c.MAC != mac || c.IP != ip || c.State != StatePending {
		t.Errorf("unexpected client data: %+v", c)
	}

	// Authenticate
	if err := db.Authenticate(mac, ip); err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	c, _ = db.GetClient(mac)
	if c.State != StateAuthenticated || c.ConnectedAt == 0 {
		t.Errorf("expected authenticated, got: %+v", c)
	}

	// List
	list, err := db.ListAuthenticated()
	if err != nil {
		t.Fatalf("ListAuthenticated failed: %v", err)
	}
	if len(list) != 1 || list[0].MAC != mac {
		t.Errorf("expected 1 authenticated client, got: %d", len(list))
	}

	// Deauthenticate
	if err := db.Deauthenticate(mac); err != nil {
		t.Fatalf("Deauthenticate failed: %v", err)
	}
	c, _ = db.GetClient(mac)
	if c.State != StatePending {
		t.Errorf("expected pending after deauth, got: %s", c.State)
	}
}
