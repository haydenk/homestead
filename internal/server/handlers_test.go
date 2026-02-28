package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"homestead/internal/checker"
	"homestead/internal/config"
)

// newTestServer builds a minimal Server suitable for handler testing,
// bypassing embed.FS and full route registration.
func newTestServer(cfg *config.Config, configPath string) *Server {
	return &Server{
		cfg:        cfg,
		configPath: configPath,
		checker:    checker.New(cfg),
	}
}

// --- GET /api/health ---

func TestHandleHealth_Returns200WithStatusOK(t *testing.T) {
	s := newTestServer(&config.Config{}, "")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf(`status field = %v, want "ok"`, body["status"])
	}
	if body["time"] == nil {
		t.Error("time field missing from health response")
	}
}

func TestHandleHealth_TimeIsRFC3339(t *testing.T) {
	s := newTestServer(&config.Config{}, "")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	timeStr, ok := body["time"].(string)
	if !ok {
		t.Fatal("time field is not a string")
	}
	if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
		t.Errorf("time %q is not valid RFC3339: %v", timeStr, err)
	}
}

// --- GET /api/config ---

func TestHandleConfig_ReturnsConfigAsJSON(t *testing.T) {
	cfg := &config.Config{
		Title:   "Test Dashboard",
		Theme:   "dark",
		Columns: 4,
	}
	s := newTestServer(cfg, "")
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	s.handleConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}

	var body config.Config
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Title != "Test Dashboard" {
		t.Errorf("Title = %q, want %q", body.Title, "Test Dashboard")
	}
	if body.Columns != 4 {
		t.Errorf("Columns = %d, want 4", body.Columns)
	}
}

func TestHandleConfig_IncludesSections(t *testing.T) {
	cfg := &config.Config{
		Sections: []config.Section{
			{Name: "Services", Items: []config.Item{{Title: "Router", URL: "http://192.168.1.1"}}},
		},
	}
	s := newTestServer(cfg, "")
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	s.handleConfig(w, req)

	var body config.Config
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Sections) != 1 {
		t.Fatalf("Sections len = %d, want 1", len(body.Sections))
	}
	if body.Sections[0].Name != "Services" {
		t.Errorf("Section name = %q, want %q", body.Sections[0].Name, "Services")
	}
}

// --- GET /api/status ---

func TestHandleStatus_EmptyMap(t *testing.T) {
	s := newTestServer(&config.Config{}, "")
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	s.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty status map, got %d entries", len(body))
	}
}

func TestHandleStatus_ReturnsCheckerResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := &config.Config{
		CheckInterval: 60,
		Sections: []config.Section{
			{Items: []config.Item{
				{ID: "s0-i0", URL: ts.URL, StatusCheck: true},
			}},
		},
	}
	chk := checker.New(cfg)
	chk.CheckNow()
	time.Sleep(150 * time.Millisecond) // wait for async probe

	s := &Server{cfg: cfg, checker: chk}
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	s.handleStatus(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := body["s0-i0"]; !ok {
		t.Error("expected s0-i0 in status response")
	}
}

// --- GET /api/status/{id} ---

func TestHandleStatusByID_NotFound(t *testing.T) {
	s := newTestServer(&config.Config{}, "")
	req := httptest.NewRequest(http.MethodGet, "/api/status/s0-i0", nil)
	req.SetPathValue("id", "s0-i0")
	w := httptest.NewRecorder()

	s.handleStatusByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleStatusByID_Found(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := &config.Config{
		CheckInterval: 60,
		Sections: []config.Section{
			{Items: []config.Item{
				{ID: "s0-i0", URL: ts.URL, StatusCheck: true},
			}},
		},
	}
	chk := checker.New(cfg)
	chk.CheckNow()
	time.Sleep(150 * time.Millisecond) // wait for async probe

	s := &Server{cfg: cfg, checker: chk}
	req := httptest.NewRequest(http.MethodGet, "/api/status/s0-i0", nil)
	req.SetPathValue("id", "s0-i0")
	w := httptest.NewRecorder()

	s.handleStatusByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body checker.Status
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.ID != "s0-i0" {
		t.Errorf("ID = %q, want %q", body.ID, "s0-i0")
	}
}

// --- POST /api/reload ---

func TestHandleReload_ValidConfig_Returns204(t *testing.T) {
	f, err := os.CreateTemp("", "homestead-reload-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(`title = "Reloaded"`)
	f.Close()

	s := newTestServer(&config.Config{Title: "Original"}, f.Name())
	req := httptest.NewRequest(http.MethodPost, "/api/reload", nil)
	w := httptest.NewRecorder()

	s.handleReload(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if s.cfg.Title != "Reloaded" {
		t.Errorf("cfg.Title = %q, want %q", s.cfg.Title, "Reloaded")
	}
}

func TestHandleReload_FileNotFound_Returns500(t *testing.T) {
	s := newTestServer(&config.Config{}, "/nonexistent/config.toml")
	req := httptest.NewRequest(http.MethodPost, "/api/reload", nil)
	w := httptest.NewRecorder()

	s.handleReload(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestHandleReload_InvalidTOML_Returns500(t *testing.T) {
	f, err := os.CreateTemp("", "homestead-bad-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString("this is [[ invalid toml !!!")
	f.Close()

	s := newTestServer(&config.Config{}, f.Name())
	req := httptest.NewRequest(http.MethodPost, "/api/reload", nil)
	w := httptest.NewRecorder()

	s.handleReload(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
