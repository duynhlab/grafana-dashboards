package kubernetes

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	k8squeries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/kubernetes"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// KEDA builds the "KEDA — Worker Autoscaling" board: what the temporal
// scaler computed per worker version, what the HPA it renders actually did,
// the Temporal server backlog and SDK schedule-to-start it is meant to
// drain, and KEDA's own health. See internal/queries/prometheus/kubernetes/keda.go
// for the source and the two selectors that became template variables.
func KEDA() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("KEDA — Worker Autoscaling").
		Description("KEDA as the Temporal workers' autoscaler (ADR-055): what the temporal scaler computed per worker version, what the HPA it renders actually did, the server backlog and SDK schedule-to-start it is meant to drain, and KEDA's own health.").
		Editable(true).
		Tags([]string{"keda", "temporal", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-3h", "now", "30s")).
		QueryVariable(panels.QueryVar("namespace", "Namespace", k8squeries.KedaNamespaceLabelValues)).
		QueryVariable(panels.QueryVar("scaledObject", "Scaled Object", k8squeries.KedaScaledObjectLabelValues)).
		QueryVariable(panels.QueryVar("deployment", "Deployment", k8squeries.KedaDeploymentLabelValues)).

		// Row: Scaler — what KEDA computed
		Panel("scaler-metric-value", panels.Series("Scaler metric value (backlog per worker version)", "short",
			panels.PromQuery(k8squeries.KedaScalerMetricValue, "{{scaledObject}}"),
		)).
		Panel("desired-vs-target", panels.Series("Desired vs target (value ÷ targetQueueSize 5)", "short",
			panels.PromQuery(k8squeries.KedaScalerDesired, "{{scaledObject}} desired"),
			panels.Query("B", k8squeries.KedaHPACurrentReplicas, "{{horizontalpodautoscaler}} current"),
		)).
		Panel("scaler-active", panels.StateTimeline("Scaler active",
			[]panels.Threshold{
				panels.Thr("transparent", nil),
				panels.Thr("green", panels.Ptr(1)),
			},
			panels.PromQuery(k8squeries.KedaScalerActive, "{{scaledObject}}"),
		)).
		Panel("scaler-metric-latency", panels.Series("Scaler metric latency (last fetch)", "s",
			panels.PromQuery(k8squeries.KedaScalerMetricLatency, "{{scaler}}"),
		)).
		Panel("scaler-errors", panels.Series("Scaler errors (rate)", "ops",
			panels.PromQuery(k8squeries.KedaScalerErrorsRate, "{{scaledObject}} / {{scaler}}"),
		)).
		Panel("scaledobject-errors", panels.Series("ScaledObject errors (rate)", "ops",
			panels.PromQuery(k8squeries.KedaScaledObjectErrorsRate, "{{scaledObject}}"),
		)).

		// Row: Scale target — what the HPA did
		Panel("current-vs-max-replicas", panels.Series("Current vs max replicas", "short",
			panels.PromQuery(k8squeries.KedaHPACurrentReplicas, "{{horizontalpodautoscaler}} current"),
			panels.Query("B", k8squeries.KedaHPAMaxReplicas, "{{horizontalpodautoscaler}} max"),
		)).
		Panel("replica-changes", panels.Series("Replica changes", "short",
			panels.PromQuery(k8squeries.KedaHPAReplicaChanges, "{{horizontalpodautoscaler}}"),
		)).

		// Row: Temporal — the signal being scaled on
		Panel("taskqueue-backlog", panels.Series("Task-queue backlog (server; summed over partition only)", "short",
			panels.PromQuery(k8squeries.KedaTaskQueueBacklog, "{{taskqueue}} {{task_type}} {{worker_version}}"),
		)).
		Panel("worker-replicas", panels.Series("Worker replicas (versioned Deployments)", "short",
			panels.PromQuery(k8squeries.KedaWorkerReplicas, "{{namespace}}/{{deployment}}"),
		)).
		Panel("schedule-to-start-p99", panels.Series("Schedule-to-start p99 by task type (SDK)", "s",
			panels.PromQuery(k8squeries.KedaScheduleToStartP99, "{{task_type}} {{task_queue}}"),
		)).

		// Row: KEDA health
		Panel("operator-up", panels.StatValue("Operator up", "", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("green", panels.Ptr(1)),
		}, k8squeries.KedaOperatorUp, "")).
		Panel("metrics-adapter-up", panels.StatValue("Metrics adapter up", "", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("green", panels.Ptr(1)),
		}, k8squeries.KedaMetricsAdapterUp, "")).
		Panel("scaledobjects-registered", panels.StatValue("ScaledObjects registered", "", []panels.Threshold{
			panels.Thr("green", nil),
		}, k8squeries.KedaScaledObjectsRegistered, "")).
		Panel("temporal-triggers-registered", panels.StatValue("temporal triggers registered", "", []panels.Threshold{
			panels.Thr("green", nil),
		}, k8squeries.KedaTemporalTriggersRegistered, "")).
		Panel("keda-version", panels.StatValue("KEDA version", "none", []panels.Threshold{
			panels.Thr("green", nil),
		}, k8squeries.KedaBuildInfo, "")).
		Panel("scale-loop-latency", panels.Series("Scale-loop latency (internal)", "s",
			panels.PromQuery(k8squeries.KedaScaleLoopLatency, "{{scaledObject}}"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Scaler — what KEDA computed",
			panels.GridItem("scaler-metric-value", 0, 0, 12, 8),
			panels.GridItem("desired-vs-target", 12, 0, 12, 8),
			panels.GridItem("scaler-active", 0, 8, 12, 6),
			panels.GridItem("scaler-metric-latency", 12, 8, 12, 6),
			panels.GridItem("scaler-errors", 0, 14, 12, 7),
			panels.GridItem("scaledobject-errors", 12, 14, 12, 7),
		),
		panels.Row("Scale target — what the HPA did",
			panels.GridItem("current-vs-max-replicas", 0, 0, 12, 8),
			panels.GridItem("replica-changes", 12, 0, 12, 8),
		),
		panels.Row("Temporal — the signal being scaled on",
			panels.GridItem("taskqueue-backlog", 0, 0, 8, 8),
			panels.GridItem("worker-replicas", 8, 0, 8, 8),
			panels.GridItem("schedule-to-start-p99", 16, 0, 8, 8),
		),
		panels.Row("KEDA health",
			panels.GridItem("operator-up", 0, 0, 4, 5),
			panels.GridItem("metrics-adapter-up", 4, 0, 4, 5),
			panels.GridItem("scaledobjects-registered", 8, 0, 4, 5),
			panels.GridItem("temporal-triggers-registered", 12, 0, 4, 5),
			panels.GridItem("keda-version", 16, 0, 8, 5),
			panels.GridItem("scale-loop-latency", 0, 5, 24, 6),
		),
	))
}
