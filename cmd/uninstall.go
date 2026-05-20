// subcommand for helm uninstall
package cmd

import (
	"fmt"
	"log"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsUninstall = &UninstallOptions{}
var newUninstallHelmClient = helmclient.NewHelmClient
var loadUninstallConfig = config.FindAndLoad

func init() {
	uninstallCmd.Flags().StringVarP(&optsUninstall.Namespace, "namespace", "n", "", "Namespace to uninstall from (overrides helmctl.yaml)")
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall a helm chart",
	Long:  "Uninstall a helm chart configured in helmctl.yaml.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("release name is required: usage: helmctl uninstall <release>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadUninstallConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		namespace := optsUninstall.Namespace
		if namespace == "" {
			namespace = cfg.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}

		settings := cli.New()
		settings.SetNamespace(namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		uninstallHelmClient, err := newUninstallHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}

		releaseName := args[0]
		if err := uninstallHelmClient.Uninstall(cmd.Context(), logger, settings, releaseName); err != nil {
			return fmt.Errorf("failed to uninstall release: %w", err)
		}

		return nil
	},
}

type UninstallOptions struct {
	Namespace   string
	ReleaseName string
}

func (o *UninstallOptions) String() string {
	return fmt.Sprintf("Namespace: %s, ReleaseName: %s", o.Namespace, o.ReleaseName)
}
