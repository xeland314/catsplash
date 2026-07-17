package state

import (
	"os"
	"testing"
	"time"
)

func TestReaper(t *testing.T) {
	dbPath := "test_reaper.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	expiredMACs := make(map[string]bool)
	onExpire := func(mac, ip string) error {
		expiredMACs[mac] = true
		return nil
	}

	reaper := NewReaper(db, 2, 1, onExpire) // 2s session, 1s idle

	mac1 := "00:11:22:33:44:55"
	mac2 := "AA:BB:CC:DD:EE:FF"

	db.UpsertClient(mac1, "1.1.1.1", true)
	db.Authenticate(mac1, "1.1.1.1")

	db.UpsertClient(mac2, "2.2.2.2", true)
	db.Authenticate(mac2, "2.2.2.2")

	// Wait 2s (should trigger idle for both since idleTimeout is 1s)
	time.Sleep(2 * time.Second)
	
	reaper.RunOnce()

	if !expiredMACs[mac1] || !expiredMACs[mac2] {
		t.Errorf("expected both to be expired due to idle timeout")
	}

	c, _ := db.GetClient(mac1)
	if c.State != StatePending {
		t.Errorf("expected state pending, got %s", c.State)
	}
}
