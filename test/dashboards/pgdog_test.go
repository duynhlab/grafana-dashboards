package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGDog(t *testing.T) {
	dash, err := postgres.PGDog().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PGDog" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 7 {
		t.Fatalf("expected 7 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 30 {
		t.Fatalf("expected 30 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("pgdog", postgres.PGDog()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "pgdog" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"pgdog_clients",
		"pgdog_total_xact_count",
		"pgdog_mirror_total_count",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}
}
