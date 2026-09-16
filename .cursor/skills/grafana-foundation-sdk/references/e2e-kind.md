# E2E — Kind 1.36

## Location

`test/e2e/kind/`

## Cluster

- Kind with `kindest/node:v1.36.0` (falls back to `v1.32.0` only if the 1.36 node image cannot be pulled)
- Grafana Operator Helm chart `5.22.2`
- Grafana `spec.version: "12.0.0"`
- Wait on Grafana CR `status.stage=complete`, not `condition=Ready`

## Assertions

1. Apply `deploy/`
2. GrafanaFolders `kubernetes` and `databases` exist
3. GrafanaDashboards `kubernetes-cluster-overview`, `pg-io-waits` exist
4. GrafanaAlertRuleGroup CRs `kubernetes` and `databases` exist (the CRD is required)

## Local run

```bash
make e2e-kind
```
