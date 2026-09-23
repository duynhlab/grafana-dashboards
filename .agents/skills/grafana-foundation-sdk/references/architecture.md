# Architecture

Observability as code. Go source is the truth, the Grafana Foundation SDK is the
technical layer, and this repository owns the domain layer: queries, panels,
dashboards, alerts, standards.

**Baseline:** Grafana 13+, `dashboardv2` (`dashboard.grafana.app/v2`). Not `dashboard`
v1, not `dashboardv2beta1`. The SDK module `github.com/grafana/grafana-foundation-sdk/go`
is pinned in `go.mod`; check the pinned version there before assuming an API exists,
and confirm with `go doc github.com/grafana/grafana-foundation-sdk/go/dashboardv2 <Symbol>`.

## Layout

```text
cmd/generate/                        entry point only, calls internal/generate.Run
internal/
  queries/prometheus/<domain>/       PromQL constants; variables.go for template-variable definitions
  panels/                            visualization + layout helpers (stat, series, table, query, variable, layout)
  standards/                         folders.go, labels.go, datasource.go, time.go
  dashboards/<domain>/<name>.go      one exported builder func per dashboard
  alerts/rule.go + <domain>.go       Rule struct and one flat file per domain
  registry/dashboards.go, alerts.go  the registration lists; Folders() is derived from them
  generate/                          deterministic renderer: spec.json, manifest.json, CRs, kustomization
generated/dashboards/<domain>/<uid>.spec.json, <uid>.manifest.json
generated/alerts/<domain>/<uid>.json
deploy/folders/                      GrafanaManifest -> folder.grafana.app/v1 Folder
deploy/dashboards/                   GrafanaManifest -> dashboard.grafana.app/v2 Dashboard,
                                     plus GrafanaAlertRuleGroup
deploy/kustomization.yaml            generated, aggregates both waves
test/dashboards/                     contract tests per dashboard + alerts_test.go (registry-wide)
test/e2e/kind/                       operator smoke test
dashboard/                           LEGACY JSON, read-only
```

Domains today: `kubernetes` (folder `Kubernetes`), `postgres` (folder `Databases`),
`observability` (folder `Observability`), `microservices` (folder `Microservices`).
A domain is the owning Go package and the directory segment under `generated/`; a
folder is the Grafana folder. They differ for Postgres, so both are declared:
`standards.Domain*` in `domains.go` and `standards.Folder*` in `folders.go`, and every
registry entry and alert rule sets both.

## Pipeline

```text
queries -> panels -> dashboards / alerts -> registry -> go run ./cmd/generate
  -> generated/dashboards/<domain>/*.spec.json   (committed; the review surface)
  -> deploy/          (flux push artifact -> ghcr.io/duynhlab/grafana-dashboards-as-code)
  -> Flux pulls the bundle -> Grafana Operator 5.25 -> Grafana 13
```

The generator writes `GrafanaManifest` CRs that carry the App Platform object inline in
`spec.template`. `GrafanaDashboard` cannot be used: it posts through the legacy
`/api/dashboards/db` envelope and Grafana answers 400 on a v2 payload in either shape.
`GrafanaManifest` in turn has no content sources at all, so there is no `spec.oci` and
no OCI environment variables; the spec is inlined and bounded by the ~1 MiB etcd limit.

Adding a dashboard, alert, or domain must not require a generator change. If it does,
stop and reconsider the abstraction.

## Abstraction rules

- Abstract observability meaning: `k8squeries.CrashLoopingPods`, `alerts.BackendsWaiting()`,
  `panels.SeriesThresholdLine`. Do not wrap every SDK method one-to-one.
- Put helpers next to their kind: query builders in `panels/query.go`, layout in
  `panels/layout.go`, thresholds in `panels/field.go`. There is no `components/` package.
- Keep alerts out of dashboard files. They share queries through `internal/queries`.
- One PromQL constant per meaning. Two dashboards showing the same signal import the
  same constant.
- Large boards may split into `<name>.go` + `<name>_sections.go`
  (see `pg_exporter_instance*.go`). Do not grow a single 3000-line file.

## Environment

Definitions are environment-independent. Datasources are logical names resolved at
deploy time (`standards.PrometheusDatasource()` returns `prometheus`). Never commit
tokens, hostnames, or cluster-specific label values.

## Official docs

- https://github.com/grafana/grafana-foundation-sdk
- https://grafana.github.io/grafana-foundation-sdk/go/Reference/
- https://grafana.com/docs/grafana/latest/as-code/observability-as-code/foundation-sdk/dashboard-automation/
- https://grafana.github.io/grafana-operator/docs/
