// tests for the uninstall subcommand
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

type fakeUninstallHelmClient struct {
	err            error
	called         bool
	gotNamespace   string
	gotReleaseName string
}

func (f *fakeUninstallHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeUninstallHelmClient) ListCharts(_ *cli.EnvSettings) error {
	return nil
}

func (f *fakeUninstallHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeUninstallHelmClient) Uninstall(_ context.Context, _ *log.Logger, settings *cli.EnvSettings, releaseName string) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	f.gotReleaseName = releaseName
	return f.err
}

func TestUninstallCmd(t *testing.T) {
	originalFactory := newUninstallHelmClient
	defer func() { newUninstallHelmClient = originalFactory }()

	tests := []struct {
		name            string
		args            []string
		namespace       string
		factoryErr      error
		uninstallErr    error
		wantErr         bool
		wantErrContains string
		wantCalled      bool
		wantRelease     string
	}{
		{
			name:        "uninstall command runs successfully",
			args:        []string{"my-release"},
			namespace:   "demo",
			wantErr:     false,
			wantCalled:  true,
			wantRelease: "my-release",
		},
		{
			name:            "uninstall command fails without release name",
			args:            []string{},
			namespace:       "demo",
			wantErr:         true,
			wantErrContains: "release name is required",
			wantCalled:      false,
		},
		{
			name:            "uninstall command fails when client creation fails",
			args:            []string{"my-release"},
			namespace:       "demo",
			factoryErr:      errors.New("factory failed"),
			wantErr:         true,
			wantErrContains: "failed to create helm client",
			wantCalled:      false,
		},
		{
			name:            "uninstall command propagates uninstall error",
			args:            []string{"my-release"},
			namespace:       "demo",
			uninstallErr:    errors.New("uninstall failed"),
			wantErr:         true,
			wantErrContains: "uninstall failed",
			wantCalled:      true,
			wantRelease:     "my-release",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			optsUninstall.Namespace = tt.namespace
			fake := &fakeUninstallHelmClient{err: tt.uninstallErr}
			newUninstallHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.AddCommand(uninstallCmd)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(append([]string{"uninstall"}, tt.args...))

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("uninstallCmd error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if fake.called != tt.wantCalled {
				t.Fatalf("expected uninstall called=%v, got %v", tt.wantCalled, fake.called)
			}
			if tt.wantCalled {
				if fake.gotNamespace != tt.namespace {
					t.Fatalf("expected namespace %q, got %q", tt.namespace, fake.gotNamespace)
				}
				if fake.gotReleaseName != tt.wantRelease {
					t.Fatalf("expected release name %q, got %q", tt.wantRelease, fake.gotReleaseName)
				}
			}
		})
	}
}

func TestUninstallOptionsString(t *testing.T) {
	o := &UninstallOptions{Namespace: "ns", ReleaseName: "my-release"}
	if got := o.String(); got != "Namespace: ns, ReleaseName: my-release" {
		t.Fatalf("unexpected options string: %s", got)
	}
}

func TestUninstallCmdMetadata(t *testing.T) {
	if uninstallCmd.Use != "uninstall" {
		t.Errorf("uninstallCmd.Use = %q, want %q", uninstallCmd.Use, "uninstall")
	}
	if uninstallCmd.Short == "" {
		t.Errorf("uninstallCmd.Short should not be empty")
	}
	if uninstallCmd.Long == "" {
		t.Errorf("uninstallCmd.Long should not be empty")
	}
	if uninstallCmd.RunE == nil {
		t.Errorf("uninstallCmd.RunE should not be nil")
	}
}
