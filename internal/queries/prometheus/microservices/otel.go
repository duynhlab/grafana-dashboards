package microservices

// PromQL for the "Microservices (OTel)" board (UID microservices-monitoring-001-otel).
//
// Ported verbatim from helm-charts, charts/grafana-dashboards/dashboards/
// microservices/microservices-dashboard-otel.json at commit 0a51fdd. The only
// edits are mechanical: the legacy $rate custom interval is replaced by
// $__rate_interval, and stray `, }` / double spaces inside selectors are
// cleaned up. $__range stays where the source uses it (the two "total in the
// selected range" panels). The source's dead $namespace variable
// (deployment_environment_name) is dropped: no panel query referenced it.
//
// Metric model: OpenTelemetry semantic conventions emitted by the Go services.
//
//	HTTP server   http_server_request_duration_seconds_{bucket,count},
//	              http_server_request_body_size_bytes_sum,
//	              http_server_response_body_size_bytes_sum
//	              labels: service_name, http_route, http_request_method,
//	                      http_response_status_code
//	gRPC RED      rpc_server_call_duration_seconds_{count,bucket},
//	              rpc_client_call_duration_seconds_{count,bucket}
//	              labels: service_name, rpc_method, server_address,
//	                      rpc_response_status_code
//	Go runtime    go_goroutine_count, go_memory_used_bytes,
//	              go_memory_allocations_total, go_memory_gc_goal_bytes
//	DB client     db_client_operation_duration_seconds_bucket,
//	              db_client_operation_errors_total (otelpgx),
//	              pgxpool_{acquired_connections,max_connections,
//	                       empty_acquire_total,
//	                       empty_acquire_wait_time_nanoseconds_total}
const (
	// OTelAppLabelValues drives the only template variable. go_goroutine_count is
	// the cheapest always-present series on every OTel Go service, so it is the
	// most reliable source of the service_name list.
	OTelAppLabelValues = `label_values(go_goroutine_count, service_name)`
)

// Overview & Key Metrics. The latency percentiles look at 2xx responses only:
// a flood of fast 4xx or slow 5xx would otherwise move the number for reasons
// that have nothing to do with how fast the service serves real work.
const (
	OTelP99SuccessLatency = `histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`
	OTelP95SuccessLatency = `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`
	OTelP50SuccessLatency = `histogram_quantile(0.5, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`

	OTelTotalRPS   = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelSuccessRPS = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval]))`
	OTelErrorRPS   = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"4..|5.."}[$__rate_interval]))`

	// OTelTotalRequests counts requests over the whole selected time range.
	OTelTotalRequests = `sum(increase(http_server_request_duration_seconds_count{service_name=~"$app"}[$__range]))`
)

// OTelSuccessRatePct is the exact complement of OTelErrorRatePct: every request
// the service handled correctly, 3xx and 4xx answers to client mistakes
// included. Success + Error always adds up to 100.
const OTelSuccessRatePct = `(
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code!~"5.."}[$__rate_interval]))
  /
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))
) * 100`

// OTelErrorRatePct counts 5xx only, so it matches the MicroserviceHighErrorRate
// alert and the availability SLO, which are both 5xx-only.
const OTelErrorRatePct = `(
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5.."}[$__rate_interval]))
  /
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))
) * 100`

// OTelApdex scores user satisfaction in [0,1]: satisfying below 0.5s, tolerating
// up to 2s, frustrated beyond. The `> 0 or vector(1)` denominator guard keeps the
// panel at 1 instead of NaN while there is no traffic at all.
const OTelApdex = `(sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="0.5"}[$__rate_interval])) + 0.5 * (sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="2"}[$__rate_interval])) - sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="0.5"}[$__rate_interval])))) / (sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval])) > 0 or vector(1))`

// Traffic & Requests. The http_route!="" guard drops the unrouted catch-all
// series that the instrumentation emits for requests which never matched a mux
// pattern; without it a single 404 scanner dominates the per-endpoint views.
const (
	// The two pies count requests over the selected range and leave out the
	// kubelet probes: a rate pie showed only the last few minutes, so an idle
	// board read ~100% 200 from /health and hid every 201 and 409 of the range.
	OTelStatusCodeDistribution   = `sum(increase(http_server_request_duration_seconds_count{service_name=~"$app", http_route!="/health"}[$__range])) by (http_response_status_code)`
	OTelTotalRequestsByRoute     = `sum(increase(http_server_request_duration_seconds_count{service_name=~"$app", http_route!="", http_route!="/health"}[$__range])) by (http_route)`
	OTelRequestRateByRoute       = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_route)`
	OTelRequestRateByMethodRoute = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_request_method, http_route)`
)

// Errors & Performance.
//
// OTelServerErrorRatioByMethodRoute divides by the full per-route request rate
// rather than filtering the denominator, which keeps a continuous 0 baseline for
// healthy routes so that legend means stay honest.
//
// The route-latency percentiles keep http_response_status_code in the grouping,
// as the source does, so a slow error path is not averaged into a fast success.
const (
	OTelServerErrorRatioByMethodRoute = `(sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5..", http_route!=""}[$__rate_interval])) by (http_request_method, http_route) / sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_request_method, http_route)) * 100`

	OTelClientErrorRPSByService = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"4.."}[$__rate_interval])) by (service_name)`
	OTelServerErrorRPSByService = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5.."}[$__rate_interval])) by (service_name)`

	OTelRouteLatencyP95 = `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route, http_response_status_code))`
	OTelRouteLatencyP50 = `histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route, http_response_status_code))`
	OTelRouteLatencyP99 = `histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route, http_response_status_code))`
)

// Go Runtime & HTTP I/O, as emitted by the OTel Go runtime instrumentation.
//
// OTelGoAllocationRate is the closest available leading indicator of GC
// pressure: the OTel runtime exposes no GC-pause metric at all.
//
// OTelGoGCPacing compares live heap against the pacer goal. Values parked near
// 1.0 mean the heap keeps hitting the goal, i.e. back-to-back GC cycles.
//
// The network pair measures HTTP body bytes only, not TCP/IP overhead.
const (
	OTelGoMemoryUsed     = `sum(go_memory_used_bytes{service_name=~"$app"}) by (service_name)`
	OTelGoAllocationRate = `sum by (service_name) (rate(go_memory_allocations_total{service_name=~"$app"}[$__rate_interval]))`
	OTelGoroutines       = `sum(go_goroutine_count{service_name=~"$app"}) by (service_name)`
	OTelGoGCPacing       = `sum by (service_name) (go_memory_used_bytes{service_name=~"$app"}) / sum by (service_name) (go_memory_gc_goal_bytes{service_name=~"$app"})`

	OTelHTTPResponseBytesRate = `sum(rate(http_server_response_body_size_bytes_sum{service_name=~"$app"}[$__rate_interval])) by (service_name)`
	OTelHTTPRequestBytesRate  = `sum(rate(http_server_request_body_size_bytes_sum{service_name=~"$app"}[$__rate_interval])) by (service_name)`
)

// gRPC East-West (RED). Server series answer "what is being asked of me",
// client series answer "what am I asking of whom", so the client side keeps
// server_address to name the upstream.
const (
	OTelGRPCServerRPS = `sum by (rpc_method) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelGRPCClientRPS = `sum by (rpc_method, server_address) (rate(rpc_client_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`

	OTelGRPCServerErrorRate = `sum by (rpc_method) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app", rpc_response_status_code!="OK"}[$__rate_interval]))`
	OTelGRPCClientErrorRate = `sum by (rpc_method, server_address) (rate(rpc_client_call_duration_seconds_count{service_name=~"$app", rpc_response_status_code!="OK"}[$__rate_interval]))`

	OTelGRPCServerP95 = `histogram_quantile(0.95, sum by (le, rpc_method) (rate(rpc_server_call_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
	OTelGRPCClientP95 = `histogram_quantile(0.95, sum by (le, rpc_method, server_address) (rate(rpc_client_call_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
)

// gRPC East-West (RED) — Per Callee: the same server series rolled up to the
// callee service, so one busy or failing service stands out without reading
// every method.
const (
	OTelGRPCServerRPSPerCallee        = `sum by (service_name) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelGRPCServerErrorRatioPerCallee = `sum by (service_name) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app", rpc_response_status_code!="OK"}[$__rate_interval])) / sum by (service_name) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelGRPCServerP95PerCallee        = `histogram_quantile(0.95, sum by (service_name, le) (rate(rpc_server_call_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
)

// Database (client — otelpgx) and the pgx connection pool.
//
// OTelDBQueryP95ByService restricts to pgx_operation_type="query" so that the
// per-service number stays a statement-latency number; acquire and connect
// operations have a completely different scale and would swamp it.
//
// The pool trio escalates in sensitivity: in-flight acquisitions show concurrent
// DB work, saturation warns once acquired/max sits near 1 (the
// PgxPoolNearExhaustion condition), and contention is the earliest signal of
// all because an acquire that had to wait already queued behind a busy pool.
// The wait-time counter is in nanoseconds, hence the /1e9.
const (
	OTelDBQueryP95ByService = `histogram_quantile(0.95, sum by (le,service_name) (rate(db_client_operation_duration_seconds_bucket{service_name=~"$app", pgx_operation_type="query"}[$__rate_interval])))`
	OTelDBP95ByOperation    = `histogram_quantile(0.95, sum by (le,pgx_operation_type) (rate(db_client_operation_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
	OTelDBOperationErrors   = `sum by (service_name) (rate(db_client_operation_errors_total{service_name=~"$app"}[$__rate_interval]))`

	OTelPoolAcquiredConns = `sum by (service_name) (pgxpool_acquired_connections{service_name=~"$app"})`
	OTelPoolSaturation    = `sum by (service_name) (pgxpool_acquired_connections{service_name=~"$app"}) / sum by (service_name) (pgxpool_max_connections{service_name=~"$app"})`

	OTelPoolEmptyAcquireRate     = `sum by (service_name) (rate(pgxpool_empty_acquire_total{service_name=~"$app"}[$__rate_interval]))`
	OTelPoolEmptyAcquireWaitSecs = `sum by (service_name) (rate(pgxpool_empty_acquire_wait_time_nanoseconds_total{service_name=~"$app"}[$__rate_interval])) / 1e9`
)

// Deploys and versions. Neither is in the helm-charts source. Both read the
// service_version resource attribute that every OTel Go service stamps on its
// runtime series, so they follow service_name exactly: workers and mockpay
// included, which a Kubernetes namespace or Deployment name would not match.
//
// OTelVersionFirstSeen is the deploy-marker annotation: a version that has
// samples now and had none two minutes earlier. A rollout that keeps the same
// version (a config-only change) does not mark, and neither does a pod
// restart; KEDA scaling cannot fire it either, unlike a Deployment generation.
// OTelRunningVersions counts the instances reporting each version, so a
// rollout in progress shows two rows for one service.
const (
	OTelVersionFirstSeen = `count by (service_name, service_version) (go_goroutine_count{service_name=~"$app"}) unless count by (service_name, service_version) (go_goroutine_count{service_name=~"$app"} offset 2m)`
	OTelRunningVersions  = `count by (service_name, service_version) (go_goroutine_count{service_name=~"$app"})`
)
