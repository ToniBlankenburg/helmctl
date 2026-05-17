# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Product Direction

helmctl is a production-focused Go CLI that abstracts Helm/Kubernetes workflows for local cloud development — an "argocd-like" operator workflow but developer-centric. V1 command surface: `install`, `upgrade`, `uninstall`, `list`.

## Commands

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
- `internal/helmclient/` — all Helm SDK and Kubernetes interactions behind the `HelmClient` interface.

**Dependency injection pattern:** each command file declares a package-level factory variable (e.g. `newInstallHelmClient`) initialized to `helmclient.NewHelmClient`. Tests override this variable to inject a fake, keeping unit tests cluster-free.

**Release name convention (V1):** `install` and `upgrade` derive the release name via `path.Base(<chart>)` (e.g. `bitnami/nginx` → `nginx`). `uninstall` takes the release name directly as a positional argument. `list` takes no positional arguments.

**Logging:** commands construct a `*log.Logger` via `log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)` and pass it into the client layer. Never instantiate loggers inside `internal/`; never use package-level `log.Printf` or `fmt.Printf` for operational output there. Use `logger.Printf` for progress messages.

**Error handling:** wrap with `fmt.Errorf("context: %w", err)` and return upward. Let Cobra print the final error.

## Testing Conventions

Use table-driven tests. Cover: argument validation, flag behavior, factory failures, propagated errors, and successful-path assertions on what was passed to the client layer. Avoid assertions on timestamps or exact log-line formatting.

## README Conventions

Keep the README user-facing. Required sections: project description, prerequisites, installation, command reference, contributing. The command reference must stay in sync with `cmd/`; do not document commands not yet wired into `rootCmd`. Code examples must use `helmctl` and real flag names.

## Roadmap

See [todos.md](todos.md) for the current focus and feature backlog. Key pending decisions: final release-name strategy for `install`, values handling (`-f/--values` + auto-values directory).

## Agent Working Rules

- Do not edit product code unless explicitly requested.
- Before editing any file, re-read its current contents — this repo frequently reverts experimental instruction changes.
- Keep one canonical source for guidance; use redirects elsewhere.
