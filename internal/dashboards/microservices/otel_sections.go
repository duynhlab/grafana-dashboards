package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
)

// The OTel board's panels, one function per row in source order. Each registers
// its panels on the shared builder and returns it for chaining; layout is
// assembled in otel.go and the panel builders live in otel_panels.go.

func otelOverview(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-p99", otelStat("99th Percentile Response Success",
			"99% of requests complete faster than this. Detects worst-case latency and outliers.",
			"s", []panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("yellow", panels.Ptr(0.5)),
				panels.Thr("red", panels.Ptr(1)),
			}, msqueries.OTelP99SuccessLatency)).
		Panel("otel-p95", otelStat("95th Percentile Response Success",
			"95% of requests complete faster than this. Key metric for user experience.",
			"s", []panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("yellow", panels.Ptr(0.3)),
				panels.Thr("red", panels.Ptr(0.5)),
			}, msqueries.OTelP95SuccessLatency)).
		Panel("otel-p50", otelStat("50th Percentile Response Success",
			"Median response time. Represents typical user experience.",
			"s", []panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("yellow", panels.Ptr(0.2)),
				panels.Thr("red", panels.Ptr(0.3)),
			}, msqueries.OTelP50SuccessLatency)).
		Panel("otel-total-rps", otelStat("Total RPS (All Requests)",
			"Total requests per second including all HTTP status codes (2xx, 4xx, 5xx). Use this to monitor overall traffic volume.",
			"reqps", []panels.Threshold{panels.Thr("blue", nil)}, msqueries.OTelTotalRPS)).
		Panel("otel-success-rps", otelStat("Success RPS (2xx)",
			"Successful requests per second (HTTP 2xx responses). This represents productive traffic.",
			"reqps", []panels.Threshold{panels.Thr("green", nil)}, msqueries.OTelSuccessRPS)).
		Panel("otel-error-rps", otelStat("Error RPS (4xx/5xx)",
			"Failed requests per second (HTTP 4xx/5xx responses). Monitor this to detect issues quickly.",
			"reqps", []panels.Threshold{panels.Thr("red", nil)}, msqueries.OTelErrorRPS)).
		Panel("otel-success-rate", otelStat("Success Rate % (non-5xx)",
			"Complement of Error Rate % (5xx): every request the service handled correctly, including 3xx/4xx responses to client mistakes. Success + Error = 100%.",
			"percent", []panels.Threshold{
				panels.Thr("red", panels.Ptr(0)),
				panels.Thr("yellow", panels.Ptr(95)),
				panels.Thr("green", panels.Ptr(99)),
			}, msqueries.OTelSuccessRatePct)).
		Panel("otel-error-rate", otelStat("Error Rate % (5xx)",
			"Server-fault error ratio — matches the MicroserviceHighErrorRate alert and the availability SLO (both 5xx-only). Client 4xx are the service answering correctly; they stay visible in Error RPS (4xx/5xx), the status pie, and Client Errors (4xx).",
			"percent", []panels.Threshold{
				panels.Thr("green", panels.Ptr(0)),
				panels.Thr("yellow", panels.Ptr(1)),
				panels.Thr("red", panels.Ptr(5)),
			}, msqueries.OTelErrorRatePct)).
		Panel("otel-apdex", otelStat("Apdex Score",
			"User satisfaction score (0-1). Satisfying: <0.5s, Tolerating: 0.5-2s, Frustrated: >2s. Formula: (satisfying + 0.5 * tolerating) / total requests. Handles zero traffic gracefully.",
			"percentunit", []panels.Threshold{
				panels.Thr("red", nil),
				panels.Thr("yellow", panels.Ptr(0.5)),
				panels.Thr("green", panels.Ptr(0.7)),
			}, msqueries.OTelApdex)).
		Panel("otel-total-requests", otelStat("Total Request",
			"Total requests in selected time range. Correlate with traffic spikes or incidents.",
			"short", []panels.Threshold{panels.Thr("blue", nil)}, msqueries.OTelTotalRequests))
}

func otelTraffic(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-status-pie", otelPie("Status Code Distribution",
			"HTTP status code distribution by rate (req/sec). Shows real-time traffic breakdown by response code. Expected: ~95% codes 2xx.",
			msqueries.OTelStatusCodeDistribution, "REST.{{http_response_status_code}}")).
		Panel("otel-route-pie", otelPie("Total Requests by Endpoint",
			"Request distribution across endpoints. Identifies hot paths and traffic patterns.",
			msqueries.OTelTotalRequestsByRoute, "{{http_route}}")).
		Panel("otel-rps-route", otelSeries("Request Rate by Endpoint",
			"Request rate per endpoint over time. Detects traffic spikes per API.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelRequestRateByRoute, "{{http_route}}")))
}

func otelErrors(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-rps-method-route", otelSeries("Request Rate by Method and Endpoint",
			"Breakdown of request rate by HTTP method (GET, POST, PUT, DELETE) and endpoint path. Helps identify traffic patterns and detect unusual method-specific spikes.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelRequestRateByMethodRoute, "{{http_request_method}} {{http_route}}"))).
		Panel("otel-5xx-ratio", otelSeries("Server Error Rate by Method and Endpoint (5xx)",
			"5xx ratio per method+route — continuous series (0 baseline) so legend means are honest. Client 4xx per endpoint live in the Client Errors (4xx) panel.",
			"percent", otelSeriesOptions{
				LineWidth: 2,
				Steps: []panels.Threshold{
					panels.Thr("green", nil),
					panels.Thr("yellow", panels.Ptr(1)),
					panels.Thr("red", panels.Ptr(5)),
				},
				ThresholdLine: true,
			},
			panels.PromQuery(msqueries.OTelServerErrorRatioByMethodRoute, "{{http_request_method}} {{http_route}}"))).
		Panel("otel-4xx-service", otelSeries("Client Errors (4xx)",
			"Client-side errors (400-499) in req/sec by service. High 4xx rates indicate API misuse, invalid requests, authentication issues, or bad client behavior. Common codes: 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 429 (Rate Limited).",
			"reqps", otelSeriesOptions{
				LineWidth: 2,
				Calcs:     []string{"mean", "last", "max"},
				Steps: []panels.Threshold{
					panels.Thr("green", nil),
					panels.Thr("yellow", panels.Ptr(1)),
					panels.Thr("orange", panels.Ptr(5)),
				},
				ColorByThresholds: true,
			},
			panels.PromQuery(msqueries.OTelClientErrorRPSByService, "{{service_name}}"))).
		Panel("otel-5xx-service", otelSeries("Server Errors (5xx)",
			"Server-side errors (500-599) in req/sec by service. High 5xx rates indicate service degradation, bugs, infrastructure issues, or dependency failures. Common codes: 500 (Internal Server Error), 502 (Bad Gateway), 503 (Service Unavailable), 504 (Gateway Timeout). Requires immediate investigation.",
			"reqps", otelSeriesOptions{
				LineWidth: 2,
				Calcs:     []string{"mean", "last", "max"},
				Steps: []panels.Threshold{
					panels.Thr("green", nil),
					panels.Thr("orange", panels.Ptr(0.5)),
					panels.Thr("red", panels.Ptr(2)),
				},
				ColorByThresholds: true,
			},
			panels.PromQuery(msqueries.OTelServerErrorRPSByService, "{{service_name}}"))).
		Panel("otel-route-p95", otelSeries("Response time 95th percentile",
			"p95 response time per endpoint. Identifies slow APIs.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelRouteLatencyP95, "{{http_response_status_code}} {{http_route}}"))).
		Panel("otel-route-p50", otelSeries("Response time 50th percentile",
			"Median response time per endpoint. Shows typical performance baseline.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelRouteLatencyP50, "{{http_response_status_code}} {{http_route}}"))).
		Panel("otel-route-p99", otelSeries("Response time 99th percentile",
			"p99 response time per endpoint. Highlights tail latency issues.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelRouteLatencyP99, "{{http_response_status_code}} {{http_route}}")))
}

func otelRuntime(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-go-memory", otelSeries("Memory In-Use",
			"OTel Go runtime: bytes currently in use (go.memory.used).",
			"bytes", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGoMemoryUsed, "{{service_name}}"))).
		Panel("otel-go-alloc", otelSeries("Memory Allocation Rate",
			"Heap allocation rate (objects/s) from the OTel Go runtime. The OTel runtime emits no GC-pause metric; rising allocation rate is the leading indicator of GC pressure.",
			"short", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGoAllocationRate, "{{service_name}}"))).
		Panel("otel-go-goroutines", otelSeries("Goroutines",
			"Goroutine and OS thread count. Steadily increasing goroutines indicates goroutine leak (forgotten defer, unclosed channels). Stable count is normal.",
			"short", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGoroutines, "{{service_name}}"))).
		Panel("otel-go-gc", otelSeries("GC Pacing Pressure (used / goal)",
			"Heap used vs the GC pacer goal from the OTel Go runtime. Sustained values near 1.0 mean the heap keeps hitting the GC goal — memory pressure / frequent GC cycles.",
			"percentunit", otelSeriesOptions{
				Steps: []panels.Threshold{
					panels.Thr("green", nil),
					panels.Thr("red", panels.Ptr(0.95)),
				},
				ThresholdLine: true,
			},
			panels.PromQuery(msqueries.OTelGoGCPacing, "{{service_name}}"))).
		Panel("otel-http-bytes", otelSeries("Total Network Traffic per Service",
			"Total HTTP traffic (TX/RX) by ALL pods in the service. Use for bandwidth planning and cost estimation. Note: Only HTTP body size, not TCP/IP overhead.",
			"Bps", otelSeriesOptions{Calcs: []string{"mean"}, SortBy: "Mean"},
			panels.PromQuery(msqueries.OTelHTTPResponseBytesRate, "{{service_name}} TX"),
			panels.Query("B", msqueries.OTelHTTPRequestBytesRate, "{{service_name}} RX")))
}

func otelGRPC(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-grpc-server-rps", otelSeries("gRPC Server RPS by Method",
			"Server-side gRPC request rate per method (rpc_method, e.g. auth.v1.AuthService/GetMe). Source: rpc_server_call_duration_seconds_count.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerRPS, "{{rpc_method}}"))).
		Panel("otel-grpc-client-rps", otelSeries("gRPC Client RPS by Method",
			"Client-side gRPC request rate per method and upstream (server_address). Source: rpc_client_call_duration_seconds_count.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCClientRPS, "{{rpc_method}} → {{server_address}}"))).
		Panel("otel-grpc-server-errors", otelSeries("gRPC Server Error Rate (non-OK)",
			"Server-side gRPC error rate per method, counting calls where rpc_response_status_code != OK. Source: rpc_server_call_duration_seconds_count.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerErrorRate, "{{rpc_method}}"))).
		Panel("otel-grpc-client-errors", otelSeries("gRPC Client Error Rate (non-OK)",
			"Client-side gRPC error rate per method and upstream, counting calls where rpc_response_status_code != OK. Source: rpc_client_call_duration_seconds_count.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCClientErrorRate, "{{rpc_method}} → {{server_address}}"))).
		Panel("otel-grpc-server-p95", otelSeries("gRPC Server P95 Latency",
			"Server-side gRPC 95th percentile call duration per method. Source: rpc_server_call_duration_seconds_bucket.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerP95, "{{rpc_method}}"))).
		Panel("otel-grpc-client-p95", otelSeries("gRPC Client P95 Latency",
			"Client-side gRPC 95th percentile call duration per method and upstream. Source: rpc_client_call_duration_seconds_bucket.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCClientP95, "{{rpc_method}} → {{server_address}}")))
}

func otelGRPCPerCallee(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-grpc-callee-rps", otelSeries("gRPC Server RPS per Callee",
			"Server-side gRPC request rate aggregated per callee service (app). Source: rpc_server_call_duration_seconds_count.",
			"reqps", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerRPSPerCallee, "{{service_name}}"))).
		Panel("otel-grpc-callee-error-ratio", otelSeries("gRPC Server Error Ratio per Callee",
			"Share of server-side gRPC calls with rpc_response_status_code != OK, per callee service (app). Source: rpc_server_call_duration_seconds_count.",
			"percentunit", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerErrorRatioPerCallee, "{{service_name}}"))).
		Panel("otel-grpc-callee-p95", otelSeries("gRPC Server P95 Latency per Callee",
			"Server-side gRPC 95th percentile call duration per callee service (app). Source: rpc_server_call_duration_seconds_bucket.",
			"s", otelSeriesOptions{},
			panels.PromQuery(msqueries.OTelGRPCServerP95PerCallee, "{{service_name}}")))
}

func otelDatabase(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("otel-db-p95-service", otelPanel("DB query p95 by service",
			"otelpgx per-statement latency (DB-scale buckets, pkg v0.24.0). Query ops only.",
			otelDBViz("s"),
			panels.PromQuery(msqueries.OTelDBQueryP95ByService, "{{service_name}}"))).
		Panel("otel-db-p95-op", otelPanel("DB p95 by operation type",
			"query / batch / prepare / connect / acquire — split by op.",
			otelDBViz("s"),
			panels.PromQuery(msqueries.OTelDBP95ByOperation, "{{pgx_operation_type}}"))).
		Panel("otel-db-errors", otelPanel("DB operation errors",
			"otelpgx non-ErrNoRows operation errors. Healthy = no series/0.",
			otelDBViz("ops"),
			panels.PromQuery(msqueries.OTelDBOperationErrors, "{{service_name}}"))).
		Panel("otel-pool-inflight", otelPanel("Pool in-flight (acquired conns)",
			"Connections currently checked out ≈ concurrent DB work in flight.",
			otelDBViz("short"),
			panels.PromQuery(msqueries.OTelPoolAcquiredConns, "{{service_name}}"))).
		// A ratio, so the axis is pinned to [0, 1] and 100% always sits at the top.
		Panel("otel-pool-saturation", otelPanel("Pool saturation (acquired / max)",
			"Sustained ≥80% = acquires start queueing (PgxPoolNearExhaustion).",
			otelDBViz("percentunit").Min(0).Max(1),
			panels.PromQuery(msqueries.OTelPoolSaturation, "{{service_name}}"))).
		// Waits per second and seconds waited per second differ in scale, so the
		// wait time (target B) gets its own axis on the right.
		Panel("otel-pool-contention", otelPanel("Pool contention (waiting acquires)",
			"Acquires that had to wait for a free conn — earliest saturation signal.",
			otelDBViz("short").OverrideByQuery("B", []dashboardv2.DynamicConfigValue{
				{Id: "custom.axisPlacement", Value: "right"},
				{Id: "custom.axisLabel", Value: "wait-time s/s"},
			}),
			panels.PromQuery(msqueries.OTelPoolEmptyAcquireRate, "waits/s {{service_name}}"),
			panels.Query("B", msqueries.OTelPoolEmptyAcquireWaitSecs, "wait-time s/s {{service_name}}")))
}
