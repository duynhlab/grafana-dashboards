# E2E — Kind 1.36

## Location

`test/e2e/kind/`

## Cluster

- Kind with `kindest/node:v1.36.0` (or latest 1.36 patch)
- Flux controllers installed
- Grafana Operator + Grafana 12+

## Stack

| Component | Purpose |
|-----------|---------|
| kube-prometheus-stack or VM + ksm + node-exporter | Kubernetes dashboard metrics |
| CNPG or metric fixtures | pg-io-waits metrics |
| Flux OCIRepository | Pull `ghcr.io/duynhlab/grafana-dashboards-as-code:latest` |
| Kustomization | Apply GrafanaFolder, GrafanaDashboard, GrafanaAlertRuleGroup |

## Assertions

1. Apply `deploy/` (or Flux reconciles the OCI artifact)
2. GrafanaFolders `kubernetes` and `databases` exist
3. GrafanaDashboards `kubernetes-cluster-overview`, `pg-io-waits` exist
4. GrafanaAlertRuleGroup CRs exist if the operator CRD is installed; if the CRD is missing, generation still succeeds and e2e logs the fallback
5. Smoke: dashboards load without panel errors

## Local run

```bash
make e2e-kind
```
