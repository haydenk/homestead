package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"homestead/internal/config"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// GET /api/config
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, s.cfg)
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
