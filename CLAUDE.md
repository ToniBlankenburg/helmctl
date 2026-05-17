# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Product Direction

helmctl turns this:
```
helm install my_app /c/dev/helm/target/my_app.tgz -f "/c/dev/helm/values/values-a.yaml" --namespace local-dev
```
into this:
```
helmctl install my_app
```

It is a **project-local configuration layer** on top of the Helm SDK. A `helmctl.yaml` file lives at the project root (committed to the repo), describes the local dev setup once, and from then on all commands resolve charts, values, and namespace from that config. No machine-specific paths, no memorized flags.

V1 command surface: `install`, `upgrade`, `uninstall`, `list`.

Future: a user-level `~/.helmctl/config.yaml` as a lower-priority override layer. Merge order: machine config → project config → CLI flags.

## Configuration

`helmctl.yaml` at the project root — all paths relative to the file's location:

```yaml
namespace: local-dev
charts_dir: target                     # target/my_app.tgz
values:
  - helm/values/values-a.yaml
```

Config loading belongs in `internal/config/`. Commands read config first, then apply CLI flag overrides on top.

## Build Commands

```bash
go build -o helmctl .                                                          # build binary
go test ./...                                                                   # run all tests
go test -run TestInstallCmd ./cmd                                               # run a focused test
go test -coverprofile=cmd.cover ./cmd                                          # coverage profile for cmd/
go test . -cover                                                                # root package with coverage
go test -tags integration -v -run TestIntegration_ManualFullLifecycle          # integration (requires cluster)
```

## Architecture

helmctl is a Cobra-based CLI wrapping the Helm SDK v4 directly — no separate `helm` binary required.

**Layer split:**
- `cmd/` — Cobra command definitions, flag wiring, argument validation. Handlers must stay thin.
- `internal/config/` — `helmctl.yaml` loading, path resolution relative to config file location.
- `internal/helmclient/` — all Helm SDK and Kubernetes interactions behind the `HelmClient` interface.

**Dependency injection pattern:** each command file declares a package-level factory variable (e.g. `newInstallHelmClient`) initialized to `helmclient.NewHelmClient`. Tests override this variable to inject a fake, keeping unit tests cluster-free. The same pattern applies to config loading.

**Release name:** the positional argument IS the release name. `install` and `upgrade` receive `my_app` and resolve the chart as `charts_dir/my_app.tgz`. `uninstall` uses the argument as the release name directly. `list` takes no positional arguments.

**Logging:** commands construct a `*log.Logger` via `log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)` and pass it into the client layer. Never instantiate loggers inside `internal/`; never use package-level `log.Printf` or `fmt.Printf` for operational output there.

**Error handling:** wrap with `fmt.Errorf("context: %w", err)` and return upward. Let Cobra print the final error.

## Testing Conventions

Use table-driven tests. Cover: argument validation, flag behavior, config loading, factory failures, propagated errors, and successful-path assertions on what was passed to the client layer. Avoid assertions on timestamps or exact log-line formatting.

## README Conventions

Keep the README user-facing. Lead with the before/after value proposition. Required sections: prerequisites, installation, configuration, command reference, contributing. Command reference must stay in sync with `cmd/`; do not document commands not yet wired into `rootCmd`.

## Roadmap

See [todos.md](todos.md) for the current focus and feature backlog.

## Agent Working Rules

- Do not edit product code unless explicitly requested.
- Before editing any file, re-read its current contents — this repo frequently reverts experimental instruction changes.
- Keep one canonical source for guidance; use redirects elsewhere.
