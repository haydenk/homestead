package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/haydenk/homestead/internal/checker"
	"github.com/haydenk/homestead/internal/config"
)

type pageData struct {
	Config   *config.Config
	Statuses map[string]*checker.Status
}

func newTemplate(webFS fs.FS) *template.Template {
	funcs := template.FuncMap{
		"iconHTML": func(icon string) template.HTML {
			if icon == "" {
				return "🔗"
			}
			if strings.HasPrefix(icon, "http://") || strings.HasPrefix(icon, "https://") || strings.HasPrefix(icon, "/") {
				return template.HTML(`<img src="` + template.HTMLEscapeString(icon) + `" alt="" loading="lazy" />`)
			}
			return template.HTML(template.HTMLEscapeString(icon))
		},
		"dotClass": func(s *checker.Status) string {
			if s == nil {
				return "unknown"
			}
			if !s.Up {
				return "down"
			}
			if s.ResponseTimeMs > 1000 {
				return "slow"
			}
			return "up"
		},
		"dotLabel": func(s *checker.Status) string {
			if s == nil {
				return "Checking\u2026"
			}
			if !s.Up {
				if s.Error != "" {
					return "Offline: " + s.Error
				}
				return "Offline"
			}
			return fmt.Sprintf("Online \u00b7 %dms", s.ResponseTimeMs)
		},
		"accentStyle": func(color string) template.HTML {
			if color == "" {
				return ""
			}
			return template.HTML(`style="--card-accent:` + template.HTMLEscapeString(color) + `"`)
		},
		"faviconURL": func(logo string) template.URL {
			return template.URL(fmt.Sprintf(
				"data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>%s</text></svg>",
				logo,
			))
		},
		"joinTags": func(tags []string) string { return strings.Join(tags, " ") },
		"lower":    strings.ToLower,
	}
	return template.Must(
		template.New("index.html").Funcs(funcs).ParseFS(webFS, "web/index.html"),
	)
}

// GET /
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := pageData{
		Config:   s.cfg,
		Statuses: s.checker.GetAll(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}

// GET /ws — upgrades to WebSocket and streams status updates.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	ch := make(chan []byte, 4)
	s.hub.register(ch)
	defer s.hub.unregister(ch)

	// Send the current snapshot immediately so the client doesn't wait for the next check.
	if data, err := json.Marshal(s.checker.GetAll()); err == nil {
		conn.WriteMessage(websocket.TextMessage, data) //nolint:errcheck
	}

	// Read pump — detects client disconnect.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Write pump — forwards broadcasts to this connection.
	for {
		select {
		case data := <-ch:
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// POST /api/reload — re-reads config from disk and restarts checks.
func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	newCfg, err := config.Load(s.configPath)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	s.cfg = newCfg
	s.checker.UpdateConfig(newCfg)
	s.checker.CheckNow()
	log.Printf("Config reloaded from %s", s.configPath)
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/health — liveness probe for Docker/k8s.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
