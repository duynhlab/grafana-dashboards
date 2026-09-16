# Grafana Observability as Code

Strongly typed Grafana dashboards and alert rules built with Go and the
[Grafana Foundation SDK](https://github.com/grafana/grafana-foundation-sdk).

The project uses the Dashboard v2 model and targets Grafana 12 or newer.
Generated Kubernetes resources are published as an OCI artifact for GitOps
delivery with Flux and the Grafana Operator.

## Architecture

```text
Go source
  internal/dashboards
  internal/alerts
  internal/queries
  internal/panels
  internal/standards
        |
        v
  cmd/generate
        |
        +--> generated/       dashboard and alert artifacts
        |
        +--> deploy/          Grafana Operator resources
                 |
                 v
       Flux OCI artifact on GHCR
                 |
                 v
           Grafana Operator
                 |
                 v
              Grafana
```

Go files under `internal/` are the source of truth. Files under `generated/`
and `deploy/` are reproducible generator output and must not be edited by hand.

## Current resources

Grafana folders:

- `Kubernetes`
- `Databases`

Dashboards:

- `kubernetes-cluster-overview`
- `pg-io-waits` — tested on PostgreSQL 18 (`pg_stat_io`)

Alert rules:

- `kubernetes_crashlooping_pods`
- `kubernetes_pending_pods`
- `kubernetes_pvcs_at_risk`
- `postgres_backends_waiting`

## Requirements

- Go 1.26 or newer
- Grafana 12 or newer
- Grafana Operator for Kubernetes delivery
- Flux CLI for OCI publishing
- Kind, kubectl, and Helm for end-to-end tests

## Quick start

Generate all artifacts:

```bash
make generate
```

Run unit tests:

```bash
make test
```

Run the repository-wide 90% coverage gate:

```bash
make coverage
```

To inspect the coverage report directly:

```bash
go test ./... -coverpkg=./... -coverprofile=/tmp/grafana-dashboards-coverage.out
go tool cover -func=/tmp/grafana-dashboards-coverage.out
```

Run formatting, vet, coverage, generation, and the generated-diff check:

```bash
make validate
```

Format Go files:

```bash
make fmt
```

## Repository layout

```text
cmd/generate/                 generation entry point
internal/
  alerts/                     alert rule definitions
  dashboards/                 dashboard composition by domain
  generate/                   deterministic artifact renderer
  panels/                     reusable visualization builders
  queries/prometheus/         reusable PromQL by domain
  registry/                   dashboard and alert registration
  standards/                  folders, labels, datasources, time settings
generated/
  alerts/                     generated alert definitions
  dashboards/                 Dashboard v2 specs and manifests
deploy/
  manifests/                  Grafana Operator custom resources
  kustomization.yaml          OCI bundle entry point
test/
  dashboards/                 cross-resource contract tests
  e2e/kind/                   Grafana Operator smoke test
```

The project Agent Skill is located at
[`.cursor/skills/grafana-foundation-sdk/`](.cursor/skills/grafana-foundation-sdk/SKILL.md).
It contains the architecture, alerting, domain, testing, CI/CD, and end-to-end
conventions used by this repository.

## Development

Use the following workflow to onboard resources while preserving shared
standards, generated output, and dashboard-alert relationships.

### Adding a dashboard

1. Add or reuse queries under `internal/queries/`.
2. Compose panels and rows under `internal/dashboards/<domain>/`.
3. Register the dashboard and its domain folder in
   `internal/registry/dashboards.go`.
4. Add operational alert rules when appropriate.
5. Run `make validate`.

### Adding an alert rule

1. Reuse the dashboard's query semantics.
2. Define the rule under `internal/alerts/`.
3. Apply standard labels and annotations, including `dashboard_uid`.
4. Register it in `internal/registry/alerts.go`.
5. Run `make validate`.

### Adding a domain

Define the folder in `internal/standards/folders.go`, then register resources
with that folder. The generator discovers folders from the registries, so adding
a domain does not require generator changes.

## Kubernetes delivery

Build the complete manifest bundle:

```bash
kubectl kustomize deploy/
```

The bundle contains:

- one `GrafanaFolder` per domain
- one `GrafanaDashboard` per dashboard (`spec.oci`, JSON is not stored in etcd)
- one `GrafanaAlertRuleGroup` per alert domain

Grafana Operator 5.24+ fetches dashboard JSON from an OCI artifact. This
repository pins Helm chart **5.25.0** in e2e.

Run the local end-to-end smoke test:

```bash
make e2e-kind
```

The test creates a Kind cluster, installs Grafana Operator 5.25.0, deploys
Grafana 12, pushes generated `*.spec.json` into an in-cluster registry, applies
CRs that reference `spec.oci`, and verifies the resources.

## OCI publishing

GitHub Actions publishes two artifacts:

```text
# Dashboard JSON for GrafanaDashboard.spec.oci (oras)
ghcr.io/duynhlab/grafana-dashboards:latest
ghcr.io/duynhlab/grafana-dashboards:sha-<commit>

# Kustomize CRs for GitOps (flux)
ghcr.io/duynhlab/grafana-dashboards-as-code:latest
ghcr.io/duynhlab/grafana-dashboards-as-code:sha-<commit>
```

Generate CRs against a tag or digest:

```bash
OCI_REFERENCE=ghcr.io/duynhlab/grafana-dashboards:latest go run ./cmd/generate
```

For a private registry, also set `OCI_PULL_SECRET=ghcr-pull` and create a
`kubernetes.io/dockerconfigjson` Secret in the Grafana Operator namespace.

The `latest` tag follows the `as-code` branch. Commit tags provide immutable
references for GitOps consumers.

## CI

Pull requests run:

- formatting verification
- `go vet`
- repository-wide test coverage with a minimum of 90%
- deterministic artifact generation and diff verification
- Kustomize rendering checks for folders, dashboards, and alert groups
- Kind end-to-end smoke tests (operator fetch via `spec.oci`)

Pushes to `as-code` additionally publish dashboard JSON (oras) and the Flux
deploy bundle to GHCR.
