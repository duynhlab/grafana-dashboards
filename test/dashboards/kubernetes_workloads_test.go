package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/kubernetes"
)

func TestKubernetesWorkloads(t *testing.T) {
	dash, err := kubernetes.Workloads().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Kubernetes Workloads" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 11 {
		t.Fatalf("expected 11 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("kubernetes-workloads", kubernetes.Workloads()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "kubernetes-workloads" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"container_cpu_usage_seconds_total",
		"container_memory_working_set_bytes",
		"kube_pod_container_status_restarts_total",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, absent := range []string{"${ds}", "[5m]"} {
		if strings.Contains(s, absent) {
			t.Fatalf("unexpected fragment present: %s", absent)
		}
	}
}
