# AGENTS for helmctl

## Product Direction
- Build a production-focused Go CLI that abstracts Helm/Kubernetes workflows for local cloud development.
- Treat helmctl as a CLI-first deployment tool ("argocd-like" operator workflow, but local/developer-centric).
- Prioritize V1 command surface: `install`, `upgrade`, `uninstall`, `list`.

## Current State (May 2026)
- Implemented command packages: `cmd/install.go`, `cmd/list.go`, `cmd/root.go`.
- Helm integration entrypoint: `internal/helmclient/client.go`.
- `upgrade` and `uninstall` commands are not implemented yet.
- Existing help/output still contains debug-oriented messages; avoid adding new debug prints unless a test requires them.

## Architecture Boundaries
- CLI orchestration belongs in `cmd/` (Cobra commands, flags, argument validation).
- Helm/Kubernetes SDK interactions belong in `internal/*/`.
- Keep command handlers thin and push behavior behind testable interfaces.
- Wire dependencies through factory variables (current pattern: `newInstallHelmClient`, `newListHelmClient`) to keep tests isolated.

## Build and Test Commands
- Run all tests: `go test ./...`
- Run root package with coverage: `go test . -cover`
- Run command tests with coverage profile: `go test -coverprofile=cmd.cover ./cmd`
- Run a focused test: `go test -run TestInstallCmd ./cmd`

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
- Known tech-debt: several `fmt.Printf` and bare `log.Printf` calls exist in `internal/helmclient/client.go`; replace them with the injected logger when touching those functions.

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
