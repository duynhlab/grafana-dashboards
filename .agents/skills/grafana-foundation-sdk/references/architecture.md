# Architecture

Observability as code. Go source is the truth, the Grafana Foundation SDK is the
technical layer, and this repository owns the domain layer: queries, panels,
dashboards, alerts, standards.

**Baseline:** Grafana 12+, `dashboardv2` (`dashboard.grafana.app/v2`). Not `dashboard`
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
generated/dashboards/<uid>.spec.json, <uid>.manifest.json
generated/alerts/<uid>.json
deploy/manifests/                    GrafanaFolder, GrafanaDashboard (spec.oci), GrafanaAlertRuleGroup
deploy/kustomization.yaml            generated, lists every manifest
test/dashboards/                     contract tests per dashboard + alerts_test.go (registry-wide)
test/e2e/kind/                       operator smoke test
dashboard/                           LEGACY JSON, read-only
```

Domains today: `kubernetes` (folder `Kubernetes`), `postgres` (folder `Databases`),
`observability` (folder `Observability`). Folder constants live in `internal/standards/folders.go`.

## Pipeline

```text
queries -> panels -> dashboards / alerts -> registry -> go run ./cmd/generate
  -> generated/dashboards/*.spec.json     (oras push -> ghcr.io/duynhlab/grafana-dashboards)
  -> deploy/                              (flux push artifact -> ghcr.io/duynhlab/grafana-dashboards-as-code)
  -> Grafana Operator >= 5.24 pulls spec.oci -> Grafana 12
```

The generator writes `GrafanaDashboard` CRs with `spec.oci.reference` and
`spec.oci.path: <uid>.spec.json`; JSON is never inlined into the CR. `OCI_REFERENCE`,
`OCI_PULL_SECRET`, and `OCI_INSECURE_PLAIN_HTTP` change the CRs at generation time.

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
