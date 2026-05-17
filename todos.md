# helmctl Roadmap

## Completed
- [x] Build/test baseline stabilized (`go test ./...` green)
- [x] CLI contract aligned for V1 (`install`/`upgrade`/`uninstall`/`list`)
- [x] Logging/error-handling cleanup in command/client layers
- [x] Command test hardening (arg validation, namespace propagation, wrapped errors, input assertions)
- [x] README synced with implemented command behavior
- [x] Manual integration-test strategy added (build-tag based)
- [x] Release name strategy decided: argument = release name; config resolves chart path

## Current Focus: Config System

The core value of helmctl — `helmctl install my_app` replacing the full helm command — depends entirely on this.

- [ ] Define `helmctl.yaml` schema (`namespace`, `charts_dir`, `values []string`)
- [ ] Implement config loader in `internal/config/` (find file, parse YAML, resolve paths relative to config location)
- [ ] Wire config into `install`: resolve chart as `charts_dir/<arg>.tgz`, apply `values`, apply `namespace`
- [ ] Wire config into `upgrade`: same resolution as install
- [ ] Wire config into `uninstall` and `list`: apply `namespace` from config
- [ ] CLI `-n` flag overrides config namespace
- [ ] Fail fast with a clear error if `helmctl.yaml` is not found
- [ ] Add table-driven tests for config loading (missing file, missing fields, relative path resolution)
- [ ] Update `cmd/install.go` to use argument as release name directly (remove `path.Base`)

## Values Handling (follows config system)

- [ ] Config `values` list applies to all commands globally
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
