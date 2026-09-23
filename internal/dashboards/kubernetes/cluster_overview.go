package kubernetes

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	k8squeries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/kubernetes"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func ClusterOverview() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Kubernetes Cluster Overview").
		Description("Kubernetes cluster overview using the USE method. Covers node/pod counts, workload health, resource utilization, and persistent volume storage.").
		Editable(true).
		Tags([]string{"kubernetes", "cluster", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-1h", "now", "30s")).
		QueryVariable(panels.QueryVar("namespace", "Namespace", k8squeries.NamespaceValues))

	panelIDs := []struct {
		id, title, expr, legend string
	}{
		{"nodes", "Nodes", k8squeries.NodeCount, "Nodes"},
		{"running-pods", "Running Pods", k8squeries.RunningPods, "Running Pods"},
		{"pending-pods", "Pending Pods", k8squeries.PendingPods, "Pending Pods"},
		{"failed-pods", "Failed Pods", k8squeries.FailedPods, "Failed Pods"},
		{"cpu-req-cap", "CPU Requests vs Capacity", k8squeries.CPURequestsVsCapacity, ""},
		{"mem-req-cap", "Memory Requests vs Capacity", k8squeries.MemoryRequestsCapacity, ""},
		{"deploy-mismatch", "Deployment Mismatches", k8squeries.DeploymentMismatches, ""},
		{"sts-mismatch", "StatefulSet Mismatches", k8squeries.StatefulSetMismatches, ""},
		{"crashloop", "CrashLooping Pods", k8squeries.CrashLoopingPods, ""},
		{"oom", "OOM Events (1h)", k8squeries.OOMEvents1h, ""},
		{"pvc-risk", "PVCs at Risk (>80% used)", k8squeries.PVCsAtRisk, ""},
	}

	for _, p := range panelIDs {
		switch p.id {
		case "cpu-req-cap", "mem-req-cap":
			b = b.Panel(p.id, panels.Gauge(p.title, p.expr, p.legend))
		default:
			b = b.Panel(p.id, panels.Stat(p.title, p.expr, p.legend))
		}
	}

	b = b.
		Panel("pod-restarts", panels.TimeSeriesExpr("Pod Restarts by Namespace", k8squeries.PodRestartsByNamespace, "{{namespace}}")).
		Panel("pod-status", panels.TimeSeriesExpr("Pod Status by Namespace", k8squeries.PodStatusByNamespace, "{{namespace}} {{phase}}")).
		Panel("cpu-usage", panels.TimeSeriesExpr("CPU Usage by Namespace", k8squeries.CPUUsageByNamespace, "{{namespace}}")).
		Panel("mem-usage", panels.TimeSeriesExpr("Memory Usage by Namespace", k8squeries.MemoryUsageByNS, "{{namespace}}")).
		Panel("cpu-throttle", panels.TimeSeriesExpr("CPU Throttled Containers", k8squeries.CPUThrottled, "{{namespace}}/{{pod}}")).
		Panel("net-io", panels.TimeSeries("Network I/O by Namespace",
			panels.PromQuery(k8squeries.NetworkReceiveByNS, "rx {{namespace}}"),
			panels.PromQuery(k8squeries.NetworkTransmitByNS, "tx {{namespace}}"),
		)).
		Panel("pvc-used-ts", panels.TimeSeriesExpr("PVC Usage Over Time", k8squeries.PVCUsedTS, "{{persistentvolumeclaim}}")).
		Panel("pvc-table", panels.TimeSeries("PVC Utilization",
			panels.PromQuery(k8squeries.PVCCapacity, "capacity"),
			panels.PromQuery(k8squeries.PVCUsed, "used"),
		))

	b = clusterOverviewNodeSectionPanels(b)

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Cluster Overview",
			panels.GridItem("nodes", 0, 0, 4, 4),
			panels.GridItem("running-pods", 4, 0, 4, 4),
			panels.GridItem("pending-pods", 8, 0, 4, 4),
			panels.GridItem("failed-pods", 12, 0, 4, 4),
			panels.GridItem("cpu-req-cap", 16, 0, 4, 4),
			panels.GridItem("mem-req-cap", 20, 0, 4, 4),
		),
		panels.Row("Workload Health",
			panels.GridItem("deploy-mismatch", 0, 0, 6, 4),
			panels.GridItem("sts-mismatch", 6, 0, 6, 4),
			panels.GridItem("crashloop", 12, 0, 6, 4),
			panels.GridItem("oom", 18, 0, 6, 4),
			panels.GridItem("pod-restarts", 0, 4, 12, 8),
			panels.GridItem("pod-status", 12, 4, 12, 8),
			panels.GridItem("node-pressure", 0, 12, 24, 8),
		),
		panels.Row("Resource Utilization",
			panels.GridItem("cpu-usage", 0, 0, 12, 8),
			panels.GridItem("mem-usage", 12, 0, 12, 8),
			panels.GridItem("cpu-throttle", 0, 8, 12, 8),
			panels.GridItem("net-io", 12, 8, 12, 8),
			panels.GridItem("cpu-req-alloc-node", 0, 16, 12, 8),
			panels.GridItem("mem-req-alloc-node", 12, 16, 12, 8),
			panels.GridItem("packets-dropped", 0, 24, 12, 8),
			panels.GridItem("pods-per-node", 12, 24, 12, 8),
		),
		panels.Row("Storage",
			panels.GridItem("pvc-table", 0, 0, 16, 8),
			panels.GridItem("pvc-used-ts", 0, 8, 16, 8),
			panels.GridItem("pvc-risk", 16, 0, 8, 4),
		),
	))
}
