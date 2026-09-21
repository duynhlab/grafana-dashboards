# Alerts

Alerts are first-class resources, same level as dashboards. Read this when adding or
changing a rule.

```text
internal/alerts/rule.go            Rule struct
internal/alerts/<domain>.go        one flat file per domain (kubernetes.go, postgres.go)
internal/registry/alerts.go        []alerts.Rule registration
generated/alerts/<domain>/<uid>.json
deploy/manifests/grafanaalertrulegroup-<group>.yaml
test/dashboards/alerts_test.go     registry-wide contract test
```

## Rule shape

`alerts.Rule` fields, all required unless noted:

| Field | Value |
|---|---|
| `UID` | `<domain>_<resource>_<condition>`, e.g. `kubernetes_crashlooping_pods`, `postgres_backends_waiting` |
| `Title` | human sentence |
| `Folder` | a `standards.Folder*` constant; must match the dashboard's folder |
| `Group` | rule-group slug, one per domain today: `kubernetes`, `databases` |
| `Expr` | a constant from `internal/queries/prometheus/<domain>/`, shared with the dashboard |
| `For` | pending duration, e.g. `5m`, `10m`, `15m` |
| `Threshold` | numeric; the generator builds a `> Threshold` classic condition |
| `Labels` | `standards.LabelSeverity`, `LabelDomain`, `LabelComponent` (add `team`, `service` when known) |
| `Annotations` | `standards.AnnotationSummary`, `AnnotationDescription`, `AnnotationDashboardUID` (+ `AnnotationRunbookURL` when a runbook exists) |

Severity values: `standards.SeverityWarning`, `standards.SeverityCritical`.

Do not encode routing in the UID or title. Routing is done with labels.

## Dashboard link

Every rule annotates `dashboard_uid` with a UID that exists in
`internal/registry/dashboards.go`. `TestAlertRegistry` fails otherwise. Operators go
dashboard -> problem -> alert -> runbook.

## Delivery

The generator groups rules by `Group` into one `GrafanaAlertRuleGroup` CR whose
`folderRef` matches the domain `GrafanaFolder`. Evaluation interval is per group
(default `1m`). A new `Group` value produces a new CR without generator changes.

## Adding an alert

1. Reuse the dashboard's query constant, or add one under `internal/queries/`.
2. Add a `func <Name>() Rule` in `internal/alerts/<domain>.go`.
3. Append it to `registry.Alerts` in `internal/registry/alerts.go`.
4. `make validate`. `TestAlertRegistry` checks UID uniqueness, required labels and
   annotations, and that `dashboard_uid` resolves. Add a specific test only when the
   rule has logic worth pinning (threshold, `for`).
5. Inspect `generated/alerts/<domain>/<uid>.json` and the changed `grafanaalertrulegroup-*.yaml`.
