package alerts

import (
	k8squeries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/kubernetes"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func CrashLoopingPods() Rule {
	return Rule{
		UID:       "kubernetes_crashlooping_pods",
		Title:     "Kubernetes crashlooping pods",
		Domain:    standards.DomainKubernetes,
		Folder:    standards.FolderKubernetes,
		Group:     "kubernetes",
		Expr:      k8squeries.CrashLoopingPods,
		For:       "5m",
		Threshold: 0,
		Labels: map[string]string{
			standards.LabelSeverity:  standards.SeverityWarning,
			standards.LabelDomain:    "kubernetes",
			standards.LabelComponent: "workloads",
		},
		Annotations: map[string]string{
			standards.AnnotationSummary:      "Pods are CrashLoopBackOff",
			standards.AnnotationDescription:  "At least one container has been in CrashLoopBackOff over the last 5m.",
			standards.AnnotationDashboardUID: "kubernetes-cluster-overview",
		},
	}
}

func PendingPods() Rule {
	return Rule{
		UID:       "kubernetes_pending_pods",
		Title:     "Kubernetes pending pods",
		Domain:    standards.DomainKubernetes,
		Folder:    standards.FolderKubernetes,
		Group:     "kubernetes",
		Expr:      k8squeries.PendingPods,
		For:       "15m",
		Threshold: 0,
		Labels: map[string]string{
			standards.LabelSeverity:  standards.SeverityWarning,
			standards.LabelDomain:    "kubernetes",
			standards.LabelComponent: "scheduler",
		},
		Annotations: map[string]string{
			standards.AnnotationSummary:      "Pods stuck Pending",
			standards.AnnotationDescription:  "One or more pods have remained Pending.",
			standards.AnnotationDashboardUID: "kubernetes-cluster-overview",
		},
	}
}

func PVCsAtRisk() Rule {
	return Rule{
		UID:       "kubernetes_pvcs_at_risk",
		Title:     "Kubernetes PVCs above 80% used",
		Domain:    standards.DomainKubernetes,
		Folder:    standards.FolderKubernetes,
		Group:     "kubernetes",
		Expr:      k8squeries.PVCsAtRisk,
		For:       "15m",
		Threshold: 0,
		Labels: map[string]string{
			standards.LabelSeverity:  standards.SeverityWarning,
			standards.LabelDomain:    "kubernetes",
			standards.LabelComponent: "storage",
		},
		Annotations: map[string]string{
			standards.AnnotationSummary:      "PersistentVolumeClaims at risk",
			standards.AnnotationDescription:  "At least one PVC is more than 80% full.",
			standards.AnnotationDashboardUID: "kubernetes-cluster-overview",
		},
	}
}
