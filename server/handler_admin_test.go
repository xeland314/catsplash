package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
	"golang.org/x/crypto/bcrypt"
)

func TestBasicAuthUsesBcryptComparison(t *testing.T) {
	db, _ := state.Open("test_admin.db")
	defer os.Remove("test_admin.db")

	// Config with a bcrypt-hashed password for "catsplash"
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{
		PortalPort: 8081,
		AdminUser:  "admin",
		AdminPass:  "$2a$10$dummyhashshouldfailcomparison", // invalid hash
	}
	srv := New(cfg, db, fw)

	// Wrap handleAdmin with basicAuth
	protected := srv.basicAuth(srv.handleAdmin)

	t.Run("rejects wrong password", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.SetBasicAuth("admin", "wrongpassword")
		rr := httptest.NewRecorder()
		protected(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("rejects correct password against invalid hash", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.SetBasicAuth("admin", "anything")
		rr := httptest.NewRecorder()
		protected(rr, req)

		// Must not panic or return 200 — the hash is invalid so bcrypt fails
		if rr.Code == http.StatusOK {
			t.Error("should not authenticate with invalid hash")
		}
	})
}

func TestBasicAuthRejectsPlaintextPassword(t *testing.T) {
	db, _ := state.Open("test_admin2.db")
	defer os.Remove("test_admin2.db")

	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{
		PortalPort: 8082,
		AdminUser:  "admin",
		AdminPass:  "$2a$10$invalid bcrypt hash that will never match", // bcrypt format but wrong
	}
	srv := New(cfg, db, fw)

	protected := srv.basicAuth(srv.handleAdmin)

	t.Run("plaintext password does not match bcrypt hash", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.SetBasicAuth("admin", "catsplash")
		rr := httptest.NewRecorder()
		protected(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("plaintext password must not authenticate, got %d", rr.Code)
		}
	})
}

func TestBasicAuthWithValidBcrypt(t *testing.T) {
	db, _ := state.Open("test_admin3.db")
	defer os.Remove("test_admin3.db")

	// Generate a real bcrypt hash for "testpass"
	hash := hashForTest(t, "testpass")

	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{
		PortalPort: 8083,
		AdminUser:  "admin",
		AdminPass:  hash,
	}
	srv := New(cfg, db, fw)

	protected := srv.basicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	t.Run("correct password authenticates", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.SetBasicAuth("admin", "testpass")
		rr := httptest.NewRecorder()
		protected(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.SetBasicAuth("admin", "wrongpass")
		rr := httptest.NewRecorder()
		protected(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})
}

func hashForTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}
	return string(hash)
}
