package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config is the root configuration structure.
type Config struct {
	Title         string    `toml:"title"`
	Subtitle      string    `toml:"subtitle"`
	Logo          string    `toml:"logo"`
	Theme         string    `toml:"theme"`          // dark | light
	Columns       int       `toml:"columns"`        // 1–6
	CheckInterval int       `toml:"check_interval"` // seconds
	Footer        string    `toml:"footer"`
	Sections      []Section `toml:"sections"`
}

// Section groups related items on the dashboard.
type Section struct {
	Name  string `toml:"name"`
	Icon  string `toml:"icon"`
	Items []Item `toml:"items"`
}

// Item represents a single dashboard card/link.
type Item struct {
	ID          string   `toml:"-"`
	Title       string   `toml:"title"`
	URL         string   `toml:"url"`
	Description string   `toml:"description"`
	Icon        string   `toml:"icon"`
	Tags        []string `toml:"tags"`
	Target      string   `toml:"target"`       // _blank | _self
	StatusCheck bool     `toml:"status_check"`
	Color       string   `toml:"color"` // optional accent hex
}

// Load reads and parses the TOML config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	applyDefaults(&cfg)
	assignIDs(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Title == "" {
		cfg.Title = "Homestead"
	}
	if cfg.Logo == "" {
		cfg.Logo = "🏠"
	}
	if cfg.Theme == "" {
		cfg.Theme = "dark"
	}
	if cfg.Columns == 0 {
		cfg.Columns = 4
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30
	}
}

func assignIDs(cfg *Config) {
	for si := range cfg.Sections {
		for ii := range cfg.Sections[si].Items {
			item := &cfg.Sections[si].Items[ii]
			item.ID = fmt.Sprintf("s%d-i%d", si, ii)
			if item.Target == "" {
				item.Target = "_blank"
			}
		}
	}
}
