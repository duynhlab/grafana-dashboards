// Package observability composes dashboards for the Observability domain.
package observability

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	obsqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/observability"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// TemporalWorker builds the Temporal workflow/activity/server RED dashboard.
// See internal/queries/prometheus/observability/temporal.go for the source
// path, the homelab commit it was ported from, and the SDK counter
// correction applied to the pre-existing queries.
func TemporalWorker() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Temporal — Workflows & Activities").
		Description("Temporal on the cluster, both halves: SDK/worker metrics (temporal_*) pushed over OTLP by the order-fulfillment worker (pkg/temporalx), and server metrics (service_*, persistence_*, approximate_backlog_*) scraped via the per-role ServiceMonitors the temporalio chart renders (job=~\".*temporal.*\"). See homelab RFC-0001.").
		Editable(true).
		Tags([]string{"temporal", "observability", "workflow", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s")).
		QueryVariable(panels.QueryVar("namespace", "Namespace", obsqueries.NamespaceLabelValues)).
		QueryVariable(panels.QueryVar("service_name", "Service", obsqueries.ServiceNameLabelValues)).
		QueryVariable(panels.QueryVar("workflow_type", "Workflow type", obsqueries.WorkflowTypeLabelValues))

	b = temporalWorkflowsPanels(b)
	b = temporalActivitiesPanels(b)
	b = temporalWorkerPanels(b)
	b = temporalServerPanels(b)

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Workflows",
			panels.GridItem("wf-completion", 0, 0, 12, 8),
			panels.GridItem("wf-failure", 12, 0, 12, 8),
			panels.GridItem("wf-latency", 0, 8, 12, 8),
			panels.GridItem("wf-sched-start", 12, 8, 12, 8),
			panels.GridItem("wf-task-exec", 0, 16, 12, 8),
			panels.GridItem("wf-poll", 12, 16, 12, 8),
		),
		panels.Row("Activities",
			panels.GridItem("act-rate", 0, 0, 12, 8),
			panels.GridItem("act-failure", 12, 0, 12, 8),
			panels.GridItem("act-latency", 0, 8, 12, 8),
			panels.GridItem("act-sched-start", 12, 8, 12, 8),
		),
		panels.Row("Worker / SDK",
			panels.GridItem("worker-slots", 0, 0, 12, 8),
			panels.GridItem("worker-pollers", 12, 0, 12, 8),
			panels.GridItem("worker-req-error", 0, 8, 8, 8),
			panels.GridItem("worker-req-latency", 8, 8, 8, 8),
			panels.GridItem("worker-sticky-cache", 16, 8, 8, 8),
		),
		panels.Row("Server",
			panels.GridItem("srv-up", 0, 0, 4, 8),
			panels.GridItem("srv-grpc-rate", 4, 0, 10, 8),
			panels.GridItem("srv-error-ratio", 14, 0, 10, 8),
			panels.GridItem("srv-latency", 0, 8, 8, 8),
			panels.GridItem("srv-persistence", 8, 8, 8, 8),
			panels.GridItem("srv-backlog", 16, 8, 8, 8),
		),
	))
}
