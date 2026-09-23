package observability

// PromQL for the OTel Collector Health dashboard: self-telemetry scraped from
// :8888 on the single local OpenTelemetry Collector every platform signal
// passes through (job=~".*otel-collector.*"). Ported from the homelab GitOps
// repo, commit 629ce6684d0d9a86d9951f57fb9a313f0e863553:
// kubernetes/infra/configs/observability/grafana/dashboards/otel-collector-health.json
const (
	// Pipeline throughput.
	OtelcolCollectorUp = `up{job=~".*otel-collector.*"}`

	OtelcolReceiverAcceptedSpans        = `sum(rate(otelcol_receiver_accepted_spans{job=~".*otel-collector.*"}[$__rate_interval])) by (receiver)`
	OtelcolReceiverAcceptedMetricPoints = `sum(rate(otelcol_receiver_accepted_metric_points{job=~".*otel-collector.*"}[$__rate_interval])) by (receiver)`
	OtelcolReceiverAcceptedLogRecords   = `sum(rate(otelcol_receiver_accepted_log_records{job=~".*otel-collector.*"}[$__rate_interval])) by (receiver)`

	// Refused = backpressure (memory_limiter); failed = decode errors. The
	// `or vector(0)` keeps the panel at zero instead of "No data" when the
	// series does not exist yet, which is the healthy state.
	OtelcolReceiverRefusedSpans        = `(sum(rate(otelcol_receiver_refused_spans{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`
	OtelcolReceiverRefusedMetricPoints = `(sum(rate(otelcol_receiver_refused_metric_points{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`
	OtelcolReceiverRefusedLogRecords   = `(sum(rate(otelcol_receiver_refused_log_records{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`
	OtelcolReceiverFailedSpans         = `(sum(rate(otelcol_receiver_failed_spans{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`
	OtelcolReceiverFailedMetricPoints  = `(sum(rate(otelcol_receiver_failed_metric_points{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`
	OtelcolReceiverFailedLogRecords    = `(sum(rate(otelcol_receiver_failed_log_records{job=~".*otel-collector.*"}[$__rate_interval])) or vector(0))`

	// Exporters.
	OtelcolExporterSentSpans        = `sum(rate(otelcol_exporter_sent_spans{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter)`
	OtelcolExporterSentMetricPoints = `sum(rate(otelcol_exporter_sent_metric_points{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter)`
	OtelcolExporterSentLogRecords   = `sum(rate(otelcol_exporter_sent_log_records{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter)`

	// No send_failed_* series exists until a backend rejects a batch, hence
	// `or vector(0)` again.
	OtelcolExporterSendFailedSpans        = `(sum(rate(otelcol_exporter_send_failed_spans{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter) or vector(0))`
	OtelcolExporterSendFailedMetricPoints = `(sum(rate(otelcol_exporter_send_failed_metric_points{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter) or vector(0))`
	OtelcolExporterSendFailedLogRecords   = `(sum(rate(otelcol_exporter_send_failed_log_records{job=~".*otel-collector.*"}[$__rate_interval])) by (exporter) or vector(0))`

	OtelcolExporterQueueSize     = `sum(otelcol_exporter_queue_size{job=~".*otel-collector.*"}) by (exporter)`
	OtelcolExporterQueueCapacity = `sum(otelcol_exporter_queue_capacity{job=~".*otel-collector.*"}) by (exporter)`

	OtelcolProcessorBatchSizeP95             = `histogram_quantile(0.95, sum(rate(otelcol_processor_batch_batch_send_size_bucket{job=~".*otel-collector.*"}[$__rate_interval])) by (le))`
	OtelcolProcessorBatchTimeoutTriggerRate  = `sum(rate(otelcol_processor_batch_timeout_trigger_send{job=~".*otel-collector.*"}[$__rate_interval]))`
	OtelcolProcessorBatchMetadataCardinality = `sum(otelcol_processor_batch_metadata_cardinality{job=~".*otel-collector.*"})`

	// Self.
	OtelcolProcessMemoryRSS      = `sum(otelcol_process_memory_rss{job=~".*otel-collector.*"})`
	OtelcolProcessHeapAllocBytes = `sum(otelcol_process_runtime_heap_alloc_bytes{job=~".*otel-collector.*"})`
	OtelcolProcessCPURate        = `sum(rate(otelcol_process_cpu_seconds{job=~".*otel-collector.*"}[$__rate_interval]))`

	// The metrics/apps pipeline converts delta OTLP to cumulative before
	// VictoriaMetrics ingest; hitting the stream limit silently drops new
	// series.
	OtelcolDeltaToCumulativeStreamsTracked = `sum(otelcol_deltatocumulative_streams_tracked{job=~".*otel-collector.*"})`
	OtelcolDeltaToCumulativeStreamsLimit   = `sum(otelcol_deltatocumulative_streams_limit{job=~".*otel-collector.*"})`
)
