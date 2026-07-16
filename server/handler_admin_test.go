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
	"golang.org/x/crypto/bcrypt"
)

func TestBasicAuthUsesBcryptComparison(t *testing.T) {
	db, _ := state.Open("test_admin.db")
	defer os.Remove("test_admin.db")

	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{
		PortalPort: 8081,
		AdminUser:  "admin",
		AdminPass:  "$2a$10$dummyhashshouldfailcomparison",
	}
	srv := New(cfg, db, fw)

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
		AdminPass:  "$2a$10$invalid bcrypt hash that will never match",
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

// getCSRF performs GET /admin and returns the CSRF token + cookie for reuse.
func getCSRF(t *testing.T, srv *Server) string {
	t.Helper()
	req, _ := http.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	for _, c := range rr.Result().Cookies() {
		if c.Name == csrfCookieName {
			return c.Value
		}
	}
	t.Fatal("GET /admin did not set CSRF cookie")
	return ""
}

// postAdmin performs POST /admin with the given form body and CSRF cookie.
func postAdmin(t *testing.T, srv *Server, body string, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	req, _ := http.NewRequest("POST", "/admin", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfToken})
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)
	return rr
}

func TestAdminGetSetsCSRFCookie(t *testing.T) {
	db, _ := state.Open("test_csrf_cookie.db")
	defer os.Remove("test_csrf_cookie.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8090, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	req, _ := http.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == csrfCookieName {
			found = true
			if c.Value == "" {
				t.Error("CSRF cookie value must not be empty")
			}
			if !c.HttpOnly {
				t.Error("CSRF cookie must be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("GET /admin must set CSRF cookie")
	}
}

func TestAdminGetEmbedsCSRFTokenInPage(t *testing.T) {
	db, _ := state.Open("test_csrf_page.db")
	defer os.Remove("test_csrf_page.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8091, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	req, _ := http.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `meta name="csrf-token"`) {
		t.Error("admin page must contain csrf-token meta tag")
	}
}

func TestAdminFormsWithClientsHaveCSRFToken(t *testing.T) {
	db, _ := state.Open("test_csrf_forms.db")
	defer os.Remove("test_csrf_forms.db")

	db.UpsertClient("AA:BB:CC:DD:EE:FF", "192.168.10.5")
	db.Authenticate("AA:BB:CC:DD:EE:FF", "192.168.10.5")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8098, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	req, _ := http.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `name="csrf_token"`) {
		t.Error("admin page with clients must contain csrf_token hidden field in forms")
	}
}

func TestAdminPostRejectsMissingCSRFToken(t *testing.T) {
	db, _ := state.Open("test_csrf_reject.db")
	defer os.Remove("test_csrf_reject.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8092, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	rr := postAdmin(t, srv, "action=kick&mac=AA:BB:CC:DD:EE:FF", "")

	if rr.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF token must return 403, got %d", rr.Code)
	}
}

func TestAdminPostRejectsWrongCSRFToken(t *testing.T) {
	db, _ := state.Open("test_csrf_wrong.db")
	defer os.Remove("test_csrf_wrong.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8093, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	rr := postAdmin(t, srv, "action=kick&mac=AA:BB:CC:DD:EE:FF&csrf_token=wrong_token", "wrong_cookie")

	if rr.Code != http.StatusForbidden {
		t.Errorf("POST with wrong CSRF token must return 403, got %d", rr.Code)
	}
}

func TestAdminPostRejectsCSRFTokenWithoutCookie(t *testing.T) {
	db, _ := state.Open("test_csrf_nocookie.db")
	defer os.Remove("test_csrf_nocookie.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8094, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	req, _ := http.NewRequest("POST", "/admin", strings.NewReader("action=kick&mac=AA:BB:CC:DD:EE:FF&csrf_token=some_token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF cookie must return 403, got %d", rr.Code)
	}
}

func TestAdminPostWithValidCSRFProcessesAction(t *testing.T) {
	db, _ := state.Open("test_csrf_valid.db")
	defer os.Remove("test_csrf_valid.db")

	db.UpsertClient("AA:BB:CC:DD:EE:FF", "192.168.10.5")
	db.Authenticate("AA:BB:CC:DD:EE:FF", "192.168.10.5")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8095, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	csrfToken := getCSRF(t, srv)
	rr := postAdmin(t, srv, "action=kick&mac=AA:BB:CC:DD:EE:FF&csrf_token="+csrfToken, csrfToken)

	if rr.Code != http.StatusFound {
		t.Errorf("POST with valid CSRF must redirect (302), got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/admin" {
		t.Errorf("must redirect to /admin, got %s", rr.Header().Get("Location"))
	}
}

func TestAdminPostUnknownActionReturns400(t *testing.T) {
	db, _ := state.Open("test_csrf_action.db")
	defer os.Remove("test_csrf_action.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8096, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	csrfToken := getCSRF(t, srv)
	rr := postAdmin(t, srv, "action=unknown&mac=AA:BB:CC:DD:EE:FF&csrf_token="+csrfToken, csrfToken)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("unknown action must return 400, got %d", rr.Code)
	}
}

func TestAdminGetIgnoresActionParams(t *testing.T) {
	db, _ := state.Open("test_csrf_getignore.db")
	defer os.Remove("test_csrf_getignore.db")

	fw := firewall.New("wlan0", "eth0", nil)
	hash := hashForTest(t, "pass")
	cfg := &config.Config{PortalPort: 8097, AdminUser: "admin", AdminPass: hash}
	srv := New(cfg, db, fw)

	req, _ := http.NewRequest("GET", "/admin?action=kick&mac=AA:BB:CC:DD:EE:FF", nil)
	rr := httptest.NewRecorder()
	srv.handleAdmin(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET with action params must render page (200), got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `meta name="csrf-token"`) {
		t.Error("GET must render page with CSRF token")
	}
}

func hashForTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}
	return string(hash)
}
