// Package observability composes dashboards for the Observability domain.
package observability

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	obsqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/observability"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// TemporalWorker builds the Temporal workflow/activity RED dashboard.
func TemporalWorker() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Temporal — Workflows & Activities").
		Description("Temporal SDK workflow/activity RED metrics from the order-fulfillment worker (pkg/temporalx). See homelab RFC-0001.").
		Editable(true).
		Tags([]string{"temporal", "observability", "workflow", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s"))

	b = b.
		Panel("wf-completion", panels.SeriesExpr("Workflow completion rate", "ops", obsqueries.WorkflowCompletionRate, "{{workflow_type}}")).
		Panel("wf-failure", panels.SeriesExpr("Workflow failure ratio", "percentunit", obsqueries.WorkflowFailureRatio, "failure ratio")).
		Panel("wf-latency", panels.Series("Workflow end-to-end latency", "s",
			panels.PromQuery(obsqueries.WorkflowLatencyP95, "p95"),
			panels.Query("B", obsqueries.WorkflowLatencyP99, "p99"),
		)).
		Panel("act-rate", panels.SeriesExpr("Activity execution rate by type", "ops", obsqueries.ActivityExecutionRate, "{{activity_type}}")).
		Panel("act-failure", panels.SeriesExpr("Activity failure ratio", "percentunit", obsqueries.ActivityFailureRatio, "failure ratio")).
		Panel("act-latency", panels.SeriesExpr("Activity execution latency p95 by type", "s", obsqueries.ActivityLatencyP95, "{{activity_type}}")).
		Panel("worker-slots", panels.Series("Worker task slots", "short",
			panels.PromQuery(obsqueries.WorkerSlotsAvailable, "available {{worker_type}}"),
			panels.Query("B", obsqueries.WorkerSlotsUsed, "used {{worker_type}}"),
		)).
		Panel("worker-errors", panels.SeriesExpr("Worker → server request error ratio", "percentunit", obsqueries.WorkerRequestErrorRatio, "request error ratio"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Workflows",
			panels.GridItem("wf-completion", 0, 0, 12, 8),
			panels.GridItem("wf-failure", 12, 0, 12, 8),
			panels.GridItem("wf-latency", 0, 8, 24, 8),
		),
		panels.Row("Activities",
			panels.GridItem("act-rate", 0, 0, 12, 8),
			panels.GridItem("act-failure", 12, 0, 12, 8),
			panels.GridItem("act-latency", 0, 8, 24, 8),
		),
		panels.Row("Worker / SDK",
			panels.GridItem("worker-slots", 0, 0, 12, 8),
			panels.GridItem("worker-errors", 12, 0, 12, 8),
		),
	))
}
