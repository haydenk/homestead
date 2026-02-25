package server

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"

	"homestead/internal/checker"
	"homestead/internal/config"
)

// Server holds all runtime state for the HTTP server.
type Server struct {
	cfg        *config.Config
	configPath string
	checker    *checker.Checker
	host       string
	port       string
	webFS      embed.FS
	mux        *http.ServeMux
}

// New creates a configured Server.
func New(cfg *config.Config, configPath string, chk *checker.Checker, host, port string, webFS embed.FS) *Server {
	s := &Server{
		cfg:        cfg,
		configPath: configPath,
		checker:    chk,
		host:       host,
		port:       port,
		webFS:      webFS,
		mux:        http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Run starts listening. Blocks until the server stops.
func (s *Server) Run() error {
	// net.JoinHostPort brackets IPv6 addresses correctly: [::1]:8080
	addr := net.JoinHostPort(s.host, s.port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("Homestead listening on http://%s", addr)
	return srv.ListenAndServe()
}

func (s *Server) registerRoutes() {
	// --- API ---
	s.mux.HandleFunc("GET /api/config", s.handleConfig)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/status/{id}", s.handleStatusByID)
	s.mux.HandleFunc("POST /api/reload", s.handleReload)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	// --- Static files (embedded) ---
	sub, err := fs.Sub(s.webFS, "web")
	if err != nil {
		log.Fatalf("web embed sub-FS: %v", err)
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
}
