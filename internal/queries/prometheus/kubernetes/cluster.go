package kubernetes

const (
	NodeCount              = `count(kube_node_info)`
	RunningPods            = `count(kube_pod_status_phase{phase="Running"} == 1)`
	PendingPods            = `count(kube_pod_status_phase{phase="Pending"} == 1)`
	FailedPods             = `count(kube_pod_status_phase{phase="Failed"} == 1)`
	CPURequestsVsCapacity  = `sum(kube_pod_container_resource_requests{resource="cpu"}) / sum(kube_node_status_allocatable{resource="cpu"})`
	MemoryRequestsCapacity = `sum(kube_pod_container_resource_requests{resource="memory"}) / sum(kube_node_status_allocatable{resource="memory"})`

	DeploymentMismatches  = `count(kube_deployment_spec_replicas != kube_deployment_status_ready_replicas)`
	StatefulSetMismatches = `count(kube_statefulset_status_replicas_ready != kube_statefulset_status_replicas)`
	CrashLoopingPods      = `count(max_over_time(kube_pod_container_status_waiting_reason{reason="CrashLoopBackOff"}[5m]) == 1)`
	OOMEvents1h           = `sum(increase(kube_pod_container_status_last_terminated_reason{reason="OOMKilled"}[1h]))`

	PodRestartsByNamespace = `sum by (namespace) (increase(kube_pod_container_status_restarts_total[1h]))`
	PodStatusByNamespace   = `count by (namespace, phase) (kube_pod_status_phase == 1)`

	CPUUsageByNamespace = `sum by (namespace) (rate(container_cpu_usage_seconds_total{container!="", image!=""}[5m]))`
	MemoryUsageByNS     = `sum by (namespace) (container_memory_working_set_bytes{container!="", image!=""})`
	CPUThrottled        = `sum by (namespace, pod) (increase(container_cpu_cfs_throttled_periods_total{container!=""}[5m])) / sum by (namespace, pod) (increase(container_cpu_cfs_periods_total{container!=""}[5m])) > 0`
	NetworkReceiveByNS  = `sum by (namespace) (rate(container_network_receive_bytes_total{image!=""}[5m]))`
	NetworkTransmitByNS = `sum by (namespace) (rate(container_network_transmit_bytes_total{image!=""}[5m]))`

	PVCCapacity = `kubelet_volume_stats_capacity_bytes`
	PVCUsed     = `kubelet_volume_stats_used_bytes`
	PVCUsedTS   = `kubelet_volume_stats_used_bytes`
	PVCsAtRisk  = `count((1 - kubelet_volume_stats_available_bytes / kubelet_volume_stats_capacity_bytes) > 0.8)`
)
