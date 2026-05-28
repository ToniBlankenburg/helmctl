// subcommand for helm upgrade
package cmd

import (
	"fmt"
	"log"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsUpgrade = &UpgradeOptions{}
var newUpgradeHelmClient = helmclient.NewHelmClient
var loadUpgradeConfig = config.FindAndLoad

func init() {
	upgradeCmd.Flags().StringVarP(&optsUpgrade.Namespace, "namespace", "n", "", "Namespace to upgrade the helm chart in (overrides helmctl.yaml)")
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a helm chart",
	Long:  "Upgrade a helm chart configured in helmctl.yaml.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("app name is required: usage: helmctl upgrade <app>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadUpgradeConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		appName := args[0]
		chartPath, err := resolveChartPath(cfg.ChartsDir, appName)
		if err != nil {
			return err
		}

		namespace := optsUpgrade.Namespace
		if namespace == "" {
			namespace = cfg.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}

		settings := cli.New()
		settings.SetNamespace(namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		upgradeHelmClient, err := newUpgradeHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}
		upgradeReq := helmclient.UpgradeRequest{
			ReleaseName:  appName,
			ChartRef:     chartPath,
			ChartVersion: "",
			ValuesFiles:  cfg.Values,
		}
		err = upgradeHelmClient.Upgrade(cmd.Context(), logger, settings, upgradeReq)
		if err != nil {
			return fmt.Errorf("failed to upgrade helm chart: %w", err)
		}
		return nil
	},
}

type UpgradeOptions struct {
	Namespace string
	ChartName string
}

func (o *UpgradeOptions) String() string {
	return fmt.Sprintf("Namespace: %s, ChartName: %s", o.Namespace, o.ChartName)
}
