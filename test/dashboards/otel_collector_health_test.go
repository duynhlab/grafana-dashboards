package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/observability"
)

func TestOTelCollectorHealth(t *testing.T) {
	dash, err := observability.OTelCollectorHealth().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "OTel Collector Health" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 10 {
		t.Fatalf("expected 10 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("otel-collector-health", observability.OTelCollectorHealth()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "otel-collector-health" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"otelcol_receiver_accepted_spans",
		"otelcol_exporter_send_failed_spans",
		"otelcol_process_memory_rss",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, absent := range []string{
		"DS_PROMETHEUS",
		"[5m]",
	} {
		if strings.Contains(s, absent) {
			t.Fatalf("unexpected fragment present: %s", absent)
		}
	}
}
