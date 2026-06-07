package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents an adapter configuration loaded from TOML.
type Config struct {
	Adapter struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
	} `toml:"adapter"`
	Spawn struct {
		Command string `toml:"command"`
	} `toml:"spawn"`
}

// Load reads an adapter config from a TOML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading adapter %s: %w", path, err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing adapter %s: %w", path, err)
	}
	if cfg.Adapter.Name == "" {
		return nil, fmt.Errorf("adapter name required in %s", path)
	}
	if cfg.Spawn.Command == "" {
		return nil, fmt.Errorf("spawn.command required in %s", path)
	}
	return &cfg, nil
}

// LoadDefault loads the opencode adapter from the adapters directory.
func LoadDefault(mcDir string) (*Config, error) {
	path := filepath.Join(mcDir, "adapters", "opencode.toml")
	return Load(path)
}
