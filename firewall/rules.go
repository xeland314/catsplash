package firewall

import (
	"fmt"
	"log"
)

// Init sets up the base rules, custom chain, and NAT.
func (f *Firewall) Init() error {
	// 0. Enable IP Forwarding in the kernel (requires root, but we run as root)
	f.exec.Execute("sysctl", "-w", "net.ipv4.ip_forward=1")

	// 1. Create custom chain
	if _, err := f.exec.Execute("iptables", "-t", "nat", "-N", "CATS_PREROUTING"); err != nil {
		f.exec.Execute("iptables", "-t", "nat", "-F", "CATS_PREROUTING")
	}

	// 2. Jump to custom chain from PREROUTING
	f.exec.Execute("iptables", "-t", "nat", "-D", "PREROUTING", "-i", f.iface, "-p", "tcp", "--dport", "80", "-j", "CATS_PREROUTING")
	if _, err := f.exec.Execute("iptables", "-t", "nat", "-A", "PREROUTING", "-i", f.iface, "-p", "tcp", "--dport", "80", "-j", "CATS_PREROUTING"); err != nil {
		return fmt.Errorf("failed to link PREROUTING: %w", err)
	}

	// 3. Setup NAT (Masquerade) to allow internet access
	f.exec.Execute("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", f.wanIface, "-j", "MASQUERADE")
	if _, err := f.exec.Execute("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", f.wanIface, "-j", "MASQUERADE"); err != nil {
		log.Printf("Warning: failed to setup NAT on %s: %v", f.wanIface, err)
	}

	// 4. Setup basic FORWARD rules
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-j", "DROP")
	if _, err := f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-j", "DROP"); err != nil {
		return fmt.Errorf("failed to set default DROP in FORWARD: %w", err)
	}

	// 5. Allow established/related traffic (important for NAT)
	f.exec.Execute("iptables", "-I", "FORWARD", "1", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")

	// 6. Allow DHCP and DNS
	f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-p", "udp", "--dport", "67:68", "-j", "ACCEPT")
	f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-p", "udp", "--dport", "53", "-j", "ACCEPT")
	f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-p", "tcp", "--dport", "53", "-j", "ACCEPT")

	// 7. Setup root qdisc for QoS
	f.ensureQdisc()

	return nil
}

// Teardown cleans up all rules.
func (f *Firewall) Teardown() {
	// Cleanup QoS
	f.cleanupQos()
	f.exec.Execute("iptables", "-t", "nat", "-D", "PREROUTING", "-i", f.iface, "-p", "tcp", "--dport", "80", "-j", "CATS_PREROUTING")
	f.exec.Execute("iptables", "-t", "nat", "-F", "CATS_PREROUTING")
	f.exec.Execute("iptables", "-t", "nat", "-X", "CATS_PREROUTING")
	f.exec.Execute("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", f.wanIface, "-j", "MASQUERADE")
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-j", "DROP")
	f.exec.Execute("iptables", "-D", "FORWARD", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
}
