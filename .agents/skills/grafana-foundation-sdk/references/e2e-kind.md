# E2E on Kind

Location: `test/e2e/kind/run.sh`, `test/e2e/kind/kind-config.yaml`. Run with
`make e2e-kind`. Needs Docker, kind, kubectl, helm, go, curl, python3. Takes several
minutes.

## What the script does

1. Creates Kind cluster `grafana-as-code-e2e` with `kindest/node:v1.36.0`, falling
   back to `v1.32.0` when that image cannot be pulled. The 1.36 node image is not
   published yet, so today every run takes the fallback path.
2. Installs Grafana Operator Helm chart `5.25.0` into `grafana-operator`.
3. Applies a `Grafana` CR (`spec.version: 13.2.0`), waits for `status.stage=complete`
   and for the Grafana Deployment to be Available. The controllers skip an instance
   whose pod is not serving yet.
4. `go run ./cmd/generate`, then applies the two waves in order: `deploy/folders`,
   wait for every folder to reach `ManifestSynchronized=True`, then `deploy/dashboards`.
5. Asserts, deriving the expected set from `deploy/folders/` and `deploy/dashboards/`:
   - every `GrafanaManifest` reaches `ManifestSynchronized=True`
   - the `GrafanaAlertRuleGroup` CRD is installed and every alert-group manifest exists
   - Grafana's `/apis` discovery lists `dashboard.grafana.app`
   - **every board reads back** from
     `/apis/dashboard.grafana.app/v2/namespaces/default/dashboards/<uid>` with the same
     `spec.title` and element count as the generated spec
6. Deletes the cluster on exit. On error it dumps CR status and operator logs first.

## Gotchas

- **The read-back in step 5 is the test.** On Grafana 12.0.0 the legacy save path
  accepts fairly arbitrary JSON and the CR reports a synchronized condition while
  storing something that is not a usable dashboard. Any check that asserts only the
  condition passes on that false green. Assert what Grafana actually stored.
- Grafana must be 13.x. 12.0 and 12.1 serve no `dashboard.grafana.app` version higher
  than `v2alpha1`, and the manifests declare `v2`. Confirm with
  `curl -s /apis | jq '.groups[] | select(.name|test("dashboard"))'` before assuming.
- A resource is only proven delivered by the specific condition type. Right after apply
  the operator sets `NoMatchingInstance=True` while the Grafana pod is still failing
  authentication, and it clears on a later reconcile. Asserting "any condition is True"
  passes on that transient state and proves nothing.
- The folders wave must complete before the dashboards wave. A board whose
  `grafana.app/folder` names a missing folder fails with
  `folders.folder.grafana.app "<uid>" not found` and then waits for the next resync.
  The script applies them in order for exactly the reason a consumer needs `dependsOn`.
- The script no longer rewrites `deploy/` in the working tree, so there is nothing to
  restore afterwards.
- Never hard-code dashboard or folder names in the script. It reads them from the
  generated manifests so a new registry entry is covered automatically.
