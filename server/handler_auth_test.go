//go:build !windows

package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func captureLog(fn func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	fn()
	return buf.String()
}

func TestHandleAuthLogsNeverContainNonce(t *testing.T) {
	dbPath := "test_auth_nonce.db"
	defer os.Remove(dbPath)
	db, _ := state.Open(dbPath)
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{PortalPort: 8084}
	srv := New(cfg, db, fw)

	nonce := "secret_nonce_value_12345"

	t.Run("nonce mismatch does not leak nonce values", func(t *testing.T) {
		logOutput := captureLog(func() {
			req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce="+nonce))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			// No cookie set -> nonce mismatch
			srv.handleAuth(rr, req)
		})

		if strings.Contains(logOutput, nonce) {
			t.Errorf("log must not contain the nonce value, got: %s", logOutput)
		}
		if strings.Contains(logOutput, "secret_nonce_value") {
			t.Error("log must not contain any part of the nonce")
		}
	})

	t.Run("missing cookie does not leak nonce values", func(t *testing.T) {
		logOutput := captureLog(func() {
			req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("nonce=abcdef"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			srv.handleAuth(rr, req)
		})

		if strings.Contains(logOutput, "abcdef") {
			t.Errorf("log must not contain form nonce, got: %s", logOutput)
		}
	})
}
