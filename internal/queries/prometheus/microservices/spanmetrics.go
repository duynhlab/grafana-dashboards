package microservices

// PromQL for the "Microservices — RED Span Metrics" board (UID red-spanmetrics).
//
// Ported from the homelab GitOps repo,
// kubernetes/infra/configs/observability/grafana/dashboards/red-spanmetrics.json
// at commit 629ce6684d0d9a86d9951f57fb9a313f0e863553.
//
// Metric model: the OTel Collector's spanmetrics connector, the local stand-in
// for Tempo's metrics-generator (ADR-057). It derives RED metrics from traces:
//
//	spanmetrics_calls_total                   counter, labels service_name,
//	                                           span_kind, status_code
//	spanmetrics_duration_milliseconds_bucket  histogram, same labels plus le
//
// Every query filters span_kind="SPAN_KIND_SERVER" so client-side spans (a
// service's outbound calls, double-counted at the callee) never inflate the
// rate or latency numbers. This board is the ADR-057 spanmetrics consumer and
// is distinct from microservices-monitoring-001-otel, which reads
// http_server_*/rpc_server_* directly from the Go SDK's semantic-convention
// instruments rather than from span-derived metrics.
//
// The legacy board's `datasource` template variable is dropped: it existed
// only to pin the homelab cluster's ambiguous choice between two
// prometheus-type datasources, which standards.PrometheusDatasource() already
// resolves for every generated board.
const (
	// SpanMetricsServiceLabelValues drives the $service template variable.
	SpanMetricsServiceLabelValues = `label_values(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER"}, service_name)`

	// SpanMetricsTotalRequestRate is the R in RED, summed across every selected
	// service's SERVER spans.
	SpanMetricsTotalRequestRate = `sum(rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval]))`

	// SpanMetricsOverallErrorRatePct is the E in RED as a percentage. The `or
	// vector(0)` keeps the panel at 0 rather than "No data" when nothing errored
	// in the window.
	SpanMetricsOverallErrorRatePct = `100 * (sum(rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",status_code="STATUS_CODE_ERROR",service_name=~"$service"}[$__rate_interval])) or vector(0)) / sum(rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval]))`

	// SpanMetricsOverallLatencyP95 is the D in RED across every selected
	// service's SERVER spans. It is reused for the p50/p95/p99 comparison panel
	// so the same "overall p95" number never drifts between the two panels.
	SpanMetricsOverallLatencyP95 = `histogram_quantile(0.95, sum by (le) (rate(spanmetrics_duration_milliseconds_bucket{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])))`

	// SpanMetricsServicesReporting counts how many selected services emitted
	// SERVER spans in the window. A quiet service is a finding, not a healthy
	// zero.
	SpanMetricsServicesReporting = `count(count by (service_name) (rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])))`

	// SpanMetricsRequestRateByService is the R in RED, broken out per service.
	SpanMetricsRequestRateByService = `sum by (service_name) (rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval]))`

	// SpanMetricsErrorRateByServicePct is the E in RED per service, as a
	// percentage. The `or (... * 0)` term keeps a service at 0% instead of
	// "No data" when it errored zero times in the window.
	SpanMetricsErrorRateByServicePct = `100 * (sum by (service_name) (rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",status_code="STATUS_CODE_ERROR",service_name=~"$service"}[$__rate_interval])) or (sum by (service_name)(rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])) * 0)) / sum by (service_name) (rate(spanmetrics_calls_total{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval]))`

	// SpanMetricsLatencyP95ByService is the D in RED per service.
	SpanMetricsLatencyP95ByService = `histogram_quantile(0.95, sum by (le, service_name) (rate(spanmetrics_duration_milliseconds_bucket{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])))`

	// SpanMetricsOverallLatencyP50 and SpanMetricsOverallLatencyP99 join
	// SpanMetricsOverallLatencyP95 in the p50/p95/p99 comparison panel.
	SpanMetricsOverallLatencyP50 = `histogram_quantile(0.50, sum by (le) (rate(spanmetrics_duration_milliseconds_bucket{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])))`
	SpanMetricsOverallLatencyP99 = `histogram_quantile(0.99, sum by (le) (rate(spanmetrics_duration_milliseconds_bucket{span_kind="SPAN_KIND_SERVER",service_name=~"$service"}[$__rate_interval])))`
)
