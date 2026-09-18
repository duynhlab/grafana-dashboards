package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// MicroservicesOTel builds the OpenTelemetry microservices board
// (UID microservices-monitoring-001-otel).
//
// It is a curated redesign rather than a schema-parity port: the legacy board's
// per-callee gRPC row only re-grouped the same server series by service_name,
// its "Total Request" and "Success RPS" stats duplicated information already
// carried by the RPS and rate stats, and its second pie repeated the per-endpoint
// breakdown that the timeseries row shows better over time. Those are dropped.
// Panels that existed as separate boxes only because v1 could not merge targets
// (client vs server errors, the three latency percentiles, server vs client gRPC)
// are merged into single multi-target panels here.
func MicroservicesOTel() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Microservices (OTel)").
		Description("RED and resource view of the Go microservices from OpenTelemetry semantic-convention metrics: HTTP server (http_server_request_duration_seconds, request/response body size), gRPC client and server call duration, the OTel Go runtime (goroutines, memory, GC pacing) and the otelpgx DB client with its pgx pool. Ported from the helm-charts grafana-dashboards chart (dashboards/microservices/microservices-dashboard-otel.json).").
		Editable(true).
		Tags([]string{"microservices", "otel", "go", "http", "grpc", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-30m", "now", "1m")).
		QueryVariable(panels.QueryVar("app", "App", msqueries.OTelAppLabelValues))

	b = b.
		// Overview.
		Panel("otel-p99", panels.StatValue("p99 latency (2xx)", "s", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.5)),
			panels.Thr("red", panels.Ptr(1)),
		}, msqueries.OTelP99SuccessLatency, "")).
		Panel("otel-p95", panels.StatValue("p95 latency (2xx)", "s", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.3)),
			panels.Thr("red", panels.Ptr(0.5)),
		}, msqueries.OTelP95SuccessLatency, "")).
		Panel("otel-p50", panels.StatValue("p50 latency (2xx)", "s", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.2)),
			panels.Thr("red", panels.Ptr(0.3)),
		}, msqueries.OTelP50SuccessLatency, "")).
		Panel("otel-total-rps", panels.StatValue("Total RPS", "reqps", []panels.Threshold{
			panels.Thr("blue", nil),
		}, msqueries.OTelTotalRPS, "")).
		Panel("otel-error-rps", panels.StatValue("Error RPS (4xx/5xx)", "reqps", []panels.Threshold{
			panels.Thr("red", nil),
		}, msqueries.OTelErrorRPS, "")).
		Panel("otel-success-rate", panels.StatValue("Success rate % (non-5xx)", "percent", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("yellow", panels.Ptr(95)),
			panels.Thr("green", panels.Ptr(99)),
		}, msqueries.OTelSuccessRatePct, "")).
		Panel("otel-error-rate", panels.StatValue("Error rate % (5xx)", "percent", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(1)),
			panels.Thr("red", panels.Ptr(5)),
		}, msqueries.OTelErrorRatePct, "")).
		Panel("otel-apdex", panels.StatValue("Apdex score", "percentunit", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("yellow", panels.Ptr(0.5)),
			panels.Thr("green", panels.Ptr(0.7)),
		}, msqueries.OTelApdex, "")).

		// Traffic.
		Panel("otel-status-pie", panels.Pie("Status code distribution",
			panels.PromQuery(msqueries.OTelStatusCodeDistribution, "{{http_response_status_code}}"),
		)).
		Panel("otel-rps-route", panels.SeriesExpr("Request rate by endpoint", "reqps",
			msqueries.OTelRequestRateByRoute, "{{http_route}}")).
		Panel("otel-rps-method-route", panels.SeriesExpr("Request rate by method and endpoint", "reqps",
			msqueries.OTelRequestRateByMethodRoute, "{{http_request_method}} {{http_route}}")).

		// Errors and latency.
		Panel("otel-5xx-ratio", panels.SeriesThresholdLine("5xx ratio by method and endpoint", "percent",
			[]panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("red", panels.Ptr(5)),
			},
			panels.PromQuery(msqueries.OTelServerErrorRatioByMethodRoute, "{{http_request_method}} {{http_route}}"),
		)).
		Panel("otel-errors-by-service", panels.Series("Client and server errors by service", "reqps",
			panels.PromQuery(msqueries.OTelClientErrorRPSByService, "4xx {{service_name}}"),
			panels.Query("B", msqueries.OTelServerErrorRPSByService, "5xx {{service_name}}"),
		)).
		Panel("otel-route-latency", panels.Series("Route latency", "s",
			panels.PromQuery(msqueries.OTelRouteLatencyP50, "p50 {{http_route}}"),
			panels.Query("B", msqueries.OTelRouteLatencyP95, "p95 {{http_route}}"),
			panels.Query("C", msqueries.OTelRouteLatencyP99, "p99 {{http_route}}"),
		)).

		// Go runtime.
		Panel("otel-go-memory", panels.SeriesExpr("Memory in use", "bytes",
			msqueries.OTelGoMemoryUsed, "{{service_name}}")).
		Panel("otel-go-alloc", panels.SeriesExpr("Allocation rate", "short",
			msqueries.OTelGoAllocationRate, "{{service_name}}")).
		Panel("otel-go-goroutines", panels.SeriesExpr("Goroutines", "short",
			msqueries.OTelGoroutines, "{{service_name}}")).
		Panel("otel-go-gc", panels.SeriesThresholdLine("GC pacing pressure (used / goal)", "percentunit",
			[]panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("red", panels.Ptr(0.95)),
			},
			panels.PromQuery(msqueries.OTelGoGCPacing, "{{service_name}}"),
		)).
		Panel("otel-http-bytes", panels.Series("HTTP body throughput per service", "Bps",
			panels.PromQuery(msqueries.OTelHTTPResponseBytesRate, "{{service_name}} TX"),
			panels.Query("B", msqueries.OTelHTTPRequestBytesRate, "{{service_name}} RX"),
		)).

		// gRPC east-west.
		Panel("otel-grpc-rps", panels.Series("gRPC RPS", "reqps",
			panels.PromQuery(msqueries.OTelGRPCServerRPS, "server {{rpc_method}}"),
			panels.Query("B", msqueries.OTelGRPCClientRPS, "client {{rpc_method}} → {{server_address}}"),
		)).
		Panel("otel-grpc-errors", panels.Series("gRPC non-OK error rate", "reqps",
			panels.PromQuery(msqueries.OTelGRPCServerErrorRate, "server {{rpc_method}}"),
			panels.Query("B", msqueries.OTelGRPCClientErrorRate, "client {{rpc_method}} → {{server_address}}"),
		)).
		Panel("otel-grpc-p95", panels.Series("gRPC p95 latency", "s",
			panels.PromQuery(msqueries.OTelGRPCServerP95, "server {{rpc_method}}"),
			panels.Query("B", msqueries.OTelGRPCClientP95, "client {{rpc_method}} → {{server_address}}"),
		)).

		// Database (otelpgx).
		Panel("otel-db-p95-service", panels.SeriesExpr("DB query p95 by service", "s",
			msqueries.OTelDBQueryP95ByService, "{{service_name}}")).
		Panel("otel-db-p95-op", panels.SeriesExpr("DB p95 by operation type", "s",
			msqueries.OTelDBP95ByOperation, "{{pgx_operation_type}}")).
		Panel("otel-db-errors", panels.SeriesExpr("DB operation errors", "ops",
			msqueries.OTelDBOperationErrors, "{{service_name}}")).
		Panel("otel-pool-inflight", panels.SeriesExpr("Pool in-flight (acquired conns)", "short",
			msqueries.OTelPoolAcquiredConns, "{{service_name}}")).
		Panel("otel-pool-saturation", panels.SeriesExpr("Pool saturation (acquired / max)", "percentunit",
			msqueries.OTelPoolSaturation, "{{service_name}}")).
		Panel("otel-pool-contention", panels.Series("Pool contention (waiting acquires)", "short",
			panels.PromQuery(msqueries.OTelPoolEmptyAcquireRate, "waits/s {{service_name}}"),
			panels.Query("B", msqueries.OTelPoolEmptyAcquireWaitSecs, "wait-time s/s {{service_name}}"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview",
			panels.GridItem("otel-p99", 0, 0, 3, 4),
			panels.GridItem("otel-p95", 3, 0, 3, 4),
			panels.GridItem("otel-p50", 6, 0, 3, 4),
			panels.GridItem("otel-total-rps", 9, 0, 3, 4),
			panels.GridItem("otel-error-rps", 12, 0, 3, 4),
			panels.GridItem("otel-success-rate", 15, 0, 3, 4),
			panels.GridItem("otel-error-rate", 18, 0, 3, 4),
			panels.GridItem("otel-apdex", 21, 0, 3, 4),
		),
		panels.Row("Traffic",
			panels.GridItem("otel-status-pie", 0, 0, 12, 8),
			panels.GridItem("otel-rps-route", 12, 0, 12, 8),
			panels.GridItem("otel-rps-method-route", 0, 8, 24, 8),
		),
		panels.Row("Errors and latency",
			panels.GridItem("otel-5xx-ratio", 0, 0, 24, 8),
			panels.GridItem("otel-errors-by-service", 0, 8, 12, 8),
			panels.GridItem("otel-route-latency", 12, 8, 12, 8),
		),
		panels.Row("Go runtime",
			panels.GridItem("otel-go-memory", 0, 0, 12, 8),
			panels.GridItem("otel-go-alloc", 12, 0, 12, 8),
			panels.GridItem("otel-go-goroutines", 0, 8, 12, 8),
			panels.GridItem("otel-go-gc", 12, 8, 12, 8),
			panels.GridItem("otel-http-bytes", 0, 16, 24, 8),
		),
		panels.Row("gRPC east-west",
			panels.GridItem("otel-grpc-rps", 0, 0, 12, 8),
			panels.GridItem("otel-grpc-errors", 12, 0, 12, 8),
			panels.GridItem("otel-grpc-p95", 0, 8, 24, 8),
		),
		panels.Row("Database (otelpgx)",
			panels.GridItem("otel-db-p95-service", 0, 0, 12, 8),
			panels.GridItem("otel-db-p95-op", 12, 0, 12, 8),
			panels.GridItem("otel-db-errors", 0, 8, 12, 8),
			panels.GridItem("otel-pool-inflight", 12, 8, 12, 8),
			panels.GridItem("otel-pool-saturation", 0, 16, 12, 8),
			panels.GridItem("otel-pool-contention", 12, 16, 12, 8),
		),
	))
}
