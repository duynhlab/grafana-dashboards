package kubernetes

// PromQL for the Kubernetes Workloads dashboard.
//
// Source: obs-as-code generated/cluster/dashboards/kubernetes-workloads.json,
// commit b3484a65fd59596e74faded0fbf13b648bc6aabd. That file is a rendered
// dashboard.grafana.app/v2 spec, not legacy v1 JSON. This port replaces the
// board homelab currently fetches from that OCI artifact and keeps its UID
// (kubernetes-workloads) and title (Kubernetes Workloads).
//
// Metric model: kube-state-metrics + cAdvisor (job="kubelet" for container_*,
// job="kube-state-metrics" for kube_pod_container_status_* and
// kube_pod_container_resource_*). Every panel joins onto
// namespace_workload_pod:kube_pod_owner:relabel, the ownership recording rule
// that attributes a pod to its owning Deployment/StatefulSet/DaemonSet/Job, so
// the $namespace/$workload_type/$workload variables can filter by owner
// rather than by raw pod name. The source's ${ds} datasource variable is
// dropped: this repository resolves the datasource through
// standards.PrometheusDatasource() instead, so every "${ds}" reference is
// gone and the datasource is not templated in PromQL at all.
const (
	// CPU.
	WorkloadsCPUByPod = `sum by (pod) (sum by (namespace, pod) (rate(container_cpu_usage_seconds_total{job="kubelet",namespace=~"$namespace",container!="",image!=""}[$__rate_interval])) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsCPUByWorkload = `sum by (workload, workload_type) (sum by (namespace, pod) (rate(container_cpu_usage_seconds_total{job="kubelet",namespace=~"$namespace",container!="",image!=""}[$__rate_interval])) * on (namespace, pod) group_left (workload, workload_type) (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsCPUThrottlingByPod = `sum by (pod) (rate(container_cpu_cfs_throttled_periods_total{job="kubelet",namespace=~"$namespace",container!=""}[$__rate_interval]) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})) / clamp_min(sum by (pod) (rate(container_cpu_cfs_periods_total{job="kubelet",namespace=~"$namespace",container!=""}[$__rate_interval]) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})), 1)`

	WorkloadsCPUUsageVsRequestsByPod = `sum by (pod) (sum by (namespace, pod) (rate(container_cpu_usage_seconds_total{job="kubelet",namespace=~"$namespace",container!="",image!=""}[$__rate_interval])) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})) / (sum by (pod) (label_replace(label_replace(kube_pod_container_resource_requests{job="kube-state-metrics",exported_namespace=~"$namespace",resource="cpu"}, "namespace", "$1", "exported_namespace", "(.+)"), "pod", "$1", "exported_pod", "(.+)") * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})) > 0)`

	// Memory.
	WorkloadsMemoryByPod = `sum by (pod) (sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet",namespace=~"$namespace",container!="",image!=""}) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsMemoryByWorkload = `sum by (workload, workload_type) (sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet",namespace=~"$namespace",container!="",image!=""}) * on (namespace, pod) group_left (workload, workload_type) (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsMemoryUsageVsLimitsByPod = `sum by (pod) (sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet",namespace=~"$namespace",container!="",image!=""}) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})) / (sum by (pod) (label_replace(label_replace(kube_pod_container_resource_limits{job="kube-state-metrics",exported_namespace=~"$namespace",resource="memory"}, "namespace", "$1", "exported_namespace", "(.+)"), "pod", "$1", "exported_pod", "(.+)") * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"})) > 0)`

	// Network.
	WorkloadsNetworkReceiveByPod = `sum by (pod) (rate(container_network_receive_bytes_total{job="kubelet",namespace=~"$namespace",namespace!=""}[$__rate_interval]) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsNetworkTransmitByPod = `sum by (pod) (rate(container_network_transmit_bytes_total{job="kubelet",namespace=~"$namespace",namespace!=""}[$__rate_interval]) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	// Reliability.
	WorkloadsRestartsByPod = `sum by (pod) (sum by (namespace, pod) (label_replace(label_replace(rate(kube_pod_container_status_restarts_total{job="kube-state-metrics",exported_namespace=~"$namespace"}[$__rate_interval]), "namespace", "$1", "exported_namespace", "(.+)"), "pod", "$1", "exported_pod", "(.+)")) * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsRestartsByWorkload = `sum by (workload, workload_type) (sum by (namespace, pod) (label_replace(label_replace(rate(kube_pod_container_status_restarts_total{job="kube-state-metrics",exported_namespace=~"$namespace"}[$__rate_interval]), "namespace", "$1", "exported_namespace", "(.+)"), "pod", "$1", "exported_pod", "(.+)")) * on (namespace, pod) group_left (workload, workload_type) (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	WorkloadsContainersWaitingByReason = `sum by (pod, reason) (label_replace(label_replace(kube_pod_container_status_waiting_reason{job="kube-state-metrics",exported_namespace=~"$namespace"}, "namespace", "$1", "exported_namespace", "(.+)"), "pod", "$1", "exported_pod", "(.+)") * on (namespace, pod) group_left () (namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type",workload=~"$workload"}))`

	// Template variables. workload_type and workload chain off $namespace (and
	// each other) so the picker only ever offers owners that actually exist in
	// the selected namespace.
	WorkloadsNamespaceLabelValues    = `label_values(kube_pod_info, exported_namespace)`
	WorkloadsWorkloadTypeLabelValues = `label_values(namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace"}, workload_type)`
	WorkloadsWorkloadLabelValues     = `label_values(namespace_workload_pod:kube_pod_owner:relabel{namespace=~"$namespace",workload_type=~"$workload_type"}, workload)`
)
