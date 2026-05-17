// subcommand for helm list
package cmd

import (
	"fmt"
	"log"

	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsList = &ListOptions{}
var newListHelmClient = helmclient.NewHelmClient

func init() {
	listCmd.Flags().StringVarP(&optsList.Namespace, "namespace", "n", "default", "Namespace to list the helm charts from")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed helm charts",
	Long:  "List installed helm charts in the specified namespace.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("list does not accept arguments: usage: helmctl list [-n namespace]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := cli.New()
		settings.SetNamespace(optsList.Namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		listHelmClient, err := newListHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}

		err = listHelmClient.ListCharts(settings)
		if err != nil {
			return fmt.Errorf("failed to list helm charts: %w", err)
		}
		return nil
	},
}

type ListOptions struct {
	Namespace string
}

func (o *ListOptions) String() string {
	return fmt.Sprintf("namespace=%s", o.Namespace)
}
