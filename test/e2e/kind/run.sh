#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER_NAME="grafana-as-code-e2e"
KIND_CONFIG="${ROOT}/test/e2e/kind/kind-config.yaml"
OPERATOR_CHART_VERSION="5.25.0"
GRAFANA_VERSION="12.0.0"
REGISTRY_HOST="registry.grafana-operator.svc.cluster.local:5000"
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
    - port: 5000
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

log "generating dashboard JSON"
(cd "${ROOT}" && go run ./cmd/generate)

log "pushing dashboard JSON to in-cluster registry"
kubectl port-forward -n grafana-operator svc/registry 5001:5000 >/tmp/grafana-e2e-registry.log 2>&1 &
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
  oras push --plain-http "localhost:5001/grafana-dashboards:${OCI_TAG}" \
    --artifact-type application/vnd.grafana.dashboard+json \
    kubernetes-cluster-overview.spec.json:application/json \
    pg-io-waits.spec.json:application/json
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

log "verifying GrafanaFolders"
kubectl get grafanafolder kubernetes databases -n grafana-operator

log "verifying GrafanaDashboards use spec.oci"
for dash in kubernetes-cluster-overview pg-io-waits; do
  kubectl get grafanadashboard "${dash}" -n grafana-operator
  ref="$(kubectl get grafanadashboard "${dash}" -n grafana-operator -o jsonpath='{.spec.oci.reference}')"
  path="$(kubectl get grafanadashboard "${dash}" -n grafana-operator -o jsonpath='{.spec.oci.path}')"
  if [[ "${ref}" != "${REGISTRY_HOST}/grafana-dashboards:${OCI_TAG}" ]]; then
    echo "unexpected oci.reference for ${dash}: ${ref}" >&2
    exit 1
  fi
  if [[ "${path}" != "${dash}.spec.json" ]]; then
    echo "unexpected oci.path for ${dash}: ${path}" >&2
    exit 1
  fi
done

log "verifying GrafanaAlertRuleGroups"
if ! kubectl get crd grafanaalertrulegroups.grafana.integreatly.org >/dev/null 2>&1; then
  echo "GrafanaAlertRuleGroup CRD is required but was not installed" >&2
  exit 1
fi
kubectl get grafanaalertrulegroup kubernetes databases -n grafana-operator

log "e2e smoke checks passed"
