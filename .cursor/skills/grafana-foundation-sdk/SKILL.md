---
name: grafana-foundation-sdk
description: Use when creating, modifying, porting, reviewing, or testing Grafana dashboards or alert rules as Go code with the Grafana Foundation SDK (dashboardv2), Flux OCI delivery, Kind e2e, GitHub Actions CI, or observability domains (Kubernetes, PostgreSQL, CNPG). Also use when working in this repo's as-code track — never edit legacy JSON in dashboard/.
---

# Grafana Foundation SDK (dashboardv2)

Build Grafana dashboards **and alert rules** as Go code for **Grafana 12+**. Legacy JSON under `dashboard/` is **read-only reference**.

Read [references/architecture.md](references/architecture.md) first. Read [references/alerts.md](references/alerts.md) when adding or changing alerts. Read [references/domains.md](references/domains.md) when opening a new domain. CI/e2e: [references/cicd.md](references/cicd.md), [references/e2e-kind.md](references/e2e-kind.md), [references/testing.md](references/testing.md).

## Core decisions

- Go + `github.com/grafana/grafana-foundation-sdk/go`
- **`dashboardv2`** (`dashboard.grafana.app/v2`) — not `dashboard` v1, not `dashboardv2beta1`
- Dashboard UID = K8s manifest `metadata.name` via `dashboardv2.Manifest(name, builder)`
- Grafana folders are **domain names** (`Kubernetes`, `Databases`), not `as-code`
- Alerts live in `internal/alerts/` and ship as `GrafanaAlertRuleGroup`
- Output: `generated/` + `deploy/manifests/`
- CI: generate → test → `flux push artifact` → `ghcr.io/duynhlab/grafana-dashboards-as-code`

## Workflow

For a dashboard:

1. Identify domain and Grafana folder
2. Reuse queries/sections/panels/standards
3. Compose the dashboard, register it
4. Add linked alerts when the behavior is operational
5. Generate, test, inspect the diff

For an alert: reuse the domain query → condition/`for` → standard labels/annotations → `dashboard_uid` → register → generate.

## dashboardv2 pattern

```go
builder := dashboardv2.NewDashboardBuilder("Title").
    Description("...").
    Tags([]string{"kubernetes"}).
    TimeSettings(standards.TimeSettings("now-1h", "now", "30s")).
    Panel("nodes", panels.Stat("Nodes", queries.NodeCount, "Nodes")).
    RowsLayout(dashboardv2.Rows().
        Row(dashboardv2.Row("Overview").GridLayout(
            dashboardv2.Grid().
                Item(dashboardv2.GridItem("nodes").X(0).Y(0).Width(4).Height(4)),
        )),
    )

manifest := dashboardv2.Manifest("kubernetes-cluster-overview", builder)
```

Prometheus: `prometheus.NewQueryV2Builder().Expr(...).LegendFormat(...)`

## Agent rules

- Never edit files under `dashboard/`
- Inspect existing code before creating abstractions
- Verify SDK APIs against the installed version
- Do not wrap every Foundation SDK method
- Do not silently change unrelated dashboards or alerts
- Keep changes scoped to the requested domain
