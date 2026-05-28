package helmclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"helm.sh/helm/v4/pkg/action"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	chartv2util "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/chart/common"
	"helm.sh/helm/v4/pkg/cli"
	kubefake "helm.sh/helm/v4/pkg/kube/fake"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// newTestConfig creates a shared in-memory action.Configuration and a factory
// that always returns it. Sharing one instance lets Install/Upgrade/Uninstall
// calls within the same test see each other's releases.
func newTestConfig(t *testing.T) (*action.Configuration, func(string) (*action.Configuration, error)) {
	t.Helper()
	cfg := &action.Configuration{
		Releases:     storage.Init(driver.NewMemory()),
		KubeClient:   &kubefake.PrintingKubeClient{Out: io.Discard},
		Capabilities: common.DefaultCapabilities,
	}
	return cfg, func(_ string) (*action.Configuration, error) { return cfg, nil }
}

// newTestClient builds a realHelmClient backed by an in-memory config.
func newTestClient(t *testing.T) (*realHelmClient, *action.Configuration) {
	t.Helper()
	cfg, factory := newTestConfig(t)
	return &realHelmClient{
		newActionConfig: factory,
		logger:          log.New(io.Discard, "", 0),
	}, cfg
}

// newTestSettings creates cli.EnvSettings with a temp dir for the registry
// config so tests never depend on ~/.config/helm.
func newTestSettings(t *testing.T) *cli.EnvSettings {
	t.Helper()
	s := cli.New()
	s.RegistryConfig = filepath.Join(t.TempDir(), "registry.json")
	return s
}

// createTestChart builds a minimal chart tgz in dir named <name>.tgz
// (no version suffix — matching helmctl's convention).
func createTestChart(t *testing.T, dir, name string) string {
	t.Helper()
	ch := &chartv2.Chart{
		Metadata: &chartv2.Metadata{
			APIVersion: chartv2.APIVersionV2,
			Name:       name,
			Version:    "0.1.0",
		},
	}
	// chartv2util.Save produces name-0.1.0.tgz; rename to name.tgz
	saved, err := chartv2util.Save(ch, dir)
	if err != nil {
		t.Fatalf("createTestChart: save: %v", err)
	}
	target := filepath.Join(dir, name+".tgz")
	if err := os.Rename(saved, target); err != nil {
		t.Fatalf("createTestChart: rename: %v", err)
	}
	return target
}

// ── mergeValuesFiles ─────────────────────────────────────────────────────────

func TestMergeValuesFiles(t *testing.T) {
	writeYAML := func(t *testing.T, dir, name, content string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatalf("setup: write %s: %v", name, err)
		}
		return path
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string) []string
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "nil slice returns nil map",
			setup: func(_ *testing.T, _ string) []string { return nil },
			want:  nil,
		},
		{
			name: "single file parsed correctly",
			setup: func(t *testing.T, dir string) []string {
				return []string{writeYAML(t, dir, "a.yaml", "key: value\ncount: 3\n")}
			},
			want: map[string]any{"key": "value", "count": 3},
		},
		{
			name: "second file overrides first",
			setup: func(t *testing.T, dir string) []string {
				a := writeYAML(t, dir, "a.yaml", "key: first\nother: kept\n")
				b := writeYAML(t, dir, "b.yaml", "key: second\n")
				return []string{a, b}
			},
			want: map[string]any{"key": "second", "other": "kept"},
		},
		{
			name: "non-overlapping files merged",
			setup: func(t *testing.T, dir string) []string {
				a := writeYAML(t, dir, "a.yaml", "alpha: 1\n")
				b := writeYAML(t, dir, "b.yaml", "beta: 2\n")
				return []string{a, b}
			},
			want: map[string]any{"alpha": 1, "beta": 2},
		},
		{
			name: "invalid YAML returns error",
			setup: func(t *testing.T, dir string) []string {
				return []string{writeYAML(t, dir, "bad.yaml", "key: [unclosed\n")}
			},
			wantErr: true,
		},
		{
			name: "missing file returns error",
			setup: func(_ *testing.T, dir string) []string {
				return []string{filepath.Join(dir, "does-not-exist.yaml")}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			paths := tt.setup(t, dir)

			got, err := mergeValuesFiles(paths)

			if (err != nil) != tt.wantErr {
				t.Fatalf("mergeValuesFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("mergeValuesFiles() returned %d keys, want %d: got %v", len(got), len(tt.want), got)
			}
			for k, want := range tt.want {
				gotVal, ok := got[k]
				if !ok {
					t.Errorf("key %q missing from result", k)
					continue
				}
				if fmt.Sprintf("%v", gotVal) != fmt.Sprintf("%v", want) {
					t.Errorf("key %q: got %v (%T), want %v (%T)", k, gotVal, gotVal, want, want)
				}
			}
		})
	}
}

// ── newRegistryClient ─────────────────────────────────────────────────────────

func TestNewRegistryClient(t *testing.T) {
	tests := []struct {
		name      string
		plainHTTP bool
	}{
		{name: "creates client with plainHTTP disabled", plainHTTP: false},
		{name: "creates client with plainHTTP enabled", plainHTTP: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := newTestSettings(t)
			client, err := newRegistryClient(settings, tt.plainHTTP)
			if err != nil {
				t.Fatalf("newRegistryClient() unexpected error: %v", err)
			}
			if client == nil {
				t.Fatal("newRegistryClient() returned nil client")
			}
		})
	}
}

// ── Install ───────────────────────────────────────────────────────────────────

func TestInstall(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string) InstallRequest
		wantErr   bool
	}{
		{
			name: "success with minimal chart and no values",
			setup: func(t *testing.T, dir string) InstallRequest {
				chartPath := createTestChart(t, dir, "echo")
				return InstallRequest{
					ReleaseName: "echo",
					ChartRef:    chartPath,
				}
			},
		},
		{
			name: "success with values file",
			setup: func(t *testing.T, dir string) InstallRequest {
				chartPath := createTestChart(t, dir, "echo")
				valsPath := filepath.Join(dir, "values.yaml")
				if err := os.WriteFile(valsPath, []byte("key: value\n"), 0600); err != nil {
					t.Fatal(err)
				}
				return InstallRequest{
					ReleaseName: "echo",
					ChartRef:    chartPath,
					ValuesFiles: []string{valsPath},
				}
			},
		},
		{
			name: "error on missing values file",
			setup: func(t *testing.T, dir string) InstallRequest {
				chartPath := createTestChart(t, dir, "echo")
				return InstallRequest{
					ReleaseName: "echo",
					ChartRef:    chartPath,
					ValuesFiles: []string{filepath.Join(dir, "nonexistent.yaml")},
				}
			},
			wantErr: true,
		},
		{
			name: "error from action config factory",
			setup: func(t *testing.T, dir string) InstallRequest {
				return InstallRequest{ReleaseName: "echo", ChartRef: filepath.Join(dir, "echo.tgz")}
			},
			wantErr: true, // chart does not exist → LocateChart fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			c, _ := newTestClient(t)
			settings := newTestSettings(t)
			req := tt.setup(t, dir)

			err := c.Install(context.Background(), log.New(io.Discard, "", 0), settings, req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstall_actionConfigError(t *testing.T) {
	c := &realHelmClient{
		newActionConfig: func(_ string) (*action.Configuration, error) {
			return nil, fmt.Errorf("cluster unreachable")
		},
		logger: log.New(io.Discard, "", 0),
	}
	err := c.Install(context.Background(), log.New(io.Discard, "", 0), newTestSettings(t), InstallRequest{})
	if err == nil {
		t.Fatal("Install() expected error from config factory, got nil")
	}
}

// ── Upgrade ───────────────────────────────────────────────────────────────────

func TestUpgrade(t *testing.T) {
	t.Run("success after prior install", func(t *testing.T) {
		dir := t.TempDir()
		c, _ := newTestClient(t)
		settings := newTestSettings(t)
		chartPath := createTestChart(t, dir, "echo")
		logger := log.New(io.Discard, "", 0)

		// install first so the release exists for upgrade
		if err := c.Install(context.Background(), logger, settings, InstallRequest{
			ReleaseName: "echo",
			ChartRef:    chartPath,
		}); err != nil {
			t.Fatalf("prerequisite Install failed: %v", err)
		}

		err := c.Upgrade(context.Background(), logger, settings, UpgradeRequest{
			ReleaseName: "echo",
			ChartRef:    chartPath,
		})
		if err != nil {
			t.Fatalf("Upgrade() unexpected error: %v", err)
		}
	})

	t.Run("error from action config factory", func(t *testing.T) {
		c := &realHelmClient{
			newActionConfig: func(_ string) (*action.Configuration, error) {
				return nil, fmt.Errorf("cluster unreachable")
			},
			logger: log.New(io.Discard, "", 0),
		}
		err := c.Upgrade(context.Background(), log.New(io.Discard, "", 0), newTestSettings(t), UpgradeRequest{})
		if err == nil {
			t.Fatal("Upgrade() expected error from config factory, got nil")
		}
	})
}

// ── Uninstall ─────────────────────────────────────────────────────────────────

func TestUninstall(t *testing.T) {
	t.Run("success after prior install", func(t *testing.T) {
		dir := t.TempDir()
		c, _ := newTestClient(t)
		settings := newTestSettings(t)
		chartPath := createTestChart(t, dir, "echo")
		logger := log.New(io.Discard, "", 0)

		if err := c.Install(context.Background(), logger, settings, InstallRequest{
			ReleaseName: "echo",
			ChartRef:    chartPath,
		}); err != nil {
			t.Fatalf("prerequisite Install failed: %v", err)
		}

		err := c.Uninstall(context.Background(), logger, settings, "echo")
		if err != nil {
			t.Fatalf("Uninstall() unexpected error: %v", err)
		}
	})

	t.Run("error when release does not exist", func(t *testing.T) {
		c, _ := newTestClient(t)
		err := c.Uninstall(context.Background(), log.New(io.Discard, "", 0), newTestSettings(t), "nonexistent")
		if err == nil {
			t.Fatal("Uninstall() expected error for missing release, got nil")
		}
	})

	t.Run("error from action config factory", func(t *testing.T) {
		c := &realHelmClient{
			newActionConfig: func(_ string) (*action.Configuration, error) {
				return nil, fmt.Errorf("cluster unreachable")
			},
			logger: log.New(io.Discard, "", 0),
		}
		err := c.Uninstall(context.Background(), log.New(io.Discard, "", 0), newTestSettings(t), "echo")
		if err == nil {
			t.Fatal("Uninstall() expected error from config factory, got nil")
		}
	})
}

// ── ListCharts ────────────────────────────────────────────────────────────────

func TestListCharts(t *testing.T) {
	t.Run("empty store returns empty list", func(t *testing.T) {
		c, _ := newTestClient(t)
		got, err := c.ListCharts(newTestSettings(t))
		if err != nil {
			t.Fatalf("ListCharts() unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("ListCharts() = %d releases, want 0", len(got))
		}
	})

	t.Run("returns installed release", func(t *testing.T) {
		dir := t.TempDir()
		c, _ := newTestClient(t)
		settings := newTestSettings(t)
		chartPath := createTestChart(t, dir, "echo")
		logger := log.New(io.Discard, "", 0)

		if err := c.Install(context.Background(), logger, settings, InstallRequest{
			ReleaseName: "echo",
			ChartRef:    chartPath,
		}); err != nil {
			t.Fatalf("prerequisite Install failed: %v", err)
		}

		got, err := c.ListCharts(settings)
		if err != nil {
			t.Fatalf("ListCharts() unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("ListCharts() = %d releases, want 1", len(got))
		}
	})

	t.Run("error from action config factory", func(t *testing.T) {
		c := &realHelmClient{
			newActionConfig: func(_ string) (*action.Configuration, error) {
				return nil, fmt.Errorf("cluster unreachable")
			},
			logger: log.New(io.Discard, "", 0),
		}
		_, err := c.ListCharts(newTestSettings(t))
		if err == nil {
			t.Fatal("ListCharts() expected error from config factory, got nil")
		}
	})
}
