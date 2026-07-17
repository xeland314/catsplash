package state_test

import (
	"testing"

	"github.com/xeland314/catsplash/state"
)

func TestMaskMAC(t *testing.T) {
	masked := state.MaskMAC("AA:BB:CC:DD:EE:FF")
	if masked == "AA:BB:CC:DD:EE:FF" {
		t.Error("MaskMAC must not return the original MAC")
	}
	if len(masked) != 8 {
		t.Errorf("MaskMAC should return 8 hex chars, got %d (%q)", len(masked), masked)
	}
}

func TestMaskMACDeterministic(t *testing.T) {
	if state.MaskMAC("AA:BB:CC:DD:EE:FF") != state.MaskMAC("AA:BB:CC:DD:EE:FF") {
		t.Error("MaskMAC must be deterministic")
	}
}

func TestMaskMACDifferentInputs(t *testing.T) {
	a := state.MaskMAC("AA:BB:CC:DD:EE:FF")
	b := state.MaskMAC("11:22:33:44:55:66")
	if a == b {
		t.Error("MaskMAC should produce different hashes for different inputs")
	}
}

func TestMaskMACEmpty(t *testing.T) {
	masked := state.MaskMAC("")
	if len(masked) != 8 {
		t.Errorf("MaskMAC of empty string should still return 8 hex chars, got %d", len(masked))
	}
}

func TestMaskIP(t *testing.T) {
	masked := state.MaskIP("192.168.1.1")
	if masked == "192.168.1.1" {
		t.Error("MaskIP must not return the original IP")
	}
	if len(masked) != 8 {
		t.Errorf("MaskIP should return 8 hex chars, got %d (%q)", len(masked), masked)
	}
}

func TestMaskIPDeterministic(t *testing.T) {
	if state.MaskIP("10.0.0.1") != state.MaskIP("10.0.0.1") {
		t.Error("MaskIP must be deterministic")
	}
}

func TestMaskIPDifferentInputs(t *testing.T) {
	a := state.MaskIP("192.168.1.1")
	b := state.MaskIP("10.0.0.1")
	if a == b {
		t.Error("MaskIP should produce different hashes for different inputs")
	}
}

func TestMaskIPEmpty(t *testing.T) {
	masked := state.MaskIP("")
	if len(masked) != 8 {
		t.Errorf("MaskIP of empty string should still return 8 hex chars, got %d", len(masked))
	}
}
