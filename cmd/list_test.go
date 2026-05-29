// test fur subcommands of list command
package cmd

import (
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

type fakeListHelmClient struct {
	err          error
	called       bool
	gotNamespace string
}

func (f *fakeListHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeListHelmClient) ListCharts(settings *cli.EnvSettings) ([]release.Accessor, error) {
	f.called = true
	f.gotNamespace = settings.Namespace()
	return nil, f.err
}

func (f *fakeListHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeListHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func (f *fakeListHelmClient) GetReleaseManifest(_ *cli.EnvSettings, _ string) (string, error) {
	return "", nil
}

func (f *fakeListHelmClient) RenderManifest(_ context.Context, _ *cli.EnvSettings, _ helmclient.RenderRequest) (string, error) {
	return "", nil
}

func TestListCmd(t *testing.T) {
	originalFactory := newListHelmClient
	originalConfig := loadListConfig
	defer func() {
		newListHelmClient = originalFactory
		loadListConfig = originalConfig
		optsList.Namespace = ""
	}()

	tests := []struct {
		name            string
		args            []string
		flagNamespace   string
		configNamespace string
		configErr       error
		factoryErr      error
		listErr         error
		wantErr         bool
		wantCalled      bool
		wantNamespace   string
		wantErrContains string
	}{
		{
			name:            "lists using config namespace",
			args:            []string{"list"},
			configNamespace: "local-dev",
			wantCalled:      true,
			wantNamespace:   "local-dev",
		},
		{
			name:            "flag namespace overrides config namespace",
			args:            []string{"list"},
			flagNamespace:   "staging",
			configNamespace: "local-dev",
			wantCalled:      true,
			wantNamespace:   "staging",
		},
		{
			name:          "falls back to default namespace when config has none",
			args:          []string{"list"},
			wantCalled:    true,
			wantNamespace: "default",
		},
		{
			name:            "fails with unexpected positional argument",
			args:            []string{"list", "extra"},
			wantErr:         true,
			wantErrContains: "list does not accept arguments",
		},
		{
			name:            "fails when config loading fails",
			args:            []string{"list"},
			configErr:       errors.New("no config found"),
			wantErr:         true,
			wantErrContains: "failed to load config",
		},
		{
			name:            "fails when client creation fails",
			args:            []string{"list"},
			factoryErr:      errors.New("factory failed"),
			wantErr:         true,
			wantErrContains: "failed to create helm client",
		},
		{
			name:            "propagates list error",
			args:            []string{"list"},
			configNamespace: "local-dev",
			listErr:         errors.New("list failed"),
			wantErr:         true,
			wantCalled:      true,
			wantNamespace:   "local-dev",
			wantErrContains: "failed to list helm charts: list failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optsList.Namespace = tt.flagNamespace

			loadListConfig = func() (*config.Config, error) {
				if tt.configErr != nil {
					return nil, tt.configErr
				}
				return &config.Config{Namespace: tt.configNamespace}, nil
			}

			fake := &fakeListHelmClient{err: tt.listErr}
			newListHelmClient = func(_ *cli.EnvSettings, _ *log.Logger) (helmclient.HelmClient, error) {
				if tt.factoryErr != nil {
					return nil, tt.factoryErr
				}
				return fake, nil
			}

			cmd := &cobra.Command{}
			cmd.AddCommand(listCmd)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
			if fake.called != tt.wantCalled {
				t.Fatalf("expected list called=%v, got %v", tt.wantCalled, fake.called)
			}
			if tt.wantCalled && fake.gotNamespace != tt.wantNamespace {
				t.Fatalf("expected namespace %q, got %q", tt.wantNamespace, fake.gotNamespace)
			}
		})
	}
}

func TestListCmdMetadata(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("listCmd.Use = %q, want %q", listCmd.Use, "list")
	}
	if listCmd.Short != "List installed helm charts" {
		t.Errorf("listCmd.Short = %q, want %q", listCmd.Short, "List installed helm charts")
	}
	if listCmd.Long != "List installed helm charts in the configured namespace." {
		t.Errorf("listCmd.Long = %q, want %q", listCmd.Long, "List installed helm charts in the configured namespace.")
	}
}

func TestListOptionsString(t *testing.T) {
	o := &ListOptions{Namespace: "ns"}
	if got := o.String(); got != "namespace=ns" {
		t.Fatalf("unexpected options string: %s", got)
	}
}
