package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGIOWaits(t *testing.T) {
	dash, err := postgres.PGIOWaits().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PostgreSQL — IO & Waits (pg_stat_io)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) < 10 {
		t.Fatalf("expected at least 10 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("pg-io-waits", postgres.PGIOWaits()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "pg-io-waits" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{"cnpg_pg_stat_io_reads", "cnpg_backends_waiting_total", "cnpg_collector_up"} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}
}
