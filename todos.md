# helmctl Roadmap

## Completed
- [x] Build/test baseline stabilized (`go test ./...` green)
- [x] CLI contract aligned for V1 (`install`/`upgrade`/`uninstall`/`list`)
- [x] Logging/error-handling cleanup in command/client layers
- [x] Command test hardening (arg validation, namespace propagation, wrapped errors, input assertions)
- [x] README synced with implemented command behavior
- [x] Manual integration-test strategy added (build-tag based)
- [x] Release name strategy decided: argument = release name; config resolves chart path
- [x] `helmctl.yaml` schema defined (`namespace`, `charts_dir`, `values []string`)
- [x] Config loader implemented in `internal/config/` (parse YAML, resolve paths relative to config file location)
- [x] `FindFrom`/`Find`/`FindAndLoad` added — walks up from CWD to locate `helmctl.yaml`
- [x] `install` wired to config: chart resolved as `charts_dir/<app>.tgz`, values from config applied, namespace from config with `-n` override, fails fast if config not found
- [x] `mergeValuesFiles` added to helmclient — loads and merges YAML values files in order (last wins)
- [x] `install` fails fast with clear error when chart `.tgz` not found on disk (before calling Helm SDK)

## Current Focus: Wire Config into Remaining Commands

- [x] Wire config into `upgrade`: same pattern as install
- [x] Wire config into `uninstall`: namespace from config, `-n` overrides
- [ ] Wire config into `list`: namespace from config, `-n` overrides

## Values Handling (follows config system)

- [ ] Config `values` list applies to all commands globally
- [ ] having one Umbrellavalues file like dev_values.yaml to make global suite like cofigs adjustments (e.g. enabel.sso = true)
- [ ] CLI `-f/--values <file>` appends on top of config values (explicit wins, applied after)
- [ ] Fail fast if an explicitly passed `-f` file does not exist
- [ ] Silently skip a configured values file only if documented behavior
- [ ] Add table-driven tests: config values only, CLI values only, both together, missing file

## Feature Backlog

- [ ] Per-app config overrides in `helmctl.yaml` (app-specific namespace, values, chart path)
- [ ] User-level `~/.helmctl/config.yaml` as lower-priority layer (merge order: machine → project → CLI)
- [ ] `list` command produces structured stdout output (table of release name, namespace, status, chart version)
- [ ] Add `releases` command (alias or replacement for `list` with richer output)
- [ ] Add `prune` command to uninstall all releases in a namespace
- [ ] Add command to show current kubeconfig context and active namespace

## Quality Backlog

- [ ] Replace `internal/helmclient/client_test.go` source-text checks with behavioral contract tests
- [ ] Fix inconsistent logger/settings injection in `HelmClient` (constructor vs method params)
- [ ] Fix `opts` package-level global in `cmd/install.go` to prevent test pollution

## Suggested Next Execution Order
1. Config loader (`internal/config/`)
2. Wire config into `install` (chart resolution + values + namespace)
3. Wire config into remaining commands
4. CLI `-f` values flag
5. Per-app config overrides
6. Machine-level config layer
