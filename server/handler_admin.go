package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AdminUser == "" || s.cfg.AdminPass == "" {
			http.Error(w, "Admin panel disabled", http.StatusNotFound)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok {
			s.requireAuth(w)
			return
		}

		userHash := sha256.Sum256([]byte(user))
		expectedUserHash := sha256.Sum256([]byte(s.cfg.AdminUser))

		userOk := subtle.ConstantTimeCompare(userHash[:], expectedUserHash[:]) == 1
		passOk := bcrypt.CompareHashAndPassword([]byte(s.cfg.AdminPass), []byte(pass)) == nil

		if !userOk || !passOk {
			s.requireAuth(w)
			return
		}

		next(w, r)
	}
}

func (s *Server) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Catsplash Admin"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	mac := r.URL.Query().Get("mac")

	if mac != "" {
		switch action {
		case "kick":
			client, err := s.db.GetClient(mac)
			if err != nil {
				http.Error(w, "Error querying client", http.StatusInternalServerError)
				return
			}
			if client == nil {
				http.Error(w, "Client not found", http.StatusNotFound)
				return
			}
			if err := s.db.Deauthenticate(mac); err != nil {
				log.Printf("Admin: deauth error for %s: %v", mac, err)
			}
			if err := s.fw.BlockClient(mac, client.IP); err != nil {
				log.Printf("Admin: block error for %s: %v", mac, err)
			}
			log.Printf("Admin kicked %s (%s)", mac, client.IP)
			http.Redirect(w, r, "/admin", http.StatusFound)
			return

		case "extend":
			minutesStr := r.URL.Query().Get("minutes")
			minutes, err := strconv.Atoi(minutesStr)
			if err != nil || minutes <= 0 {
				minutes = 30
			}
			client, err := s.db.GetClient(mac)
			if err != nil || client == nil {
				http.Error(w, "Client not found", http.StatusNotFound)
				return
			}
			newConnAt := client.ConnectedAt + int64(minutes*60)
			if _, err := s.db.Conn.Exec("UPDATE clients SET connected_at = ? WHERE mac = ?", newConnAt, mac); err != nil {
				log.Printf("Admin: extend error for %s: %v", mac, err)
			}
			log.Printf("Admin extended %s by %d min", mac, minutes)
			http.Redirect(w, r, "/admin", http.StatusFound)
			return

		case "limit":
			mbStr := r.URL.Query().Get("mb")
			mb, err := strconv.ParseInt(mbStr, 10, 64)
			if err != nil || mb < 0 {
				mb = 0
			}
			bytesLimit := mb * 1024 * 1024
			if err := s.db.UpdateMaxBytes(mac, bytesLimit); err != nil {
				log.Printf("Admin: limit error for %s: %v", mac, err)
			}
			log.Printf("Admin set limit %d MB for %s", mb, mac)
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}

	clients, err := s.db.ListAll()
	if err != nil {
		http.Error(w, "Error listing clients", http.StatusInternalServerError)
		return
	}

	activeClients := make([]*clientView, 0)
	pendingClients := make([]*clientView, 0)
	now := time.Now().Unix()

	for _, c := range clients {
		cv := &clientView{
			MAC:     c.MAC,
			IP:      c.IP,
			State:   c.State,
			DataIn:  formatBytes(c.BytesIn),
			DataOut: formatBytes(c.BytesOut),
			Total:   formatBytes(c.BytesIn + c.BytesOut),
		}

		if c.MaxBytes > 0 {
			pct := int64(0)
			if c.MaxBytes > 0 {
				pct = (c.BytesIn + c.BytesOut) * 100 / c.MaxBytes
			}
			cv.QuotaUsed = formatBytes(c.BytesIn + c.BytesOut)
			cv.QuotaLimit = formatBytes(c.MaxBytes)
			cv.QuotaPct = int(pct)
		} else {
			cv.QuotaUsed = "—"
			cv.QuotaLimit = "∞"
		}

		if c.State == "authenticated" {
			cv.ConnectedAt = time.Unix(c.ConnectedAt, 0).Format("15:04:05")

			remSession := int64(s.cfg.SessionTimeout) - (now - c.ConnectedAt)
			remIdle := int64(s.cfg.IdleTimeout) - (now - c.LastSeen)
			rem := remSession
			if s.cfg.IdleTimeout > 0 && (remIdle < rem || s.cfg.SessionTimeout <= 0) {
				rem = remIdle
			}
			if rem <= 0 {
				cv.ExpiresIn = "EXPIRED"
			} else {
				cv.ExpiresIn = (time.Duration(rem) * time.Second).Round(time.Second).String()
			}

			activeClients = append(activeClients, cv)
		} else {
			cv.ConnectedAt = "—"
			cv.ExpiresIn = "—"
			pendingClients = append(pendingClients, cv)
		}
	}

	data := struct {
		Active  []*clientView
		Pending []*clientView
	}{
		Active:  activeClients,
		Pending: pendingClients,
	}

	s.adminTmpl.Execute(w, data)
}

type clientView struct {
	MAC        string
	IP         string
	State      string
	ConnectedAt string
	ExpiresIn  string
	DataIn     string
	DataOut    string
	Total      string
	QuotaUsed  string
	QuotaLimit string
	QuotaPct   int
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return strconv.FormatInt(b, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(b)/float64(div), 'f', 2, 64) + " " + string("KMGTPE"[exp]) + "iB"
}


