package server

import (
	"log"
	"net/http"
)

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/portal", http.StatusFound)
		return
	}

	// Validate Nonce
	cookie, err := r.Cookie("catsplash_nonce")
	formNonce := r.FormValue("nonce")

	if err != nil {
		log.Printf("Auth failed: missing cookie 'catsplash_nonce' for IP %s", r.RemoteAddr)
		s.renderError(w, "Sesión inválida (falta cookie). Por favor, habilita las cookies e intenta de nuevo.")
		return
	}

	if cookie.Value != formNonce {
		log.Printf("Auth failed: nonce mismatch for IP %s", r.RemoteAddr)
		s.renderError(w, "Sesión inválida (mismatch). Por favor, recarga la página e intenta de nuevo.")
		return
	}

	// Validate consent
	consent := r.FormValue("consent") == "true"
	if !consent {
		log.Printf("Auth failed: no consent for IP %s", r.RemoteAddr)
		s.renderError(w, "Debes aceptar la política de privacidad para conectarte.")
		return
	}

	ip := getIPFromRemoteAddr(r.RemoteAddr)
	mac, err := getMACFromIP(ip)
	if err != nil {
		log.Printf("Auth failed: could not resolve MAC for IP %s", ip)
		s.renderError(w, "No se pudo identificar tu dispositivo.")
		return
	}

	// Update DB and Firewall
	if err := s.db.UpsertClient(mac, ip, consent); err != nil {
		log.Printf("Auth failed: DB upsert error for %s: %v", maskMAC(mac), err)
		s.renderError(w, "Error al registrar dispositivo.")
		return
	}

	if err := s.db.Authenticate(mac, ip); err != nil {
		log.Printf("Auth failed: DB authenticate error for %s: %v", maskMAC(mac), err)
		s.renderError(w, "Error al autenticar sesión.")
		return
	}

	if err := s.fw.AllowClient(mac, ip); err != nil {
		log.Printf("Auth failed: Firewall allow error for %s: %v", maskMAC(mac), err)
		s.renderError(w, "Error al configurar el firewall.")
		return
	}

	log.Printf("Auth success for %s", maskMAC(mac))

	// Success
	data := struct {
		RedirectURL string
	}{
		RedirectURL: "http://www.google.com", // Default redirect
	}
	s.templates.ExecuteTemplate(w, "success.html", data)
}

func (s *Server) renderError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	data := struct {
		Message string
	}{
		Message: msg,
	}
	s.templates.ExecuteTemplate(w, "error.html", data)
}
