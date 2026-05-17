// test fur subcommands of list command
package cmd

import (
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/helmctl/internal/helmclient"

	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

type fakeListHelmClient struct {
	err          error
	called       bool
	gotNamespace string
}

func (f *fakeListHelmClient) Install(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.InstallRequest) error {
	return nil
}

func (f *fakeListHelmClient) ListCharts(settings *cli.EnvSettings) error {
	f.called = true
	f.gotNamespace = settings.Namespace()
	return f.err
}

func (f *fakeListHelmClient) Upgrade(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ helmclient.UpgradeRequest) error {
	return nil
}

func (f *fakeListHelmClient) Uninstall(_ context.Context, _ *log.Logger, _ *cli.EnvSettings, _ string) error {
	return nil
}

func TestListCmd(t *testing.T) {
	originalFactory := newListHelmClient
	defer func() { newListHelmClient = originalFactory }()

	tests := []struct {
		name            string
		args            []string
		namespace       string
		factoryErr      error
		listErr         error
		wantErr         bool
		wantCalled      bool
		wantCalledNS    string
		wantErrContains string
	}{
		{
			name:         "list command runs successfully with default namespace",
			args:         []string{"list"},
			namespace:    "default",
			wantErr:      false,
			wantCalled:   true,
			wantCalledNS: "default",
		},
		{
			name:            "list command fails with unexpected positional argument",
			args:            []string{"list", "extra"},
			namespace:       "default",
			wantErr:         true,
			wantCalled:      false,
			wantErrContains: "list does not accept arguments",
		},
		{
			name:            "list command fails when client creation fails",
			args:            []string{"list"},
			namespace:       "default",
			factoryErr:      errors.New("factory failed"),
			wantErr:         true,
			wantCalled:      false,
			wantErrContains: "failed to create helm client",
		},
		{
			name:            "list command propagates list error",
			args:            []string{"list"},
			namespace:       "demo",
			listErr:         errors.New("list failed"),
			wantErr:         true,
			wantCalled:      true,
			wantCalledNS:    "demo",
			wantErrContains: "failed to list helm charts: list failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optsList.Namespace = tt.namespace
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
			if tt.wantCalled && fake.gotNamespace != tt.wantCalledNS {
				t.Fatalf("expected namespace %q, got %q", tt.wantCalledNS, fake.gotNamespace)
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
	if listCmd.Long != "List installed helm charts in the specified namespace." {
		t.Errorf("listCmd.Long = %q, want %q", listCmd.Long, "List installed helm charts in the specified namespace.")
	}
}

func TestListOptionsString(t *testing.T) {
	o := &ListOptions{Namespace: "ns"}
	if got := o.String(); got != "namespace=ns" {
		t.Fatalf("unexpected options string: %s", got)
	}
}
