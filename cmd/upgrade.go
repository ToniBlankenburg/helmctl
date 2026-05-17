// subcommand for helm upgrade
package cmd

import (
	"fmt"
	"log"
	"path"

	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsUpgrade = &UpgradeOptions{}
var newUpgradeHelmClient = helmclient.NewHelmClient

func init() {
	upgradeCmd.Flags().StringVarP(&optsUpgrade.Namespace, "namespace", "n", "default", "Namespace to upgrade the helm chart in")
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a helm chart",
	Long:  "Upgrade a helm chart with specified parameters.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("chart name is required: usage: helmctl upgrade <chart>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := cli.New()
		settings.SetNamespace(optsUpgrade.Namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		upgradeHelmClient, err := newUpgradeHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}
		upgradeReq := helmclient.UpgradeRequest{
			ReleaseName:   path.Base(args[0]),
			ChartRef:      args[0],
			ChartVersion:  "",
			ReleaseValues: nil,
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
