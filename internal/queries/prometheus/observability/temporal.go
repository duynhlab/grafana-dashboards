// Package observability holds reusable PromQL for the Observability domain
// dashboards.
package observability

// PromQL for the Temporal worker dashboard.
//
// Source: homelab
// kubernetes/infra/configs/observability/grafana/dashboards/temporal.json,
// commit 7a003f8d865dbcb28a65327fa1e73cb2c2edfec1. The board covers both
// halves of Temporal: Go SDK worker metrics (temporal_*) pushed over OTLP by
// the order-fulfillment worker, and server metrics (service_*, persistence_*,
// approximate_backlog_*) scraped from the four per-role ServiceMonitors the
// temporalio chart renders (job=~".*temporal.*"). The as-code track resolves
// the logical "prometheus" datasource at deploy time instead of the source's
// ${DS_PROMETHEUS} template variable.
//
// Counter correction: the SDK counters were previously queried without their
// "_total" suffix (temporal_workflow_completed, temporal_workflow_failed,
// temporal_request_failure, temporal_request), which does not match what the
// Go SDK actually exports. homelab's own alerting rule
// (kubernetes/infra/configs/temporal/prometheusrule.yaml), its runbook
// (docs/observability/runbooks/temporal/TemporalWorkflowFailureRateHigh.md),
// and this source dashboard all use the "_total" form. The shipped board was
// therefore rendering empty for every SDK counter panel; every such metric
// below is corrected to "_total". Histogram *_bucket/*_count suffixes are not
// counters and were left as-is.
const (
	// Template variables.
	NamespaceLabelValues    = `label_values(temporal_worker_task_slots_available, namespace)`
	ServiceNameLabelValues  = `label_values(temporal_worker_task_slots_available{namespace=~"$namespace"}, service_name)`
	WorkflowTypeLabelValues = `label_values(temporal_workflow_completed_total{namespace=~"$namespace"}, workflow_type)`

	// Workflows.
	WorkflowCompletionRate = `sum(rate(temporal_workflow_completed_total{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) by (workflow_type, service_name)`
	WorkflowFailureRatio   = `(sum(rate(temporal_workflow_failed_total{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(temporal_workflow_completed_total{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) + (sum(rate(temporal_workflow_failed_total{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) or vector(0)), 1)`
	WorkflowLatencyP95     = `histogram_quantile(0.95, sum(rate(temporal_workflow_endtoend_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) by (le))`
	WorkflowLatencyP99     = `histogram_quantile(0.99, sum(rate(temporal_workflow_endtoend_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name",workflow_type=~"$workflow_type"}[$__rate_interval])) by (le))`

	WorkflowTaskScheduleToStartLatencyP95 = `histogram_quantile(0.95, sum(rate(temporal_workflow_task_schedule_to_start_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, task_queue))`
	WorkflowTaskScheduleToStartLatencyP99 = `histogram_quantile(0.99, sum(rate(temporal_workflow_task_schedule_to_start_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, task_queue))`
	WorkflowTaskExecutionLatencyP95       = `histogram_quantile(0.95, sum(rate(temporal_workflow_task_execution_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, task_queue))`
	WorkflowTaskPollSucceedRate           = `sum(rate(temporal_workflow_task_queue_poll_succeed_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (task_queue)`
	WorkflowTaskPollEmptyRate             = `sum(rate(temporal_workflow_task_queue_poll_empty_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (task_queue)`

	// Activities.
	ActivityExecutionRate             = `sum(rate(temporal_activity_execution_latency_seconds_count{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (activity_type)`
	ActivityFailureRatio              = `(sum(rate(temporal_activity_execution_failed_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(temporal_activity_execution_latency_seconds_count{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])), 1)`
	ActivityLatencyP95                = `histogram_quantile(0.95, sum(rate(temporal_activity_execution_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, activity_type))`
	ActivityScheduleToStartLatencyP95 = `histogram_quantile(0.95, sum(rate(temporal_activity_schedule_to_start_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, task_queue))`

	// Worker / SDK.
	WorkerSlotsAvailable          = `sum(temporal_worker_task_slots_available{namespace=~"$namespace",service_name=~"$service_name"}) by (service_name, worker_type)`
	WorkerSlotsUsed               = `sum(temporal_worker_task_slots_used{namespace=~"$namespace",service_name=~"$service_name"}) by (service_name, worker_type)`
	WorkerPollers                 = `sum(temporal_num_pollers{namespace=~"$namespace",service_name=~"$service_name"}) by (service_name, poller_type)`
	WorkerRequestErrorRatio       = `(sum(rate(temporal_request_failure_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(temporal_request_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])), 1)`
	WorkerRequestLatencyP95       = `histogram_quantile(0.95, sum(rate(temporal_request_latency_seconds_bucket{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (le, service_name))`
	WorkerStickyCacheSize         = `sum(temporal_sticky_cache_size{namespace=~"$namespace",service_name=~"$service_name"}) by (service_name)`
	WorkerStickyCacheHitRate      = `sum(rate(temporal_sticky_cache_hit_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (service_name)`
	WorkerStickyCacheEvictionRate = `sum(rate(temporal_sticky_cache_forced_eviction_total{namespace=~"$namespace",service_name=~"$service_name"}[$__rate_interval])) by (service_name)`

	// Server. These read server-side metrics selected by job=~".*temporal.*",
	// not by the SDK namespace/service_name/workflow_type variables above,
	// because the server ServiceMonitors do not carry those labels.
	ServerUp                     = `min(up{job=~".*temporal.*"})`
	ServerGRPCRequestRateByRole  = `sum(rate(service_requests{job=~".*temporal.*"}[$__rate_interval])) by (service_name)`
	ServerErrorRatio             = `(sum(rate(service_error_with_type{job=~".*temporal.*"}[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(service_requests{job=~".*temporal.*"}[$__rate_interval])), 1)`
	ServerLatencyP95ByRole       = `histogram_quantile(0.95, sum(rate(service_latency_bucket{job=~".*temporal.*"}[$__rate_interval])) by (le, service_name))`
	ServerPersistenceRequestRate = `sum(rate(persistence_requests{job=~".*temporal.*"}[$__rate_interval]))`
	ServerPersistenceErrorRatio  = `(sum(rate(persistence_error_with_type{job=~".*temporal.*"}[$__rate_interval])) or vector(0)) / clamp_min(sum(rate(persistence_requests{job=~".*temporal.*"}[$__rate_interval])), 1)`
	ServerTaskQueueBacklogCount  = `sum by (taskqueue, task_type, worker_version) (approximate_backlog_count{job=~".*temporal.*"})`
	ServerTaskQueueBacklogAge    = `max by (taskqueue, task_type, worker_version) (approximate_backlog_age_seconds{job=~".*temporal.*"})`
)
