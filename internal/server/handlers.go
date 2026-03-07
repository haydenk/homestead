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

	"homestead/internal/checker"
	"homestead/internal/config"
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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// GET /api/status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, s.checker.GetAll())
}

// GET /api/status/{id}
func (s *Server) handleStatusByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	status := s.checker.Get(id)
	if status == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, status)
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
