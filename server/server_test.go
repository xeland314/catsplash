package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func TestHandlers(t *testing.T) {
	// Setup dependencies
	dbPath := "test_server.db"
	defer os.Remove(dbPath)
	db, _ := state.Open(dbPath)
	
	fw := firewall.New("wlan0", nil)
	cfg := &config.Config{PortalPort: 8080}
	
	srv := New(cfg, db, fw)

	t.Run("Portal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/portal", nil)
		rr := httptest.NewRecorder()
		srv.handlePortal(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}
		
		if rr.Header().Get("Set-Cookie") == "" {
			t.Error("expected nonce cookie to be set")
		}
	})

	t.Run("Redirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/generate_204", nil)
		rr := httptest.NewRecorder()
		srv.handleRedirect(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
		}
		
		if rr.Header().Get("Location") != "/portal" {
			t.Errorf("expected redirect to /portal, got %s", rr.Header().Get("Location"))
		}
	})
}
