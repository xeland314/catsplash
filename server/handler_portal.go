package server

import (
	"net/http"
)

func (s *Server) handlePortal(w http.ResponseWriter, r *http.Request) {
	var nonce string
	
	// Try to reuse existing valid nonce from cookie
	if cookie, err := r.Cookie("catsplash_nonce"); err == nil && cookie.Value != "" {
		nonce = cookie.Value
	} else {
		nonce = generateNonce()
		// Set nonce in cookie for validation
		http.SetCookie(w, &http.Cookie{
			Name:     "catsplash_nonce",
			Value:    nonce,
			Path:     "/",
			HttpOnly: true,
			// SameSite Lax helps with some mobile browsers
			SameSite: http.SameSiteLaxMode,
		})
	}

	data := struct {
		Nonce string
	}{
		Nonce: nonce,
	}

	if err := s.templates.ExecuteTemplate(w, "portal.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
