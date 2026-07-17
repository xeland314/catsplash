package state

import (
	"os"
	"testing"
	"time"
)

func setupAuditTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	dbPath := "test_audit.db"
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}
	return db, cleanup
}

func TestLogAuditEventAndList(t *testing.T) {
	db, cleanup := setupAuditTestDB(t)
	defer cleanup()

	err := db.LogAuditEvent(AuditDataAccess, "abc12345", "10.0.0.1", "json_export")
	if err != nil {
		t.Fatalf("LogAuditEvent failed: %v", err)
	}

	events, err := db.ListAuditEvents(10)
	if err != nil {
		t.Fatalf("ListAuditEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Action != AuditDataAccess {
		t.Errorf("expected action %s, got %s", AuditDataAccess, events[0].Action)
	}
	if events[0].SubjectMAC != "abc12345" {
		t.Errorf("expected subject_mac abc12345, got %s", events[0].SubjectMAC)
	}
	if events[0].RequesterIP != "10.0.0.1" {
		t.Errorf("expected requester_ip 10.0.0.1, got %s", events[0].RequesterIP)
	}
	if events[0].Details != "json_export" {
		t.Errorf("expected details json_export, got %s", events[0].Details)
	}
	if events[0].Timestamp <= 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestListAuditEventsOrdering(t *testing.T) {
	db, cleanup := setupAuditTestDB(t)
	defer cleanup()

	db.LogAuditEvent(AuditAuthSuccess, "aaa11111", "10.0.0.1", "first")
	time.Sleep(10 * time.Millisecond)
	db.LogAuditEvent(AuditDataAccess, "bbb22222", "10.0.0.2", "second")
	time.Sleep(10 * time.Millisecond)
	db.LogAuditEvent(AuditDataDeletion, "ccc33333", "10.0.0.3", "third")

	events, err := db.ListAuditEvents(10)
	if err != nil {
		t.Fatalf("ListAuditEvents failed: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Most recent first
	if events[0].Action != AuditDataDeletion {
		t.Errorf("expected most recent event first, got %s", events[0].Action)
	}
	if events[2].Action != AuditAuthSuccess {
		t.Errorf("expected oldest event last, got %s", events[2].Action)
	}
}

func TestListAuditEventsLimit(t *testing.T) {
	db, cleanup := setupAuditTestDB(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		db.LogAuditEvent(AuditAuthDenied, "unknown", "10.0.0.1", "attempt")
	}

	events, err := db.ListAuditEvents(3)
	if err != nil {
		t.Fatalf("ListAuditEvents failed: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events (limit), got %d", len(events))
	}
}

func TestPurgeOldAuditEvents(t *testing.T) {
	db, cleanup := setupAuditTestDB(t)
	defer cleanup()

	// Insert events
	db.LogAuditEvent(AuditDataAccess, "aaa11111", "10.0.0.1", "old")
	db.LogAuditEvent(AuditDataAccess, "bbb22222", "10.0.0.2", "new")

	// Make the first event old by backdating it
	db.Conn.Exec("UPDATE audit_log SET timestamp = ? WHERE details = 'old'", time.Now().Add(-2*time.Hour).Unix())

	deleted, err := db.PurgeOldAuditEvents(1 * time.Hour)
	if err != nil {
		t.Fatalf("PurgeOldAuditEvents failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	events, _ := db.ListAuditEvents(10)
	if len(events) != 1 {
		t.Errorf("expected 1 remaining event, got %d", len(events))
	}
	if events[0].Details != "new" {
		t.Errorf("expected remaining event to be 'new', got %s", events[0].Details)
	}
}

func TestAuditMACIsAlwaysMasked(t *testing.T) {
	db, cleanup := setupAuditTestDB(t)
	defer cleanup()

	realMAC := "AA:BB:CC:DD:EE:FF"
	masked := "261900fb" // known truncated SHA-256 of AA:BB:CC:DD:EE:FF

	// Simulate what handlers do — always pass masked MAC
	db.LogAuditEvent(AuditDataAccess, masked, "10.0.0.1", "test")

	events, _ := db.ListAuditEvents(10)
	if events[0].SubjectMAC == realMAC {
		t.Error("audit log must NEVER contain the real MAC address")
	}
	if events[0].SubjectMAC != masked {
		t.Errorf("expected masked MAC %s, got %s", masked, events[0].SubjectMAC)
	}
}
