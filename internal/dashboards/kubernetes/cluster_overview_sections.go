package kubernetes

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	k8squeries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/kubernetes"
)

// clusterOverviewNodeSectionPanels registers the five panels ported from the
// retiring homelab `kubernetes-cluster-overview` board that this board did not
// already cover under a different title (see the doc comment on
// internal/queries/prometheus/kubernetes/cluster.go for the source and
// normalisation notes). Kept in its own file so cluster_overview.go stays
// under the line-count guideline.
func clusterOverviewNodeSectionPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("cpu-req-alloc-node", panels.SeriesExpr("CPU Requests vs Allocatable by Node", "percentunit", k8squeries.CPURequestsVsAllocatableByNode, "{{node}}")).
		Panel("mem-req-alloc-node", panels.SeriesExpr("Memory Requests vs Allocatable by Node", "percentunit", k8squeries.MemoryRequestsVsAllocatableByNode, "{{node}}")).
		Panel("node-pressure", panels.Series("Node Pressure Conditions", "short",
			panels.PromQuery(k8squeries.NodeMemoryPressure, "{{node}} MemoryPressure"),
			panels.Query("B", k8squeries.NodeDiskPressure, "{{node}} DiskPressure"),
			panels.Query("C", k8squeries.NodePIDPressure, "{{node}} PIDPressure"),
		)).
		Panel("packets-dropped", panels.Series("Packets Dropped by Namespace", "short",
			panels.PromQuery(k8squeries.PacketsDroppedReceiveByNamespace, "RX {{namespace}}"),
			panels.Query("B", k8squeries.PacketsDroppedTransmitByNamespace, "TX {{namespace}}"),
		)).
		Panel("pods-per-node", panels.Series("Pods per Node", "short",
			panels.PromQuery(k8squeries.PodsPerNode, "{{node}}"),
			panels.Query("B", k8squeries.NodePodCapacity, "capacity {{node}}"),
		))
}
