package firewall

import "fmt"

// SetupRedirect adds the DNAT rule to the custom chain.
func (f *Firewall) SetupRedirect(portalIP string, portalPort int) error {
	target := fmt.Sprintf("%s:%d", portalIP, portalPort)
	_, err := f.exec.Execute("iptables", "-t", "nat", "-A", "CATS_PREROUTING", "-p", "tcp", "--dport", "80", "-j", "DNAT", "--to-destination", target)
	return err
}
