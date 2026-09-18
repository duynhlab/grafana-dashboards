package microservices

// PromQL for the "Microservices (OTel)" board (UID microservices-monitoring-001-otel).
//
// Ported from helm-charts, charts/grafana-dashboards/dashboards/microservices/
// microservices-dashboard-otel.json at commit 969fc56.
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
//
// Every selector is scoped by the single $app template variable. The legacy
// board's $rate custom interval is replaced by $__rate_interval, and its dead
// $namespace variable (deployment_environment_name) is dropped: no panel query
// ever referenced it.
const (
	// OTelAppLabelValues drives the only template variable. go_goroutine_count is
	// the cheapest always-present series on every OTel Go service, so it is the
	// most reliable source of the service_name list.
	OTelAppLabelValues = `label_values(go_goroutine_count, service_name)`
)

// Overview stats. Latency percentiles deliberately look at 2xx responses only:
// a flood of fast 4xx or slow 5xx would otherwise move the number for reasons
// that have nothing to do with how fast the service serves real work.
const (
	OTelP99SuccessLatency = `histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`
	OTelP95SuccessLatency = `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`
	OTelP50SuccessLatency = `histogram_quantile(0.5, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_response_status_code=~"2.."}[$__rate_interval])) by (le))`

	OTelTotalRPS = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelErrorRPS = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"4..|5.."}[$__rate_interval]))`
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
// alert and the availability SLO, which are both 5xx-only. Client 4xx are the
// service answering correctly and stay visible in Error RPS, the status-code
// pie and the client-error series.
const OTelErrorRatePct = `(
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5.."}[$__rate_interval]))
  /
  sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))
) * 100`

// OTelApdex scores user satisfaction in [0,1]: satisfying below 0.5s, tolerating
// up to 2s, frustrated beyond. The `> 0 or vector(1)` denominator guard keeps the
// panel at 1 instead of NaN while there is no traffic at all.
const OTelApdex = `(sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="0.5"}[$__rate_interval])) + 0.5 * (sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="2"}[$__rate_interval])) - sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", le="0.5"}[$__rate_interval])))) / (sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval])) > 0 or vector(1))`

// Traffic. The http_route!="" guard drops the unrouted catch-all series that the
// instrumentation emits for requests which never matched a mux pattern; without
// it a single 404 scanner dominates the per-endpoint breakdowns.
const (
	OTelStatusCodeDistribution   = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app"}[$__rate_interval])) by (http_response_status_code)`
	OTelRequestRateByRoute       = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_route)`
	OTelRequestRateByMethodRoute = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_request_method, http_route)`
)

// Errors and latency.
//
// OTelServerErrorRatioByMethodRoute divides by the full per-route request rate
// rather than filtering the denominator, which keeps a continuous 0 baseline for
// healthy routes so that legend means stay honest.
//
// The route-latency percentiles aggregate by (le, http_route) only. The legacy
// board also kept http_response_status_code in the grouping, but this board
// draws p50, p95 and p99 in one panel, and a third dimension would triple an
// already tripled series count for no diagnostic gain.
const (
	OTelServerErrorRatioByMethodRoute = `(sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5..", http_route!=""}[$__rate_interval])) by (http_request_method, http_route) / sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_route!=""}[$__rate_interval])) by (http_request_method, http_route)) * 100`

	OTelClientErrorRPSByService = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"4.."}[$__rate_interval])) by (service_name)`
	OTelServerErrorRPSByService = `sum(rate(http_server_request_duration_seconds_count{service_name=~"$app", http_response_status_code=~"5.."}[$__rate_interval])) by (service_name)`

	OTelRouteLatencyP50 = `histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route))`
	OTelRouteLatencyP95 = `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route))`
	OTelRouteLatencyP99 = `histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket{service_name=~"$app", http_route!=""}[$__rate_interval])) by (le, http_route))`
)

// Go runtime, as emitted by the OTel Go runtime instrumentation.
//
// OTelGoAllocationRate is the closest available leading indicator of GC
// pressure: the OTel runtime exposes no GC-pause metric at all, so a rising
// allocation rate is what there is to watch.
//
// OTelGoGCPacing compares live heap against the pacer goal. Values parked near
// 1.0 mean the heap keeps hitting the goal, i.e. back-to-back GC cycles.
//
// The network pair measures HTTP body bytes only, not TCP/IP overhead, so it is
// a bandwidth-planning signal rather than a NIC-level one.
const (
	OTelGoMemoryUsed     = `sum(go_memory_used_bytes{service_name=~"$app"}) by (service_name)`
	OTelGoAllocationRate = `sum by (service_name) (rate(go_memory_allocations_total{service_name=~"$app"}[$__rate_interval]))`
	OTelGoroutines       = `sum(go_goroutine_count{service_name=~"$app"}) by (service_name)`
	OTelGoGCPacing       = `sum by (service_name) (go_memory_used_bytes{service_name=~"$app"}) / sum by (service_name) (go_memory_gc_goal_bytes{service_name=~"$app"})`

	OTelHTTPResponseBytesRate = `sum(rate(http_server_response_body_size_bytes_sum{service_name=~"$app"}[$__rate_interval])) by (service_name)`
	OTelHTTPRequestBytesRate  = `sum(rate(http_server_request_body_size_bytes_sum{service_name=~"$app"}[$__rate_interval])) by (service_name)`
)

// gRPC east-west RED. Server series answer "what is being asked of me", client
// series answer "what am I asking of whom", so each panel carries both and the
// client side keeps server_address to name the upstream.
const (
	OTelGRPCServerRPS = `sum by (rpc_method) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`
	OTelGRPCClientRPS = `sum by (rpc_method, server_address) (rate(rpc_client_call_duration_seconds_count{service_name=~"$app"}[$__rate_interval]))`

	OTelGRPCServerErrorRate = `sum by (rpc_method) (rate(rpc_server_call_duration_seconds_count{service_name=~"$app", rpc_response_status_code!="OK"}[$__rate_interval]))`
	OTelGRPCClientErrorRate = `sum by (rpc_method, server_address) (rate(rpc_client_call_duration_seconds_count{service_name=~"$app", rpc_response_status_code!="OK"}[$__rate_interval]))`

	OTelGRPCServerP95 = `histogram_quantile(0.95, sum by (le, rpc_method) (rate(rpc_server_call_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
	OTelGRPCClientP95 = `histogram_quantile(0.95, sum by (le, rpc_method, server_address) (rate(rpc_client_call_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
)

// Database client (otelpgx) and the pgx connection pool.
//
// OTelDBQueryP95ByService restricts to pgx_operation_type="query" so that the
// per-service number stays a statement-latency number; acquire and connect
// operations have a completely different scale and would swamp it.
//
// The pool trio escalates in sensitivity: in-flight acquisitions show concurrent
// DB work, saturation warns once acquired/max sits near 1 (the
// PgxPoolNearExhaustion condition), and contention is the earliest signal of
// all because an acquire that had to wait already queued behind a busy pool.
// The wait-time counter is in nanoseconds, hence the /1e9 to read as seconds
// of waiting accrued per second.
const (
	OTelDBQueryP95ByService = `histogram_quantile(0.95, sum by (le, service_name) (rate(db_client_operation_duration_seconds_bucket{service_name=~"$app", pgx_operation_type="query"}[$__rate_interval])))`
	OTelDBP95ByOperation    = `histogram_quantile(0.95, sum by (le, pgx_operation_type) (rate(db_client_operation_duration_seconds_bucket{service_name=~"$app"}[$__rate_interval])))`
	OTelDBOperationErrors   = `sum by (service_name) (rate(db_client_operation_errors_total{service_name=~"$app"}[$__rate_interval]))`

	OTelPoolAcquiredConns = `sum by (service_name) (pgxpool_acquired_connections{service_name=~"$app"})`
	OTelPoolSaturation    = `sum by (service_name) (pgxpool_acquired_connections{service_name=~"$app"}) / sum by (service_name) (pgxpool_max_connections{service_name=~"$app"})`

	OTelPoolEmptyAcquireRate     = `sum by (service_name) (rate(pgxpool_empty_acquire_total{service_name=~"$app"}[$__rate_interval]))`
	OTelPoolEmptyAcquireWaitSecs = `sum by (service_name) (rate(pgxpool_empty_acquire_wait_time_nanoseconds_total{service_name=~"$app"}[$__rate_interval])) / 1e9`
)
