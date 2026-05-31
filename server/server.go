package server

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	cfg       *config.Config
	db        *state.DB
	fw        *firewall.Firewall
	templates *template.Template
}

func New(cfg *config.Config, db *state.DB, fw *firewall.Firewall) *Server {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	return &Server{
		cfg:       cfg,
		db:        db,
		fw:        fw,
		templates: tmpl,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("/portal", s.handlePortal)
	mux.HandleFunc("/auth", s.handleAuth)
	mux.HandleFunc("/", s.handleRedirect) // Catch-all for intercepcion

	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.PortalPort), s.logMiddleware(mux))
}
