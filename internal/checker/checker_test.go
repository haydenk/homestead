package checker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/haydenk/homestead/internal/config"
)

// makeConfig builds a minimal Config containing the given items in a single section.
func makeConfig(items ...config.Item) *config.Config {
	return &config.Config{
		CheckInterval: 60,
		Sections:      []config.Section{{Items: items}},
	}
}

// --- New ---

func TestNew(t *testing.T) {
	cfg := makeConfig()
	c := New(cfg)

	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.cfg != cfg {
		t.Error("cfg not stored correctly")
	}
	if c.results == nil {
		t.Error("results map is nil")
	}
	if c.client == nil {
		t.Error("http client is nil")
	}
}

// --- probe ---

func TestProbe_Up200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	s := c.probe(item)

	if !s.Up {
		t.Errorf("Up = false, want true (200 OK)")
	}
	if s.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", s.StatusCode)
	}
	if s.Error != "" {
		t.Errorf("Error = %q, want empty", s.Error)
	}
	if s.ResponseTimeMs < 0 {
		t.Error("ResponseTimeMs should be >= 0")
	}
	if s.LastChecked == "" {
		t.Error("LastChecked should not be empty")
	}
}

func TestProbe_Down500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	s := c.probe(item)

	if s.Up {
		t.Errorf("Up = true, want false (500 should be down)")
	}
	if s.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", s.StatusCode)
	}
}

func TestProbe_AuthProtected401_IsUp(t *testing.T) {
	// Auth-protected pages return 401 but the service is running.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	s := c.probe(item)

	if !s.Up {
		t.Errorf("Up = false, want true (401 < 500 counts as up)")
	}
}

func TestProbe_HeadNotAllowed_FallsBackToGet(t *testing.T) {
	var getCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		getCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	s := c.probe(item)

	if !s.Up {
		t.Errorf("Up = false after GET fallback, want true")
	}
	if getCount != 1 {
		t.Errorf("GET called %d time(s), want 1", getCount)
	}
}

func TestProbe_SetsUserAgentHeader(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	c.probe(item)

	if gotUA != "Homestead-StatusChecker/1.0" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "Homestead-StatusChecker/1.0")
	}
}

func TestProbe_ConnectionError(t *testing.T) {
	c := New(makeConfig())
	// Port 1 is not bindable by user processes; connection should be refused.
	item := config.Item{ID: "s0-i0", URL: "http://127.0.0.1:1", StatusCheck: true}
	s := c.probe(item)

	if s.Up {
		t.Error("Up = true for unreachable host, want false")
	}
	if s.Error == "" {
		t.Error("Error should be set for connection error")
	}
}

func TestProbe_InvalidURL(t *testing.T) {
	c := New(makeConfig())
	item := config.Item{ID: "s0-i0", URL: "://invalid-url", StatusCheck: true}
	s := c.probe(item)

	if s.Up {
		t.Error("Up = true for invalid URL, want false")
	}
	if s.Error == "" {
		t.Error("Error should be set for invalid URL")
	}
}

func TestProbe_SetsIDAndURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(makeConfig())
	item := config.Item{ID: "s1-i2", URL: ts.URL, StatusCheck: true}
	s := c.probe(item)

	if s.ID != "s1-i2" {
		t.Errorf("ID = %q, want %q", s.ID, "s1-i2")
	}
	if s.URL != ts.URL {
		t.Errorf("URL = %q, want %q", s.URL, ts.URL)
	}
}

// --- GetAll / Get ---

func TestGetAll_Empty(t *testing.T) {
	c := New(makeConfig())
	all := c.GetAll()
	if len(all) != 0 {
		t.Errorf("GetAll on empty checker returned %d entries, want 0", len(all))
	}
}

func TestGet_NotFound(t *testing.T) {
	c := New(makeConfig())
	if s := c.Get("nonexistent"); s != nil {
		t.Errorf("Get unknown ID = %v, want nil", s)
	}
}

func TestGetAll_ReturnsCopy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true})
	c := New(cfg)
	c.checkAll()

	all := c.GetAll()
	if len(all) != 1 {
		t.Fatalf("GetAll len = %d, want 1", len(all))
	}

	// Mutate the returned copy.
	snapshot := all["s0-i0"]
	originalUp := snapshot.Up
	snapshot.Up = !originalUp

	// Original should be unaffected.
	all2 := c.GetAll()
	if all2["s0-i0"].Up != originalUp {
		t.Error("GetAll did not return a copy; mutation affected the stored value")
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true})
	c := New(cfg)
	c.checkAll()

	s := c.Get("s0-i0")
	if s == nil {
		t.Fatal("Get returned nil after checkAll")
	}

	originalUp := s.Up
	s.Up = !originalUp // mutate the returned copy

	s2 := c.Get("s0-i0")
	if s2.Up != originalUp {
		t.Error("Get did not return a copy; mutation affected the stored value")
	}
}

// --- UpdateConfig ---

func TestUpdateConfig_RemovesStaleResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	item1 := config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true}
	item2 := config.Item{ID: "s0-i1", URL: ts.URL, StatusCheck: true}
	cfg := makeConfig(item1, item2)
	c := New(cfg)
	c.checkAll()

	if got := len(c.GetAll()); got != 2 {
		t.Fatalf("expected 2 results after checkAll, got %d", got)
	}

	// New config only contains item1; item2 should be removed.
	newCfg := makeConfig(item1)
	c.UpdateConfig(newCfg)

	all := c.GetAll()
	if _, ok := all["s0-i0"]; !ok {
		t.Error("s0-i0 should still exist after UpdateConfig")
	}
	if _, ok := all["s0-i1"]; ok {
		t.Error("s0-i1 should be removed after UpdateConfig")
	}
}

func TestUpdateConfig_ReplacesConfig(t *testing.T) {
	cfg := makeConfig()
	c := New(cfg)

	newCfg := &config.Config{CheckInterval: 10}
	c.UpdateConfig(newCfg)

	if c.cfg != newCfg {
		t.Error("cfg was not replaced by UpdateConfig")
	}
}

// --- checkAll ---

func TestCheckAll_OnlyChecksStatusCheckItems(t *testing.T) {
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(
		config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: false}, // should be skipped
		config.Item{ID: "s0-i1", URL: ts.URL, StatusCheck: true},  // should be checked
	)
	c := New(cfg)
	c.checkAll()

	// Only one item should produce a result.
	all := c.GetAll()
	if _, ok := all["s0-i0"]; ok {
		t.Error("s0-i0 (StatusCheck=false) should not have a result")
	}
	if _, ok := all["s0-i1"]; !ok {
		t.Error("s0-i1 (StatusCheck=true) should have a result")
	}
}

func TestCheckAll_SkipsEmptyURL(t *testing.T) {
	cfg := makeConfig(
		config.Item{ID: "s0-i0", URL: "", StatusCheck: true},
	)
	c := New(cfg)
	c.checkAll() // must not panic or attempt a request

	if got := len(c.GetAll()); got != 0 {
		t.Errorf("expected 0 results for empty URL, got %d", got)
	}
}

func TestCheckAll_MultipleSections(t *testing.T) {
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
			{Items: []config.Item{
				{ID: "s1-i0", URL: ts.URL, StatusCheck: true},
			}},
		},
	}
	c := New(cfg)
	c.checkAll()

	all := c.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 results across sections, got %d", len(all))
	}
}

// --- Notify ---

func TestNotify_FiresAfterCheckAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true})
	c := New(cfg)

	notify := c.Notify()
	c.checkAll()

	select {
	case <-notify:
		// passed
	case <-time.After(500 * time.Millisecond):
		t.Error("Notify channel did not fire after checkAll")
	}
}

func TestNotify_DoesNotBlockWhenUnread(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(config.Item{ID: "s0-i0", URL: ts.URL, StatusCheck: true})
	c := New(cfg)

	// Call checkAll twice without reading from Notify — must not block.
	done := make(chan struct{})
	go func() {
		c.checkAll()
		c.checkAll()
		close(done)
	}()

	select {
	case <-done:
		// passed
	case <-time.After(2 * time.Second):
		t.Error("checkAll blocked on unread Notify channel")
	}
}

// --- Start / Stop ---

func TestStartStop_DoesNotPanic(t *testing.T) {
	cfg := &config.Config{CheckInterval: 1}
	c := New(cfg)
	c.Start()
	time.Sleep(10 * time.Millisecond)
	c.Stop() // must not hang or panic
}

// --- CheckNow ---

func TestCheckNow_IsNonBlocking(t *testing.T) {
	cfg := makeConfig()
	c := New(cfg)

	done := make(chan struct{})
	go func() {
		c.CheckNow()
		close(done)
	}()

	select {
	case <-done:
		// passed
	case <-time.After(100 * time.Millisecond):
		t.Error("CheckNow blocked for too long")
	}
}
