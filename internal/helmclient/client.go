package helmclient

import (
	"context"
	"fmt"
	"log"
	"os"

	"dario.cat/mergo"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart"
	"helm.sh/helm/v4/pkg/chart/loader"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	"helm.sh/helm/v4/pkg/registry"
	yaml "gopkg.in/yaml.v3"
)

type HelmClient interface {
	Install(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, req InstallRequest) error
	ListCharts(settings *cli.EnvSettings) error
	Uninstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, releaseName string) error
	Upgrade(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, req UpgradeRequest) error
}

type InstallRequest struct {
	ReleaseName  string
	ChartRef     string
	ChartVersion string
	ValuesFiles  []string
}

type UpgradeRequest struct {
	ReleaseName  string
	ChartRef     string
	ChartVersion string
	ValuesFiles  []string
}

type realHelmClient struct {
	actionConfig   *action.Configuration
	logger         *log.Logger
	registryClient *registry.Client
	settings       *cli.EnvSettings
}

func NewHelmClient(settings *cli.EnvSettings, logger *log.Logger) (HelmClient, error) {

	initialActionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	registryClient, err := newRegistryClient(
		settings,
		false, // plainHTTP
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	return &realHelmClient{
		settings:       settings,
		actionConfig:   initialActionConfig,
		registryClient: registryClient,
		logger:         logger,
	}, nil
}

func (c *realHelmClient) Install(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, req InstallRequest) error {
	logger.Printf("Installing chart %s in namespace %s", req.ChartRef, settings.Namespace())

	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRunStrategy = "none"
	installClient.WaitStrategy = "watcher"
	installClient.ReleaseName = req.ReleaseName
	installClient.Namespace = settings.Namespace()
	installClient.Version = req.ChartVersion

	registryClient, err := newRegistryClient(
		settings,
		installClient.PlainHTTP)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}
	installClient.SetRegistryClient(registryClient)

	chartPath, err := installClient.ChartPathOptions.LocateChart(req.ChartRef, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	providers := getter.All(settings)

	charter, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	chartAccessor, err := chart.NewDefaultAccessor(charter)
	if err != nil {
		return fmt.Errorf("failed to create chart accessor: %w", err)
	}

	// check chart dependencies to make sure all are present in /charts
	if chartDependencies := chartAccessor.MetaDependencies(); chartDependencies != nil {
		if err := action.CheckDependencies(charter, chartDependencies); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %w", err)
			if !installClient.DependencyUpdate {
				return err
			}

			manager := &downloader.Manager{
				Out:              logger.Writer(),
				ChartPath:        chartPath,
				Keyring:          installClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
				RegistryClient:   installClient.GetRegistryClient(),
			}
			if err := manager.Update(); err != nil {
				return fmt.Errorf("failed to update chart dependencies: %w", err)
			}

			// reload the chart after updating dependencies
			if charter, err = loader.Load(chartPath); err != nil {
				return fmt.Errorf("failed to reload chart after updating repos: %w", err)
			}
		}
	}

	vals, err := mergeValuesFiles(req.ValuesFiles)
	if err != nil {
		return fmt.Errorf("failed to load values files: %w", err)
	}

	_, err = installClient.RunWithContext(ctx, charter, vals)
	if err != nil {
		return fmt.Errorf("failed to run install: %w", err)
	}

	logger.Printf("Chart %s installed successfully in namespace %s\n", req.ChartRef, settings.Namespace())

	return nil
}

func (c *realHelmClient) ListCharts(settings *cli.EnvSettings) error {
	c.logger.Printf("Listing charts in namespace %s", settings.Namespace())

	actionConfig, err := initActionConfigList(settings, c.logger, false)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	client := action.NewList(actionConfig)
	// Only list deployed
	client.Deployed = true
	results, err := client.Run()
	if err != nil {
		return fmt.Errorf("failed to list releases: %w", err)
	}

	c.logger.Printf("Listed %d releases in namespace %s", len(results), settings.Namespace())
	return nil

}

func (c *realHelmClient) Uninstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, releaseName string) error {
	logger.Printf("Uninstalling release %s from namespace %s", releaseName, settings.Namespace())
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	uninstallClient := action.NewUninstall(actionConfig)
	uninstallClient.WaitStrategy = "watcher"

	_, err = uninstallClient.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall release: %w", err)
	}

	logger.Printf("Release %s uninstalled successfully from namespace %s\n", releaseName, settings.Namespace())
	return nil
}

func (c *realHelmClient) Upgrade(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, req UpgradeRequest) error {
	logger.Printf("Upgrading release %s with chart %s in namespace %s", req.ReleaseName, req.ChartRef, settings.Namespace())

	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	upgradeClient := action.NewUpgrade(actionConfig)

	upgradeClient.DryRunStrategy = "none"
	upgradeClient.WaitStrategy = "watcher"
	upgradeClient.Namespace = settings.Namespace()
	upgradeClient.Version = req.ChartVersion

	registryClient, err := newRegistryClient(
		settings,
		upgradeClient.PlainHTTP)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}
	upgradeClient.SetRegistryClient(registryClient)

	chartPath, err := upgradeClient.ChartPathOptions.LocateChart(req.ChartRef, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	charter, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	vals, err := mergeValuesFiles(req.ValuesFiles)
	if err != nil {
		return fmt.Errorf("failed to load values files: %w", err)
	}

	_, err = upgradeClient.RunWithContext(ctx, req.ReleaseName, charter, vals)
	if err != nil {
		return fmt.Errorf("failed to run upgrade: %w", err)
	}

	logger.Printf("Release %s upgraded successfully with chart %s in namespace %s\n", req.ReleaseName, req.ChartRef, settings.Namespace())
	return nil
}

func initActionConfig(settings *cli.EnvSettings, logger *log.Logger) (*action.Configuration, error) {
	return initActionConfigList(settings, logger, false)
}

func initActionConfigList(settings *cli.EnvSettings, logger *log.Logger, allNamespaces bool) (*action.Configuration, error) {

	actionConfig := new(action.Configuration)

	namespace := func() string {
		// For list action, you can pass an empty string instead of settings.Namespace() to list
		// all namespaces
		if allNamespaces {
			return ""
		}
		return settings.Namespace()
	}()

	helmDriver := os.Getenv("HELM_DRIVER")
	if helmDriver == "" {
		helmDriver = "secrets"
	}
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		namespace,
		helmDriver); err != nil {
		return nil, err
	}

	return actionConfig, nil
}

func newRegistryClient(settings *cli.EnvSettings, plainHTTP bool) (*registry.Client, error) {

	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}

	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry client: %w", err)
	}

	return registryClient, nil
}

func mergeValuesFiles(files []string) (map[string]interface{}, error) {
	if len(files) == 0 {
		return nil, nil
	}
	merged := map[string]interface{}{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", f, err)
		}
		vals := map[string]interface{}{}
		if err := yaml.Unmarshal(data, &vals); err != nil {
			return nil, fmt.Errorf("failed to parse values file %s: %w", f, err)
		}
		if err := mergo.Merge(&merged, vals, mergo.WithOverride); err != nil {
			return nil, fmt.Errorf("failed to merge values from %s: %w", f, err)
		}
	}
	return merged, nil
}
