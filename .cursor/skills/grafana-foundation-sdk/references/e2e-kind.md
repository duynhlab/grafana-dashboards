# E2E — Kind 1.36

## Location

`test/e2e/kind/`

## Cluster

- Kind with `kindest/node:v1.36.0` (falls back to `v1.32.0` only if the 1.36 node image cannot be pulled)
- Grafana Operator Helm chart `5.25.0` (`>= 5.24.0`, OCI dashboard source)
- Grafana `spec.version: "12.0.0"`
- Wait on Grafana CR `status.stage=complete`, not `condition=Ready`
- In-cluster `registry:2` plus `oras push --plain-http`
- GrafanaDashboard CRs generated with `OCI_REFERENCE` pointing at the in-cluster registry and `OCI_INSECURE_PLAIN_HTTP=true`

## Assertions

1. Apply `deploy/` after pushing spec JSON into the in-cluster registry
2. GrafanaFolders `kubernetes` and `databases` exist
3. GrafanaDashboards `kubernetes-cluster-overview`, `pg-io-waits` exist with `spec.oci`
4. GrafanaAlertRuleGroup CRs `kubernetes` and `databases` exist (the CRD is required)

## Local run

Requires Kind, kubectl, Helm, oras, Go, and Docker.

```bash
make e2e-kind
```
