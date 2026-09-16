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

	if len(dash.Elements) < 19 {
		t.Fatalf("expected at least 19 panels, got %d", len(dash.Elements))
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
	for _, needle := range []string{"kube_node_info", "kube_pod_status_phase", "kubelet_volume_stats_available_bytes"} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}
}
