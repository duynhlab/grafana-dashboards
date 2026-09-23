#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER_NAME="grafana-as-code-e2e"
KIND_CONFIG="${ROOT}/test/e2e/kind/kind-config.yaml"
OPERATOR_CHART_VERSION="5.25.0"
# Grafana 12.0.0 is not good enough to test against. Its legacy save path
# accepts fairly arbitrary JSON, so a v2 payload posted through the classic
# GrafanaDashboard kind reports DashboardSynchronized=True while storing
# something that is not a usable dashboard. Any assertion on the CR condition
# alone passes there and proves nothing. 13.x serves dashboard.grafana.app/v2,
# which is what the generated manifests declare.
GRAFANA_VERSION="13.2.0"
GRAFANA_PORT="3000"
PF_PID=""

log() { echo "[e2e] $*"; }

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1"; exit 1; }
}

require kind
require kubectl
require helm
require go
require curl
require python3

dump() {
  log "dumping cluster state"
  kubectl get grafana,grafanamanifest,grafanaalertrulegroup,deploy,pod,event -n grafana-operator || true
  kubectl describe grafana grafana -n grafana-operator || true
  kubectl describe grafanamanifest -n grafana-operator || true
  kubectl logs -n grafana-operator -l app.kubernetes.io/name=grafana-operator --tail=200 || true
  kubectl logs -n grafana-operator -l app=grafana --tail=200 || true
}

cleanup() {
  if [[ -n "${PF_PID}" ]]; then
    kill "${PF_PID}" >/dev/null 2>&1 || true
  fi
  log "cleaning up kind cluster ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}" >/dev/null 2>&1 || true
}

trap dump ERR
trap cleanup EXIT

log "creating kind cluster (Kubernetes 1.36)"
if ! create_error="$(kind create cluster --name "${CLUSTER_NAME}" --config "${KIND_CONFIG}" 2>&1)"; then
  log "kindest/node:v1.36.0 failed:"
  echo "${create_error}"
  kind delete cluster --name "${CLUSTER_NAME}" >/dev/null 2>&1 || true
  log "falling back to kindest/node:v1.32.0"
  sed 's/kindest\/node:v1.36.0/kindest\/node:v1.32.0/' "${KIND_CONFIG}" |
    kind create cluster --name "${CLUSTER_NAME}" --config /dev/stdin
else
  echo "${create_error}"
fi

log "waiting for control plane"
kubectl wait --for=condition=Ready nodes --all --timeout=120s

log "installing Grafana Operator ${OPERATOR_CHART_VERSION}"
helm upgrade --install grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator \
  --version "${OPERATOR_CHART_VERSION}" \
  --namespace grafana-operator --create-namespace \
  --wait --timeout 5m

kubectl wait --for=condition=Available deploy --all \
  -n grafana-operator --timeout=180s

log "deploying Grafana ${GRAFANA_VERSION}"
kubectl apply -f - <<EOF
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  namespace: grafana-operator
  labels:
    dashboards: grafana
spec:
  version: "${GRAFANA_VERSION}"
  config:
    log:
      mode: console
    security:
      admin_user: admin
      admin_password: admin
EOF

log "waiting for Grafana status.stage=complete"
kubectl wait --for=jsonpath='{.status.stage}'=complete \
  grafana/grafana -n grafana-operator --timeout=300s

# stage=complete only means the operator finished reconciling; the dashboard and
# folder controllers skip any instance whose pod is not serving yet.
log "waiting for the Grafana deployment to become available"
for _ in $(seq 1 60); do
  if kubectl get deploy grafana-deployment -n grafana-operator >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
kubectl wait --for=condition=Available deploy/grafana-deployment \
  -n grafana-operator --timeout=300s

log "generating dashboard JSON"
(cd "${ROOT}" && go run ./cmd/generate)

log "applying the folders wave"
kubectl apply -k "${ROOT}/deploy/folders"

# The expected resource set is derived from the generated manifests so a newly
# registered dashboard, folder, or alert group is covered without editing this script.
manifest_names() {
  local directory="$1" prefix="$2"
  local file
  shopt -s nullglob
  for file in "${ROOT}"/deploy/"${directory}"/"${prefix}"-*.yaml; do
    basename "${file}" .yaml | sed "s/^${prefix}-//"
  done
}
mapfile -t FOLDERS < <(manifest_names folders grafanamanifest)
mapfile -t DASHBOARDS < <(manifest_names dashboards grafanamanifest)
mapfile -t ALERT_GROUPS < <(manifest_names dashboards grafanaalertrulegroup)
if [[ ${#DASHBOARDS[@]} -eq 0 || ${#FOLDERS[@]} -eq 0 ]]; then
  echo "no generated GrafanaManifest resources found under deploy/folders and deploy/dashboards" >&2
  exit 1
fi
log "expecting ${#FOLDERS[@]} folders, ${#DASHBOARDS[@]} dashboards, ${#ALERT_GROUPS[@]} alert groups"

# wait_condition <kind> <name> <condition type> waits until the named condition is
# True. NoMatchingInstance=True means the operator has not paired the resource with the
# Grafana instance yet; it is transient right after apply and must not count as success.
wait_condition() {
  local kind="$1" name="$2" want="$3"
  local conditions
  for _ in $(seq 1 60); do
    conditions="$(kubectl get "${kind}" "${name}" -n grafana-operator \
      -o jsonpath='{range .status.conditions[*]}{.type}={.status} {end}')"
    if [[ "${conditions}" == *"${want}=True"* ]]; then
      log "${kind}/${name} ready (${conditions})"
      return 0
    fi
    sleep 5
  done
  echo "${kind}/${name} never reached ${want}=True (last: ${conditions})" >&2
  kubectl get "${kind}" "${name}" -n grafana-operator -o yaml >&2
  return 1
}

# The folders must exist before the boards. A board whose grafana.app/folder
# names a missing folder fails outright rather than retrying inside the deadline,
# so the two waves are applied in order here exactly as a consumer's Flux
# Kustomizations would with dependsOn.
log "waiting for the folders to reach ManifestSynchronized"
for folder in "${FOLDERS[@]}"; do
  wait_condition grafanamanifest "${folder}" ManifestSynchronized
done

log "applying the dashboards wave"
kubectl apply -k "${ROOT}/deploy/dashboards"

log "waiting for the dashboards to reach ManifestSynchronized"
for dash in "${DASHBOARDS[@]}"; do
  wait_condition grafanamanifest "${dash}" ManifestSynchronized
done

log "verifying GrafanaAlertRuleGroups"
if ! kubectl get crd grafanaalertrulegroups.grafana.integreatly.org >/dev/null 2>&1; then
  echo "GrafanaAlertRuleGroup CRD is required but was not installed" >&2
  exit 1
fi
for group in "${ALERT_GROUPS[@]}"; do
  kubectl get grafanaalertrulegroup "${group}" -n grafana-operator
done

# Reading the board back is the point of this whole section. A CR condition only
# says the operator finished its own work; on 12.x the legacy path reported
# success for payloads Grafana could not render. Ask Grafana's own apiserver what
# it stored, and compare it with what was generated.
log "port-forwarding Grafana"
kubectl port-forward -n grafana-operator svc/grafana-service "${GRAFANA_PORT}:3000" \
  >/tmp/grafana-e2e-portforward.log 2>&1 &
PF_PID=$!
for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${GRAFANA_PORT}/api/health" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl -fsS "http://127.0.0.1:${GRAFANA_PORT}/api/health" >/dev/null

log "checking Grafana serves dashboard.grafana.app/v2"
if ! curl -fsS -u admin:admin "http://127.0.0.1:${GRAFANA_PORT}/apis" |
  grep -q '"dashboard.grafana.app"'; then
  echo "this Grafana does not serve dashboard.grafana.app; the generated manifests cannot apply" >&2
  exit 1
fi

log "reading every dashboard back through /apis"
for dash in "${DASHBOARDS[@]}"; do
  domain_spec="$(find "${ROOT}/generated/dashboards" -name "${dash}.spec.json" -print -quit)"
  if [[ -z "${domain_spec}" ]]; then
    echo "no generated spec for ${dash}" >&2
    exit 1
  fi
  want_title="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["title"])' "${domain_spec}")"
  want_elements="$(python3 -c 'import json,sys; print(len(json.load(open(sys.argv[1]))["elements"]))' "${domain_spec}")"

  stored="$(curl -fsS -u admin:admin \
    "http://127.0.0.1:${GRAFANA_PORT}/apis/dashboard.grafana.app/v2/namespaces/default/dashboards/${dash}")"
  got_title="$(printf '%s' "${stored}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["spec"]["title"])')"
  got_elements="$(printf '%s' "${stored}" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)["spec"]["elements"]))')"

  if [[ "${got_title}" != "${want_title}" || "${got_elements}" != "${want_elements}" ]]; then
    echo "${dash} read back as title=${got_title} elements=${got_elements}, want title=${want_title} elements=${want_elements}" >&2
    exit 1
  fi
  log "${dash}: ${got_elements} elements, title ${got_title}"
done

log "e2e smoke checks passed"
