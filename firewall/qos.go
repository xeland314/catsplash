package firewall

import (
	"fmt"
	"strconv"
	"strings"
)

func ipToClassID(ip string, base int) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Sprintf("1:%d", base)
	}
	lastOctet, _ := strconv.Atoi(parts[3])
	return fmt.Sprintf("1:%d", base+lastOctet)
}

func (f *Firewall) ensureQdisc() error {
	if f.qdiscInit {
		return nil
	}

	f.exec.Execute("tc", "qdisc", "del", "dev", f.iface, "root")

	if _, err := f.exec.Execute("tc", "qdisc", "add", "dev", f.iface, "root", "handle", "1:", "htb", "default", "30000"); err != nil {
		return fmt.Errorf("failed to create root qdisc: %w", err)
	}

	if _, err := f.exec.Execute("tc", "class", "add", "dev", f.iface, "parent", "1:", "classid", "1:1", "htb", "rate", "1000mbit"); err != nil {
		return fmt.Errorf("failed to create root class: %w", err)
	}

	f.exec.Execute("tc", "class", "add", "dev", f.iface, "parent", "1:1", "classid", "1:30000", "htb", "rate", "1000mbit")

	f.qdiscInit = true
	return nil
}

func (f *Firewall) applyClientQos(mac, ip, downloadSpeed, uploadSpeed string) error {
	if downloadSpeed == "0" && uploadSpeed == "0" {
		return nil
	}
	if downloadSpeed == "" {
		downloadSpeed = f.DownloadSpeed
	}
	if uploadSpeed == "" {
		uploadSpeed = f.UploadSpeed
	}
	if downloadSpeed == "0" && uploadSpeed == "0" {
		return nil
	}

	if err := f.ensureQdisc(); err != nil {
		return err
	}

	upClass := ipToClassID(ip, 100)
	downClass := ipToClassID(ip, 200)

	f.exec.Execute("tc", "class", "del", "dev", f.iface, "parent", "1:1", "classid", upClass)
	f.exec.Execute("tc", "class", "del", "dev", f.iface, "parent", "1:1", "classid", downClass)

	if uploadSpeed != "0" {
		if _, err := f.exec.Execute("tc", "class", "add", "dev", f.iface, "parent", "1:1", "classid", upClass, "htb", "rate", uploadSpeed); err != nil {
			return fmt.Errorf("failed to create upload class for %s: %w", ip, err)
		}
		if _, err := f.exec.Execute("tc", "filter", "add", "dev", f.iface, "parent", "1:", "protocol", "ip", "prio", "1", "u32", "match", "ip", "src", ip, "flowid", upClass); err != nil {
			return fmt.Errorf("failed to add upload filter for %s: %w", ip, err)
		}
	}

	if downloadSpeed != "0" {
		if _, err := f.exec.Execute("tc", "class", "add", "dev", f.iface, "parent", "1:1", "classid", downClass, "htb", "rate", downloadSpeed); err != nil {
			return fmt.Errorf("failed to create download class for %s: %w", ip, err)
		}
		if _, err := f.exec.Execute("tc", "filter", "add", "dev", f.iface, "parent", "1:", "protocol", "ip", "prio", "2", "u32", "match", "ip", "dst", ip, "flowid", downClass); err != nil {
			return fmt.Errorf("failed to add download filter for %s: %w", ip, err)
		}
	}

	return nil
}

func (f *Firewall) removeClientQos(ip string) {
	if !f.qdiscInit {
		return
	}
	upClass := ipToClassID(ip, 100)
	downClass := ipToClassID(ip, 200)

	f.exec.Execute("tc", "filter", "del", "dev", f.iface, "parent", "1:", "protocol", "ip", "prio", "1", "u32", "match", "ip", "src", ip)
	f.exec.Execute("tc", "filter", "del", "dev", f.iface, "parent", "1:", "protocol", "ip", "prio", "2", "u32", "match", "ip", "dst", ip)

	f.exec.Execute("tc", "class", "del", "dev", f.iface, "parent", "1:1", "classid", upClass)
	f.exec.Execute("tc", "class", "del", "dev", f.iface, "parent", "1:1", "classid", downClass)
}

func (f *Firewall) cleanupQos() {
	f.exec.Execute("tc", "qdisc", "del", "dev", f.iface, "root")
	f.qdiscInit = false
}
