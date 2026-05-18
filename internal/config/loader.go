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

// FindFrom walks up from startDir looking for helmctl.yaml.
func FindFrom(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "helmctl.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("helmctl.yaml not found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

// Find walks up from the current working directory looking for helmctl.yaml.
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return FindFrom(dir)
}

// FindAndLoad finds the nearest helmctl.yaml and loads it.
func FindAndLoad() (*Config, error) {
	path, err := Find()
	if err != nil {
		return nil, err
	}
	return Load(path)
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
