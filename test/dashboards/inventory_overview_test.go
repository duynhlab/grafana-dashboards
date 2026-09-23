package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

func TestInventoryOverview(t *testing.T) {
	dash, err := microservices.InventoryOverview().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Inventory Service — Stock Authority" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 7 {
		t.Fatalf("expected 7 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("inventory-overview", microservices.InventoryOverview()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "inventory-overview" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"inventory:reservation:rate5m",
		"inventory:rpc_duration:p95_5m",
		"rpc_server_call_duration_seconds_count",
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
			t.Fatalf("unexpected query fragment present: %s", needle)
		}
	}
}
