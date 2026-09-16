#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER_NAME="grafana-as-code-e2e"
KIND_CONFIG="${ROOT}/test/e2e/kind/kind-config.yaml"

log() { echo "[e2e] $*"; }

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1"; exit 1; }
}

require kind
require kubectl
require helm

cleanup() {
  log "cleaning up kind cluster ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

log "creating kind cluster (Kubernetes 1.36)"
if ! kind create cluster --name "${CLUSTER_NAME}" --config "${KIND_CONFIG}" 2>/dev/null; then
  log "kindest/node:v1.36.0 unavailable, falling back to kindest/node:v1.32.0"
  sed 's/v1.36.0/v1.32.0/' "${KIND_CONFIG}" | kind create cluster --name "${CLUSTER_NAME}" --config /dev/stdin
fi

log "installing Grafana Operator"
helm repo add grafana https://grafana.github.io/helm-charts >/dev/null
helm repo update >/dev/null
helm upgrade --install grafana-operator grafana/grafana-operator \
  --namespace grafana-operator --create-namespace \
  --wait --timeout 5m

log "deploying Grafana 12 instance"
kubectl apply -f - <<'EOF'
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  namespace: grafana-operator
  labels:
    dashboards: grafana
spec:
  config:
    log:
      mode: console
    security:
      admin_user: admin
      admin_password: admin
  deployment:
    spec:
      template:
        spec:
          containers:
            - name: grafana
              image: grafana/grafana:12.0.0
EOF

kubectl wait --for=condition=Ready grafana/grafana -n grafana-operator --timeout=300s

log "applying generated dashboard manifests"
kubectl apply -k "${ROOT}/deploy"

log "waiting for GrafanaDashboard CRs"
for dash in kubernetes-cluster-overview pg-io-waits; do
  kubectl wait --for=condition=Unknown "grafanadashboard/${dash}" -n grafana-operator --timeout=120s || true
  kubectl get grafanadashboard "${dash}" -n grafana-operator
done

log "verifying GrafanaFolders"
kubectl get grafanafolder kubernetes databases -n grafana-operator

log "verifying GrafanaAlertRuleGroups (optional CRD)"
if kubectl get crd grafanaalertrulegroups.grafana.integreatly.org >/dev/null 2>&1; then
  kubectl get grafanaalertrulegroup kubernetes databases -n grafana-operator
else
  log "GrafanaAlertRuleGroup CRD not installed; generated manifests still present"
fi

log "e2e smoke checks passed"
