package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

func TestMicroservicesOTel(t *testing.T) {
	dash, err := microservices.MicroservicesOTel().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Microservices (OTel)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 28 {
		t.Fatalf("expected 28 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("microservices-monitoring-001-otel", microservices.MicroservicesOTel()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "microservices-monitoring-001-otel" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"http_server_request_duration_seconds_bucket",
		"http_server_response_body_size_bytes_sum",
		"rpc_server_call_duration_seconds_count",
		"rpc_client_call_duration_seconds_count",
		"go_memory_gc_goal_bytes",
		"db_client_operation_errors_total",
		"pgxpool_acquired_connections",
		"$__rate_interval",
		"label_values(go_goroutine_count, service_name)",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	// The legacy $rate custom interval and the dead $namespace variable must not
	// survive the port.
	for _, forbidden := range []string{"$rate", "$namespace", "DS_PROMETHEUS"} {
		if strings.Contains(s, forbidden) {
			t.Fatalf("dashboard still references %s", forbidden)
		}
	}
}
