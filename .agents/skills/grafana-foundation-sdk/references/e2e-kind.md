# E2E on Kind

Location: `test/e2e/kind/run.sh`, `test/e2e/kind/kind-config.yaml`. Run with
`make e2e-kind`. Needs Docker, kind, kubectl, helm, oras, go, curl. Takes several minutes.

## What the script does

1. Creates Kind cluster `grafana-as-code-e2e` with `kindest/node:v1.36.0`, falling
   back to `v1.32.0` when that image cannot be pulled. The 1.36 node image is not
   published yet, so today every run takes the fallback path.
2. Installs Grafana Operator Helm chart `5.25.0` into `grafana-operator`.
3. Deploys an in-cluster `registry:2` Service on **port 80** and port-forwards it.
4. Applies a `Grafana` CR (`spec.version: 12.0.0`), waits for `status.stage=complete`
   and for the Grafana Deployment to be Available. The dashboard and folder controllers
   skip an instance whose pod is not serving yet.
5. `oras push --plain-http` of every `generated/dashboards/<domain>/*.spec.json` into the
   registry, pushed from `generated/dashboards` so each layer title keeps its domain prefix.
6. Regenerates `deploy/` with `OCI_REFERENCE=<in-cluster registry>/grafana-dashboards:e2e`,
   `OCI_INSECURE_PLAIN_HTTP=true`, `OCI_PULL_SECRET=` and applies `deploy/`.
7. Asserts, deriving the expected set from `deploy/manifests/`:
   - every `GrafanaFolder` manifest exists in the cluster
   - every `GrafanaDashboard` exists with the expected `spec.oci.reference` and a
     `spec.oci.path` of the form `<domain>/<uid>.spec.json` that names a file which was
     actually generated
   - every `GrafanaDashboard` reaches a `status.conditions` entry with `status=True`,
     which proves the operator pulled the JSON out of the registry
   - the `GrafanaAlertRuleGroup` CRD is installed and every alert-group manifest exists
8. Deletes the cluster on exit. On error it dumps CR status and operator logs first.

## Gotchas

- A dashboard or folder is only proven delivered by `DashboardSynchronized=True` or
  `FolderSynchronized=True`. Right after apply the operator sets
  `NoMatchingInstance=True` while the Grafana pod is still failing authentication, and
  it clears on a later reconcile. Asserting "any condition is True" passes on that
  transient state and proves nothing, so always assert the specific condition type.

- The `GrafanaDashboard` CRD validates `spec.oci.reference` against
  `^[^:@]+(:[^:@/]+|@sha256:[a-fA-F0-9]{64})$`, which rejects a port in the registry
  host. That is why the registry Service listens on 80. `internal/generate` enforces the
  same pattern so a bad `OCI_REFERENCE` fails before reaching the cluster.
- Step 6 rewrites `deploy/` in the working tree. Run `make generate` afterwards to
  restore the committed GHCR reference before committing.
- Never hard-code dashboard or folder names in the script. It reads them from the
  generated manifests so a new registry entry is covered automatically.
