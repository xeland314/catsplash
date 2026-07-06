package firewall

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

// TrafficStats represents the parsed bytes for a client.
type TrafficStats struct {
	MAC      string
	IP       string
	BytesIn  int64 // Download
	BytesOut int64 // Upload
}

// QueryTrafficCounters runs iptables to get byte counters for all forwarding rules.
// It parses the output and aggregates statistics per client IP/MAC.
func (f *Firewall) QueryTrafficCounters() (map[string]*TrafficStats, error) {
	output, err := f.exec.Execute("iptables", "-L", "FORWARD", "-v", "-n", "-x")
	if err != nil {
		return nil, err
	}

	statsByIP := make(map[string]*TrafficStats) // IP -> Stats
	ipToMAC := make(map[string]string)          // IP -> MAC

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// Expected format:
		// pkts bytes target prot opt in out source destination [options...]
		// Index:
		// 0: pkts
		// 1: bytes
		// 2: target
		// 3: prot
		// 4: opt
		// 5: in
		// 6: out
		// 7: source
		// 8: destination

		if fields[2] != "ACCEPT" {
			continue
		}

		bytesVal, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}

		inIface := fields[5]
		outIface := fields[6]
		srcIP := fields[7]
		dstIP := fields[8]

		// 1. Check if it's an Upload rule (e.g. in = f.iface, source = client IP, options contain MAC)
		if inIface == f.iface && srcIP != "0.0.0.0/0" && dstIP == "0.0.0.0/0" {
			// Find MAC in the options. Look for "MAC" token
			var mac string
			for i, field := range fields {
				if field == "MAC" && i+1 < len(fields) {
					mac = strings.ToLower(fields[i+1])
					break
				}
			}

			if mac != "" {
				ipToMAC[srcIP] = mac
				stats, exists := statsByIP[srcIP]
				if !exists {
					stats = &TrafficStats{IP: srcIP, MAC: mac}
					statsByIP[srcIP] = stats
				}
				stats.BytesOut = bytesVal
			}
		}

		// 2. Check if it's a Download rule (e.g. out = f.iface, destination = client IP)
		if outIface == f.iface && dstIP != "0.0.0.0/0" && srcIP == "0.0.0.0/0" {
			stats, exists := statsByIP[dstIP]
			if !exists {
				stats = &TrafficStats{IP: dstIP}
				statsByIP[dstIP] = stats
			}
			stats.BytesIn = bytesVal
		}
	}

	// Build the final map keyed by MAC address
	statsByMAC := make(map[string]*TrafficStats)
	for ip, stats := range statsByIP {
		mac := stats.MAC
		if mac == "" {
			mac = ipToMAC[ip]
		}
		if mac != "" {
			stats.MAC = mac
			statsByMAC[mac] = stats
		}
	}

	return statsByMAC, nil
}
