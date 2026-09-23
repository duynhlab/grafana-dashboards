package observability

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	obsqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/observability"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// OTelCollectorHealth builds the health board for the single OpenTelemetry
// Collector every local signal passes through. It pairs with the
// OtelCollectorDown and ClickHouseExporterUnhealthy alert rules.
func OTelCollectorHealth() cog.Builder[dashboardv2.Dashboard] {
	upSteps := []panels.Threshold{
		panels.Thr("red", nil),
		panels.Thr("green", panels.Ptr(1)),
	}

	b := dashboardv2.NewDashboardBuilder("OTel Collector Health").
		Description("Health of the OpenTelemetry Collector self-telemetry (job=~\".*otel-collector.*\"): pipeline throughput, exporter fan-out, and collector self resource usage.").
		Editable(true).
		Tags([]string{"otel-collector", "observability", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s"))

	b = b.
		Panel("collector-up", panels.StatValue("Collector up", "", upSteps, obsqueries.OtelcolCollectorUp, "")).
		Panel("receiver-accepted", panels.Series("Receiver accepted rate", "ops",
			panels.PromQuery(obsqueries.OtelcolReceiverAcceptedSpans, "spans {{receiver}}"),
			panels.Query("B", obsqueries.OtelcolReceiverAcceptedMetricPoints, "metric points {{receiver}}"),
			panels.Query("C", obsqueries.OtelcolReceiverAcceptedLogRecords, "log records {{receiver}}"),
		)).
		Panel("receiver-refused", panels.Series("Receiver refused / failed rate", "ops",
			panels.PromQuery(obsqueries.OtelcolReceiverRefusedSpans, "refused spans"),
			panels.Query("B", obsqueries.OtelcolReceiverRefusedMetricPoints, "refused metric points"),
			panels.Query("C", obsqueries.OtelcolReceiverRefusedLogRecords, "refused log records"),
			panels.Query("D", obsqueries.OtelcolReceiverFailedSpans, "failed spans"),
			panels.Query("E", obsqueries.OtelcolReceiverFailedMetricPoints, "failed metric points"),
			panels.Query("F", obsqueries.OtelcolReceiverFailedLogRecords, "failed log records"),
		)).
		Panel("exporter-sent", panels.Series("Exporter sent rate by destination", "ops",
			panels.PromQuery(obsqueries.OtelcolExporterSentSpans, "spans {{exporter}}"),
			panels.Query("B", obsqueries.OtelcolExporterSentMetricPoints, "metric points {{exporter}}"),
			panels.Query("C", obsqueries.OtelcolExporterSentLogRecords, "log records {{exporter}}"),
		)).
		Panel("exporter-failed", panels.Series("Exporter send-failure rate", "ops",
			panels.PromQuery(obsqueries.OtelcolExporterSendFailedSpans, "spans {{exporter}}"),
			panels.Query("B", obsqueries.OtelcolExporterSendFailedMetricPoints, "metric points {{exporter}}"),
			panels.Query("C", obsqueries.OtelcolExporterSendFailedLogRecords, "log records {{exporter}}"),
		)).
		Panel("exporter-queue", panels.Series("Exporter queue utilisation", "",
			panels.PromQuery(obsqueries.OtelcolExporterQueueSize, "size {{exporter}}"),
			panels.Query("B", obsqueries.OtelcolExporterQueueCapacity, "capacity {{exporter}}"),
		)).
		Panel("batch-processor", panels.Series("Batch processor", "",
			panels.PromQuery(obsqueries.OtelcolProcessorBatchSizeP95, "batch size p95"),
			panels.Query("B", obsqueries.OtelcolProcessorBatchTimeoutTriggerRate, "timeout-triggered sends/s"),
			panels.Query("C", obsqueries.OtelcolProcessorBatchMetadataCardinality, "metadata cardinality"),
		)).
		Panel("collector-memory", panels.Series("Collector memory", "bytes",
			panels.PromQuery(obsqueries.OtelcolProcessMemoryRSS, "RSS"),
			panels.Query("B", obsqueries.OtelcolProcessHeapAllocBytes, "heap alloc"),
		)).
		Panel("collector-cpu", panels.SeriesExpr("Collector CPU", "percentunit", obsqueries.OtelcolProcessCPURate, "cpu")).
		Panel("deltatocumulative", panels.Series("deltatocumulative streams", "",
			panels.PromQuery(obsqueries.OtelcolDeltaToCumulativeStreamsTracked, "tracked"),
			panels.Query("B", obsqueries.OtelcolDeltaToCumulativeStreamsLimit, "limit"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Pipeline throughput",
			panels.GridItem("collector-up", 0, 0, 4, 8),
			panels.GridItem("receiver-accepted", 4, 0, 10, 8),
			panels.GridItem("receiver-refused", 14, 0, 10, 8),
		),
		panels.Row("Exporters",
			panels.GridItem("exporter-sent", 0, 0, 12, 8),
			panels.GridItem("exporter-failed", 12, 0, 12, 8),
			panels.GridItem("exporter-queue", 0, 8, 12, 8),
			panels.GridItem("batch-processor", 12, 8, 12, 8),
		),
		panels.Row("Self",
			panels.GridItem("collector-memory", 0, 0, 8, 8),
			panels.GridItem("collector-cpu", 8, 0, 8, 8),
			panels.GridItem("deltatocumulative", 16, 0, 8, 8),
		),
	))
}
