package firewall

// BlockClient removes the ACCEPT rule and the NAT bypass for the specific MAC and IP.
func (f *Firewall) BlockClient(mac, ip string) error {
	// 1. Remove FORWARD rule
	f.exec.Execute("iptables", "-D", "FORWARD", "-i", f.iface, "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "ACCEPT")

	// 2. Remove NAT bypass rule
	_, err := f.exec.Execute("iptables", "-t", "nat", "-D", "CATS_PREROUTING", "-s", ip, "-m", "mac", "--mac-source", mac, "-j", "RETURN")
	return err
}
