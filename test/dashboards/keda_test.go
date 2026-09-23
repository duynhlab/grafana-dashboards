package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/kubernetes"
)

func TestKEDA(t *testing.T) {
	dash, err := kubernetes.KEDA().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "KEDA — Worker Autoscaling" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 17 {
		t.Fatalf("expected 17 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("keda", kubernetes.KEDA()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "keda" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"keda_scaler_metrics_value",
		"kube_horizontalpodautoscaler_status_current_replicas",
		"approximate_backlog_count",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, forbidden := range []string{
		"DS_PROMETHEUS",
		"[5m]",
		"order|checkout",
		"order-fulfillment",
	} {
		if strings.Contains(s, forbidden) {
			t.Fatalf("marshalled JSON must not contain %q", forbidden)
		}
	}
}
