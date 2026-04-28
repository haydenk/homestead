package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/haydenk/homestead/internal/checker"
	"github.com/haydenk/homestead/internal/config"
)

// Server holds all runtime state for the HTTP server.
type Server struct {
	// cfgMu guards cfg so that concurrent request handlers (handleIndex
	// reads s.cfg, handleReload swaps it) don't race. Before this, go
	// test -race flagged a data race between GET / and POST /api/reload.
	cfgMu      sync.RWMutex
	cfg        *config.Config
	configPath string
	checker    *checker.Checker
	host       string
	port       string
	webFS      fs.FS
	tmpl       *template.Template
	mux        *http.ServeMux
	hub        *hub
	upgrader   websocket.Upgrader
}

// getCfg returns the current config under the server's read lock.
func (s *Server) getCfg() *config.Config {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.cfg
}

// setCfg swaps the active config under the server's write lock.
func (s *Server) setCfg(cfg *config.Config) {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	s.cfg = cfg
}

// New creates a configured Server.
func New(cfg *config.Config, configPath string, chk *checker.Checker, host, port string, webFS fs.FS) *Server {
	s := &Server{
		cfg:        cfg,
		configPath: configPath,
		checker:    chk,
		host:       host,
		port:       port,
		webFS:      webFS,
		tmpl:       newTemplate(webFS),
		mux:        http.NewServeMux(),
		hub:        newHub(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	s.registerRoutes()
	go s.runHub()
	return s
}

// runHub forwards checker notifications to all connected WebSocket clients.
func (s *Server) runHub() {
	for range s.checker.Notify() {
		data, err := json.Marshal(s.checker.GetAll())
		if err != nil {
			log.Printf("ws marshal: %v", err)
			continue
		}
		s.hub.broadcast(data)
	}
}

// Run starts listening. Blocks until the server stops.
func (s *Server) Run() error {
	srv := &http.Server{
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if s.host != "" {
		addr := net.JoinHostPort(s.host, s.port)
		srv.Addr = addr
		log.Printf("Homestead listening on http://%s", addr)
		return srv.ListenAndServe()
	}

	// No host specified — listen on all IPv4 and IPv6 interfaces.
	var listeners []net.Listener
	for _, network := range []string{"tcp4", "tcp6"} {
		ln, err := net.Listen(network, ":"+s.port)
		if err != nil {
			log.Printf("skipping %s: %v", network, err)
			continue
		}
		listeners = append(listeners, ln)
	}
	if len(listeners) == 0 {
		return fmt.Errorf("could not bind to any address on port %s", s.port)
	}

	log.Printf("Homestead listening on 0.0.0.0:%s and [::]:%s", s.port, s.port)
	errc := make(chan error, len(listeners))
	for _, ln := range listeners {
		go func(l net.Listener) { errc <- srv.Serve(l) }(ln)
	}
	return <-errc
}

func (s *Server) registerRoutes() {
	// --- API ---
	s.mux.HandleFunc("POST /api/reload", s.handleReload)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	// --- WebSocket ---
	s.mux.HandleFunc("GET /ws", s.handleWS)

	// --- Dashboard (server-side rendered) ---
	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	// --- Static files (embedded) ---
	sub, err := fs.Sub(s.webFS, "web")
	if err != nil {
		log.Fatalf("web embed sub-FS: %v", err)
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
}
