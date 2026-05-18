package config

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: write a helmctl.yaml into a temp dir and return its path
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "helmctl.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		wantCfg     *Config
	}{
		{
			name: "fully populated config",
			yaml: `
namespace: local-dev
charts_dir: target
values:
  - helm/values/values-a.yaml
  - helm/values/values-b.yaml
`,
			wantCfg: &Config{
				Namespace: "local-dev",
				ChartsDir: "target",
				Values:    []string{"helm/values/values-a.yaml", "helm/values/values-b.yaml"},
			},
		},
		{
			name: "namespace only",
			yaml: `namespace: staging`,
			wantCfg: &Config{
				Namespace: "staging",
				ChartsDir: "",
				Values:    nil,
			},
		},
		{
			name: "empty values list",
			yaml: `
namespace: local-dev
charts_dir: target
values: []
`,
			wantCfg: &Config{
				Namespace: "local-dev",
				ChartsDir: "target",
				Values:    []string{},
			},
		},
		{
			name:    "invalid yaml",
			yaml:    `namespace: [unclosed`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.yaml)
			configDir := filepath.Dir(path)

			cfg, err := Load(path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if cfg.Namespace != tt.wantCfg.Namespace {
				t.Errorf("Namespace = %q, want %q", cfg.Namespace, tt.wantCfg.Namespace)
			}

			wantChartsDir := ""
			if tt.wantCfg.ChartsDir != "" {
				wantChartsDir = filepath.Join(configDir, tt.wantCfg.ChartsDir)
			}
			if cfg.ChartsDir != wantChartsDir {
				t.Errorf("ChartsDir = %q, want %q", cfg.ChartsDir, wantChartsDir)
			}

			if len(cfg.Values) != len(tt.wantCfg.Values) {
				t.Fatalf("Values len = %d, want %d", len(cfg.Values), len(tt.wantCfg.Values))
			}
			for i, v := range cfg.Values {
				wantValue := filepath.Join(configDir, tt.wantCfg.Values[i])
				if v != wantValue {
					t.Errorf("Values[%d] = %q, want %q", i, v, wantValue)
				}
			}
		})
	}
}

func TestFindFrom(t *testing.T) {
	t.Run("finds config in start directory", func(t *testing.T) {
		path := writeConfig(t, `namespace: local-dev`)
		got, err := FindFrom(filepath.Dir(path))
		if err != nil {
			t.Fatalf("FindFrom() unexpected error: %v", err)
		}
		if got != path {
			t.Errorf("FindFrom() = %q, want %q", got, path)
		}
	})

	t.Run("finds config in parent directory", func(t *testing.T) {
		path := writeConfig(t, `namespace: local-dev`)
		subDir := filepath.Join(filepath.Dir(path), "sub", "dir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}
		got, err := FindFrom(subDir)
		if err != nil {
			t.Fatalf("FindFrom() unexpected error: %v", err)
		}
		if got != path {
			t.Errorf("FindFrom() = %q, want %q", got, path)
		}
	})

	t.Run("returns error when not found", func(t *testing.T) {
		_, err := FindFrom(t.TempDir())
		if err == nil {
			t.Fatal("expected error when config not found, got nil")
		}
	})
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/does/not/exist/helmctl.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// Paths in charts_dir and values are resolved relative to the config file's directory.
func TestLoad_ResolvesRelativePaths(t *testing.T) {
	path := writeConfig(t, `
namespace: local-dev
charts_dir: target
values:
  - helm/values/values-a.yaml
`)
	configDir := filepath.Dir(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	wantChartsDir := filepath.Join(configDir, "target")
	if cfg.ChartsDir != wantChartsDir {
		t.Errorf("ChartsDir = %q, want %q", cfg.ChartsDir, wantChartsDir)
	}

	wantValue := filepath.Join(configDir, "helm/values/values-a.yaml")
	if len(cfg.Values) != 1 || cfg.Values[0] != wantValue {
		t.Errorf("Values = %v, want [%q]", cfg.Values, wantValue)
	}
}
