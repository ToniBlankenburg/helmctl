package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveChartPath(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		appName   string
		wantSuffix string // expected filename suffix (not full path)
		wantErr   bool
	}{
		{
			name: "plain <name>.tgz preferred when present",
			setup: func(t *testing.T, dir string) {
				touch(t, filepath.Join(dir, "echo.tgz"))
				touch(t, filepath.Join(dir, "echo-0.1.0.tgz"))
			},
			appName:    "echo",
			wantSuffix: "echo.tgz",
		},
		{
			name: "versioned fallback when plain is absent",
			setup: func(t *testing.T, dir string) {
				touch(t, filepath.Join(dir, "echo-0.1.0.tgz"))
			},
			appName:    "echo",
			wantSuffix: "echo-0.1.0.tgz",
		},
		{
			name: "latest version chosen when multiple versioned files exist",
			setup: func(t *testing.T, dir string) {
				touch(t, filepath.Join(dir, "echo-0.1.0.tgz"))
				touch(t, filepath.Join(dir, "echo-0.2.0.tgz"))
			},
			appName:    "echo",
			wantSuffix: "echo-0.2.0.tgz",
		},
		{
			name: "hyphenated chart name does not match wrong versioned file",
			setup: func(t *testing.T, dir string) {
				touch(t, filepath.Join(dir, "echo-db-0.1.0.tgz"))
			},
			appName:    "echo",
			wantErr:    true, // echo-db-0.1.0 must not match appName "echo"
		},
		{
			name:    "no chart file returns error",
			setup:   func(_ *testing.T, _ string) {},
			appName: "echo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got, err := resolveChartPath(dir, tt.appName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveChartPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if filepath.Base(got) != tt.wantSuffix {
				t.Errorf("resolveChartPath() = %q, want filename %q", got, tt.wantSuffix)
			}
		})
	}
}

// touch creates an empty file at path, failing the test on error.
func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatalf("touch %s: %v", path, err)
	}
}
