package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Commands    []string      `yaml:"commands"`
	OutputDir   string        `yaml:"output_dir"`
	LineLimit   int           `yaml:"line_limit"`
	Timeout     time.Duration `yaml:"timeout"`
	Parallelism int           `yaml:"parallelism"`
}

func DefaultConfig() *Config {
	return &Config{
		OutputDir:   filepath.Join(homeDir(), ".claude", "cli-help"),
		LineLimit:   60,
		Timeout:     3 * time.Second,
		Parallelism: 8,
	}
}

// Load reads the config file from the default location.
// If the file does not exist, it returns the default config.
func Load() (*Config, error) {
	return LoadFrom(defaultConfigPath())
}

// LoadFrom reads a config from the given path.
// If the file does not exist, it returns the default config.
func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(expandTilde(path))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for zero values
	if cfg.LineLimit == 0 {
		cfg.LineLimit = 60
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.Parallelism == 0 {
		cfg.Parallelism = 8
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = filepath.Join(homeDir(), ".claude", "cli-help")
	}

	cfg.OutputDir = expandTilde(cfg.OutputDir)
	return cfg, nil
}

func defaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "cli-help-db", "config.yaml")
	}
	return filepath.Join(homeDir(), ".config", "cli-help-db", "config.yaml")
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), path[2:])
	}
	return path
}
