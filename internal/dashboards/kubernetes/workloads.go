package kubernetes

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	k8squeries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/kubernetes"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// Workloads renders the Kubernetes Workloads dashboard: per-workload and
// per-pod CPU, memory, network, and reliability, scoped by the
// namespace/workload_type/workload template variables. It replaces the board
// homelab currently fetches from the obs-as-code OCI artifact and keeps the
// same UID and title.
func Workloads() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Kubernetes Workloads").
		Description("Per-workload and per-pod CPU, memory, network, and reliability, filtered by namespace and owning workload.").
		Editable(true).
		Tags([]string{"kubernetes", "workloads", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "1m")).
		QueryVariable(panels.QueryVar("namespace", "Namespace", k8squeries.WorkloadsNamespaceLabelValues)).
		QueryVariable(panels.QueryVar("workload_type", "Type", k8squeries.WorkloadsWorkloadTypeLabelValues)).
		QueryVariable(panels.QueryVar("workload", "Workload", k8squeries.WorkloadsWorkloadLabelValues)).
		Panel("cpu-by-pod", panels.SeriesExpr("CPU by Pod", "short", k8squeries.WorkloadsCPUByPod, "{{pod}}")).
		Panel("cpu-by-workload", panels.SeriesExpr("CPU by Workload", "short", k8squeries.WorkloadsCPUByWorkload, "{{workload_type}}/{{workload}}")).
		Panel("cpu-throttling", panels.SeriesExpr("CPU Throttling by Pod", "percentunit", k8squeries.WorkloadsCPUThrottlingByPod, "{{pod}}")).
		Panel("cpu-vs-requests", panels.SeriesExpr("CPU Usage vs Requests by Pod", "percentunit", k8squeries.WorkloadsCPUUsageVsRequestsByPod, "{{pod}}")).
		Panel("memory-by-pod", panels.SeriesExpr("Memory by Pod", "decbytes", k8squeries.WorkloadsMemoryByPod, "{{pod}}")).
		Panel("memory-by-workload", panels.SeriesExpr("Memory by Workload", "decbytes", k8squeries.WorkloadsMemoryByWorkload, "{{workload_type}}/{{workload}}")).
		Panel("memory-vs-limits", panels.SeriesExpr("Memory Usage vs Limits by Pod", "percentunit", k8squeries.WorkloadsMemoryUsageVsLimitsByPod, "{{pod}}")).
		Panel("network-io", panels.Series("Network I/O by Pod", "Bps",
			panels.PromQuery(k8squeries.WorkloadsNetworkReceiveByPod, "RX {{pod}}"),
			panels.Query("B", k8squeries.WorkloadsNetworkTransmitByPod, "TX {{pod}}"),
		)).
		Panel("restarts-by-pod", panels.SeriesExpr("Restarts by Pod", "short", k8squeries.WorkloadsRestartsByPod, "{{pod}}")).
		Panel("restarts-by-workload", panels.SeriesExpr("Restarts by Workload", "short", k8squeries.WorkloadsRestartsByWorkload, "{{workload_type}}/{{workload}}")).
		Panel("waiting-reasons", panels.SeriesExpr("Containers Waiting by Reason", "short", k8squeries.WorkloadsContainersWaitingByReason, "{{pod}} · {{reason}}"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Workloads in $namespace",
			panels.GridItem("cpu-by-workload", 0, 0, 8, 8),
			panels.GridItem("memory-by-workload", 8, 0, 8, 8),
			panels.GridItem("restarts-by-workload", 16, 0, 8, 8),
		),
		panels.Row("CPU",
			panels.GridItem("cpu-by-pod", 0, 0, 8, 8),
			panels.GridItem("cpu-vs-requests", 8, 0, 8, 8),
			panels.GridItem("cpu-throttling", 16, 0, 8, 8),
		),
		panels.Row("Memory",
			panels.GridItem("memory-by-pod", 0, 0, 12, 8),
			panels.GridItem("memory-vs-limits", 12, 0, 12, 8),
		),
		panels.Row("Network",
			panels.GridItem("network-io", 0, 0, 24, 8),
		),
		panels.Row("Health",
			panels.GridItem("restarts-by-pod", 0, 0, 12, 8),
			panels.GridItem("waiting-reasons", 12, 0, 12, 8),
		),
	))
}
