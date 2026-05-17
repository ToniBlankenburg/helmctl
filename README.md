# helmctl

A CLI tool that abstracts Helm/Kubernetes workflows for local cloud development.
helmctl wraps the Helm SDK to provide a concise, scriptable interface for managing releases — designed for developer-centric workflows.

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go          | 1.26+   |
| kubectl     | any recent stable |
| A reachable Kubernetes cluster (kubeconfig configured) | — |

> helmctl uses the Helm SDK directly; a separate Helm installation is **not** required.

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

## Commands

### install

Install a chart into a namespace.

```bash
helmctl install <chart> [flags]
```

For V1, release name is derived from the chart argument using the base path.
Examples: `bitnami/nginx` -> `nginx`, `podinfo/podinfo` -> `podinfo`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | `default` | Namespace to install into |

**Example**

```bash
helmctl install bitnami/nginx -n production
```

---

### upgrade

Upgrade an existing release with a new chart version.

```bash
helmctl upgrade <chart> [flags]
```

For V1, release name is derived from the chart argument using the base path.
Examples: `bitnami/nginx` -> `nginx`, `podinfo/podinfo` -> `podinfo`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | `default` | Namespace of the release |

**Example**

```bash
helmctl upgrade bitnami/nginx -n production
```

---

### uninstall

Remove a release from a namespace.

```bash
helmctl uninstall <release> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | `default` | Namespace of the release |

**Example**

```bash
helmctl uninstall nginx -n production
```

---

### list

List all installed releases in a namespace.

```bash
helmctl list [flags]
```

`list` does not accept positional arguments.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | `default` | Namespace to query |

**Example**

```bash
helmctl list -n production
```

---

## Development

### Run all tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test -coverprofile=cmd.cover ./cmd
go tool cover -html=cmd.cover
```

### Run a focused test

```bash
go test -run TestInstallCmd ./cmd
```

---

## Project Structure

```
helmctl/
├── main.go                        # Entry point
├── go.mod
├── cmd/                           # Cobra command definitions and flag wiring
│   ├── root.go
│   ├── install.go
│   ├── list.go
│   ├── upgrade.go
│   └── uninstall.go
└── internal/
    └── helmclient/
        └── client.go              # HelmClient interface + Helm SDK integration
```

---

## Contributing

See [AGENTS.md](AGENTS.md) for architecture boundaries, coding conventions, and agent working rules.

