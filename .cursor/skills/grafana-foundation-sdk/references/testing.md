# Testing and CI

## Commands

```bash
gofmt -w .
go vet ./...
go test ./...
go run ./cmd/generate
```

CI also asserts a clean git diff on `generated/` and `deploy/`, then kustomize-builds the bundle.

## What to assert

Dashboards: UID/manifest name, title, folder, variable names, panel count, key PromQL fragments.

Alerts: UID uniqueness, name, condition/`for`, labels (`severity`, `domain`, `component`), annotations (`summary`, `description`, `dashboard_uid` present in dashboard registry).

Queries: metric, aggregation, labels, no environment-specific hardcoding.

Golden JSON only when output is stable enough that diffs stay useful.

## CI

See [cicd.md](cicd.md): validate on PR; `flux push artifact` to `ghcr.io/duynhlab/grafana-dashboards-as-code` on push to `as-code`.

## E2E

See [e2e-kind.md](e2e-kind.md): Kind 1.36, Grafana 12+, Grafana Operator, apply `deploy/`.
