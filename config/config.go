package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level logpipe configuration.
type Config struct {
	Inputs  []InputConfig  `yaml:"inputs"`
	Filters []FilterConfig `yaml:"filters"`
	Outputs []OutputConfig `yaml:"outputs"`
}

// InputConfig defines a log source.
type InputConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // e.g. "file", "stdin", "tcp"
	Path string `yaml:"path,omitempty"`
	Addr string `yaml:"addr,omitempty"`
}

// FilterConfig defines a filtering rule applied to log entries.
type FilterConfig struct {
	Name    string            `yaml:"name"`
	Field   string            `yaml:"field"`
	Match   string            `yaml:"match"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

// OutputConfig defines a log destination.
type OutputConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"` // e.g. "stdout", "file", "http"
	Target string `yaml:"target,omitempty"`
	Format string `yaml:"format,omitempty"` // e.g. "json", "text"
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parsing YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is semantically valid.
func (c *Config) Validate() error {
	if len(c.Inputs) == 0 {
		return fmt.Errorf("at least one input must be defined")
	}
	if len(c.Outputs) == 0 {
		return fmt.Errorf("at least one output must be defined")
	}
	for i, inp := range c.Inputs {
		if inp.Name == "" {
			return fmt.Errorf("input[%d]: name is required", i)
		}
		if inp.Type == "" {
			return fmt.Errorf("input %q: type is required", inp.Name)
		}
	}
	for i, out := range c.Outputs {
		if out.Name == "" {
			return fmt.Errorf("output[%d]: name is required", i)
		}
		if out.Type == "" {
			return fmt.Errorf("output %q: type is required", out.Name)
		}
	}
	return nil
}
