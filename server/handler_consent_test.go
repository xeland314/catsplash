package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func TestHandleAuthRejectsWithoutConsent(t *testing.T) {
	dbPath := "test_auth_consent.db"
	defer os.Remove(dbPath)
	db, _ := state.Open(dbPath)
	defer db.Close()
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{PortalPort: 8085}
	srv := New(cfg, db, fw)

	nonce := "test_nonce_consent"

	t.Run("missing consent field is rejected", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce="+nonce))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: nonce})

		srv.handleAuth(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "política de privacidad") {
			t.Errorf("expected privacy policy error message, got: %s", rr.Body.String())
		}
	})

	t.Run("consent=false is rejected", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce="+nonce+"&consent=false"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: nonce})

		srv.handleAuth(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("consent=other value is rejected", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce="+nonce+"&consent=yes"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: nonce})

		srv.handleAuth(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("consent=true passes consent check", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce="+nonce+"&consent=true"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "catsplash_nonce", Value: nonce})

		srv.handleAuth(rr, req)

		// Should NOT be 400 for consent reasons — may fail later on MAC resolution
		if rr.Code == http.StatusBadRequest {
			body := rr.Body.String()
			if strings.Contains(body, "política de privacidad") {
				t.Errorf("consent=true should pass consent check, got: %s", body)
			}
		}
	})
}
