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

	if len(dash.Variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 21 {
		t.Fatalf("expected 21 panels, got %d", len(dash.Elements))
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
		"temporal_workflow_completed_total",
		"temporal_workflow_failed_total",
		"temporal_activity_execution_latency_seconds_bucket",
		"temporal_worker_task_slots_available",
		"temporal_request_total",
		"temporal_request_failure_total",
		"service_requests",
		"service_error_with_type",
		"persistence_requests",
		"persistence_error_with_type",
		"approximate_backlog_count",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, needle := range []string{
		"DS_PROMETHEUS",
		"[5m]",
		`temporal_workflow_completed"`,
		"temporal_workflow_completed[",
	} {
		if strings.Contains(s, needle) {
			t.Fatalf("unexpected query fragment present: %s", needle)
		}
	}
}
