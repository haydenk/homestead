package checker

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"homestead/internal/config"
)

// Status holds the result of a single health check.
type Status struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Up             bool   `json:"up"`
	StatusCode     int    `json:"statusCode"`
	ResponseTimeMs int64  `json:"responseTimeMs"`
	LastChecked    string `json:"lastChecked"`
	Error          string `json:"error,omitempty"`
}

// Checker runs background health checks for all configured items.
type Checker struct {
	cfg     *config.Config
	mu      sync.RWMutex
	results map[string]*Status
	stopCh  chan struct{}
	client  *http.Client
}

// New creates a Checker. Call Start() to begin background checks.
func New(cfg *config.Config) *Checker {
	return &Checker{
		cfg:     cfg,
		results: make(map[string]*Status),
		stopCh:  make(chan struct{}),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				DisableKeepAlives:   true,
				MaxIdleConnsPerHost: 1,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// Start runs an initial check immediately then re-checks on the configured interval.
func (c *Checker) Start() {
	go c.checkAll() // non-blocking initial check

	interval := time.Duration(c.cfg.CheckInterval) * time.Second
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.checkAll()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop halts background checks.
func (c *Checker) Stop() {
	close(c.stopCh)
}

// CheckNow triggers an immediate check of all items (non-blocking).
func (c *Checker) CheckNow() {
	go c.checkAll()
}

// UpdateConfig replaces the current config (used after a live reload).
func (c *Checker) UpdateConfig(cfg *config.Config) {
	c.mu.Lock()
	c.cfg = cfg
	// Clear stale results for items that no longer exist.
	valid := make(map[string]bool)
	for _, sec := range cfg.Sections {
		for _, item := range sec.Items {
			valid[item.ID] = true
		}
	}
	for id := range c.results {
		if !valid[id] {
			delete(c.results, id)
		}
	}
	c.mu.Unlock()
}

// GetAll returns a snapshot of current statuses keyed by item ID.
func (c *Checker) GetAll() map[string]*Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*Status, len(c.results))
	for k, v := range c.results {
		cp := *v
		out[k] = &cp
	}
	return out
}

// Get returns the status for a single item by ID, or nil if not found.
func (c *Checker) Get(id string) *Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.results[id]; ok {
		cp := *s
		return &cp
	}
	return nil
}

func (c *Checker) checkAll() {
	var wg sync.WaitGroup
	for _, section := range c.cfg.Sections {
		for _, item := range section.Items {
			if !item.StatusCheck || item.URL == "" {
				continue
			}
			wg.Add(1)
			go func(item config.Item) {
				defer wg.Done()
				s := c.probe(item)
				c.mu.Lock()
				c.results[item.ID] = s
				c.mu.Unlock()
			}(item)
		}
	}
	wg.Wait()
}

func (c *Checker) probe(item config.Item) *Status {
	start := time.Now()
	s := &Status{
		ID:          item.ID,
		URL:         item.URL,
		LastChecked: start.UTC().Format(time.RFC3339),
	}

	// Try HEAD first; fall back to GET (some servers reject HEAD).
	resp, err := c.doRequest(http.MethodHead, item.URL)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed {
		if resp != nil {
			resp.Body.Close()
		}
		resp, err = c.doRequest(http.MethodGet, item.URL)
	}

	s.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		s.Error = fmt.Sprintf("%v", err)
		return s
	}
	defer resp.Body.Close()

	s.StatusCode = resp.StatusCode
	// Anything below 500 is considered "up" — even auth-protected pages.
	s.Up = resp.StatusCode < 500
	return s
}

func (c *Checker) doRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Homestead-StatusChecker/1.0")
	return c.client.Do(req)
}
