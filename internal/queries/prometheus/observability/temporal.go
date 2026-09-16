// Package observability holds reusable PromQL for the Observability domain
// dashboards.
package observability

// PromQL for the Temporal worker dashboard. These are Temporal Go SDK RED
// metrics emitted by the order-fulfillment worker. The source dashboard used a
// VictoriaMetrics datasource; the as-code track resolves the logical
// "prometheus" datasource at deploy time.
const (
	// Workflows.
	WorkflowCompletionRate = `sum(rate(temporal_workflow_completed[$__rate_interval])) by (workflow_type)`
	WorkflowFailureRatio   = `sum(rate(temporal_workflow_failed[$__rate_interval])) / clamp_min(sum(rate(temporal_workflow_completed[$__rate_interval])) + sum(rate(temporal_workflow_failed[$__rate_interval])), 1)`
	WorkflowLatencyP95     = `histogram_quantile(0.95, sum(rate(temporal_workflow_endtoend_latency_seconds_bucket[$__rate_interval])) by (le))`
	WorkflowLatencyP99     = `histogram_quantile(0.99, sum(rate(temporal_workflow_endtoend_latency_seconds_bucket[$__rate_interval])) by (le))`

	// Activities.
	ActivityExecutionRate = `sum(rate(temporal_activity_execution_latency_seconds_count[$__rate_interval])) by (activity_type)`
	ActivityFailureRatio  = `sum(rate(temporal_activity_execution_failed[$__rate_interval])) / clamp_min(sum(rate(temporal_activity_execution_latency_seconds_count[$__rate_interval])), 1)`
	ActivityLatencyP95    = `histogram_quantile(0.95, sum(rate(temporal_activity_execution_latency_seconds_bucket[$__rate_interval])) by (le, activity_type))`

	// Worker / SDK.
	WorkerSlotsAvailable    = `sum(temporal_worker_task_slots_available) by (worker_type)`
	WorkerSlotsUsed         = `sum(temporal_worker_task_slots_used) by (worker_type)`
	WorkerRequestErrorRatio = `sum(rate(temporal_request_failure[$__rate_interval])) / clamp_min(sum(rate(temporal_request[$__rate_interval])), 1)`
)
