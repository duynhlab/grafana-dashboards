#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER_NAME="grafana-as-code-e2e"
KIND_CONFIG="${ROOT}/test/e2e/kind/kind-config.yaml"
OPERATOR_CHART_VERSION="5.25.0"
GRAFANA_VERSION="12.0.0"
# The GrafanaDashboard CRD rejects a port in spec.oci.reference, so the
# in-cluster registry is exposed on port 80.
REGISTRY_HOST="registry.grafana-operator.svc.cluster.local"
OCI_TAG="e2e"
PF_PID=""

log() { echo "[e2e] $*"; }

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1"; exit 1; }
}

require kind
require kubectl
require helm
require oras
require go
require curl

dump() {
  log "dumping cluster state"
  kubectl get grafana,grafanadashboard,grafanafolder,grafanaalertrulegroup,deploy,pod,event -n grafana-operator || true
  kubectl describe grafana grafana -n grafana-operator || true
  kubectl describe grafanadashboard -n grafana-operator || true
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

log "deploying in-cluster OCI registry"
kubectl apply -f - <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
  namespace: grafana-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
        - name: registry
          image: registry:2
          ports:
            - containerPort: 5000
---
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: grafana-operator
spec:
  selector:
    app: registry
  ports:
    - port: 80
      targetPort: 5000
EOF

kubectl wait --for=condition=Available deploy/registry -n grafana-operator --timeout=180s

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

log "pushing dashboard JSON to in-cluster registry"
kubectl port-forward -n grafana-operator svc/registry 5001:80 >/tmp/grafana-e2e-registry.log 2>&1 &
PF_PID=$!
for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:5001/v2/ >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS http://127.0.0.1:5001/v2/ >/dev/null

(
  cd "${ROOT}/generated/dashboards"
  # Paths stay domain-relative (kubernetes/foo.spec.json): oras records them as
  # the layer title, which is what spec.oci.path must equal.
  spec_files=()
  while IFS= read -r spec; do
    spec_files+=("${spec}:application/json")
  done < <(find . -name '*.spec.json' -printf '%P\n' | sort)
  if [[ ${#spec_files[@]} -eq 0 ]]; then
    echo "no dashboard spec files found under generated/dashboards" >&2
    exit 1
  fi
  log "pushing ${#spec_files[@]} dashboard spec files"
  oras push --plain-http "localhost:5001/grafana-dashboards:${OCI_TAG}" \
    --artifact-type application/vnd.grafana.dashboard+json \
    "${spec_files[@]}"
)

log "generating GrafanaDashboard CRs that fetch spec.oci"
(
  cd "${ROOT}"
  OCI_REFERENCE="${REGISTRY_HOST}/grafana-dashboards:${OCI_TAG}" \
    OCI_INSECURE_PLAIN_HTTP=true \
    OCI_PULL_SECRET= \
    go run ./cmd/generate
)

log "applying generated dashboard manifests"
kubectl apply -k "${ROOT}/deploy"

# The expected resource set is derived from the generated manifests so a newly
# registered dashboard, folder, or alert group is covered without editing this script.
manifest_names() {
  local kind="$1"
  local file
  shopt -s nullglob
  for file in "${ROOT}"/deploy/manifests/"${kind}"-*.yaml; do
    basename "${file}" .yaml | sed "s/^${kind}-//"
  done
}
mapfile -t FOLDERS < <(manifest_names grafanafolder)
mapfile -t DASHBOARDS < <(manifest_names grafanadashboard)
mapfile -t ALERT_GROUPS < <(manifest_names grafanaalertrulegroup)
if [[ ${#DASHBOARDS[@]} -eq 0 || ${#FOLDERS[@]} -eq 0 ]]; then
  echo "no generated GrafanaDashboard or GrafanaFolder manifests found under deploy/manifests" >&2
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

log "verifying GrafanaFolders"
for folder in "${FOLDERS[@]}"; do
  kubectl get grafanafolder "${folder}" -n grafana-operator
done

log "verifying GrafanaDashboards use spec.oci"
for dash in "${DASHBOARDS[@]}"; do
  kubectl get grafanadashboard "${dash}" -n grafana-operator
  ref="$(kubectl get grafanadashboard "${dash}" -n grafana-operator -o jsonpath='{.spec.oci.reference}')"
  path="$(kubectl get grafanadashboard "${dash}" -n grafana-operator -o jsonpath='{.spec.oci.path}')"
  if [[ "${ref}" != "${REGISTRY_HOST}/grafana-dashboards:${OCI_TAG}" ]]; then
    echo "unexpected oci.reference for ${dash}: ${ref}" >&2
    exit 1
  fi
  # The path is <domain>/<uid>.spec.json. The domain cannot be derived from the
  # CR name, so assert the shape and that the file the CR points at is one the
  # push actually shipped.
  if [[ ! "${path}" =~ ^[a-z0-9-]+/${dash}\.spec\.json$ ]]; then
    echo "unexpected oci.path for ${dash}: ${path}" >&2
    exit 1
  fi
  if [[ ! -f "${ROOT}/generated/dashboards/${path}" ]]; then
    echo "oci.path for ${dash} points at a file that was not generated: ${path}" >&2
    exit 1
  fi
done

log "waiting for the operator to create the folders in Grafana"
for folder in "${FOLDERS[@]}"; do
  wait_condition grafanafolder "${folder}" FolderSynchronized
done

log "waiting for the operator to fetch the dashboards from the registry"
for dash in "${DASHBOARDS[@]}"; do
  wait_condition grafanadashboard "${dash}" DashboardSynchronized
done

log "verifying GrafanaAlertRuleGroups"
if ! kubectl get crd grafanaalertrulegroups.grafana.integreatly.org >/dev/null 2>&1; then
  echo "GrafanaAlertRuleGroup CRD is required but was not installed" >&2
  exit 1
fi
for group in "${ALERT_GROUPS[@]}"; do
  kubectl get grafanaalertrulegroup "${group}" -n grafana-operator
done

log "e2e smoke checks passed"
