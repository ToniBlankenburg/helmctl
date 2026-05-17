# AGENTS for helmctl

## Product Direction
- Build a production-focused Go CLI that abstracts Helm/Kubernetes workflows for local cloud development.
- Treat helmctl as a CLI-first deployment tool ("argocd-like" operator workflow, but local/developer-centric).
- Prioritize V1 command surface: `install`, `upgrade`, `uninstall`, `list`.

## Current State (May 2026)
- Implemented command packages: `cmd/install.go`, `cmd/list.go`, `cmd/upgrade.go`, `cmd/uninstall.go`, `cmd/root.go`.
- Helm integration entrypoint: `internal/helmclient/client.go`.
- V1 command surface is wired in `rootCmd`: `install`, `list`, `upgrade`, `uninstall`.
- Release name behavior for V1:
  - `install` and `upgrade`: derive release name from `path.Base(<chart>)`.
  - `uninstall`: release name comes from the positional argument.
- `list` does not accept positional arguments.

## Architecture Boundaries
- CLI orchestration belongs in `cmd/` (Cobra commands, flags, argument validation).
- Helm/Kubernetes SDK interactions belong in `internal/*/`.
- Keep command handlers thin and push behavior behind testable interfaces.
- Wire dependencies through factory variables (current pattern: `newInstallHelmClient`, `newListHelmClient`) to keep tests isolated.

## Build and Test Commands
- Build binary: `go build -o helmctl .`
- Run all tests: `go test ./...`
- Run root package with coverage: `go test . -cover`
- Run command tests with coverage profile: `go test -coverprofile=cmd.cover ./cmd`
- Run a focused test: `go test -run TestInstallCmd ./cmd`
- Run manual integration lifecycle tests (requires cluster/helm repo setup): `go test -tags integration -v -run TestIntegration_ManualFullLifecycle`

## Test and Coverage Expectations
- Goal is practical confidence: "as high as possible, as low as needed".
- use when ever possible table-driven tests for command behavior and edge cases.
- Add tests for:
  - argument validation and flag behavior
  - SDK/client factory failures
  - propagated Helm/Kubernetes errors
  - successful execution paths with assertions on inputs passed to client layer
- Avoid brittle tests coupled to timestamps or full log-line formatting.

## Logging
- Use stdlib `log.Logger` (no third-party logging library).
- Commands construct the logger via `log.New(cmd.ErrOrStderr(), "helmctl: ", log.LstdFlags)` and pass it into the client layer. Never create a logger inside `internal/`.
- Always write through the injected `*log.Logger`; never call the package-level `log.Printf` or `fmt.Printf` for operational output inside `internal/helmclient/`.
- Use `logger.Printf` for informational progress messages (e.g. "Installing chart %s…").
- Use `logger.Printf` (not `cmd.PrintErrln`) for error context before returning an error; let Cobra print the final error to the user.
- Do not add debug-level prints unless a test explicitly requires observable output.
- Keep errors wrapped with context via `fmt.Errorf("...: %w", err)` and return them upward.

## README
- Keep the README user-facing and production-oriented; 
- Required sections: project description, prerequisites (Go version, Helm, kubectl), installation, command reference, and contributing.
- Command reference must stay in sync with the actual `cmd/` implementations; do not document commands that are not yet wired into `rootCmd`.
- Code examples in the README must use the `helmctl` binary name and reflect real flag names (verify against `cmd/*.go` before updating).
- Do not duplicate information already in `AGENTS.md`; link to it for contributor/agent guidance instead.

## Agent Working Rules
- Do not edit product code unless explicitly requested by the user.
- When updating chat customization files, keep guidance concise and action-oriented.
- If instructions exist in multiple locations, keep one canonical source and use redirects elsewhere.
- Before editing, re-read current file contents because this repo frequently reverts experimental instruction changes.

## Recommended Next Customizations
- Add a prompt in `.github/prompts/` for "add new command scaffolding" (flags, tests, wiring checklist).
- Add file-scoped instructions in `.github/instructions/` for `cmd/**` and `internal/helmclient/**` with layer-specific conventions.
- Add a skill for release-readiness checks (coverage, smoke tests, command contract checks).
