# CI/CD — Flux OCI to GHCR

## Pipeline (`.github/workflows/as-code.yml`)

Actions: `actions/checkout@v7`, `actions/setup-go@v7`, Go `1.26`.

1. gofmt check, `go vet ./...`, tests with repository-wide coverage ≥ 90%
2. `go run ./cmd/generate`
3. Assert no unexpected diff in `generated/` and `deploy/`
4. Kustomize render must include GrafanaFolder, GrafanaDashboard, GrafanaAlertRuleGroup
5. Kind e2e applies the deploy bundle against Grafana Operator
6. On push to `as-code`: `flux push artifact` → `ghcr.io/duynhlab/grafana-dashboards-as-code`

## Tags

- `:latest` — mutable, tracks `as-code` branch head
- `:sha-<short>` — immutable per commit

## Artifact contents

`deploy/` kustomize bundle:

- `GrafanaFolder` — one per domain (`Kubernetes`, `Databases`, …)
- `GrafanaDashboard` CRs — one per dashboard UID
- `GrafanaAlertRuleGroup` CRs — one per domain with seed rules
- `Grafana` instance selector labels (`dashboards: grafana`)

## Tools

- [gcx](https://grafana.com/docs/grafana-cloud/ai-tools/gcx/overview/) — Grafana 12+ resource management (optional validate in e2e)
- [flux push artifact](https://fluxcd.io/flux/cmd/flux_push_artifact/) — OCI bundle publish

## Secrets

- `GITHUB_TOKEN` — GHCR push (packages:write)
- `GRAFANA_TOKEN` — e2e Grafana API checks (optional)
