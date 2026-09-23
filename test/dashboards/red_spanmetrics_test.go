package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

func TestRedSpanMetrics(t *testing.T) {
	dash, err := microservices.RedSpanMetrics().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Microservices — RED Span Metrics" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 8 {
		t.Fatalf("expected 8 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("red-spanmetrics", microservices.RedSpanMetrics()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "red-spanmetrics" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"spanmetrics_calls_total",
		"spanmetrics_duration_milliseconds_bucket",
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
