package server

import (
	"strings"
	"testing"
)

func TestMaskMAC(t *testing.T) {
 mac := "AA:BB:CC:DD:EE:FF"
 masked := maskMAC(mac)

 if masked == mac {
 	t.Error("maskMAC must not return the original MAC")
 }
 if len(masked) != 8 {
 	t.Errorf("maskMAC should return 8 hex chars, got %d (%q)", len(masked), masked)
 }
 if !strings.ContainsAny(masked, "0123456789abcdef") {
 	t.Errorf("maskMAC should return hex, got %q", masked)
 }
}

func TestMaskMACDeterministic(t *testing.T) {
 mac := "00:11:22:33:44:55"
 if maskMAC(mac) != maskMAC(mac) {
 	t.Error("maskMAC must be deterministic")
 }
}

func TestMaskMACDifferentInputs(t *testing.T) {
 a := maskMAC("AA:BB:CC:DD:EE:FF")
 b := maskMAC("11:22:33:44:55:66")
 if a == b {
 	t.Error("different MACs must produce different masks")
 }
}

func TestMaskMACNoColonFormat(t *testing.T) {
 masked := maskMAC("AABBCCDDEEFF")
 if masked == "AABBCCDDEEFF" {
 	t.Error("maskMAC must not return the original value")
 }
 if len(masked) != 8 {
 	t.Errorf("expected 8 hex chars, got %d", len(masked))
 }
}

func TestMaskMACEmptyString(t *testing.T) {
 masked := maskMAC("")
 if len(masked) != 8 {
 	t.Errorf("maskMAC of empty string should still return 8 hex chars, got %d", len(masked))
 }
}
