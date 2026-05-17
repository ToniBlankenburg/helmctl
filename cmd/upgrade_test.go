// tests for the upgrade subcommand
package cmd

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

type fakeUpgradeHelmClient struct {
	err             error
	called          bool
	gotNamespace    string
	gotReleaseName  string
	gotChartRef     string
	gotChartVersion string
}

func (f *fakeUpgradeHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeUpgradeHelmClient) ListCharts(_ *cli.EnvSettings) error {
	return nil
}

func (f *fakeUpgradeHelmClient) Upgrade(_ context.Context, _ *log.Logger, settings *cli.EnvSettings, req helmclient.UpgradeRequest) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	f.gotReleaseName = req.ReleaseName
	f.gotChartRef = req.ChartRef
	f.gotChartVersion = req.ChartVersion
	return f.err
}

func (f *fakeUpgradeHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func TestUpgradeCmd(t *testing.T) {
	originalFactory := newUpgradeHelmClient
	defer func() { newUpgradeHelmClient = originalFactory }()

	tests := []struct {
		name            string
		args            []string
		namespace       string
		factoryErr      error
		upgradeErr      error
		wantErr         bool
		wantErrContains string
		wantCalled      bool
		wantRelease     string
		wantChartRef    string
	}{
		{
			name:         "upgrade command runs successfully",
			args:         []string{"my-chart"},
			namespace:    "demo",
			wantErr:      false,
			wantCalled:   true,
			wantRelease:  "my-chart",
			wantChartRef: "my-chart",
		},
		{
			name:         "upgrade command derives release name from repo chart",
			args:         []string{"podinfo/podinfo"},
			namespace:    "demo",
			wantErr:      false,
			wantCalled:   true,
			wantRelease:  "podinfo",
			wantChartRef: "podinfo/podinfo",
		},
		{
			name:            "upgrade command fails without arguments",
			args:            []string{},
			namespace:       "demo",
			wantErr:         true,
			wantErrContains: "chart name is required",
			wantCalled:      false,
		},
		{
			name:            "upgrade command fails with too many arguments",
			args:            []string{"my-chart", "extra"},
			namespace:       "demo",
			wantErr:         true,
			wantErrContains: "chart name is required",
			wantCalled:      false,
		},
		{
			name:            "upgrade command fails when client creation fails",
			args:            []string{"my-chart"},
			namespace:       "demo",
			factoryErr:      errors.New("factory failed"),
			wantErr:         true,
			wantErrContains: "failed to create helm client",
			wantCalled:      false,
		},
		{
			name:            "upgrade command propagates upgrade error",
			args:            []string{"my-chart"},
			namespace:       "demo",
			upgradeErr:      errors.New("upgrade failed"),
			wantErr:         true,
			wantErrContains: "failed to upgrade helm chart: upgrade failed",
			wantCalled:      true,
			wantRelease:     "my-chart",
			wantChartRef:    "my-chart",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			optsUpgrade.Namespace = tt.namespace
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
			if tt.wantCalled {
				if fake.gotNamespace != tt.namespace {
					t.Fatalf("expected namespace %q, got %q", tt.namespace, fake.gotNamespace)
				}
				if fake.gotReleaseName != tt.wantRelease {
					t.Fatalf("expected release name %q, got %q", tt.wantRelease, fake.gotReleaseName)
				}
				if fake.gotChartRef != tt.wantChartRef {
					t.Fatalf("expected chart ref %q, got %q", tt.wantChartRef, fake.gotChartRef)
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
