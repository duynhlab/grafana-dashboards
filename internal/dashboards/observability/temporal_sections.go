package observability

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	obsqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/observability"
)

// The Temporal worker dashboard carries 21 panels across four rows; panel
// registration is split by row to keep each function readable. Every function
// registers its panels on the shared builder and returns it for chaining.
// Layout is assembled in temporal.go.

func temporalWorkflowsPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("wf-completion", panels.SeriesExpr("Workflow completion rate", "ops", obsqueries.WorkflowCompletionRate, "{{workflow_type}} ({{service_name}})")).
		Panel("wf-failure", panels.SeriesExpr("Workflow failure ratio", "percentunit", obsqueries.WorkflowFailureRatio, "failure ratio")).
		Panel("wf-latency", panels.Series("Workflow end-to-end latency", "s",
			panels.PromQuery(obsqueries.WorkflowLatencyP95, "p95"),
			panels.Query("B", obsqueries.WorkflowLatencyP99, "p99"),
		)).
		Panel("wf-sched-start", panels.Series("Workflow task schedule-to-start latency", "s",
			panels.PromQuery(obsqueries.WorkflowTaskScheduleToStartLatencyP95, "p95 {{task_queue}}"),
			panels.Query("B", obsqueries.WorkflowTaskScheduleToStartLatencyP99, "p99 {{task_queue}}"),
		)).
		Panel("wf-task-exec", panels.SeriesExpr("Workflow task execution latency (p95)", "s", obsqueries.WorkflowTaskExecutionLatencyP95, "{{task_queue}}")).
		Panel("wf-poll", panels.Series("Workflow task poll outcome rate", "ops",
			panels.PromQuery(obsqueries.WorkflowTaskPollSucceedRate, "succeed {{task_queue}}"),
			panels.Query("B", obsqueries.WorkflowTaskPollEmptyRate, "empty {{task_queue}}"),
		))
}

func temporalActivitiesPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("act-rate", panels.SeriesExpr("Activity execution rate by type", "ops", obsqueries.ActivityExecutionRate, "{{activity_type}}")).
		Panel("act-failure", panels.SeriesExpr("Activity failure ratio", "percentunit", obsqueries.ActivityFailureRatio, "failure ratio")).
		Panel("act-latency", panels.SeriesExpr("Activity execution latency (p95) by type", "s", obsqueries.ActivityLatencyP95, "{{activity_type}}")).
		Panel("act-sched-start", panels.SeriesExpr("Activity schedule-to-start latency (p95)", "s", obsqueries.ActivityScheduleToStartLatencyP95, "{{task_queue}}"))
}

func temporalWorkerPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("worker-slots", panels.Series("Worker task slots", "short",
			panels.PromQuery(obsqueries.WorkerSlotsAvailable, "available {{service_name}}/{{worker_type}}"),
			panels.Query("B", obsqueries.WorkerSlotsUsed, "used {{service_name}}/{{worker_type}}"),
		)).
		Panel("worker-pollers", panels.SeriesExpr("Pollers", "short", obsqueries.WorkerPollers, "{{service_name}}/{{poller_type}}")).
		Panel("worker-req-error", panels.SeriesExpr("Worker → server request error ratio", "percentunit", obsqueries.WorkerRequestErrorRatio, "error ratio")).
		Panel("worker-req-latency", panels.SeriesExpr("Worker → server request latency (p95)", "s", obsqueries.WorkerRequestLatencyP95, "{{service_name}}")).
		Panel("worker-sticky-cache", panels.Series("Sticky cache", "short",
			panels.PromQuery(obsqueries.WorkerStickyCacheSize, "size {{service_name}}"),
			panels.Query("B", obsqueries.WorkerStickyCacheHitRate, "hit/s {{service_name}}"),
			panels.Query("C", obsqueries.WorkerStickyCacheEvictionRate, "evictions/s {{service_name}}"),
		))
}

func temporalServerPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("srv-up", panels.StatValue("Server up", "", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("green", panels.Ptr(1)),
		}, obsqueries.ServerUp, "up")).
		Panel("srv-grpc-rate", panels.SeriesExpr("gRPC request rate by role", "reqps", obsqueries.ServerGRPCRequestRateByRole, "{{service_name}}")).
		Panel("srv-error-ratio", panels.SeriesExpr("Service error ratio", "percentunit", obsqueries.ServerErrorRatio, "error ratio")).
		Panel("srv-latency", panels.SeriesExpr("Service latency (p95) by role", "s", obsqueries.ServerLatencyP95ByRole, "{{service_name}}")).
		Panel("srv-persistence", panels.Series("Persistence request rate / error ratio", "short",
			panels.PromQuery(obsqueries.ServerPersistenceRequestRate, "requests/s"),
			panels.Query("B", obsqueries.ServerPersistenceErrorRatio, "error ratio"),
		)).
		Panel("srv-backlog", panels.Series("Task-queue backlog (server view)", "short",
			panels.PromQuery(obsqueries.ServerTaskQueueBacklogCount, "backlog {{taskqueue}} {{task_type}} {{worker_version}}"),
			panels.Query("B", obsqueries.ServerTaskQueueBacklogAge, "age(s) {{taskqueue}} {{task_type}} {{worker_version}}"),
		))
}
