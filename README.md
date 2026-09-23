# Grafana Observability as Code

Strongly typed Grafana dashboards and alert rules built with Go and the
[Grafana Foundation SDK](https://github.com/grafana/grafana-foundation-sdk).

The project uses the Dashboard v2 model and targets Grafana 13 or newer:
`dashboard.grafana.app/v2` is not served on 12.0 or 12.1.
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
- `Platform`
- `API Gateway`

Dashboards:

- `kubernetes-cluster-overview`
- `pg-io-waits` — tested on PostgreSQL 18 (`pg_stat_io`)
- `pg-maintenance` — CNPG locks, checkpointer, autovacuum/bloat, long transactions
- `pg-query-performance` — CNPG `pg_stat_statements` throughput, latency, top statements
- `pg-exporter-instance` — Pigsty PGRDS instance board (uses the Pigsty
  `postgres_exporter` metric model: `pg_*` metrics and `pg:*` recording rules
  with `ins`/`cls` labels, not `cnpg_*`)
- `pgdog` — PGDog connection pooler (ported from Grafana.com dashboard 24583)
- `keda` — KEDA scaler value and errors, the HPA it drives, and the Temporal backlog
  it scales on. The source board hardcoded the namespaces and Deployments it watched;
  those are template variables here
- `temporal-worker` — Temporal workflow/activity RED metrics plus the server row
  (gRPC by role, persistence, task-queue backlog)
- `otel-collector-health` — OTel Collector receiver/exporter/processor pipeline
- `microservices-monitoring-001-otel` — Go services RED, runtime, gRPC east-west and
  otelpgx pool metrics, using the OpenTelemetry semantic conventions
  (`http_server_*`, `rpc_*`, `go_*`, `db_client_*`) keyed on `service_name`
- `business-otel` — per-domain business KPIs (payments, orders and saga, auth, product,
  cart, shipping, user, review, notification, checkout)
- `red-spanmetrics` — RED from the OTel spanmetrics connector, keyed on `service_name`.
  Distinct from `microservices-monitoring-001-otel`, which reads `http_server_*` and
  `rpc_server_*` directly
- `rfc0021-baseline` — order saga and payment cutover gate
- `inventory-overview` — inventory reservation FSM and gRPC RED
- `cert-manager` — certificate expiry and renewal, controller sync, ACME, workqueue
- `keycloak-identity` — login and token KPIs, realm latency, auth events, JVM and DB pool
- `eg-edge` — Envoy Gateway edge golden signals, data plane, and control plane

`microservices-monitoring-001-otel` and `business-otel` were ported from the
`duynhlab/helm-charts` `grafana-dashboards` chart, which still serves its own copies
through ConfigMaps. See the cutover note below. The rest of the boards above were ported
from JSON vendored in `duynhlab/homelab`; each query file records the source path and the
commit it was read at.

`rfc0021-baseline` and `inventory-overview` read `rfc0021:*` and `inventory:*` **recording
rules**, not raw metrics. Those rules live in homelab under
`kubernetes/infra/configs/observability/metrics/prometheusrules/microservices/`; without
them the two boards are empty.

The PostgreSQL dashboards are tested on **PostgreSQL 18**.

Alert rules:

- `kubernetes_crashlooping_pods`
- `kubernetes_pending_pods`
- `kubernetes_pvcs_at_risk`
- `postgres_backends_waiting`

## Requirements

- Go 1.26 or newer
- Grafana 13 or newer (12.x does not serve `dashboard.grafana.app/v2`)
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
  alerts/<domain>/            generated alert definitions, grouped by Go package
  dashboards/<domain>/        Dashboard v2 specs and manifests, grouped by Go package
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

- one `GrafanaManifest` wrapping a `folder.grafana.app/v1` `Folder` per domain
- one `GrafanaManifest` wrapping a `dashboard.grafana.app/v2` `Dashboard` per dashboard
- one `GrafanaAlertRuleGroup` per alert domain

### Why `GrafanaManifest` and not `GrafanaDashboard`

`GrafanaDashboard` posts through the legacy `/api/dashboards/db` envelope, and
that is what the controller does rather than a setting — its content sources
(`oci`, `url`, `configMapRef`, `json`, `grafanaCom`, `jsonnet`) are only
transports for the bytes. A Dashboard v2 payload is refused twice over: the bare
spec returns `400 dashboard appears to be in v2 format`, and the wrapped object
that error asks for returns `400 The k8s style dashboard must not include an id
on the root element` — unfixable from here, because the operator's content
resolver sets an `id` on every model before posting.

`GrafanaManifest` applies its payload with a discovery-based dynamic client
against `/apis`, so it speaks the same protocol as the schema. The trade is that
it has **no content sources at all**, so the dashboard spec is inlined and the
object is bounded by the etcd limit near 1 MiB. Grafana must serve
`dashboard.grafana.app/v2`, which means **13.x**; 12.0 and 12.1 serve no higher
than `v2alpha1`.

### Two waves

`deploy/` is split because a dashboard whose `grafana.app/folder` annotation
names a folder that does not exist yet fails outright and then waits for the
operator's next resync — long enough that a Flux wave times out red. Listing the
folder first inside one kustomization does not help; kustomize ordering is not
apply ordering for independent controllers.

```text
deploy/folders/      GrafanaManifest -> folder.grafana.app/v1 Folder
deploy/dashboards/   GrafanaManifest -> dashboard.grafana.app/v2 Dashboard, plus alert groups
deploy/              both, for a single apply where nothing is racing
```

A consumer points one Flux `Kustomization` at `./deploy/folders` and a second at
`./deploy/dashboards` with `dependsOn`. The second also needs `healthCheckExprs`:
`GrafanaManifest` reports `ManifestSynchronized`, not the `Ready` condition
kstatus expects.

Run the local end-to-end smoke test:

```bash
make e2e-kind
```

The test creates a Kind cluster, installs Grafana Operator 5.25.0, deploys
Grafana 13, applies the two waves in order, waits for every `GrafanaManifest` to
report `ManifestSynchronized`, and then **reads each board back** through
`/apis/dashboard.grafana.app/v2/namespaces/default/dashboards/<uid>`, comparing
title and element count against the generated spec. The read-back is the point:
on Grafana 12.0.0 the legacy save path accepts fairly arbitrary JSON and reports
a synchronized condition while storing something unusable, so an assertion on
the CR condition alone proves nothing. The expected resource set is read from
`deploy/`, so a newly registered dashboard is covered automatically.

## OCI publishing

GitHub Actions publishes one artifact, the `deploy/` bundle, on every push to
`main` and on `v*` tags:

```text
ghcr.io/duynhlab/grafana-dashboards-as-code:latest
ghcr.io/duynhlab/grafana-dashboards-as-code:sha-<commit>   # branch push
ghcr.io/duynhlab/grafana-dashboards-as-code:v<x.y.z>       # tag push, pin this
```

There is no separate dashboard-JSON artifact. `GrafanaManifest` cannot fetch
content, so nothing would pull one; the specs under `generated/dashboards/` stay
committed as the review surface, not as a delivery mechanism.


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
- Kustomize rendering checks on each of the three overlays, including a guard
  that `deploy/` never emits a `GrafanaDashboard` again
- a guard that every generated `*.spec.json` has a matching `GrafanaManifest` CR
- Kind end-to-end smoke tests that read every board back through `/apis`

Pushes to `main` and `v*` tags additionally publish the Flux deploy bundle to
GHCR.
