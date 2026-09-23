# CI/CD: the deploy bundle to GHCR

Workflow: `.github/workflows/as-code.yml`. Actions `actions/checkout@v7`,
`actions/setup-go@v7`, Go `1.26`, Flux CLI, Kind.

## Jobs

`validate` (pull requests to `main`, pushes to `main` and `v*` tags):
1. gofmt check, `go vet ./...`, tests with repository-wide coverage >= 90%
2. `go run ./cmd/generate`
3. `git diff --exit-code -- generated/ deploy/`
4. `kubectl kustomize` on each of `deploy/`, `deploy/folders/`, `deploy/dashboards/`.
   The bundle must contain `GrafanaManifest` and `GrafanaAlertRuleGroup`, the folders
   overlay must carry `folder.grafana.app/v1` and the dashboards overlay
   `dashboard.grafana.app/v2`, and a `GrafanaDashboard` anywhere in the render is a
   hard failure: that kind cannot carry a v2 board.
5. the number of `deploy/dashboards/grafanamanifest-*.yaml` files equals the number of
   `generated/dashboards/**/*.spec.json` files

`e2e-kind` (needs `validate`): runs `test/e2e/kind/run.sh`, see [e2e-kind.md](e2e-kind.md).

`publish-oci` (any push, needs `validate`):
- `flux push artifact` of `deploy/` to `ghcr.io/duynhlab/grafana-dashboards-as-code`

There is no second artifact. `GrafanaManifest` has no content sources at all, so a
separate dashboard-JSON artifact would have no consumer; the specs under
`generated/dashboards/` stay committed as the review surface. Nothing in the workflow
names a dashboard, so a newly registered board ships without a workflow change.

## Tags

- `:latest` mutable, tracks `main`
- `:sha-<short>` immutable per commit on a branch push
- `:v<x.y.z>` on a tag push. Consumers pin this rather than chasing `:latest`

## Artifact contents

The `deploy/` kustomization, three overlays:

```text
deploy/folders/      GrafanaManifest -> folder.grafana.app/v1 Folder, one per domain
deploy/dashboards/   GrafanaManifest -> dashboard.grafana.app/v2 Dashboard, one per
                     board, plus one GrafanaAlertRuleGroup per group
deploy/              both, for a single apply where nothing is racing
```

Instance selector `dashboards: grafana` throughout. A consumer points one Flux
`Kustomization` at `./deploy/folders` and a second at `./deploy/dashboards` with
`dependsOn`, because a board naming a folder that does not exist yet fails outright and
kustomize ordering is not apply ordering. The second Kustomization also needs
`healthCheckExprs`: `GrafanaManifest` reports `ManifestSynchronized`, not `Ready`.

Private GHCR is a Flux concern now, not a generator one: there is no `spec.oci` and no
`OCI_PULL_SECRET`.

## Secrets

- `GITHUB_TOKEN` with `packages: write` for GHCR pushes
- no other secret is required

## Tools

- flux push artifact: https://fluxcd.io/flux/cmd/flux_push_artifact/
- Grafana Operator Helm chart 5.25.0, Grafana 13.x (12.x does not serve
  `dashboard.grafana.app/v2`)
