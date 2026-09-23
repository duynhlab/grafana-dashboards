package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/kubernetes"
)

func TestKubernetesClusterOverview(t *testing.T) {
	dash, err := kubernetes.ClusterOverview().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Kubernetes Cluster Overview" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Elements) != 24 {
		t.Fatalf("expected 24 panels, got %d", len(dash.Elements))
	}

	if len(dash.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(dash.Variables))
	}

	manifest, err := dashboardv2.Manifest("kubernetes-cluster-overview", kubernetes.ClusterOverview()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "kubernetes-cluster-overview" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}
	if manifest.ApiVersion != "dashboard.grafana.app/v2" {
		t.Fatalf("unexpected apiVersion: %s", manifest.ApiVersion)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"kube_node_info",
		"kube_pod_status_phase",
		"kubelet_volume_stats_available_bytes",
		"kube_node_status_condition",
		"container_network_receive_packets_dropped_total",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, needle := range []string{"${ds}", "[5m]"} {
		if strings.Contains(s, needle) {
			t.Fatalf("unexpected query fragment present: %s", needle)
		}
	}
}
