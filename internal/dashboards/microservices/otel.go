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
// It is a 1:1 port of the helm-charts JSON: all 40 panels in the source order,
// the same seven rows and the same panel sizes, titles, descriptions, units,
// thresholds and legends. Only the plumbing differs: the logical prometheus
// datasource replaces the DS_PROMETHEUS picker, $__rate_interval replaces the
// custom $rate window, and the unused namespace variable is dropped, so app is
// the single template variable. Row titles drop the source's leading emoji.
func MicroservicesOTel() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Microservices (OTel)").
		Description("RED and resource view of the Go microservices from OpenTelemetry semantic-convention metrics: HTTP server (http_server_request_duration_seconds, request/response body size), gRPC client and server call duration, the OTel Go runtime (goroutines, memory, GC pacing) and the otelpgx DB client with its pgx pool. Ported from the helm-charts grafana-dashboards chart (dashboards/microservices/microservices-dashboard-otel.json).").
		Editable(true).
		Tags([]string{"prometheus", "kubernetes", "microservices", "go", "observability", "rfc-0014", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-30m", "now", "")).
		QueryVariable(panels.QueryVar("app", "App", msqueries.OTelAppLabelValues))

	b = otelOverview(b)
	b = otelTraffic(b)
	b = otelErrors(b)
	b = otelRuntime(b)
	b = otelGRPC(b)
	b = otelGRPCPerCallee(b)
	b = otelDatabase(b)

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview & Key Metrics",
			panels.GridItem("otel-p99", 0, 0, 5, 4),
			panels.GridItem("otel-p95", 5, 0, 5, 4),
			panels.GridItem("otel-p50", 10, 0, 5, 4),
			panels.GridItem("otel-total-rps", 15, 0, 5, 4),
			panels.GridItem("otel-success-rps", 20, 0, 4, 4),
			panels.GridItem("otel-error-rps", 0, 4, 5, 4),
			panels.GridItem("otel-success-rate", 5, 4, 5, 4),
			panels.GridItem("otel-error-rate", 10, 4, 5, 4),
			panels.GridItem("otel-apdex", 15, 4, 5, 4),
			panels.GridItem("otel-total-requests", 20, 4, 4, 4),
		),
		panels.Row("Traffic & Requests",
			panels.GridItem("otel-status-pie", 0, 0, 12, 9),
			panels.GridItem("otel-route-pie", 12, 0, 12, 9),
			panels.GridItem("otel-rps-route", 0, 9, 24, 8),
		),
		panels.Row("Errors & Performance",
			panels.GridItem("otel-rps-method-route", 0, 0, 24, 8),
			panels.GridItem("otel-5xx-ratio", 0, 8, 24, 8),
			panels.GridItem("otel-4xx-service", 0, 16, 12, 8),
			panels.GridItem("otel-5xx-service", 12, 16, 12, 8),
			panels.GridItem("otel-route-p95", 0, 24, 12, 8),
			panels.GridItem("otel-route-p50", 12, 24, 12, 8),
			panels.GridItem("otel-route-p99", 0, 32, 24, 8),
		),
		panels.Row("Go Runtime & HTTP I/O (OTel)",
			panels.GridItem("otel-go-memory", 0, 0, 12, 8),
			panels.GridItem("otel-go-alloc", 12, 0, 12, 8),
			panels.GridItem("otel-go-goroutines", 0, 8, 12, 8),
			panels.GridItem("otel-go-gc", 12, 8, 12, 8),
			panels.GridItem("otel-http-bytes", 0, 16, 24, 8),
		),
		panels.Row("gRPC East-West (RED)",
			panels.GridItem("otel-grpc-server-rps", 0, 0, 12, 8),
			panels.GridItem("otel-grpc-client-rps", 12, 0, 12, 8),
			panels.GridItem("otel-grpc-server-errors", 0, 8, 12, 8),
			panels.GridItem("otel-grpc-client-errors", 12, 8, 12, 8),
			panels.GridItem("otel-grpc-server-p95", 0, 16, 12, 8),
			panels.GridItem("otel-grpc-client-p95", 12, 16, 12, 8),
		),
		panels.Row("gRPC East-West (RED) — Per Callee",
			panels.GridItem("otel-grpc-callee-rps", 0, 0, 12, 8),
			panels.GridItem("otel-grpc-callee-error-ratio", 12, 0, 12, 8),
			panels.GridItem("otel-grpc-callee-p95", 0, 8, 24, 8),
		),
		panels.Row("Database (client — otelpgx)",
			panels.GridItem("otel-db-p95-service", 0, 0, 12, 8),
			panels.GridItem("otel-db-p95-op", 12, 0, 12, 8),
			panels.GridItem("otel-db-errors", 0, 8, 12, 8),
			panels.GridItem("otel-pool-inflight", 12, 8, 12, 8),
			panels.GridItem("otel-pool-saturation", 0, 16, 12, 8),
			panels.GridItem("otel-pool-contention", 12, 16, 12, 8),
		),
	))
}
