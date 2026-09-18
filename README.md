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
- `Observability`
- `Microservices`

Dashboards:

- `kubernetes-cluster-overview`
- `pg-io-waits` — tested on PostgreSQL 18 (`pg_stat_io`)
- `pg-maintenance` — CNPG locks, checkpointer, autovacuum/bloat, long transactions
- `pg-query-performance` — CNPG `pg_stat_statements` throughput, latency, top statements
- `pg-exporter-instance` — Pigsty PGRDS instance board (uses the Pigsty
  `postgres_exporter` metric model: `pg_*` metrics and `pg:*` recording rules
  with `ins`/`cls` labels, not `cnpg_*`)
- `pgdog` — PGDog connection pooler (ported from Grafana.com dashboard 24583)
- `temporal-worker` — Temporal workflow/activity RED metrics
- `microservices-monitoring-001-otel` — Go services RED, runtime, gRPC east-west and
  otelpgx pool metrics, using the OpenTelemetry semantic conventions
  (`http_server_*`, `rpc_*`, `go_*`, `db_client_*`) keyed on `service_name`
- `business-otel` — per-domain business KPIs (payments, orders and saga, auth, product,
  cart, shipping, user, review, notification, checkout)

The last two were ported from the `duynhlab/helm-charts` `grafana-dashboards` chart,
which still serves its own copies through ConfigMaps. See the cutover note below.

The PostgreSQL dashboards are tested on **PostgreSQL 18**.

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

## Agent configuration

The project Agent Skill is located at
[`.agents/skills/grafana-foundation-sdk/`](.agents/skills/grafana-foundation-sdk/SKILL.md).
It follows the [Agent Skills](https://agentskills.io) format and contains the
architecture, domain, alerting, testing, CI/CD, end-to-end, and multi-agent
conventions used by this repository.

`.agents/skills/` is the canonical location, which OpenAI Codex discovers natively.
`.claude/skills/grafana-foundation-sdk` and `.cursor/skills/grafana-foundation-sdk`
are symlinks to it, so Claude Code and Cursor load the same files.

[`AGENTS.md`](AGENTS.md) holds the repository-wide agent instructions; `CLAUDE.md`
imports it.

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
Grafana 12, pushes every generated `*.spec.json` into an in-cluster registry,
applies CRs that reference `spec.oci`, and waits for each folder and dashboard
to report a synchronized condition. The expected resource set is read from
`deploy/manifests/`, so a newly registered dashboard is covered automatically.

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

### Microservices boards: two delivery paths

`microservices-monitoring-001-otel` and `business-otel` are also served by the
`duynhlab/helm-charts` `grafana-dashboards` chart, which renders them as ConfigMaps
that homelab consumes through `GrafanaDashboard.configMapRef`. Nothing consumes the
as-code copies yet, so the two paths coexist safely today.

They must never both be active against the same Grafana: the UIDs collide. A cutover
means pointing a Flux `OCIRepository` at the `deploy/` bundle, removing the boards from
the chart values, and deleting the two `configMapRef` CRs. Two differences to expect:
the as-code copies land in the `Microservices` folder rather than
"Microservices / Golden Signals" and "Business & Product", and they resolve a datasource
named `prometheus` instead of the chart's `DS_PROMETHEUS` input mapped to
`VictoriaMetrics`.

## CI

Pull requests run:

- formatting verification
- `go vet`
- repository-wide test coverage with a minimum of 90%
- deterministic artifact generation and diff verification
- Kustomize rendering checks for folders, dashboards, and alert groups
- a guard that every generated `*.spec.json` has a matching `GrafanaDashboard` CR
- Kind end-to-end smoke tests (operator fetch via `spec.oci`)

Pushes to `as-code` additionally publish dashboard JSON (oras) and the Flux
deploy bundle to GHCR.
