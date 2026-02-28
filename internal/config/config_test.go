package config

import (
	"os"
	"testing"
)

// writeTempConfig writes content to a temporary TOML file and returns its path.
// The caller is responsible for removing the file.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "homestead-config-*.toml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
title    = "My Home"
subtitle = "A subtitle"
theme    = "light"
columns  = 3

[[sections]]
name = "Services"
icon = "🔧"

[[sections.items]]
title        = "Router"
url          = "http://192.168.1.1"
description  = "Home router"
status_check = true
tags         = ["network"]
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	if cfg.Title != "My Home" {
		t.Errorf("Title = %q, want %q", cfg.Title, "My Home")
	}
	if cfg.Subtitle != "A subtitle" {
		t.Errorf("Subtitle = %q, want %q", cfg.Subtitle, "A subtitle")
	}
	if cfg.Theme != "light" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "light")
	}
	if cfg.Columns != 3 {
		t.Errorf("Columns = %d, want 3", cfg.Columns)
	}
	if len(cfg.Sections) != 1 {
		t.Fatalf("Sections len = %d, want 1", len(cfg.Sections))
	}
	if cfg.Sections[0].Name != "Services" {
		t.Errorf("Section.Name = %q, want %q", cfg.Sections[0].Name, "Services")
	}
	if len(cfg.Sections[0].Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(cfg.Sections[0].Items))
	}

	item := cfg.Sections[0].Items[0]
	if item.Title != "Router" {
		t.Errorf("Item.Title = %q, want %q", item.Title, "Router")
	}
	if item.URL != "http://192.168.1.1" {
		t.Errorf("Item.URL = %q, want %q", item.URL, "http://192.168.1.1")
	}
	if !item.StatusCheck {
		t.Error("Item.StatusCheck = false, want true")
	}
	if len(item.Tags) != 1 || item.Tags[0] != "network" {
		t.Errorf("Item.Tags = %v, want [network]", item.Tags)
	}
}

func TestLoad_IDsAndTargetAssigned(t *testing.T) {
	path := writeTempConfig(t, `
[[sections]]
name = "A"
[[sections.items]]
title = "First"
[[sections.items]]
title = "Second"
target = "_self"

[[sections]]
name = "B"
[[sections.items]]
title = "Third"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	tests := []struct {
		si, ii int
		wantID string
		wantTarget string
	}{
		{0, 0, "s0-i0", "_blank"},
		{0, 1, "s0-i1", "_self"},
		{1, 0, "s1-i0", "_blank"},
	}
	for _, tt := range tests {
		item := cfg.Sections[tt.si].Items[tt.ii]
		if item.ID != tt.wantID {
			t.Errorf("[%d][%d] ID = %q, want %q", tt.si, tt.ii, item.ID, tt.wantID)
		}
		if item.Target != tt.wantTarget {
			t.Errorf("[%d][%d] Target = %q, want %q", tt.si, tt.ii, item.Target, tt.wantTarget)
		}
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	path := writeTempConfig(t, "this is [[ not valid toml !!!")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestLoad_EmptyFile_AppliesDefaults(t *testing.T) {
	path := writeTempConfig(t, "")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Title != "Homestead" {
		t.Errorf("Title = %q, want %q", cfg.Title, "Homestead")
	}
	if cfg.Logo != "🏠" {
		t.Errorf("Logo = %q, want %q", cfg.Logo, "🏠")
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "dark")
	}
	if cfg.Columns != 4 {
		t.Errorf("Columns = %d, want 4", cfg.Columns)
	}
	if cfg.CheckInterval != 30 {
		t.Errorf("CheckInterval = %d, want 30", cfg.CheckInterval)
	}
}

// --- applyDefaults ---

func TestApplyDefaults_EmptyConfig(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	if cfg.Title != "Homestead" {
		t.Errorf("Title = %q, want %q", cfg.Title, "Homestead")
	}
	if cfg.Logo != "🏠" {
		t.Errorf("Logo = %q, want %q", cfg.Logo, "🏠")
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "dark")
	}
	if cfg.Columns != 4 {
		t.Errorf("Columns = %d, want 4", cfg.Columns)
	}
	if cfg.CheckInterval != 30 {
		t.Errorf("CheckInterval = %d, want 30", cfg.CheckInterval)
	}
}

func TestApplyDefaults_ExistingValuesPreserved(t *testing.T) {
	cfg := &Config{
		Title:         "Custom",
		Logo:          "🌟",
		Theme:         "light",
		Columns:       2,
		CheckInterval: 60,
	}
	applyDefaults(cfg)

	if cfg.Title != "Custom" {
		t.Errorf("Title = %q, want %q", cfg.Title, "Custom")
	}
	if cfg.Logo != "🌟" {
		t.Errorf("Logo = %q, want %q", cfg.Logo, "🌟")
	}
	if cfg.Theme != "light" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "light")
	}
	if cfg.Columns != 2 {
		t.Errorf("Columns = %d, want 2", cfg.Columns)
	}
	if cfg.CheckInterval != 60 {
		t.Errorf("CheckInterval = %d, want 60", cfg.CheckInterval)
	}
}

// --- assignIDs ---

func TestAssignIDs_Format(t *testing.T) {
	cfg := &Config{
		Sections: []Section{
			{Items: []Item{{Title: "A"}, {Title: "B"}}},
			{Items: []Item{{Title: "C"}}},
		},
	}
	assignIDs(cfg)

	want := [][]string{
		{"s0-i0", "s0-i1"},
		{"s1-i0"},
	}
	for si, sec := range cfg.Sections {
		for ii, item := range sec.Items {
			if item.ID != want[si][ii] {
				t.Errorf("Section %d Item %d: ID = %q, want %q", si, ii, item.ID, want[si][ii])
			}
		}
	}
}

func TestAssignIDs_DefaultTarget(t *testing.T) {
	cfg := &Config{
		Sections: []Section{
			{Items: []Item{
				{Title: "No target"},
				{Title: "Has target", Target: "_self"},
			}},
		},
	}
	assignIDs(cfg)

	if got := cfg.Sections[0].Items[0].Target; got != "_blank" {
		t.Errorf("empty target: Target = %q, want _blank", got)
	}
	if got := cfg.Sections[0].Items[1].Target; got != "_self" {
		t.Errorf("explicit target: Target = %q, want _self", got)
	}
}

func TestAssignIDs_NoSections(t *testing.T) {
	cfg := &Config{}
	// Should not panic
	assignIDs(cfg)
}
