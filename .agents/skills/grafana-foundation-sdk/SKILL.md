---
name: grafana-foundation-sdk
description: >-
  Build, port, review, or test Grafana dashboards and alert rules as Go code with the
  Grafana Foundation SDK (dashboardv2) in this repository. Use for any change under
  internal/, cmd/generate, generated/, deploy/, test/, or .github/workflows/as-code.yml;
  for the Kubernetes, PostgreSQL (CNPG or Pigsty postgres_exporter), PGDog, and Temporal
  domains; for Flux OCI delivery, Grafana Operator CRs, or the Kind e2e. Also use when
  several agents split dashboard work in parallel. Never edit legacy JSON in dashboard/.
metadata:
  short-description: Grafana dashboards and alerts as Go (Foundation SDK dashboardv2)
  version: "2"
---

# Grafana Foundation SDK (dashboardv2)

This repository defines Grafana dashboards and alert rules as Go code, renders them
with `go run ./cmd/generate`, and ships the result to Grafana through Flux OCI
artifacts and the Grafana Operator. Go under `internal/` is the only source of truth.

## Invariants

- Never edit files under `dashboard/`. They are frozen legacy JSON kept for porting.
- Never hand-edit `generated/` or `deploy/`. Run `make generate`; commit the output.
- Use `dashboardv2` (`dashboard.grafana.app/v2`), not `dashboard` v1 or `dashboardv2beta1`.
- Dashboard UID equals the K8s manifest name passed to `dashboardv2.Manifest(uid, builder)`.
- Grafana folders come from `internal/standards/folders.go` (`Kubernetes`, `Databases`,
  `Observability`, `Microservices`). A resource also declares a **domain**, its owning Go
  package, from `internal/standards/domains.go` (`kubernetes`, `postgres`,
  `observability`, `microservices`). The two differ for Postgres, and the domain is the
  directory segment under `generated/` and the prefix of `spec.oci.path`.
- `make validate` must pass before any change is done (fmt, vet, coverage >= 90%,
  generate, clean diff on `generated/` and `deploy/`).

## Reference map

Read only what the task needs.

| Read | When |
|---|---|
| [references/architecture.md](references/architecture.md) | first time in the repo, or before adding abstractions |
| [references/domains.md](references/domains.md) | choosing a folder, porting legacy JSON, metric model (CNPG vs Pigsty) |
| [references/alerts.md](references/alerts.md) | adding or changing alert rules |
| [references/testing.md](references/testing.md) | writing tests, before running `make validate` |
| [references/multi-agent.md](references/multi-agent.md) | splitting work across several agents or subagents |
| [references/cicd.md](references/cicd.md) | touching the GitHub Actions workflow or OCI publishing |
| [references/e2e-kind.md](references/e2e-kind.md) | touching `deploy/`, `spec.oci`, or `test/e2e/kind/` |

## Repository shape

```text
internal/queries/prometheus/<domain>/   PromQL constants, one file per board or topic
internal/panels/                        panel and layout helpers (see below)
internal/standards/                     folders, labels, datasource, time settings
internal/dashboards/<domain>/           one Go file per dashboard, returns a builder
internal/alerts/<domain>.go             alert rules, flat, one file per domain
internal/registry/dashboards.go|alerts.go   the only place resources are registered
internal/generate/                      renderer; do not touch for new resources
generated/dashboards/<domain>/          spec.json + manifest.json, mirrors the packages
generated/alerts/<domain>/              one JSON per rule
deploy/manifests/                       CRs, flat; filenames carry the UID
test/dashboards/                        one contract test per dashboard + alerts_test.go
```

SDK is pinned in `go.mod` (`github.com/grafana/grafana-foundation-sdk/go`). Verify an
API before using it:

```bash
go doc github.com/grafana/grafana-foundation-sdk/go/dashboardv2 DashboardBuilder
```

## Helpers to reuse (do not re-wrap the SDK)

| Need | Use |
|---|---|
| single-stat / gauge | `panels.Stat`, `panels.Gauge`, `panels.StatValue` (with thresholds) |
| time series | `panels.SeriesExpr(title, unit, expr, legend)`, `panels.Series(title, unit, queries...)`, `panels.TimeSeriesExpr` |
| threshold line | `panels.SeriesThresholdLine` with `panels.Thr`, `panels.Ptr` |
| table | `panels.Table` + `panels.QueryTable`, `panels.Merge`, `panels.Organize`, `panels.SortBy` |
| multi-target panel | `panels.PromQuery(expr, legend)` for ref A, `panels.Query("B", expr, legend)` for the rest |
| template variable | `panels.QueryVar(name, label, definition)` |
| layout | `panels.RowsLayout(panels.Row(title, panels.GridItem(id, x, y, w, h)...))` |
| time / datasource | `standards.TimeSettings(from, to, refresh)`, `standards.PrometheusDatasource()` |
| alert | `alerts.Rule{...}` with `standards.Label*` and `standards.Annotation*` keys |

## Dashboard skeleton

```go
package postgres

func PGExample() cog.Builder[dashboardv2.Dashboard] {
    b := dashboardv2.NewDashboardBuilder("PostgreSQL Example").
        Description("...").
        Editable(true).
        Tags([]string{"postgresql", "generated", "foundation-sdk"}).
        TimeSettings(standards.TimeSettings("now-1h", "now", "1m")).
        QueryVariable(panels.QueryVar("cluster", "Cluster", pgqueries.ClusterLabelValues)).
        QueryVariable(panels.QueryVar("datname", "Database", pgqueries.DatnameLabelValues)).
        Panel("calls", panels.SeriesExpr("Calls/s", "ops", pgqueries.CallsPerSec, "calls")).
        Panel("latency", panels.SeriesExpr("Mean latency", "ms", pgqueries.MeanLatency, "mean"))

    return b.RowsLayout(panels.RowsLayout(
        panels.Row("Throughput",
            panels.GridItem("calls", 0, 0, 12, 8),
            panels.GridItem("latency", 12, 0, 12, 8),
        ),
    ))
}
```

Register it in `internal/registry/dashboards.go` with `UID`, `Folder`, `Build`.

## Workflows

**Add a dashboard**
1. Pick the domain and folder (see `domains.md`). Check the metric model the target
   exporter exposes before writing PromQL.
2. Add PromQL constants under `internal/queries/prometheus/<domain>/`. Reuse existing ones.
3. Compose the board in `internal/dashboards/<domain>/<name>.go` using the helpers above.
4. Register in `internal/registry/dashboards.go` with `UID`, `Domain`, `Folder`, `Build`.
   The generator refuses an empty `Domain`.
5. Add `test/dashboards/<name>_test.go` (title, variable count, panel count, manifest
   name, key PromQL fragments). Add alerts if the board has operational thresholds.
6. `make validate`. Inspect the diff in `generated/` and `deploy/`.

**Add an alert**: reuse a domain query, define `alerts.Rule` in `internal/alerts/<domain>.go`,
set `Group`, labels `severity|domain|component`, annotations `summary|description|dashboard_uid`
(the UID must exist in the dashboard registry), register in `internal/registry/alerts.go`,
`make validate`. Details in `alerts.md`.

**Port legacy JSON**: read the JSON under `dashboard/` for titles, PromQL, variables, and
layout; rewrite as dashboardv2 with the helpers; keep the same UID when the board is a
replacement; do not aim for v1 schema parity. Recipe in `domains.md`.

**Add a domain**: add a `Folder*` constant in `internal/standards/folders.go`, create the
`queries`, `dashboards`, and `alerts` packages, register resources with that folder. The
generator discovers folders from the registries; do not change `cmd/generate`.

## Verification

```bash
make validate     # gofmt check, go vet, coverage >= 90%, generate, git diff --exit-code generated/ deploy/
make e2e-kind     # only when deploy/, spec.oci, or test/e2e/kind changed; needs Docker
```

## Working with several agents

One orchestrator owns `internal/registry/`, `internal/standards/`, `internal/panels/`,
`generated/`, `deploy/`, the Makefile, and the workflow. Each worker owns exactly one
dashboard or one alert group and touches only its own query, dashboard, and test files.
Workers report back instead of editing shared files; the orchestrator registers, runs
`make generate` once, and runs `make validate` once. Full contract, report format, and
per-harness delegation hints in `references/multi-agent.md`.

## Agent rules

- Inspect existing queries, panels, and dashboards before creating new abstractions.
- Abstract observability meaning (a named query, a domain section), not SDK method names.
- Keep changes scoped to the requested dashboard, alert, or domain. Do not touch other
  boards, and do not reformat generated output by hand.
- Do not duplicate PromQL across dashboards unless the semantic difference is intentional.
- Keep definitions environment-independent: logical datasource names, no secrets,
  no cluster-specific label values.
- Report what `make validate` printed. A failing gate is a result, not a detail to omit.
