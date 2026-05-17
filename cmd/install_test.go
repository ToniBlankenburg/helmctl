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

type fakeInstallHelmClient struct {
	err             error
	called          bool
	gotNamespace    string
	gotChartRef     string
	gotReleaseName  string
	gotChartVersion string
}

func (f *fakeInstallHelmClient) Install(_ context.Context, _ *log.Logger, settings *cli.EnvSettings, req helmclient.InstallRequest) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	f.gotChartRef = req.ChartRef
	f.gotReleaseName = req.ReleaseName
	f.gotChartVersion = req.ChartVersion
	return f.err
}

func (f *fakeInstallHelmClient) ListCharts(_ *cli.EnvSettings) error {
	return nil
}

func (f *fakeInstallHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeInstallHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func TestInstallCmd(t *testing.T) {
	originalFactory := newInstallHelmClient
	defer func() { newInstallHelmClient = originalFactory }()

	tests := []struct {
		name             string
		args             []string
		namespace        string
		factoryErr       error
		installErr       error
		wantErr          bool
		wantErrContains  string
		wantChartRef     string
		wantReleaseName  string
		wantInstallCalls int
	}{
		{
			name:             "install command runs successfully",
			args:             []string{"my-chart"},
			namespace:        "demo",
			wantErr:          false,
			wantChartRef:     "my-chart",
			wantReleaseName:  "my-chart",
			wantInstallCalls: 1,
		},
		{
			name:             "install command derives release name from repo chart",
			args:             []string{"podinfo/podinfo"},
			namespace:        "demo",
			wantErr:          false,
			wantChartRef:     "podinfo/podinfo",
			wantReleaseName:  "podinfo",
			wantInstallCalls: 1,
		},
		{
			name:             "install command fails without chart name",
			args:             []string{},
			namespace:        "demo",
			wantErr:          true,
			wantErrContains:  "chart name is required",
			wantInstallCalls: 0,
		},
		{
			name:             "install command fails with too many arguments",
			args:             []string{"my-chart", "extra"},
			namespace:        "demo",
			wantErr:          true,
			wantErrContains:  "chart name is required",
			wantInstallCalls: 0,
		},
		{
			name:             "install command fails when client creation fails",
			args:             []string{"my-chart"},
			namespace:        "demo",
			factoryErr:       errors.New("boom"),
			wantErr:          true,
			wantErrContains:  "failed to create helm client",
			wantInstallCalls: 0,
		},
		{
			name:             "install command propagates install error",
			args:             []string{"my-chart"},
			namespace:        "demo",
			installErr:       errors.New("install failed"),
			wantErr:          true,
			wantErrContains:  "failed to install helm chart: install failed",
			wantChartRef:     "my-chart",
			wantReleaseName:  "my-chart",
			wantInstallCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts.Namespace = tt.namespace
			fake := &fakeInstallHelmClient{err: tt.installErr}
			newInstallHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.AddCommand(installCmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			cmd.SetArgs(append([]string{"install"}, tt.args...))

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("installCmd error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if fake.called != (tt.wantInstallCalls == 1) {
				t.Fatalf("expected install called=%v, got %v", tt.wantInstallCalls == 1, fake.called)
			}
			if tt.wantInstallCalls == 1 {
				if fake.gotNamespace != tt.namespace {
					t.Fatalf("expected namespace %q, got %q", tt.namespace, fake.gotNamespace)
				}
				if fake.gotChartRef != tt.wantChartRef {
					t.Fatalf("expected chart ref %q, got %q", tt.wantChartRef, fake.gotChartRef)
				}
				if fake.gotReleaseName != tt.wantReleaseName {
					t.Fatalf("expected release name %q, got %q", tt.wantReleaseName, fake.gotReleaseName)
				}
			}

			_ = buf.String()
		})
	}
}

func TestInstallOptionsString(t *testing.T) {
	o := &InstallOptions{Namespace: "ns", ChartName: "chart"}
	if got := o.String(); got != "namespace=ns, chartName=chart" {
		t.Fatalf("unexpected options string: %s", got)
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
