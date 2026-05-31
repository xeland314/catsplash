package firewall

// AllowClient adds an ACCEPT rule for the specific MAC and IP, and a bypass for NAT redirection.
// It is idempotent: it removes existing rules before adding them to avoid duplicates.
func (f *Firewall) AllowClient(mac, ip string) error {
	// 1. Clean up existing rules to avoid duplicates
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")
	f.exec.Execute("iptables", "-t", "nat", "-D", "CATS_PREROUTING", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN")

	// 2. Allow FORWARD traffic
	if _, err := f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT"); err != nil {
		return err
	}

	// 3. Bypass redirection in NAT table
	_, err := f.exec.Execute("iptables", "-t", "nat", "-I", "CATS_PREROUTING", "1", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN")
	return err
}
