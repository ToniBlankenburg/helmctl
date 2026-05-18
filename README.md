# helmctl

helmctl is a configured shorthand for Helm — describe your local dev setup once, and manage releases with minimal commands.

**Before:**
```bash
helm install my_app /c/dev/helm/target/my_app.tgz -f "/c/dev/helm/values/values-a.yaml" --namespace local-dev
```

**After:**
```bash
helmctl install my_app
```

helmctl wraps the Helm SDK directly (no separate `helm` installation required) and reads a project-local `helmctl.yaml` to resolve charts, values files, and namespace automatically.

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go          | 1.26+   |
| kubectl     | any recent stable |
| A reachable Kubernetes cluster (kubeconfig configured) | — |

---

## Installation

```bash
git clone https://github.com/ToniBlankenburg/helmctl.git
cd helmctl
go install .
```

The binary will be placed in `$GOPATH/bin/helmctl` (ensure it is on your `$PATH`).

To build without installing:

```bash
go build -o helmctl .
```

---

## Configuration

Place a `helmctl.yaml` at your project root and commit it to the repo. All paths are relative to the config file.

```yaml
namespace: local-dev
charts_dir: target                     # resolves my_app → target/my_app.tgz
values:
  - helm/values/values-a.yaml
```

| Field | Description |
|-------|-------------|
| `namespace` | Default Kubernetes namespace for all commands |
| `charts_dir` | Directory containing packaged chart files (`.tgz`) |
| `values` | List of values files applied to every install/upgrade |

CLI flags always override config values (e.g. `-n` overrides `namespace`).

### Naming conventions

**Charts:** helmctl resolves `helmctl install <app>` to `charts_dir/<app>.tgz`. The app name you pass must match the chart filename exactly (without the `.tgz` extension).

```
charts_dir: target

target/
  my_app.tgz        →  helmctl install my_app
  payment_service.tgz  →  helmctl install payment_service
```

**Values files:** any `.yaml` files listed under `values` are applied to every install and upgrade in the order listed — later files override earlier ones for conflicting keys.

```yaml
values:
  - helm/values/common.yaml        # applied first (base)
  - helm/values/local-overrides.yaml  # applied second (wins on conflict)
```

---

## Commands

### install

Install a chart into the configured namespace.

```bash
helmctl install <app> [flags]
```

Resolves chart as `charts_dir/<app>.tgz`. Release name is `<app>`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | from config | Namespace to install into |

**Example**

```bash
helmctl install my_app
helmctl install my_app -n staging
```

---

### upgrade

Upgrade an existing release.

```bash
helmctl upgrade <app> [flags]
```

Resolves chart as `charts_dir/<app>.tgz`. Release name is `<app>`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | from config | Namespace of the release |

**Example**

```bash
helmctl upgrade my_app
```

---

### uninstall

Remove a release.

```bash
helmctl uninstall <release> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | from config | Namespace of the release |

**Example**

```bash
helmctl uninstall my_app
```

---

### list

List all installed releases in a namespace.

```bash
helmctl list [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | from config | Namespace to query |

**Example**

```bash
helmctl list
helmctl list -n staging
```

---

## Development

```bash
go test ./...                              # run all tests
go test -run TestInstallCmd ./cmd          # run a focused test
go test -coverprofile=cmd.cover ./cmd      # coverage profile
go tool cover -html=cmd.cover             # open coverage in browser
```

---

## Contributing

See [CLAUDE.md](CLAUDE.md) for architecture boundaries, coding conventions, and agent working rules.
