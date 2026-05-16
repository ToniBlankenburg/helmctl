# helmctl Open TODOs

## 1. Get Build Green (Highest Priority) ✅
- [x] Fix upgrade command contract and tests (release name = chart name)
- [x] Fix undefined `opts` symbol in install_test.go (line 99)
- [x] Fix `ReleaseName` struct field in uninstall_test.go (line 138)
- [x] Fix `ReleaseName` struct field in upgrade_test.go (line 155)
- [x] Run full suite: `go test ./...` and ensure all packages pass.
- **STATUS**: All tests passing ✅

## 2. Align CLI Contracts (Implementation vs Tests)
- [x] **DECIDED**: `upgrade` contract: `helmctl upgrade <chart> [-n namespace]`
  - Chart name is mandatory positional argument
  - Release name = chart name (V1 MVP simple approach)
- [x] Update `cmd/upgrade.go` to set ReleaseName from args[0] (the chart).
- [x] Update `cmd/upgrade_test.go` test cases to expect 1 arg only; fix expected error messages.

## 3. Logging and Error Handling Cleanup ✅
- [x] Remove debug help output from root command help hook.
- [x] Replace root command execution `fmt.Printf` error output with Cobra-consistent behavior.
- [x] Remove `DEBUG` print from install command execution path.
- [x] Replace list command `fmt.Printf` error output with Cobra-consistent behavior.
- [x] In `internal/helmclient/client.go`, replace operational `fmt.Printf`/package-level `log.Printf` with injected logger usage.
- [x] Remove `os.Exit(...)` from internal client layer; return errors upward instead.
- [x] Updated tests to not expect debug output.
- **STATUS**: All logging cleaned up for production ✅

## 4. Test Hardening
- [ ] Add/adjust table-driven tests for argument validation edge cases (too few/too many args).
- [ ] Add/adjust tests for namespace propagation to Helm client requests.
- [ ] Add/adjust tests asserting wrapped error context from factory/client failures.
- [ ] Add/adjust tests for successful command paths with expected client inputs.

## 5. Docs Consistency
- [ ] Ensure README command examples exactly match final CLI contracts after implementation changes.
- [ ] Verify README flags/defaults are in sync with `cmd/*.go`.

## 6. Optional Quality Improvements
- [ ] Add lightweight output-focused tests for user-facing command output where stable and useful.
- [ ] Consider a small integration test strategy around Helm interaction boundaries (without making tests brittle).

## 7. Future Enhancements (Post-MVP)
- [ ] Implement release-name derivation from chart name (strip version suffix and `.tgz`)
  - Example: `my-app-v2.tgz` → release name `my-app`
- [ ] Decide and document release-name behavior for `install`:
  - release from positional argument, or
  - release from explicit flag.
- [ ] Align `cmd/install.go` request mapping to chosen release-name strategy.
- [ ] Update command tests to verify release/chart mapping and arg validation.

## Suggested Execution Order
1. Build green fixes
2. CLI contract decisions + command/test alignment
3. Logging/error cleanup
4. Test hardening
5. README sync
