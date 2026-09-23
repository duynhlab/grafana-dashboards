package kubernetes

// Ported from the homelab GitOps repo, read-only source:
// kubernetes/infra/configs/observability/grafana/dashboards/keda.json
// at commit 7a003f8d865dbcb28a65327fa1e73cb2c2edfec1.
//
// KEDA as the Temporal workers' autoscaler (ADR-055): what the temporal
// scaler computed per worker version, what the HPA it renders actually did,
// the server backlog and SDK schedule-to-start it is meant to drain, and
// KEDA's own health.
//
// Two selectors in the source "Worker replicas (versioned Deployments)"
// panel hardcoded cluster-specific values:
//
//	namespace=~"order|checkout"
//	deployment=~"order-fulfillment.*|checkout-abandon.*"
//
// This repository forbids environment-specific label values in committed
// PromQL (see the skill's architecture.md, Environment section), so both
// became template variables: the namespace selector reuses the board's
// existing $namespace variable, and a new $deployment variable
// (KedaDeploymentLabelValues) drives the workload selector instead.
const (
	// KedaNamespaceLabelValues drives the $namespace variable. KEDA stamps
	// its own `namespace` label on scraped metrics (the ScaledObject's
	// namespace); under a Prometheus-operator scrape the target's namespace
	// wins that key, so KEDA's copy is exposed as `exported_namespace`. That
	// is the label the official KEDA dashboard uses and the one this board
	// follows for every keda_* query.
	KedaNamespaceLabelValues = `label_values(keda_scaler_active, exported_namespace)`

	// KedaScaledObjectLabelValues drives the $scaledObject variable, chained
	// off $namespace. One value per running worker version: the controller's
	// hashed per-build-id ScaledObject name.
	KedaScaledObjectLabelValues = `label_values(keda_scaler_active{exported_namespace=~"$namespace"}, scaledObject)`

	// KedaDeploymentLabelValues drives the $deployment variable used by
	// KedaWorkerReplicas. It replaces the source's hardcoded
	// `deployment=~"order-fulfillment.*|checkout-abandon.*"` selector.
	KedaDeploymentLabelValues = `label_values(kube_deployment_status_replicas{namespace=~"$namespace"}, deployment)`
)

const (
	// KedaScalerMetricValue is keda_scaler_metrics_value — the number the
	// temporal scaler handed the HPA: this version's ApproximateBacklogCount
	// from DescribeTaskQueue(stats).
	KedaScalerMetricValue = `sum by (scaledObject) (keda_scaler_metrics_value{exported_namespace=~"$namespace", scaledObject=~"$scaledObject"})`

	// KedaScalerDesired is the scaler's raw desired count (value / targetQueueSize 5).
	KedaScalerDesired = `sum by (scaledObject) (keda_scaler_metrics_value{exported_namespace=~"$namespace", scaledObject=~"$scaledObject"}) / 5`

	// KedaHPACurrentReplicas is the kube-state-metrics view of what the HPA
	// KEDA renders behind every ScaledObject actually set. Uses the plain
	// `namespace` label — kube-state-metrics series carry the object's
	// namespace natively.
	KedaHPACurrentReplicas = `sum by (horizontalpodautoscaler) (kube_horizontalpodautoscaler_status_current_replicas{namespace=~"$namespace", horizontalpodautoscaler=~"keda-hpa-.*"})`

	// KedaHPAMaxReplicas is the configured ceiling for the same HPA.
	KedaHPAMaxReplicas = `sum by (horizontalpodautoscaler) (kube_horizontalpodautoscaler_spec_max_replicas{namespace=~"$namespace", horizontalpodautoscaler=~"keda-hpa-.*"})`

	// KedaScalerActive is keda_scaler_active — 1 while the backlog is above
	// activationTargetQueueSize.
	KedaScalerActive = `max by (scaledObject) (keda_scaler_active{exported_namespace=~"$namespace", scaledObject=~"$scaledObject"})`

	// KedaScalerMetricLatency is keda_scaler_metrics_latency_seconds — how
	// long one DescribeTaskQueue round trip to temporal-frontend takes.
	KedaScalerMetricLatency = `max by (scaler, scaledObject) (keda_scaler_metrics_latency_seconds{exported_namespace=~"$namespace"})`

	// KedaScalerErrorsRate is keda_scaler_detail_errors_total (the 2.20
	// name; keda_scaler_errors_total does not exist).
	KedaScalerErrorsRate = `sum by (scaledObject, scaler) (rate(keda_scaler_detail_errors_total{exported_namespace=~"$namespace", scaledObject=~"$scaledObject"}[$__rate_interval]))`

	// KedaScaledObjectErrorsRate is keda_scaled_object_errors_total —
	// reconcile errors on the ScaledObject itself.
	KedaScaledObjectErrorsRate = `sum by (scaledObject) (rate(keda_scaled_object_errors_total{exported_namespace=~"$namespace", scaledObject=~"$scaledObject"}[$__rate_interval]))`

	// KedaHPAReplicaChanges is the delta of current replicas, positive for
	// scale-outs and negative for scale-ins.
	KedaHPAReplicaChanges = `delta(kube_horizontalpodautoscaler_status_current_replicas{namespace=~"$namespace", horizontalpodautoscaler=~"keda-hpa-.*"}[$__rate_interval])`

	// KedaTaskQueueBacklog is approximate_backlog_count, a SERVER metric
	// from the matching service, summed over partitions only (keep
	// task_type and worker_version apart).
	KedaTaskQueueBacklog = `sum by (taskqueue, task_type, worker_version) (approximate_backlog_count{job=~".*temporal.*"})`

	// KedaWorkerReplicas is one Deployment per worker version (Temporal
	// Worker Controller). Both selectors are template variables: $namespace
	// (shared with the rest of the board) and $deployment (new, see
	// KedaDeploymentLabelValues) replacing the source's hardcoded
	// `namespace=~"order|checkout"` / `deployment=~"order-fulfillment.*|checkout-abandon.*"`.
	KedaWorkerReplicas = `sum by (namespace, deployment) (kube_deployment_status_replicas{namespace=~"$namespace", deployment=~"$deployment"})`

	// KedaScheduleToStartP99 is the SDK's leading indicator for both task
	// types: time a workflow task or an activity waited for a poller.
	KedaScheduleToStartP99 = `label_replace(histogram_quantile(0.99, sum by (le, task_queue) (rate(temporal_workflow_task_schedule_to_start_latency_seconds_bucket[$__rate_interval]))), "task_type", "Workflow", "", "") or label_replace(histogram_quantile(0.99, sum by (le, task_queue) (rate(temporal_activity_schedule_to_start_latency_seconds_bucket[$__rate_interval]))), "task_type", "Activity", "", "")`

	// KedaOperatorUp is min() over the keda-operator scrape targets.
	KedaOperatorUp = `min(up{job=~".*keda-operator.*"})`

	// KedaMetricsAdapterUp covers the external.metrics.k8s.io adapter the
	// HPAs read from.
	KedaMetricsAdapterUp = `min(up{job=~".*keda-metrics-apiserver.*|.*keda-operator-metrics-apiserver.*"})`

	// KedaScaledObjectsRegistered is keda_resource_registered_total{type="scaled_object"}.
	KedaScaledObjectsRegistered = `sum(keda_resource_registered_total{type="scaled_object"})`

	// KedaTemporalTriggersRegistered is keda_trigger_registered_total{type="temporal"}.
	KedaTemporalTriggersRegistered = `sum(keda_trigger_registered_total{type="temporal"})`

	// KedaBuildInfo is keda_build_info — the running operator's version label.
	KedaBuildInfo = `max(keda_build_info) by (version)`

	// KedaScaleLoopLatency is keda_internal_scale_loop_latency_seconds — how
	// late each polling loop ran versus its 15 s schedule.
	KedaScaleLoopLatency = `max by (scaledObject) (keda_internal_scale_loop_latency_seconds{exported_namespace=~"$namespace"})`
)
