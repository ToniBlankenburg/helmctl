package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var optsDiff = &DiffOptions{}
var newDiffHelmClient = helmclient.NewHelmClient
var loadDiffConfig = config.FindAndLoad

func init() {
	diffCmd.Flags().StringVarP(&optsDiff.Namespace, "namespace", "n", "", "Namespace (overrides helmctl.yaml)")
}

var diffCmd = &cobra.Command{
	Use:   "diff <app>",
	Short: "Diff deployed manifest against locally rendered chart",
	Long:  "Show a unified diff between the manifest of the running release and the locally rendered chart.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("app name is required: usage: helmctl diff <app>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadDiffConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		appName := args[0]
		chartPath, err := resolveChartPath(cfg.ChartsDir, appName)
		if err != nil {
			return err
		}

		namespace := optsDiff.Namespace
		if namespace == "" {
			namespace = cfg.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}

		settings := cli.New()
		settings.SetNamespace(namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		helmClient, err := newDiffHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}

		deployed, err := helmClient.GetReleaseManifest(settings, appName)
		if err != nil {
			return fmt.Errorf("failed to get deployed manifest: %w", err)
		}

		local, err := helmClient.RenderManifest(cmd.Context(), settings, helmclient.RenderRequest{
			ReleaseName: appName,
			ChartRef:    chartPath,
			ValuesFiles: cfg.Values,
		})
		if err != nil {
			return fmt.Errorf("failed to render local manifest: %w", err)
		}

		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(strings.TrimSpace(deployed) + "\n"),
			B:        difflib.SplitLines(strings.TrimSpace(local) + "\n"),
			FromFile: "deployed",
			ToFile:   "local",
			Context:  3,
		}
		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return fmt.Errorf("failed to compute diff: %w", err)
		}

		if text == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No differences found.")
			return nil
		}
		fmt.Fprint(cmd.OutOrStdout(), text)
		return nil
	},
}

type DiffOptions struct {
	Namespace string
}
