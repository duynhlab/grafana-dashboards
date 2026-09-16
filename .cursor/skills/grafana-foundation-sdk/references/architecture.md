# Architecture

Observability as code: Go source is truth. Grafana Foundation SDK is the technical layer. This repo owns the domain layer (queries, sections, dashboards, alerts, standards).

**Baseline:** Grafana 12+, `dashboardv2` (`dashboard.grafana.app/v2`), not `dashboard` v1, not `dashboardv2beta1`. Pin the SDK in `go.mod`. UID for a dashboard is K8s `metadata.name` via `dashboardv2.Manifest(name, builder)`.

## Layout

```text
cmd/generate/                 # orchestration only
internal/
  dashboards/<domain>/
  alerts/<domain>/
  queries/prometheus/<domain>/
  panels/                     # visualization primitives
  standards/                  # datasource, time, units, labels
  registry/                   # dashboards.go + alerts.go
generated/dashboards/
generated/alerts/
deploy/manifests/             # GrafanaFolder + GrafanaDashboard + GrafanaAlertRuleGroup
dashboard/                    # LEGACY JSON — do not edit
```

Grafana folders are **domain names** (`Kubernetes`, `Databases`), not a catch-all `as-code` folder. Adding a domain means new queries/sections/dashboard/alerts + a registry entry — do not change the generator.

## Pipeline

```text
queries -> sections -> dashboards/alerts -> registry -> cmd/generate
  -> generated/ + deploy/ -> GitHub Actions -> flux push artifact -> GHCR
  -> Flux / Grafana Operator -> Grafana
```

Generated JSON/YAML is an artifact. Never hand-edit it as the source of truth.

## Abstraction

Do not wrap every Foundation SDK method (`NewPanel`, `NewStat`). Abstract observability meaning (`kubernetes.CrashLoopingPods()`, query constants, standards).

Do not dump helpers into `components/`. Do not mix alerts into dashboard files. Do not duplicate PromQL across dashboards unless the semantic difference is intentional.

## Environment

Keep definitions environment-independent. Use logical datasource names (`prometheus`, `loki`, `tempo`) resolved at deploy time. Never commit tokens or secrets.

## Official docs

- https://github.com/grafana/grafana-foundation-sdk
- https://grafana.github.io/grafana-foundation-sdk/go/Reference/
- https://grafana.com/docs/grafana/latest/as-code/observability-as-code/foundation-sdk/dashboard-automation/
