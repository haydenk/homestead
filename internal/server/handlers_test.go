package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/haydenk/homestead/internal/checker"
	"github.com/haydenk/homestead/internal/config"
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

// newIntegrationServer builds a full Server with a fake filesystem and returns
// a running httptest.Server. The caller need not close it — t.Cleanup handles it.
func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()
	webFS := fstest.MapFS{
		"web/index.html": &fstest.MapFile{Data: []byte(`<html><body>INDEX</body></html>`)},
		"web/style.css":  &fstest.MapFile{Data: []byte(`body { color: red; }`)},
		"web/app.js":     &fstest.MapFile{Data: []byte(`"use strict";`)},
	}
	cfg := &config.Config{CheckInterval: 30}
	srv := New(cfg, "", checker.New(cfg), "127.0.0.1", "0", webFS)
	ts := httptest.NewServer(srv.mux)
	t.Cleanup(ts.Close)
	return ts
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

// --- GET / and static files ---

// TestRoute_RootServesHTML verifies GET / returns the HTML index page.
func TestRoute_RootServesHTML(t *testing.T) {
	ts := newIntegrationServer(t)
	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "INDEX") {
		t.Error("response body does not contain expected index content")
	}
}

// TestRoute_StyleCSSNotInterceptedByIndex verifies that GET /style.css is served
// as a static file and not swallowed by the index handler (regression for GET /{$} fix).
func TestRoute_StyleCSSNotInterceptedByIndex(t *testing.T) {
	ts := newIntegrationServer(t)
	resp, err := ts.Client().Get(ts.URL + "/style.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q: style.css was intercepted by the index handler", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "INDEX") {
		t.Error("style.css response contains index HTML — static file routing is broken")
	}
}

// TestRoute_AppJSNotInterceptedByIndex verifies that GET /app.js is served
// as a static file and not swallowed by the index handler.
func TestRoute_AppJSNotInterceptedByIndex(t *testing.T) {
	ts := newIntegrationServer(t)
	resp, err := ts.Client().Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q: app.js was intercepted by the index handler", ct)
	}
}

// TestStyleCSS_EmptyStateHiddenOverride verifies the CSS fix that prevents
// display:flex on .empty-state from overriding the HTML hidden attribute.
func TestStyleCSS_EmptyStateHiddenOverride(t *testing.T) {
	css, err := os.ReadFile("../../web/style.css")
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	if !strings.Contains(string(css), ".empty-state[hidden]") {
		t.Error("style.css is missing .empty-state[hidden] rule — empty state will show when search is blank")
	}
}

// --- GET /ws ---

// TestWS_ReceivesSnapshotOnConnect verifies that a new WebSocket client immediately
// receives the current status snapshot without waiting for the next check round.
func TestWS_ReceivesSnapshotOnConnect(t *testing.T) {
	ts := newIntegrationServer(t)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial message: %v", err)
	}

	var statuses map[string]any
	if err := json.Unmarshal(msg, &statuses); err != nil {
		t.Fatalf("unmarshal: %v — body: %s", err, msg)
	}
}

// TestWS_ReceivesBroadcastAfterCheck verifies that connected clients receive a push
// after a check round completes.
func TestWS_ReceivesBroadcastAfterCheck(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	webFS := fstest.MapFS{
		"web/index.html": &fstest.MapFile{Data: []byte(`<html><body>INDEX</body></html>`)},
		"web/style.css":  &fstest.MapFile{Data: []byte(`body{}`)},
		"web/app.js":     &fstest.MapFile{Data: []byte(`"use strict";`)},
	}
	cfg := &config.Config{
		CheckInterval: 60,
		Sections: []config.Section{
			{Items: []config.Item{{ID: "s0-i0", URL: backend.URL, StatusCheck: true}}},
		},
	}
	chk := checker.New(cfg)
	srv := New(cfg, "", chk, "127.0.0.1", "0", webFS)
	ts := httptest.NewServer(srv.mux)
	t.Cleanup(ts.Close)

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	// Drain the initial snapshot.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}

	// Trigger a check and expect a broadcast.
	chk.CheckNow()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read broadcast after check: %v", err)
	}

	var statuses map[string]any
	if err := json.Unmarshal(msg, &statuses); err != nil {
		t.Fatalf("unmarshal broadcast: %v — body: %s", err, msg)
	}
	if _, ok := statuses["s0-i0"]; !ok {
		t.Error("broadcast did not include s0-i0 status")
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
