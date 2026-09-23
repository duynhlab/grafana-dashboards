package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// RedSpanMetrics builds the "Microservices — RED Span Metrics" board
// (UID red-spanmetrics), ported from the homelab GitOps repo's
// red-spanmetrics.json. See the query file for the source path and commit.
func RedSpanMetrics() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Microservices — RED Span Metrics").
		Description("RED (rate, errors, duration) per service from the spanmetrics connector in the local OTel Collector — the local stand-in for Tempo's metrics-generator. Server-span request rate, error ratio, and latency quantiles derived from spanmetrics_calls_total / spanmetrics_duration_milliseconds_bucket.").
		Editable(true).
		Tags([]string{"red", "tracing", "spanmetrics", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-30m", "now", "10s")).
		QueryVariable(panels.QueryVar("service", "Service", msqueries.SpanMetricsServiceLabelValues))

	b = b.
		// Overview.
		Panel("spanmetrics-total-rps", panels.StatValue("Total request rate (SERVER)", "reqps", []panels.Threshold{
			panels.Thr("green", nil),
		}, msqueries.SpanMetricsTotalRequestRate, "")).
		Panel("spanmetrics-error-rate", panels.StatValue("Overall error rate", "percent", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(1)),
			panels.Thr("red", panels.Ptr(5)),
		}, msqueries.SpanMetricsOverallErrorRatePct, "")).
		Panel("spanmetrics-p95", panels.StatValue("Overall latency p95", "ms", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(250)),
			panels.Thr("red", panels.Ptr(1000)),
		}, msqueries.SpanMetricsOverallLatencyP95, "")).
		Panel("spanmetrics-services-reporting", panels.StatValue("Services reporting", "short", []panels.Threshold{
			panels.Thr("blue", nil),
		}, msqueries.SpanMetricsServicesReporting, "")).

		// RED by service.
		Panel("spanmetrics-rps-by-service", panels.SeriesExpr("Request rate (req/s) by service", "reqps",
			msqueries.SpanMetricsRequestRateByService, "{{service_name}}")).
		Panel("spanmetrics-error-by-service", panels.SeriesExpr("Error rate (%) by service", "percent",
			msqueries.SpanMetricsErrorRateByServicePct, "{{service_name}}")).
		Panel("spanmetrics-p95-by-service", panels.SeriesExpr("Latency p95 (ms) by service", "ms",
			msqueries.SpanMetricsLatencyP95ByService, "{{service_name}}")).
		Panel("spanmetrics-p50-p95-p99", panels.Series("Latency p50 / p95 / p99 (selected services)", "ms",
			panels.PromQuery(msqueries.SpanMetricsOverallLatencyP50, "p50"),
			panels.Query("B", msqueries.SpanMetricsOverallLatencyP95, "p95"),
			panels.Query("C", msqueries.SpanMetricsOverallLatencyP99, "p99"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview",
			panels.GridItem("spanmetrics-total-rps", 0, 0, 6, 4),
			panels.GridItem("spanmetrics-error-rate", 6, 0, 6, 4),
			panels.GridItem("spanmetrics-p95", 12, 0, 6, 4),
			panels.GridItem("spanmetrics-services-reporting", 18, 0, 6, 4),
		),
		panels.Row("RED by service",
			panels.GridItem("spanmetrics-rps-by-service", 0, 0, 12, 8),
			panels.GridItem("spanmetrics-error-by-service", 12, 0, 12, 8),
			panels.GridItem("spanmetrics-p95-by-service", 0, 8, 12, 8),
			panels.GridItem("spanmetrics-p50-p95-p99", 12, 8, 12, 8),
		),
	))
}
