package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGQueryPerformance(t *testing.T) {
	dash, err := postgres.PGQueryPerformance().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PostgreSQL — Query Performance (pg_stat_statements)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 14 {
		t.Fatalf("expected 14 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("pg-query-performance", postgres.PGQueryPerformance()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "pg-query-performance" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"cnpg_pg_stat_statements_calls",
		"cnpg_pg_stat_statements_shared_blks_hit",
		"merge",
		"Exec ms/s",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing fragment: %s", needle)
		}
	}
}
