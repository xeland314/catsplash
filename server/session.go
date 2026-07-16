package server

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

// getMACFromIP retrieves the MAC address for a given IP from /proc/net/arp.
func getMACFromIP(ip string) (string, error) {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == ip {
			return fields[3], nil
		}
	}

	return "", fmt.Errorf("MAC not found for IP: %s", ip)
}

// getIPFromRemoteAddr extracts the IP address from a RemoteAddr string.
func getIPFromRemoteAddr(remoteAddr string) string {
	ip, _, _ := net.SplitHostPort(remoteAddr)
	return ip
}

// generateNonce creates a random string for CSRF protection.
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// maskMAC returns a truncated SHA-256 hash of the MAC address for safe logging.
func maskMAC(mac string) string {
	h := sha256.Sum256([]byte(mac))
	return hex.EncodeToString(h[:4])
}
