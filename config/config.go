package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Filter defines criteria for matching log entries.
type Filter struct {
	Level    string            `yaml:"level"`
	Contains string            `yaml:"contains"`
	Fields   map[string]string `yaml:"fields"`
}

// Input describes a log source.
type Input struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Path string `yaml:"path,omitempty"`
	Addr string `yaml:"addr,omitempty"`
}

// Output describes a log destination.
type Output struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Path    string   `yaml:"path,omitempty"`
	Addr    string   `yaml:"addr,omitempty"`
	Filters []Filter `yaml:"filters,omitempty"`
}

// Config is the top-level configuration structure.
type Config struct {
	Inputs  []Input  `yaml:"inputs"`
	Outputs []Output `yaml:"outputs"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if len(cfg.Inputs) == 0 {
		return nil, errors.New("config: at least one input is required")
	}
	if len(cfg.Outputs) == 0 {
		return nil, errors.New("config: at least one output is required")
	}

	return &cfg, nil
}
