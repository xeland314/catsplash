package server

import "testing"

func TestIsValidMAC(t *testing.T) {
	valid := []string{
		"AA:BB:CC:DD:EE:FF",
		"aa:bb:cc:dd:ee:ff",
		"00:11:22:33:44:55",
		"AA-BB-CC-DD-EE-FF",
		"AABBCCDDEEFF",
	}
	for _, mac := range valid {
		if !isValidMAC(mac) {
			t.Errorf("expected valid MAC: %s", mac)
		}
	}

	invalid := []string{
		"",
		"not-a-mac",
		"AA:BB:CC:DD:EE",       // too short
		"AA:BB:CC:DD:EE:FF:00", // too long
		"GG:HH:II:JJ:KK:LL",   // invalid hex
	}
	for _, mac := range invalid {
		if isValidMAC(mac) {
			t.Errorf("expected invalid MAC: %s", mac)
		}
	}
}

func TestIsValidMACRejectsShellInjection(t *testing.T) {
	payloads := []string{
		"AA:BB:CC:DD:EE:FF; rm -rf /",
		"AA:BB:CC:DD:EE:FF| cat /etc/passwd",
		"AA:BB:CC:DD:EE:FF`id`",
		"AA:BB:CC:DD:EE:FF$(id)",
		"AA:BB:CC:DD:EE:FF&& whoami",
		"AA:BB:CC:DD:EE:FF|| reboot",
		"AA:BB:CC:DD:EE:FF\n/etc/passwd",
		"AA:BB:CC:DD:EE:FF\r/etc/shadow",
		"; echo pwned",
		"| cat /etc/shadow",
		"`reboot`",
		"$(rm -rf /)",
		"AA:BB:CC:DD:EE:FF\x00AAAA",
		"AA:BB:CC:DD:EE' OR 1=1--",
		"AA:BB:CC:DD:EE\" OR \"1\"=\"1",
		"A]B:C[D:E:F:G:H",
	}
	for _, payload := range payloads {
		if isValidMAC(payload) {
			t.Errorf("MAC validation must reject injection: %q", payload)
		}
	}
}
