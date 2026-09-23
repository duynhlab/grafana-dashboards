// Package gateway holds PromQL for the API Gateway folder.
//
// EGEdge* below ports the "Envoy Gateway — Edge Overview" board from the
// homelab GitOps repo:
//
//	source: kubernetes/infra/configs/observability/grafana/dashboards/eg-edge.json
//	repo:   github.com/duynhlab/homelab
//	commit: d895ea58a8a2c28b13da9a1efd3e913805ffdb8f
//
// The board reads two jobs from one Envoy Gateway install: `envoy-gateway`
// (the Envoy data-plane proxy, gateway:19005 /stats/prometheus) and
// `envoy-gateway-controller` (the EG control plane, gateway:19001). Those two
// job values and the `envoy-gateway` namespace name the component's own
// install, not a caller-supplied environment, so they stay hardcoded here the
// same way the source dashboard hardcodes them.
//
// This board does not supersede the four vendored envoyproxy/gateway
// dashboards (Global, Cluster Backends, gRPC, Rate Limiting) shipped as
// upstream JSON in the API Gateway folder. Those stay vendored; eg-edge is
// the local-stack SRE view that sits next to them.
package gateway

const (
	// Edge Overview: golden signals on the edge (downstream) traffic listener.
	// envoy_http_conn_manager_prefix=~"https?-.*" scopes every query in this
	// group to the edge HTTP(S) listeners, excluding admin/readiness/stats.

	EGEdgeRequestsPerSec = `sum(rate(envoy_http_downstream_rq_total{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval]))`

	EGEdgeErrorRate5xx = `sum(rate(envoy_http_downstream_rq_xx{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*",envoy_response_code_class="5"}[$__rate_interval])) / sum(rate(envoy_http_downstream_rq_xx{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval]))`

	EGEdgeP99Latency = `histogram_quantile(0.99, sum by (le) (rate(envoy_http_downstream_rq_time_bucket{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval])))`

	// EGEdgeAvailabilityWindow is a dashboard-window aggregate ($__range, not
	// $__rate_interval): 1 - (5xx / all responses) over the whole selected range.
	EGEdgeAvailabilityWindow = `1 - (sum(increase(envoy_http_downstream_rq_xx{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*",envoy_response_code_class="5"}[$__range])) / sum(increase(envoy_http_downstream_rq_xx{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__range])))`

	// Data Plane: per-route and per-cluster detail. envoy_cluster_name=~"httproute/.*"
	// scopes upstream-cluster queries to clusters Envoy Gateway programs for an
	// HTTPRoute (httproute/<ns>/<route>/rule/<n>), excluding internal clusters.

	EGEdgeRequestRateByRoute = `sum by (envoy_cluster_name) (rate(envoy_cluster_upstream_rq_total{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval]))`

	EGEdgeResponseRate4xx5xx = `sum by (envoy_response_code_class) (rate(envoy_http_downstream_rq_xx{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*",envoy_response_code_class=~"4|5"}[$__rate_interval]))`

	EGEdgeDownstreamLatencyP50 = `histogram_quantile(0.50, sum by (le) (rate(envoy_http_downstream_rq_time_bucket{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval])))`
	EGEdgeDownstreamLatencyP95 = `histogram_quantile(0.95, sum by (le) (rate(envoy_http_downstream_rq_time_bucket{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval])))`
	EGEdgeDownstreamLatencyP99 = `histogram_quantile(0.99, sum by (le) (rate(envoy_http_downstream_rq_time_bucket{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"}[$__rate_interval])))`

	EGEdgeUpstreamLatencyP95ByRoute = `histogram_quantile(0.95, sum by (le, envoy_cluster_name) (rate(envoy_cluster_upstream_rq_time_bucket{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval])))`

	EGEdgeUpstreamRetries        = `sum(rate(envoy_cluster_upstream_rq_retry{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval]))`
	EGEdgeUpstreamRetrySuccesses = `sum(rate(envoy_cluster_upstream_rq_retry_success{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval]))`
	EGEdgeUpstreamTimeouts       = `sum(rate(envoy_cluster_upstream_rq_timeout{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval]))`
	EGEdgeUpstreamPerTryTimeouts = `sum(rate(envoy_cluster_upstream_rq_per_try_timeout{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"}[$__rate_interval]))`

	EGEdgeActiveConnectionsDownstream = `sum(envoy_http_downstream_cx_active{job="envoy-gateway",envoy_http_conn_manager_prefix=~"https?-.*"})`
	EGEdgeActiveConnectionsListener   = `sum by (envoy_listener_address) (envoy_listener_downstream_cx_active{job="envoy-gateway",envoy_listener_address="0.0.0.0_8000"})`
	EGEdgeActiveConnectionsUpstream   = `sum(envoy_cluster_upstream_cx_active{job="envoy-gateway",envoy_cluster_name=~"httproute/.*"})`

	// Control Plane (Envoy Gateway): the `watchable` package is EG's internal
	// pub/sub between the provider, gateway-api, xds-translator, xds-server and
	// infrastructure runners, keyed by the `runner` label.

	EGEdgeWatchableEvents    = `sum by (runner) (rate(watchable_event_total{job="envoy-gateway-controller"}[$__rate_interval]))`
	EGEdgeWatchablePublishes = `sum by (runner) (rate(watchable_publish_total{job="envoy-gateway-controller"}[$__rate_interval]))`

	EGEdgeWatchableSubscribeDurationP99 = `histogram_quantile(0.99, sum by (le, runner) (rate(watchable_subscribe_duration_seconds_bucket{job="envoy-gateway-controller"}[$__rate_interval])))`

	EGEdgeWatchableQueueDepth = `sum by (runner) (watchable_depth{job="envoy-gateway-controller"})`

	EGEdgeXDSSnapshotCreates = `sum by (status) (increase(xds_snapshot_create_total{job="envoy-gateway-controller"}[$__rate_interval]))`
	EGEdgeXDSSnapshotUpdates = `sum by (nodeID, status) (increase(xds_snapshot_update_total{job="envoy-gateway-controller"}[$__rate_interval]))`

	// EGEdgeGatewayStatusUpdates reads status_update_total{kind}, which the
	// Kubernetes provider emits when it writes Gateway/HTTPRoute status; it stays
	// empty when Envoy Gateway runs outside Kubernetes.
	EGEdgeGatewayStatusUpdates = `sum by (kind) (rate(status_update_total{job="envoy-gateway-controller"}[$__rate_interval]))`
	// EGEdgeProviderStatusQueueDepth is the fallback for the query above: the
	// provider runner's own *-status watchable queue depths, which exist in
	// every install and read 0 once status work drains.
	EGEdgeProviderStatusQueueDepth = `sum by (message) (watchable_depth{job="envoy-gateway-controller",runner="provider",message=~".*-status"})`

	// EGEdgeWatchablePanicsRecovered falls back to vector(0) so the series stays
	// drawable before the first (hopefully nonexistent) recovered panic.
	EGEdgeWatchablePanicsRecovered = `sum(increase(watchable_panics_recovered_total{job="envoy-gateway-controller"}[$__rate_interval])) or vector(0)`
	EGEdgeCertWatcherReadErrors    = `sum(increase(certwatcher_read_certificate_errors_total{job="envoy-gateway-controller"}[$__rate_interval]))`
	EGEdgeCertWatcherReadsTotal    = `sum(increase(certwatcher_read_certificate_total{job="envoy-gateway-controller"}[$__rate_interval]))`

	// Infrastructure: process-level metrics from both jobs, plus the
	// kube-state-metrics/cAdvisor pair for the workload's own namespace.

	EGEdgeControlPlaneCPU = `rate(process_cpu_seconds_total{job="envoy-gateway-controller"}[$__rate_interval])`

	EGEdgeControlPlaneMemoryRSS = `process_resident_memory_bytes{job="envoy-gateway-controller"}`
	EGEdgeEnvoyMemoryAllocated  = `envoy_server_memory_allocated{job="envoy-gateway"}`
	EGEdgeEnvoyMemoryHeapSize   = `envoy_server_memory_heap_size{job="envoy-gateway"}`
	EGEdgeEnvoyMemoryPhysical   = `envoy_server_memory_physical_size{job="envoy-gateway"}`

	EGEdgeEnvoyLive     = `envoy_server_live{job="envoy-gateway"}`
	EGEdgeEGScrapeUp    = `up{job="envoy-gateway-controller"}`
	EGEdgeEnvoyScrapeUp = `up{job="envoy-gateway"}`

	// EGEdgePodRestarts and EGEdgeContainerCPU read kube-state-metrics/cAdvisor
	// for the envoy-gateway namespace; both are empty off-cluster (local-stack
	// runs neither exporter), so EGEdgeEnvoyUptimeFallback (a sawtooth on
	// restart) is kept as the always-live signal in the same panel.
	EGEdgePodRestarts         = `sum by (pod) (kube_pod_container_status_restarts_total{exported_namespace="envoy-gateway"})`
	EGEdgeContainerCPU        = `sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="envoy-gateway"}[$__rate_interval]))`
	EGEdgeEnvoyUptimeFallback = `envoy_server_uptime{job="envoy-gateway"}`
)
