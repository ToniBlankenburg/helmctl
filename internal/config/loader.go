package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Namespace string   `yaml:"namespace"`
	ChartsDir string   `yaml:"charts_dir"`
	Values    []string `yaml:"values"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	err := readYAML(path, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func readYAML(path string, c *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	dir := filepath.Dir(path)
	if c.ChartsDir != "" {
		c.ChartsDir = filepath.Join(dir, c.ChartsDir)
	}
	for i, valPath := range c.Values {
		c.Values[i] = filepath.Join(dir, valPath)
	}

	return nil
}
