#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER_NAME="grafana-as-code-e2e"
KIND_CONFIG="${ROOT}/test/e2e/kind/kind-config.yaml"
OPERATOR_CHART_VERSION="5.22.2"
GRAFANA_VERSION="12.0.0"

log() { echo "[e2e] $*"; }

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1"; exit 1; }
}

require kind
require kubectl
require helm

dump() {
  log "dumping cluster state"
  kubectl get grafana,deploy,pod,event -n grafana-operator || true
  kubectl describe grafana grafana -n grafana-operator || true
  kubectl logs -n grafana-operator -l app.kubernetes.io/name=grafana-operator --tail=200 || true
  kubectl logs -n grafana-operator -l app=grafana --tail=200 || true
}

cleanup() {
  log "cleaning up kind cluster ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}" >/dev/null 2>&1 || true
}

on_error() {
  dump
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

log "applying generated dashboard manifests"
kubectl apply -k "${ROOT}/deploy"

log "verifying GrafanaFolders"
kubectl get grafanafolder kubernetes databases -n grafana-operator

log "verifying GrafanaDashboards"
for dash in kubernetes-cluster-overview pg-io-waits; do
  kubectl get grafanadashboard "${dash}" -n grafana-operator
done

log "verifying GrafanaAlertRuleGroups"
if ! kubectl get crd grafanaalertrulegroups.grafana.integreatly.org >/dev/null 2>&1; then
  echo "GrafanaAlertRuleGroup CRD is required but was not installed" >&2
  exit 1
fi
kubectl get grafanaalertrulegroup kubernetes databases -n grafana-operator

log "e2e smoke checks passed"
