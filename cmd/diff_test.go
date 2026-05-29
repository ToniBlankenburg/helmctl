package cmd

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/release"
)

type fakeDiffHelmClient struct {
	deployedManifest string
	deployedErr      error
	localManifest    string
	renderErr        error

	gotReleaseName string
	gotChartRef    string
	gotValuesFiles []string
}

func (f *fakeDiffHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeDiffHelmClient) ListCharts(_ *cli.EnvSettings) ([]release.Accessor, error) {
	return nil, nil
}

func (f *fakeDiffHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeDiffHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func (f *fakeDiffHelmClient) GetReleaseManifest(_ *cli.EnvSettings, releaseName string) (string, error) {
	f.gotReleaseName = releaseName
	return f.deployedManifest, f.deployedErr
}

func (f *fakeDiffHelmClient) RenderManifest(_ context.Context, _ *cli.EnvSettings, req helmclient.RenderRequest) (string, error) {
	f.gotChartRef = req.ChartRef
	f.gotValuesFiles = req.ValuesFiles
	return f.localManifest, f.renderErr
}

func TestDiffCmd(t *testing.T) {
	originalFactory := newDiffHelmClient
	originalConfig := loadDiffConfig
	defer func() {
		newDiffHelmClient = originalFactory
		loadDiffConfig = originalConfig
		optsDiff.Namespace = ""
	}()

	tests := []struct {
		name             string
		args             []string
		flagNamespace    string
		configNamespace  string
		configValues     []string
		configErr        error
		factoryErr       error
		needsChart       bool
		deployedManifest string
		deployedErr      error
		localManifest    string
		renderErr        error
		wantErr          bool
		wantErrContains  string
		wantOutputContains string
		wantNoDiff       bool
	}{
		{
			name:             "shows no differences when manifests match",
			args:             []string{"my_app"},
			configNamespace:  "local-dev",
			needsChart:       true,
			deployedManifest: "apiVersion: v1\nkind: Pod\n",
			localManifest:    "apiVersion: v1\nkind: Pod\n",
			wantNoDiff:       true,
		},
		{
			name:             "shows unified diff when manifests differ",
			args:             []string{"my_app"},
			configNamespace:  "local-dev",
			needsChart:       true,
			deployedManifest: "apiVersion: v1\nkind: Pod\nreplicas: 1\n",
			localManifest:    "apiVersion: v1\nkind: Pod\nreplicas: 2\n",
			wantOutputContains: "-replicas: 1",
		},
		{
			name:            "flag namespace overrides config namespace",
			args:            []string{"my_app"},
			flagNamespace:   "staging",
			configNamespace: "local-dev",
			needsChart:      true,
			wantNoDiff:      true,
		},
		{
			name:            "falls back to default namespace when config has none",
			args:            []string{"my_app"},
			needsChart:      true,
			wantNoDiff:      true,
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
			name:            "fails when chart not found",
			args:            []string{"my_app"},
			configNamespace: "local-dev",
			wantErr:         true,
			wantErrContains: "chart not found",
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
			name:            "fails when get deployed manifest fails",
			args:            []string{"my_app"},
			needsChart:      true,
			deployedErr:     errors.New("release not found"),
			wantErr:         true,
			wantErrContains: "failed to get deployed manifest",
		},
		{
			name:            "fails when render manifest fails",
			args:            []string{"my_app"},
			needsChart:      true,
			renderErr:       errors.New("chart load error"),
			wantErr:         true,
			wantErrContains: "failed to render local manifest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optsDiff.Namespace = tt.flagNamespace

			var chartsDir string
			if tt.needsChart {
				chartsDir = createTempChart(t, "my_app")
			}

			loadDiffConfig = func() (*config.Config, error) {
				if tt.configErr != nil {
					return nil, tt.configErr
				}
				return &config.Config{
					Namespace: tt.configNamespace,
					ChartsDir: chartsDir,
					Values:    tt.configValues,
				}, nil
			}

			fake := &fakeDiffHelmClient{
				deployedManifest: tt.deployedManifest,
				deployedErr:      tt.deployedErr,
				localManifest:    tt.localManifest,
				renderErr:        tt.renderErr,
			}
			newDiffHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.AddCommand(diffCmd)
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(new(bytes.Buffer))
			cmd.SetArgs(append([]string{"diff"}, tt.args...))

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if tt.wantNoDiff && !strings.Contains(out.String(), "No differences found.") {
				t.Errorf("expected 'No differences found.' in output, got: %q", out.String())
			}
			if tt.wantOutputContains != "" && !strings.Contains(out.String(), tt.wantOutputContains) {
				t.Errorf("expected output to contain %q, got: %q", tt.wantOutputContains, out.String())
			}
		})
	}
}

func TestDiffCmdMetadata(t *testing.T) {
	if diffCmd.Use != "diff <app>" {
		t.Errorf("diffCmd.Use = %q, want %q", diffCmd.Use, "diff <app>")
	}
	if diffCmd.Short == "" {
		t.Errorf("diffCmd.Short should not be empty")
	}
	if diffCmd.Long == "" {
		t.Errorf("diffCmd.Long should not be empty")
	}
	if diffCmd.RunE == nil {
		t.Errorf("diffCmd.RunE should not be nil")
	}
}
