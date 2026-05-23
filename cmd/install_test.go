package cmd

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/release"
)

// createTempChart creates an empty .tgz file in a temp dir and returns the dir path.
func createTempChart(t *testing.T, appName string) string {
	t.Helper()
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, appName+".tgz"))
	if err != nil {
		t.Fatalf("failed to create temp chart: %v", err)
	}
	f.Close()
	return dir
}

type fakeInstallHelmClient struct {
	err            error
	called         bool
	gotNamespace   string
	gotChartRef    string
	gotReleaseName string
	gotValuesFiles []string
}

func (f *fakeInstallHelmClient) Install(_ context.Context, _ *log.Logger, settings *cli.EnvSettings, req helmclient.InstallRequest) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	f.gotChartRef = req.ChartRef
	f.gotReleaseName = req.ReleaseName
	f.gotValuesFiles = req.ValuesFiles
	return f.err
}

func (f *fakeInstallHelmClient) ListCharts(_ *cli.EnvSettings) ([]release.Accessor, error) {
	return nil, nil
}

func (f *fakeInstallHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeInstallHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func TestInstallCmd(t *testing.T) {
	originalFactory := newInstallHelmClient
	originalConfig := loadInstallConfig
	defer func() {
		newInstallHelmClient = originalFactory
		loadInstallConfig = originalConfig
		opts.Namespace = ""
	}()

	tests := []struct {
		name            string
		args            []string
		flagNamespace   string
		configNamespace string
		configChartsDir string
		configValues    []string
		configErr       error
		factoryErr      error
		installErr      error
		needsChart      bool // create a real .tgz so os.Stat passes
		wantErr         bool
		wantErrContains string
		wantReleaseName string
		wantNamespace   string
		wantValuesFiles []string
		wantCalled      bool
	}{
		{
			name:            "installs using config chart path and namespace",
			args:            []string{"my_app"},
			configNamespace: "local-dev",
			configValues:    []string{"/fake/values/common.yaml"},
			needsChart:      true,
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "local-dev",
			wantValuesFiles: []string{"/fake/values/common.yaml"},
		},
		{
			name:            "flag namespace overrides config namespace",
			args:            []string{"my_app"},
			flagNamespace:   "staging",
			configNamespace: "local-dev",
			needsChart:      true,
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "staging",
		},
		{
			name:            "falls back to default namespace when config has none",
			args:            []string{"my_app"},
			needsChart:      true,
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "default",
		},
		{
			name:            "passes empty values when config has none",
			args:            []string{"my_app"},
			needsChart:      true,
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "default",
			wantValuesFiles: nil,
		},
		{
			name:            "fails when chart tgz does not exist",
			args:            []string{"my_app"},
			configChartsDir: "/nonexistent/charts",
			wantErr:         true,
			wantErrContains: "chart not found",
		},
		{
			name:            "fails with clear path in error when charts_dir is empty",
			args:            []string{"my_app"},
			configChartsDir: "",
			wantErr:         true,
			wantErrContains: "chart not found",
		},
		{
			name:            "fails without app name",
			args:            []string{},
			wantErr:         true,
			wantErrContains: "app name is required",
		},
		{
			name:            "fails with too many arguments",
			args:            []string{"my_app", "extra"},
			wantErr:         true,
			wantErrContains: "app name is required",
		},
		{
			name:            "fails when config loading fails",
			args:            []string{"my_app"},
			configErr:       errors.New("no config found"),
			wantErr:         true,
			wantErrContains: "failed to load config",
		},
		{
			name:            "fails when client creation fails",
			args:            []string{"my_app"},
			needsChart:      true,
			factoryErr:      errors.New("boom"),
			wantErr:         true,
			wantErrContains: "failed to create helm client",
		},
		{
			name:            "propagates install error",
			args:            []string{"my_app"},
			needsChart:      true,
			installErr:      errors.New("install failed"),
			wantErr:         true,
			wantErrContains: "failed to install helm chart: install failed",
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts.Namespace = tt.flagNamespace

			chartsDir := tt.configChartsDir
			if tt.needsChart {
				chartsDir = createTempChart(t, "my_app")
			}

			loadInstallConfig = func() (*config.Config, error) {
				if tt.configErr != nil {
					return nil, tt.configErr
				}
				return &config.Config{
					Namespace: tt.configNamespace,
					ChartsDir: chartsDir,
					Values:    tt.configValues,
				}, nil
			}

			fake := &fakeInstallHelmClient{err: tt.installErr}
			newInstallHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.AddCommand(installCmd)
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(append([]string{"install"}, tt.args...))

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if fake.called != tt.wantCalled {
				t.Fatalf("called = %v, want %v", fake.called, tt.wantCalled)
			}
			if !tt.wantCalled {
				return
			}
			if fake.gotReleaseName != tt.wantReleaseName {
				t.Errorf("ReleaseName = %q, want %q", fake.gotReleaseName, tt.wantReleaseName)
			}
			wantChartRef := filepath.Join(chartsDir, tt.wantReleaseName+".tgz")
			if fake.gotChartRef != wantChartRef {
				t.Errorf("ChartRef = %q, want %q", fake.gotChartRef, wantChartRef)
			}
			if fake.gotNamespace != tt.wantNamespace {
				t.Errorf("Namespace = %q, want %q", fake.gotNamespace, tt.wantNamespace)
			}
			if len(fake.gotValuesFiles) != len(tt.wantValuesFiles) {
				t.Fatalf("ValuesFiles = %v, want %v", fake.gotValuesFiles, tt.wantValuesFiles)
			}
			for i, v := range fake.gotValuesFiles {
				if v != tt.wantValuesFiles[i] {
					t.Errorf("ValuesFiles[%d] = %q, want %q", i, v, tt.wantValuesFiles[i])
				}
			}
		})
	}
}

func TestInstallCmdMetadata(t *testing.T) {
	if installCmd.Use != "install" {
		t.Errorf("installCmd.Use = %q, want %q", installCmd.Use, "install")
	}
	if installCmd.Short == "" {
		t.Errorf("installCmd.Short should not be empty")
	}
	if installCmd.Long == "" {
		t.Errorf("installCmd.Long should not be empty")
	}
	if installCmd.RunE == nil {
		t.Errorf("installCmd.RunE should not be nil")
	}
}
