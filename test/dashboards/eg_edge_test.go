package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/gateway"
)

func TestEGEdge(t *testing.T) {
	dash, err := gateway.EGEdge().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Envoy Gateway — Edge Overview" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 21 {
		t.Fatalf("expected 21 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("eg-edge", gateway.EGEdge()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "eg-edge" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"envoy_http_downstream_rq_time_bucket",
		"watchable_subscribe_duration_seconds_bucket",
		"envoy_cluster_upstream_rq_retry",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, needle := range []string{
		"DS_PROMETHEUS",
		"[5m]",
	} {
		if strings.Contains(s, needle) {
			t.Fatalf("unexpected fragment present: %s", needle)
		}
	}
}
