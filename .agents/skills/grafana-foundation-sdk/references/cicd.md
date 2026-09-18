# CI/CD: OCI artifacts to GHCR

Workflow: `.github/workflows/as-code.yml`. Actions `actions/checkout@v7`,
`actions/setup-go@v7`, Go `1.26`, Flux CLI, oras, Kind.

## Jobs

`validate` (pull requests to `main` or `as-code`, pushes to `as-code`):
1. gofmt check, `go vet ./...`, tests with repository-wide coverage >= 90%
2. `go run ./cmd/generate`
3. `git diff --exit-code -- generated/ deploy/`
4. `kubectl kustomize deploy/` must contain GrafanaFolder, GrafanaDashboard, GrafanaAlertRuleGroup
5. the number of `deploy/manifests/grafanadashboard-*.yaml` files equals the number of
   `generated/dashboards/*.spec.json` files

`e2e-kind` (needs `validate`): runs `test/e2e/kind/run.sh`, see [e2e-kind.md](e2e-kind.md).

`publish-oci` (push to `as-code` only, needs `validate`):
- `oras push` of **every** `generated/dashboards/*.spec.json` to
  `ghcr.io/duynhlab/grafana-dashboards` with artifact type
  `application/vnd.grafana.dashboard+json`
- `flux push artifact` of `deploy/` to `ghcr.io/duynhlab/grafana-dashboards-as-code`

Never list spec files by name in the workflow. The file set is derived from the
generated directory so a newly registered dashboard ships without a workflow change.

## Tags

- `:latest` mutable, tracks the `as-code` branch head
- `:sha-<short>` immutable per commit

Both tags are applied to both artifacts.

## Artifact contents

Dashboard JSON (oras): `generated/dashboards/<uid>.spec.json`, one layer per file, the
`GrafanaDashboard.spec.oci.path` value equals the file name.

Deploy bundle (Flux): the `deploy/` kustomization with one `GrafanaFolder` per domain,
one `GrafanaDashboard` per dashboard (`spec.oci`, no inline JSON), one
`GrafanaAlertRuleGroup` per group, instance selector `dashboards: grafana`.

Private GHCR: generate with `OCI_PULL_SECRET=ghcr-pull` and create a
`kubernetes.io/dockerconfigjson` Secret with that name in the operator namespace.

## Secrets

- `GITHUB_TOKEN` with `packages: write` for GHCR pushes
- no other secret is required

## Tools

- oras: https://oras.land/
- flux push artifact: https://fluxcd.io/flux/cmd/flux_push_artifact/
- Grafana Operator Helm chart >= 5.24.0 (e2e pins 5.25.0)
