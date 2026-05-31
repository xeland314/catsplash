package firewall

import (
	"slices"
	"strings"
	"testing"
)

type MockExecutor struct {
	Commands []string
}

func (m *MockExecutor) Execute(name string, arg ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(arg, " ")
	m.Commands = append(m.Commands, cmd)
	return []byte(""), nil
}

func TestFirewallInit(t *testing.T) {
	mock := &MockExecutor{}
	fw := New("wlan0", mock)

	if err := fw.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	expectedCmds := []string{
		"iptables -t nat -N CATS_PREROUTING",
		"iptables -t nat -D PREROUTING -i wlan0 -p tcp --dport 80 -j CATS_PREROUTING",
		"iptables -t nat -A PREROUTING -i wlan0 -p tcp --dport 80 -j CATS_PREROUTING",
		"iptables -D FORWARD -i wlan0 -j DROP",
		"iptables -I FORWARD 1 -i wlan0 -j DROP",
	}

	for _, expected := range expectedCmds {
		found := slices.Contains(mock.Commands, expected)
		if !found {
			t.Errorf("expected command not found: %s", expected)
		}
	}
}

func TestFirewallAllowBlock(t *testing.T) {
	mock := &MockExecutor{}
	fw := New("wlan0", mock)

	mac := "00:11:22:33:44:55"
	ip := "192.168.1.50"

	fw.AllowClient(mac, ip)
	fw.BlockClient(mac, ip)

	expectedCmds := []string{
		// Idempotency cleanup
		"iptables -D FORWARD -i wlan0 -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j ACCEPT",
		"iptables -t nat -D CATS_PREROUTING -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j RETURN",
		// Add rules
		"iptables -I FORWARD 1 -i wlan0 -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j ACCEPT",
		"iptables -t nat -I CATS_PREROUTING 1 -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j RETURN",
		// Block client
		"iptables -D FORWARD -i wlan0 -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j ACCEPT",
		"iptables -t nat -D CATS_PREROUTING -s 192.168.1.50 -m mac --mac-source 00:11:22:33:44:55 -j RETURN",
	}

	if len(mock.Commands) != len(expectedCmds) {
		t.Errorf("expected %d commands, got %d", len(expectedCmds), len(mock.Commands))
	}

	for i, expected := range expectedCmds {
		if i < len(mock.Commands) && mock.Commands[i] != expected {
			t.Errorf("expected command %d to be %s, got %s", i, expected, mock.Commands[i])
		}
	}
}
