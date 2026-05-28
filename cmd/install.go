package cmd

import (
	"fmt"
	"log"

	"github.com/ToniBlankenburg/helmctl/internal/config"
	"github.com/ToniBlankenburg/helmctl/internal/helmclient"
	"github.com/ToniBlankenburg/helmctl/internal/kube"
	"github.com/spf13/cobra"
	"helm.sh/helm/v4/pkg/cli"
)

var opts = &InstallOptions{}
var newInstallHelmClient = helmclient.NewHelmClient
var loadInstallConfig = config.FindAndLoad
var newKubeClient = kube.NewKubeClient

func init() {
	installCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "Namespace to install into (overrides helmctl.yaml)")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a helm chart",
	Long:  "Install a helm chart configured in helmctl.yaml.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("app name is required: usage: helmctl install <app>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadInstallConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		appName := args[0]
		chartPath, err := resolveChartPath(cfg.ChartsDir, appName)
		if err != nil {
			return err
		}

		namespace := opts.Namespace
		if namespace == "" {
			namespace = cfg.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}

		settings := cli.New()
		settings.SetNamespace(namespace)

		logger := log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)
		installHelmClient, err := newInstallHelmClient(settings, logger)
		if err != nil {
			return fmt.Errorf("failed to create helm client: %w", err)
		}

		kubeClient, err := newKubeClient(logger)
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}
		if err := kubeClient.EnsureNamespace(cmd.Context(), namespace); err != nil {
			return fmt.Errorf("failed to ensure namespace: %w", err)
		}
		installReq := helmclient.InstallRequest{
			ReleaseName:  appName,
			ChartRef:     chartPath,
			ChartVersion: "",
			ValuesFiles:  cfg.Values,
		}
		if err := installHelmClient.Install(cmd.Context(), logger, settings, installReq); err != nil {
			return fmt.Errorf("failed to install helm chart: %w", err)
		}

		return nil
	},
}

type InstallOptions struct {
	Namespace string
}
