# Alerts

Alerts are first-class, same level as dashboards. Read this file when adding or changing alert rules.

```text
internal/alerts/<domain>/
internal/registry/alerts.go
generated/alerts/
deploy/manifests/grafanaalertrulegroup-*.yaml
```

## Rule shape

```text
UID, title
query (reuse internal/queries)
condition + for
labels: severity, domain, component  (plus team/service when known)
annotations: summary, description, dashboard_uid  (runbook_url when it exists)
```

Naming: `<domain>_<resource>_<condition>` — e.g. `kubernetes_crashlooping_pods`, `postgres_backends_waiting`.

Do not encode routing in the name; use labels.

## Dashboard link

Every production alert should annotate `dashboard_uid` with a dashboard that exists in the dashboard registry. Operators go dashboard → problem → alert → runbook.

## Delivery

Generator groups rules by Grafana folder/domain into `GrafanaAlertRuleGroup` (grafana-operator). `folderUID` / `folderRef` must match the domain `GrafanaFolder`. Evaluation interval is per group (default `1m`).

If a Kind/operator version lacks the CRD, still generate the artifact; e2e may skip reconcile and document the fallback.

## Adding an alert

1. Reuse or add a query in `internal/queries/`
2. Define the rule in `internal/alerts/<domain>/`
3. Register in `internal/registry/alerts.go`
4. Generate + test UID uniqueness, required labels/annotations, dashboard_uid exists
