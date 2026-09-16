package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGExporterInstance(t *testing.T) {
	dash, err := postgres.PGExporterInstance().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PG Exporter Instance" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 61 {
		t.Fatalf("expected 61 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("pg-exporter-instance", postgres.PGExporterInstance()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "pg-exporter-instance" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	// Faithful to the Pigsty metric model (pg_* / pg:* with ins/cls labels).
	for _, needle := range []string{
		"pg:ins:xact_commit_rate1m",
		"pg_activity_count",
		"pg:query:call_rate1m",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing Pigsty fragment: %s", needle)
		}
	}
	// Must not have been rewritten onto the CNPG model.
	if strings.Contains(s, "cnpg_") {
		t.Fatal("pg-exporter-instance unexpectedly contains cnpg_ metrics")
	}
}
