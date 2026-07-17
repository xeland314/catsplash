package state

import (
	"crypto/sha256"
	"encoding/hex"
)

// MaskMAC returns a truncated SHA-256 hash of the MAC address for safe logging.
func MaskMAC(mac string) string {
	h := sha256.Sum256([]byte(mac))
	return hex.EncodeToString(h[:4])
}

// MaskIP returns a truncated SHA-256 hash of the IP address for safe logging.
func MaskIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:4])
}
