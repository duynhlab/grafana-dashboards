# As-Code Audit Report

Audit date: 2026-03-16  
Scope: Grafana Foundation SDK docs (Context7 + Grafana docs MCP) packaged into [`.cursor/skills/grafana-foundation-sdk/`](../.cursor/skills/grafana-foundation-sdk/SKILL.md)

## Executive summary

The repository has a parallel **as-code** track. Legacy JSON under `dashboard/` stays frozen. New work uses **Go + Grafana Foundation SDK `dashboardv2`** targeting **Grafana 12+**, with generated artifacts delivered via **GitHub Actions → Flux OCI → GHCR**. Grafana folders are **domain names** (`Kubernetes`, `Databases`). Alerts are first-class Go resources.

## Architecture (kept in the skill)

See skill references — do not keep a root architecture markdown file:

- [architecture.md](../.cursor/skills/grafana-foundation-sdk/references/architecture.md)
- [alerts.md](../.cursor/skills/grafana-foundation-sdk/references/alerts.md)
- [domains.md](../.cursor/skills/grafana-foundation-sdk/references/domains.md)
- [testing.md](../.cursor/skills/grafana-foundation-sdk/references/testing.md)
- [cicd.md](../.cursor/skills/grafana-foundation-sdk/references/cicd.md)
- [e2e-kind.md](../.cursor/skills/grafana-foundation-sdk/references/e2e-kind.md)

### Gaps addressed

| Gap | Resolution |
|-----|------------|
| CI/CD specifics | `.github/workflows/as-code.yml`: lint, test, generate, diff check, `flux push artifact` |
| OCI delivery | `deploy/` kustomize bundle → `ghcr.io/duynhlab/grafana-dashboards-as-code` |
| Grafana version | **Grafana 12+** required |
| dashboardv2 vs v1 | Greenfield `dashboardv2`; legacy JSON is reference only |
| Domain folders | `Kubernetes` and `Databases` GrafanaFolders, not `as-code` |
| Alerts | `internal/alerts/` + `GrafanaAlertRuleGroup` CRs |
| E2E | `test/e2e/kind/` with Kubernetes 1.36 |

### API version note

Official dashboard automation examples use `dashboard` (v1) + `dashboard.grafana.app/v1`. This project uses **`dashboardv2` + `dashboard.grafana.app/v2`**. UID lives in K8s manifest `metadata.name`.

## Target dashboards

### kubernetes-cluster-overview

- Folder: Kubernetes
- Seed alerts: crashlooping pods, pending pods, PVCs at risk

### pg-io-waits

- Folder: Databases
- Seed alert: backends waiting

## References

- [Grafana Foundation SDK](https://github.com/grafana/grafana-foundation-sdk)
- [Dashboard automation](https://grafana.com/docs/grafana/latest/as-code/observability-as-code/foundation-sdk/dashboard-automation/)
- [gcx CLI](https://grafana.com/docs/grafana-cloud/ai-tools/gcx/overview/)
- [Flux OCI source](https://fluxcd.io/flux/cmd/flux_create_source_oci/)
- [Grafana Operator alert groups](https://grafana.github.io/grafana-operator/docs/examples/alertrulegroup/)
