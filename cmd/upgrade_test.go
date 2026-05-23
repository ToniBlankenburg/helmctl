// tests for the upgrade subcommand
package cmd

import (
	"bytes"
	"context"
	"errors"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/release"
)

type fakeUpgradeHelmClient struct {
	err             error
	called          bool
	gotNamespace    string
	gotReleaseName  string
	gotChartRef     string
	gotChartVersion string
	gotValuesFiles  []string
}

func (f *fakeUpgradeHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeUpgradeHelmClient) ListCharts(_ *cli.EnvSettings) ([]release.Accessor, error) {
	return nil, nil
}

func (f *fakeUpgradeHelmClient) Upgrade(_ context.Context, _ *log.Logger, settings *cli.EnvSettings, req helmclient.UpgradeRequest) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	f.gotReleaseName = req.ReleaseName
	f.gotChartRef = req.ChartRef
	f.gotChartVersion = req.ChartVersion
	f.gotValuesFiles = req.ValuesFiles
	return f.err
}

func (f *fakeUpgradeHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func TestUpgradeCmd(t *testing.T) {
	originalFactory := newUpgradeHelmClient
	originalConfig := loadUpgradeConfig
	defer func() {
		newUpgradeHelmClient = originalFactory
		loadUpgradeConfig = originalConfig
		optsUpgrade.Namespace = ""
	}()

	tests := []struct {
		name            string
		args            []string
		flagNamespace   string
		configNamespace string
		configValues    []string
		configErr       error
		factoryErr      error
		upgradeErr      error
		needsChart      bool
		wantErr         bool
		wantErrContains string
		wantCalled      bool
		wantReleaseName string
		wantNamespace   string
		wantValuesFiles []string
	}{
		{
			name:            "upgrades using config chart path and namespace",
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
			configNamespace: "local-dev",
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
			name:            "propagates upgrade error",
			args:            []string{"my_app"},
			needsChart:      true,
			upgradeErr:      errors.New("upgrade failed"),
			wantErr:         true,
			wantErrContains: "failed to upgrade helm chart: upgrade failed",
			wantCalled:      true,
			wantReleaseName: "my_app",
			wantNamespace:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optsUpgrade.Namespace = tt.flagNamespace

			var chartsDir string
			if tt.needsChart {
				chartsDir = createTempChart(t, "my_app")
			}

			loadUpgradeConfig = func() (*config.Config, error) {
				if tt.configErr != nil {
					return nil, tt.configErr
				}
				return &config.Config{
					Namespace: tt.configNamespace,
					ChartsDir: chartsDir,
					Values:    tt.configValues,
				}, nil
			}

			fake := &fakeUpgradeHelmClient{err: tt.upgradeErr}
			newUpgradeHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.AddCommand(upgradeCmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(append([]string{"upgrade"}, tt.args...))

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("upgradeCmd error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if fake.called != tt.wantCalled {
				t.Fatalf("expected upgrade called=%v, got %v", tt.wantCalled, fake.called)
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

func TestUpgradeOptionsString(t *testing.T) {
	o := &UpgradeOptions{Namespace: "ns", ChartName: "my-chart"}
	if got := o.String(); got != "Namespace: ns, ChartName: my-chart" {
		t.Fatalf("unexpected options string: %s", got)
	}
}

func TestUpgradeCmdMetadata(t *testing.T) {
	if upgradeCmd.Use != "upgrade" {
		t.Errorf("upgradeCmd.Use = %q, want %q", upgradeCmd.Use, "upgrade")
	}
	if upgradeCmd.Short == "" {
		t.Errorf("upgradeCmd.Short should not be empty")
	}
	if upgradeCmd.Long == "" {
		t.Errorf("upgradeCmd.Long should not be empty")
	}
	if upgradeCmd.RunE == nil {
		t.Errorf("upgradeCmd.RunE should not be nil")
	}
}
