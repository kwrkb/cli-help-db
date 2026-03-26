package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LineLimit != 60 {
		t.Errorf("LineLimit = %d, want 60", cfg.LineLimit)
	}
	if cfg.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want 3s", cfg.Timeout)
	}
	if cfg.Parallelism != 8 {
		t.Errorf("Parallelism = %d, want 8", cfg.Parallelism)
	}
}

func TestLoadFrom_NotExist(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LineLimit != 60 {
		t.Errorf("expected default LineLimit, got %d", cfg.LineLimit)
	}
}

func TestLoadFrom_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	data := []byte(`commands:
  - docker
  - kubectl
output_dir: /tmp/test-help
line_limit: 100
timeout: 5s
parallelism: 4
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Commands) != 2 {
		t.Errorf("Commands len = %d, want 2", len(cfg.Commands))
	}
	if cfg.OutputDir != "/tmp/test-help" {
		t.Errorf("OutputDir = %q, want /tmp/test-help", cfg.OutputDir)
	}
	if cfg.LineLimit != 100 {
		t.Errorf("LineLimit = %d, want 100", cfg.LineLimit)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", cfg.Timeout)
	}
	if cfg.Parallelism != 4 {
		t.Errorf("Parallelism = %d, want 4", cfg.Parallelism)
	}
}

func TestLoadFrom_MinimalYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	data := []byte("commands:\n  - jq\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Commands) != 1 || cfg.Commands[0] != "jq" {
		t.Errorf("Commands = %v, want [jq]", cfg.Commands)
	}
	// Defaults should be applied for omitted fields
	if cfg.LineLimit != 60 {
		t.Errorf("LineLimit = %d, want default 60", cfg.LineLimit)
	}
	if cfg.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want default 3s", cfg.Timeout)
	}
}

func TestExpandTilde(t *testing.T) {
	home := homeDir()
	got := expandTilde("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Errorf("expandTilde(~/foo/bar) = %q, want %q", got, want)
	}

	// No tilde — unchanged
	got = expandTilde("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("expandTilde(/absolute/path) = %q, want /absolute/path", got)
	}
}
