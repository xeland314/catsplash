package firewall

// BlockClient removes the ACCEPT rule, NAT bypass, and QoS limits for the specific MAC and IP.
func (f *Firewall) BlockClient(mac, ip string) error {
	// 1. Remove FORWARD upload rule
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")

	// 2. Remove FORWARD download rule
	f.exec.Execute("iptables", "-D", "FORWARD", "-o", f.iface, "-d", ip, "-j", "ACCEPT")

	// 3. Remove NAT bypass rule
	f.exec.Execute("iptables", "-t", "nat", "-D", "CATS_PREROUTING", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN")

	// 4. Remove QoS bandwidth limits
	f.removeClientQos(ip)

	return nil
}
