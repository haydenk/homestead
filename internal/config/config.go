package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config is the root configuration structure.
type Config struct {
	Title         string    `toml:"title"          json:"title"`
	Subtitle      string    `toml:"subtitle"       json:"subtitle"`
	Logo          string    `toml:"logo"           json:"logo"`
	Theme         string    `toml:"theme"          json:"theme"`         // dark | light
	Columns       int       `toml:"columns"        json:"columns"`       // 1–6
	CheckInterval int       `toml:"check_interval" json:"checkInterval"` // seconds
	Footer        string    `toml:"footer"         json:"footer"`
	Sections      []Section `toml:"sections"       json:"sections"`
}

// Section groups related items on the dashboard.
type Section struct {
	Name  string `toml:"name"  json:"name"`
	Icon  string `toml:"icon"  json:"icon"`
	Items []Item `toml:"items" json:"items"`
}

// Item represents a single dashboard card/link.
type Item struct {
	ID          string   `toml:"-"            json:"id"`
	Title       string   `toml:"title"        json:"title"`
	URL         string   `toml:"url"          json:"url"`
	Description string   `toml:"description"  json:"description"`
	Icon        string   `toml:"icon"         json:"icon"`
	Tags        []string `toml:"tags"         json:"tags"`
	Target      string   `toml:"target"       json:"target"` // _blank | _self
	StatusCheck bool     `toml:"status_check" json:"statusCheck"`
	Color       string   `toml:"color"        json:"color"` // optional accent hex
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
