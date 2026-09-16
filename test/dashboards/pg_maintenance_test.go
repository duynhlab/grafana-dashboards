package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGMaintenance(t *testing.T) {
	dash, err := postgres.PGMaintenance().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PostgreSQL — Maintenance (CNPG)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 15 {
		t.Fatalf("expected 15 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("pg-maintenance", postgres.PGMaintenance()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "pg-maintenance" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"cnpg_pg_locks_count_count",
		"cnpg_pg_stat_checkpointer_checkpoints_timed",
		"cnpg_pg_stat_progress_vacuum_heap_blks_scanned",
		"organize",
		"sortBy",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing fragment: %s", needle)
		}
	}
}
