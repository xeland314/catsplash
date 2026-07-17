package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func setupDataTestServer(t *testing.T) (*Server, func()) {
	t.Helper()
	dbPath := "test_data_handler.db"
	db, err := state.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{PortalPort: 8090}
	srv := New(cfg, db, fw)
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}
	return srv, cleanup
}

func insertTestClientWithSession(t *testing.T, db *state.DB, mac, ip, token string) {
	t.Helper()
	err := db.UpsertClient(mac, ip, true)
	if err != nil {
		t.Fatalf("UpsertClient failed: %v", err)
	}
	err = db.Authenticate(mac, ip)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	err = db.SetSessionToken(mac, token)
	if err != nil {
		t.Fatalf("SetSessionToken failed: %v", err)
	}
}

func TestHandleDataRequestReturnsClientData(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	token := "valid_session_token_abc123"
	insertTestClientWithSession(t, srv.db, "AA:BB:CC:DD:EE:FF", "10.0.0.5", token)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data-request", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: token})

	srv.handleDataRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp dataResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC AA:BB:CC:DD:EE:FF, got %s", resp.MAC)
	}
	if resp.IP != "10.0.0.5" {
		t.Errorf("expected IP 10.0.0.5, got %s", resp.IP)
	}
	if resp.State != "authenticated" {
		t.Errorf("expected state authenticated, got %s", resp.State)
	}
	if !resp.ConsentGiven {
		t.Error("expected consent_given=true")
	}
}

func TestHandleDataRequestWithoutCookie(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data-request", nil)

	srv.handleDataRequest(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleDataRequestWithInvalidToken(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data-request", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: "nonexistent_token"})

	srv.handleDataRequest(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleDataRequestRejectsMacParameter(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	insertTestClientWithSession(t, srv.db, "AA:BB:CC:DD:EE:FF", "10.0.0.5", "legit_token")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data-request?mac=AA:BB:CC:DD:EE:FF", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: "nonexistent_token"})

	srv.handleDataRequest(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("IDOR: expected 404 when using MAC param with wrong token, got %d", rr.Code)
	}
}

func TestHandleDataRequestMethodNotAllowed(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-request", nil)

	srv.handleDataRequest(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleDataDeletionDeletesClient(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	mac := "AA:BB:CC:DD:EE:FF"
	token := "delete_session_token_xyz"
	insertTestClientWithSession(t, srv.db, mac, "10.0.0.6", token)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-deletion", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: token})

	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp["ok"] != true {
		t.Error("expected ok=true in response")
	}

	// Verify client is actually deleted from DB
	client, err := srv.db.GetClient(mac)
	if err != nil {
		t.Fatalf("DB error: %v", err)
	}
	if client != nil {
		t.Error("client should be deleted from DB")
	}
}

func TestHandleDataDeletionWithoutCookie(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-deletion", nil)

	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleDataDeletionWithInvalidToken(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-deletion", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: "fake_token"})

	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleDataDeletionRejectsMacParameter(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	insertTestClientWithSession(t, srv.db, "AA:BB:CC:DD:EE:FF", "10.0.0.5", "legit_delete_token")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-deletion", strings.NewReader("mac=AA:BB:CC:DD:EE:FF"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: "wrong_token"})

	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("IDOR: expected 404 when using MAC body param with wrong token, got %d", rr.Code)
	}

	// Verify original client is NOT deleted
	client, err := srv.db.GetClient("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("DB error: %v", err)
	}
	if client == nil {
		t.Error("IDOR: client should NOT be deleted when using wrong token")
	}
}

func TestHandleDataDeletionMethodNotAllowed(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data-deletion", nil)

	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleDataDeletionIsIrreversible(t *testing.T) {
	srv, cleanup := setupDataTestServer(t)
	defer cleanup()

	mac := "AA:BB:CC:DD:EE:FF"
	token := " irreversible_token_123"
	insertTestClientWithSession(t, srv.db, mac, "10.0.0.7", token)

	// First deletion — should succeed
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-deletion", nil)
	req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: token})
	srv.handleDataDeletion(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("first deletion: expected 200, got %d", rr.Code)
	}

	// Second deletion — same token should return 404 (already deleted)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/data-deletion", nil)
	req2.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: token})
	srv.handleDataDeletion(rr2, req2)

	if rr2.Code != http.StatusNotFound {
		t.Errorf("second deletion: expected 404 (already deleted), got %d", rr2.Code)
	}
}
