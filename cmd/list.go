// subcommand for helm list
package cmd

import (
	"fmt"
	"log"
	"text/tabwriter"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsList = &ListOptions{}
var newListHelmClient = helmclient.NewHelmClient
var loadListConfig = config.FindAndLoad

func init() {
	listCmd.Flags().StringVarP(&optsList.Namespace, "namespace", "n", "", "Namespace to list releases from (overrides helmctl.yaml)")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed helm charts",
	Long:  "List installed helm charts in the configured namespace.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("list does not accept arguments: usage: helmctl list [-n namespace]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadListConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		namespace := optsList.Namespace
		if namespace == "" {
			namespace = cfg.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}

		settings := cli.New()
		settings.SetNamespace(namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		listHelmClient, err := newListHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}

		releases, err := listHelmClient.ListCharts(settings)
		if err != nil {
			return fmt.Errorf("failed to list helm charts: %w", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tNAMESPACE\tSTATUS\tREVISION")
		for _, r := range releases {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.Name(), r.Namespace(), r.Status(), r.Version())
		}
		w.Flush()
		return nil
	},
}

type ListOptions struct {
	Namespace string
}

func (o *ListOptions) String() string {
	return fmt.Sprintf("namespace=%s", o.Namespace)
}
