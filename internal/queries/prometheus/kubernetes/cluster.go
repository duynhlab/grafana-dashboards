package kubernetes

// The per-node request/allocatable pair, the node pressure conditions, the
// packets-dropped-by-namespace pair, the pods-per-node pair, and the
// `namespace` template variable were ported from the retiring homelab
// `kubernetes-cluster-overview` board (a separate delivery this repo's board
// is about to replace):
// /home/duydo/Working/Me/duynhlab/obs-as-code/generated/cluster/dashboards/kubernetes-cluster-overview.json
// at obs-as-code commit b3484a65fd59596e74faded0fbf13b648bc6aabd. PromQL was
// normalised to this repo's conventions: the `job="kube-state-metrics"` /
// `job="kubelet"` scrape-job selectors were dropped (this repo's other
// kube-state-metrics/cAdvisor queries below carry no job filter either), the
// federation-only `exported_namespace`/`exported_pod` join labels were
// rewritten to the plain `namespace`/`pod` labels documented for this metric
// model, and the literal `[$__rate_interval]` window was kept as-is. The
// pre-existing literal `[5m]` windows elsewhere in this file were also
// normalised to `$__rate_interval` while touching this file, per the "rate
// windows never a literal" convention; the `[1h]` restart/OOM windows are a
// different, deliberate semantic (a fixed lookback) and are left alone.
const (
	NodeCount              = `count(kube_node_info)`
	RunningPods            = `count(kube_pod_status_phase{phase="Running"} == 1)`
	PendingPods            = `count(kube_pod_status_phase{phase="Pending"} == 1)`
	FailedPods             = `count(kube_pod_status_phase{phase="Failed"} == 1)`
	CPURequestsVsCapacity  = `sum(kube_pod_container_resource_requests{resource="cpu"}) / sum(kube_node_status_allocatable{resource="cpu"})`
	MemoryRequestsCapacity = `sum(kube_pod_container_resource_requests{resource="memory"}) / sum(kube_node_status_allocatable{resource="memory"})`

	DeploymentMismatches  = `count(kube_deployment_spec_replicas != kube_deployment_status_ready_replicas)`
	StatefulSetMismatches = `count(kube_statefulset_status_replicas_ready != kube_statefulset_status_replicas)`
	CrashLoopingPods      = `count(max_over_time(kube_pod_container_status_waiting_reason{reason="CrashLoopBackOff"}[$__rate_interval]) == 1)`
	OOMEvents1h           = `sum(increase(kube_pod_container_status_last_terminated_reason{reason="OOMKilled"}[1h]))`

	PodRestartsByNamespace = `sum by (namespace) (increase(kube_pod_container_status_restarts_total{namespace=~"$namespace"}[1h]))`
	PodStatusByNamespace   = `count by (namespace, phase) (kube_pod_status_phase{namespace=~"$namespace"} == 1)`

	CPUUsageByNamespace = `sum by (namespace) (rate(container_cpu_usage_seconds_total{container!="", image!="", namespace=~"$namespace"}[$__rate_interval]))`
	MemoryUsageByNS     = `sum by (namespace) (container_memory_working_set_bytes{container!="", image!="", namespace=~"$namespace"})`
	CPUThrottled        = `sum by (namespace, pod) (increase(container_cpu_cfs_throttled_periods_total{container!="", namespace=~"$namespace"}[$__rate_interval])) / sum by (namespace, pod) (increase(container_cpu_cfs_periods_total{container!="", namespace=~"$namespace"}[$__rate_interval])) > 0`
	NetworkReceiveByNS  = `sum by (namespace) (rate(container_network_receive_bytes_total{image!="", namespace=~"$namespace"}[$__rate_interval]))`
	NetworkTransmitByNS = `sum by (namespace) (rate(container_network_transmit_bytes_total{image!="", namespace=~"$namespace"}[$__rate_interval]))`

	PVCCapacity = `kubelet_volume_stats_capacity_bytes`
	PVCUsed     = `kubelet_volume_stats_used_bytes`
	PVCUsedTS   = `kubelet_volume_stats_used_bytes`
	PVCsAtRisk  = `count((1 - kubelet_volume_stats_available_bytes / kubelet_volume_stats_capacity_bytes) > 0.8)`

	// NamespaceValues drives the `namespace` template variable.
	NamespaceValues = `label_values(kube_pod_info, namespace)`

	// CPURequestsVsAllocatableByNode and MemoryRequestsVsAllocatableByNode are
	// the per-node breakdown of CPURequestsVsCapacity / MemoryRequestsCapacity
	// above (which stay cluster-wide). Node-scoped, so left unfiltered by
	// $namespace.
	CPURequestsVsAllocatableByNode    = `sum by (node) (kube_pod_container_resource_requests{resource="cpu"} * on (namespace, pod) group_left () (max by (namespace, pod) (kube_pod_status_phase{phase="Running"}) == 1)) / (sum by (node) (kube_node_status_allocatable{resource="cpu"}) > 0)`
	MemoryRequestsVsAllocatableByNode = `sum by (node) (kube_pod_container_resource_requests{resource="memory"} * on (namespace, pod) group_left () (max by (namespace, pod) (kube_pod_status_phase{phase="Running"}) == 1)) / (sum by (node) (kube_node_status_allocatable{resource="memory"}) > 0)`

	// NodeMemoryPressure, NodeDiskPressure, and NodePIDPressure feed the Node
	// Pressure Conditions panel as three separate targets, one per condition.
	// Node-scoped, so left unfiltered by $namespace.
	NodeMemoryPressure = `sum by (node) (kube_node_status_condition{condition="MemoryPressure", status="true"})`
	NodeDiskPressure   = `sum by (node) (kube_node_status_condition{condition="DiskPressure", status="true"})`
	NodePIDPressure    = `sum by (node) (kube_node_status_condition{condition="PIDPressure", status="true"})`

	// PacketsDroppedReceiveByNamespace and PacketsDroppedTransmitByNamespace
	// feed the Packets Dropped by Namespace panel.
	PacketsDroppedReceiveByNamespace  = `sum by (namespace) (rate(container_network_receive_packets_dropped_total{namespace=~"$namespace", namespace!=""}[$__rate_interval]))`
	PacketsDroppedTransmitByNamespace = `sum by (namespace) (rate(container_network_transmit_packets_dropped_total{namespace=~"$namespace", namespace!=""}[$__rate_interval]))`

	// PodsPerNode and NodePodCapacity feed the Pods per Node panel. Node-scoped,
	// so left unfiltered by $namespace.
	PodsPerNode     = `count by (node) (kube_pod_info)`
	NodePodCapacity = `sum by (node) (kube_node_status_allocatable{resource="pods"})`
)
