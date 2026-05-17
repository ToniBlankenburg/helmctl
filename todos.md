# helmctl Roadmap

## Completed
- [x] Build/test baseline stabilized (`go test ./...` green)
- [x] CLI contract aligned for V1 (`upgrade <chart> [-n namespace]`)
- [x] Logging/error-handling cleanup in command/client layers
- [x] Command test hardening (arg validation, namespace propagation, wrapped errors, input assertions)
- [x] README synced with implemented command behavior
- [x] Lightweight output-focused tests added
- [x] Manual integration-test strategy added (build-tag based)

## Current Focus (Post-MVP Enhancements)
- [ ] Decide final release-name strategy for `install`
  - Option A: derive from chart path base (current V1 behavior)
  - Option B: explicit release-name flag
- [ ] If keeping derived names, implement optional normalization for packaged charts
  - Example: `my-app-v2.tgz` -> `my-app`
- [ ] Align `cmd/install.go` mapping and tests to the selected strategy

## Values Handling (Next)
- [ ] Add repeatable `-f, --values <file>` support for `install` (optionally `upgrade` later)
- [ ] Add configurable auto-values directory (user-defined folder)
- [ ] On `install`, scan auto-values directory for values files matching release/chart (matching rules TBD)
- [ ] If match exists, append auto values file(s) automatically to Helm SDK values inputs
- [ ] Implement merge order so explicit values override auto defaults
  - Auto values first
  - User `-f/--values` files after
- [ ] Define missing file behavior
  - Missing explicit `-f/--values` file: fail fast with clear error
  - Missing auto-match file: continue without failure
- [ ] Add table-driven tests for values resolution and merge order
  - explicit `-f` only
  - auto values only
  - auto + explicit together (explicit wins)
  - missing explicit file vs missing auto match
- [ ] Define and document matching strategy
  - candidate keys: release name and chart name
  - candidate patterns: exact file names, glob, regex
- [ ] Prepare future `helmctlconfig` extension points (design note only)
  - Configurable default values directory
  - Shared/org values file paths and naming conventions

## Feature Backlog
- [ ] Add command to show current Kubernetes context and namespace metadata from kubeconfig
- [ ] Add `releases` command to list all releases in a given namespace
- [ ] Add `prune` command to uninstall all Helm releases in a given namespace

## Quality Backlog
- [ ] Add focused tests for future release-name normalization logic (`.tgz` + version suffix cases)
- [ ] Add command-contract regression tests for any new command (`releases`, `prune`, metadata)

## Suggested Next Execution Order
1. Finalize `install` release-name strategy decision
2. Implement and test strategy in `cmd/install.go`
3. Implement values handling (`-f/--values` + auto-values directory lookup)
4. Add kubeconfig metadata command
5. Add `releases` command
6. Add `prune` command
