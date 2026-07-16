package server

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

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
	adminTmpl *template.Template
	rl        *RateLimiter
}

func New(cfg *config.Config, db *state.DB, fw *firewall.Firewall) *Server {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/portal.html", "templates/success.html", "templates/error.html"))
	adminTmpl := template.Must(template.ParseFS(templateFS, "templates/admin.html"))
	return &Server{
		cfg:       cfg,
		db:        db,
		fw:        fw,
		templates: tmpl,
		adminTmpl: adminTmpl,
		rl:        NewRateLimiter(5, 60*time.Second),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("/admin", s.basicAuth(s.handleAdmin))
	mux.HandleFunc("/portal", s.handlePortal)
	mux.Handle("/auth", s.rl.Middleware(http.HandlerFunc(s.handleAuth)))
	mux.HandleFunc("/", s.handleRedirect) // Catch-all for intercepcion

	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.PortalPort), s.logMiddleware(mux))
}
