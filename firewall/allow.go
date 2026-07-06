package firewall

// AllowClient adds an ACCEPT rule for the specific MAC and IP, and a bypass for NAT redirection.
// It is idempotent: it removes existing rules before adding them to avoid duplicates.
func (f *Firewall) AllowClient(mac, ip string) error {
	// 1. Clean up existing rules to avoid duplicates
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")
	f.exec.Execute("iptables", "-D", "FORWARD", "-o", f.iface, "-d", ip, "-j", "ACCEPT")
	f.exec.Execute("iptables", "-t", "nat", "-D", "CATS_PREROUTING", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN")

	// 2. Allow FORWARD upload traffic (from client)
	if _, err := f.exec.Execute("iptables", "-I", "FORWARD", "1", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT"); err != nil {
		return err
	}

	// 3. Allow FORWARD download traffic (to client)
	if _, err := f.exec.Execute("iptables", "-I", "FORWARD", "1", "-o", f.iface, "-d", ip, "-j", "ACCEPT"); err != nil {
		f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")
		return err
	}

	// 4. Bypass redirection in NAT table
	if _, err := f.exec.Execute("iptables", "-t", "nat", "-I", "CATS_PREROUTING", "1", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN"); err != nil {
		f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")
		f.exec.Execute("iptables", "-D", "FORWARD", "-o", f.iface, "-d", ip, "-j", "ACCEPT")
		return err
	}

	// 5. Apply QoS bandwidth limits
	if err := f.applyClientQos(mac, ip, "", ""); err != nil {
		return err
	}

	return nil
}

// AllowClientWithSpeed allows a client with specific bandwidth limits.
func (f *Firewall) AllowClientWithSpeed(mac, ip, downloadSpeed, uploadSpeed string) error {
	if err := f.AllowClient(mac, ip); err != nil {
		return err
	}
	return f.applyClientQos(mac, ip, downloadSpeed, uploadSpeed)
}
