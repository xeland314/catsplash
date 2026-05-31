package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	// Detect CNA or direct requests
	if strings.Contains(r.URL.Path, "generate_204") ||
		strings.Contains(r.URL.Path, "hotspot-detect") ||
		strings.Contains(r.URL.Path, "ncsi.txt") {
		http.Redirect(w, r, "/portal", http.StatusFound)
		return
	}

	// For any other request, if not authenticated, redirect to portal
	ip := getIPFromRemoteAddr(r.RemoteAddr)
	mac, err := getMACFromIP(ip)
	if err != nil {
		http.Redirect(w, r, "/portal", http.StatusFound)
		return
	}

	client, err := s.db.GetClient(mac)
	if err != nil || client == nil || client.State != "authenticated" {
		http.Redirect(w, r, "/portal", http.StatusFound)
		return
	}

	// If authenticated, we shouldn't even be here (iptables would allow FORWARD),
	// but maybe they are hitting the portal IP directly.
	http.NotFound(w, r)
}
