// subcommand for helm install
package cmd

import (
	"fmt"
	"log"

	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var opts = &InstallOptions{}
var newInstallHelmClient = helmclient.NewHelmClient

func init() {
	installCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "default", "Namespace to install the helm chart into")

}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a helm chart",
	Long:  "Install a helm chart with specified parameters.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("chart name is required: usage: helmctl install <chart>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := cli.New()
		settings.SetNamespace(opts.Namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		installHelmClient, err := newInstallHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}
		installReq := helmclient.InstallRequest{
			ReleaseName:   opts.ChartName,
			ChartRef:      args[0],
			ChartVersion:  "",
			ReleaseValues: nil,
		}
		err = installHelmClient.Install(cmd.Context(), logger, settings, installReq)
		if err != nil {
			return err
		}


		return nil
	},
}

type InstallOptions struct {
	Namespace string
	ChartName string
}

func (o *InstallOptions) String() string {
	return fmt.Sprintf("namespace=%s, chartName=%s", o.Namespace, o.ChartName)
}
