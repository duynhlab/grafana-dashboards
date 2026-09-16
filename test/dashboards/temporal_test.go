package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/observability"
)

func TestTemporalWorker(t *testing.T) {
	dash, err := observability.TemporalWorker().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Temporal — Workflows & Activities" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Elements) != 8 {
		t.Fatalf("expected 8 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("temporal-worker", observability.TemporalWorker()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "temporal-worker" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"temporal_workflow_completed",
		"temporal_activity_execution_latency_seconds_bucket",
		"temporal_worker_task_slots_available",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}
}
