// Package platform holds PromQL constants for cluster-support boards: the
// things every workload depends on but no workload owns.
//
// Source: homelab kubernetes/infra/configs/observability/grafana/dashboards/cert-manager.json
// at commit 644d706a5f748c7dd5d5018ee5940bc00da2e1e6.
package platform

// PromQL for the cert-manager dashboard. Metrics come from the cert-manager
// controller (certmanager_* namespace), client-go work queues
// (workqueue_*), and standard process/container metrics, scraped from any
// job whose name contains "cert-manager".
const (
	// Certificate health.
	CertManagerCertificatesReady    = `count(certmanager_certificate_ready_status{condition="True"} == 1)`
	CertManagerCertificatesNotReady = `count(certmanager_certificate_ready_status{condition="False"} == 1) OR on() vector(0)`
	CertManagerCertsExpiringIn7d    = `count((certmanager_certificate_expiration_timestamp_seconds - time()) < 7*24*3600) OR on() vector(0)`
	CertManagerCertsExpiringIn24h   = `count((certmanager_certificate_expiration_timestamp_seconds - time()) < 24*3600) OR on() vector(0)`
	CertManagerTimeToExpiry         = `certmanager_certificate_expiration_timestamp_seconds - time()`
	CertManagerTimeToRenewal        = `certmanager_certificate_renewal_timestamp_seconds - time()`
	CertManagerReadyStatusByCond    = `sum(certmanager_certificate_ready_status) by (condition)`
	CertManagerNotReadyByNamespace  = `(sum(certmanager_certificate_ready_status{condition="False"}) by (namespace)) or vector(0)`

	// Controller.
	CertManagerControllerSyncRate  = `sum(rate(certmanager_controller_sync_call_count[$__rate_interval])) by (controller)`
	CertManagerControllerErrorRate = `(sum(rate(certmanager_controller_sync_error_count[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(certmanager_controller_sync_call_count[$__rate_interval])), 1)`
	CertManagerWorkqueueDepth      = `sum(workqueue_depth{job=~".*cert-manager.*"}) by (name)`
	CertManagerWorkqueueAddRate    = `sum(rate(workqueue_adds_total{job=~".*cert-manager.*"}[$__rate_interval])) by (name)`

	// ACME.
	CertManagerACMERequestRate = `sum(rate(certmanager_http_acme_client_request_count[$__rate_interval])) by (status, path)`
	CertManagerACMEAvgDuration = `sum(rate(certmanager_http_acme_client_request_duration_seconds_sum[$__rate_interval])) / clamp_min(sum(rate(certmanager_http_acme_client_request_duration_seconds_count[$__rate_interval])), 1)`

	// Resources.
	CertManagerCPUByComponent    = `sum(rate(process_cpu_seconds_total{job=~".*cert-manager.*"}[$__rate_interval])) by (job)`
	CertManagerMemoryByComponent = `sum(process_resident_memory_bytes{job=~".*cert-manager.*"}) by (job)`
)
