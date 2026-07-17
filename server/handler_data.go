package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/xeland314/catsplash/state"
)

// dataResponse is the JSON structure returned by /data-request.
type dataResponse struct {
	MAC              string `json:"mac"`
	IP               string `json:"ip"`
	State            string `json:"state"`
	ConnectedAt      int64  `json:"connected_at"`
	LastSeen         int64  `json:"last_seen"`
	BytesIn          int64  `json:"bytes_in"`
	BytesOut         int64  `json:"bytes_out"`
	ConsentGiven     bool   `json:"consent_given"`
	ConsentTimestamp int64  `json:"consent_timestamp"`
}

// resolveSession reads the catsplash_nonce cookie and looks up the client in the DB.
// Returns the client's data if found, or nil if no valid session exists.
// Identity is never accepted from URL parameters — only from the cookie.
func (s *Server) resolveSession(r *http.Request) (*state.Client, error) {
	cookie, err := r.Cookie("catsplash_nonce")
	if err != nil || cookie.Value == "" {
		return nil, nil
	}

	client, err := s.db.GetClientBySessionToken(cookie.Value)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// handleDataRequest handles GET /data-request — ARCO+ data access.
// Identity is resolved from the catsplash_nonce cookie, never from URL parameters.
func (s *Server) handleDataRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := s.resolveSession(r)
	if err != nil {
		log.Printf("DataRequest: DB error: %v", err)
		http.Error(w, "Error consultando datos", http.StatusInternalServerError)
		return
	}
	if client == nil {
		http.Error(w, "Sesión no encontrada", http.StatusNotFound)
		return
	}

	resp := dataResponse{
		MAC:              client.MAC,
		IP:               client.IP,
		State:            client.State,
		ConnectedAt:      client.ConnectedAt,
		LastSeen:         client.LastSeen,
		BytesIn:          client.BytesIn,
		BytesOut:         client.BytesOut,
		ConsentGiven:     client.ConsentGiven,
		ConsentTimestamp: client.ConsentTimestamp,
	}

	log.Printf("DataRequest: data exported for %s", maskMAC(client.MAC))

	// Audit trail — LOPDP traceability
	if logErr := s.db.LogAuditEvent(state.AuditDataAccess, maskMAC(client.MAC), getIPFromRemoteAddr(r.RemoteAddr), "json_export"); logErr != nil {
		log.Printf("Failed to log audit event: %v", logErr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("DataRequest: encode error: %v", err)
	}
}

// handleDataDeletion handles POST /data-deletion — ARCO+ data deletion (cancelación).
// Identity is resolved from the catsplash_nonce cookie, never from URL parameters.
// Deletes the client record from DB and removes firewall rules.
func (s *Server) handleDataDeletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := s.resolveSession(r)
	if err != nil {
		log.Printf("DataDeletion: DB error: %v", err)
		http.Error(w, "Error procesando solicitud", http.StatusInternalServerError)
		return
	}
	if client == nil {
		http.Error(w, "Sesión no encontrada", http.StatusNotFound)
		return
	}

	mac := client.MAC
	ip := client.IP

	// Remove firewall rules
	if err := s.fw.BlockClient(mac, ip); err != nil {
		log.Printf("DataDeletion: firewall block error for %s: %v", maskMAC(mac), err)
	}

	// Audit trail — LOPDP traceability (log BEFORE deletion, since record won't exist after)
	if logErr := s.db.LogAuditEvent(state.AuditDataDeletion, maskMAC(mac), getIPFromRemoteAddr(r.RemoteAddr), "arco_plus_cancelacion"); logErr != nil {
		log.Printf("Failed to log audit event: %v", logErr)
	}

	// Delete client record from DB
	if err := s.db.DeleteClient(mac); err != nil {
		log.Printf("DataDeletion: DB delete error for %s: %v", maskMAC(mac), err)
		http.Error(w, "Error eliminando datos", http.StatusInternalServerError)
		return
	}

	log.Printf("DataDeletion: data deleted for %s", maskMAC(mac))

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"ok":true,"message":"Datos eliminados correctamente"}`)); err != nil {
		log.Printf("DataDeletion: write response error: %v", err)
	}
}
