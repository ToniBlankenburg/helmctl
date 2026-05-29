// tests for the uninstall subcommand
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

type fakeUninstallHelmClient struct {
	err            error
	called         bool
	gotNamespace   string
	gotReleaseName string
}

func (f *fakeUninstallHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeUninstallHelmClient) ListCharts(_ *cli.EnvSettings) ([]release.Accessor, error) {
	return nil, nil
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

func (f *fakeUninstallHelmClient) GetReleaseManifest(_ *cli.EnvSettings, _ string) (string, error) {
	return "", nil
}

func (f *fakeUninstallHelmClient) RenderManifest(_ context.Context, _ *cli.EnvSettings, _ helmclient.RenderRequest) (string, error) {
	return "", nil
}

func TestUninstallCmd(t *testing.T) {
	originalFactory := newUninstallHelmClient
	originalConfig := loadUninstallConfig
	defer func() {
		newUninstallHelmClient = originalFactory
		loadUninstallConfig = originalConfig
		optsUninstall.Namespace = ""
	}()

	tests := []struct {
		name            string
		args            []string
		flagNamespace   string
		configNamespace string
		configErr       error
		factoryErr      error
		uninstallErr    error
		wantErr         bool
		wantErrContains string
		wantCalled      bool
		wantRelease     string
		wantNamespace   string
	}{
		{
			name:            "uninstalls using config namespace",
			args:            []string{"my-release"},
			configNamespace: "local-dev",
			wantCalled:      true,
			wantRelease:     "my-release",
			wantNamespace:   "local-dev",
		},
		{
			name:            "flag namespace overrides config namespace",
			args:            []string{"my-release"},
			flagNamespace:   "staging",
			configNamespace: "local-dev",
			wantCalled:      true,
			wantRelease:     "my-release",
			wantNamespace:   "staging",
		},
		{
			name:          "falls back to default namespace when config has none",
			args:          []string{"my-release"},
			wantCalled:    true,
			wantRelease:   "my-release",
			wantNamespace: "default",
		},
		{
			name:            "fails without release name",
			args:            []string{},
			wantErr:         true,
			wantErrContains: "release name is required",
		},
		{
			name:            "fails with too many arguments",
			args:            []string{"my-release", "extra"},
			wantErr:         true,
			wantErrContains: "release name is required",
		},
		{
			name:            "fails when config loading fails",
			args:            []string{"my-release"},
			configErr:       errors.New("no config found"),
			wantErr:         true,
			wantErrContains: "failed to load config",
		},
		{
			name:            "fails when client creation fails",
			args:            []string{"my-release"},
			factoryErr:      errors.New("boom"),
			wantErr:         true,
			wantErrContains: "failed to create helm client",
		},
		{
			name:            "propagates uninstall error",
			args:            []string{"my-release"},
			uninstallErr:    errors.New("uninstall failed"),
			wantErr:         true,
			wantErrContains: "failed to uninstall release: uninstall failed",
			wantCalled:      true,
			wantRelease:     "my-release",
			wantNamespace:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optsUninstall.Namespace = tt.flagNamespace

			loadUninstallConfig = func() (*config.Config, error) {
				if tt.configErr != nil {
					return nil, tt.configErr
				}
				return &config.Config{Namespace: tt.configNamespace}, nil
			}

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
			if !tt.wantCalled {
				return
			}
			if fake.gotNamespace != tt.wantNamespace {
				t.Errorf("Namespace = %q, want %q", fake.gotNamespace, tt.wantNamespace)
			}
			if fake.gotReleaseName != tt.wantRelease {
				t.Errorf("ReleaseName = %q, want %q", fake.gotReleaseName, tt.wantRelease)
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
