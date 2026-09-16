# CI/CD — OCI artifacts to GHCR

## Pipeline (`.github/workflows/as-code.yml`)

Actions: `actions/checkout@v7`, `actions/setup-go@v7`, Go `1.26`.

1. gofmt check, `go vet ./...`, tests with repository-wide coverage ≥ 90%
2. `go run ./cmd/generate`
3. Assert no unexpected diff in `generated/` and `deploy/`
4. Kustomize render must include GrafanaFolder, GrafanaDashboard (`spec.oci`), GrafanaAlertRuleGroup
5. Kind e2e: Grafana Operator **5.25.0**, in-cluster registry, `oras push` dashboard JSON, operator fetches `spec.oci`
6. On push to `as-code`:
   - `oras push` dashboard JSON → `ghcr.io/duynhlab/grafana-dashboards`
   - `flux push artifact` of `deploy/` → `ghcr.io/duynhlab/grafana-dashboards-as-code`

## Tags

- `:latest` — mutable, tracks `as-code` branch head
- `:sha-<short>` — immutable per commit

Apply to both artifacts.

## Artifact contents

Dashboard JSON (`oras`, `application/vnd.grafana.dashboard+json`):

- `generated/dashboards/<uid>.spec.json` files

`deploy/` kustomize bundle (Flux):

- `GrafanaFolder` — one per domain
- `GrafanaDashboard` CRs with `spec.oci.reference` + `spec.oci.path` (no inline JSON)
- `GrafanaAlertRuleGroup` CRs
- `Grafana` instance selector labels (`dashboards: grafana`)

Private GHCR: set `OCI_PULL_SECRET=ghcr-pull` when generating and create a `kubernetes.io/dockerconfigjson` Secret in the operator namespace.

## Tools

- [oras](https://oras.land/) — dashboard JSON OCI
- [flux push artifact](https://fluxcd.io/flux/cmd/flux_push_artifact/) — GitOps kustomize bundle
- Grafana Operator Helm chart `>= 5.24.0` (CI uses `5.25.0`)

## Secrets

- `GITHUB_TOKEN` — GHCR push (packages:write)
- `GRAFANA_TOKEN` — e2e Grafana API checks (optional)
