package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
)

func TestPGQueryTelemetry(t *testing.T) {
	dash, err := postgres.PGQueryTelemetry().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "PostgreSQL — Query telemetry (pg_stat_ch)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) < 20 {
		t.Fatalf("expected at least 20 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("postgresql-query-telemetry", postgres.PGQueryTelemetry()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "postgresql-query-telemetry" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		// ClickHouse targets are routed to the ClickHouse plugin and datasource.
		`"group":"grafana-clickhouse-datasource"`,
		`"name":"clickhouse"`,
		`"editorType":"sql"`,
		// The raw store and the rollups.
		"FROM pg_stat_ch.events_raw",
		"FROM pg_stat_ch.query_stats_5m",
		"FROM pg_stat_ch.errors_recent",
		"quantilesTDigestMerge(0.5, 0.95, 0.99)",
		// Both variables scope every events_raw query.
		"db_name IN (${db_name:singlequote})",
		"app IN (${app:singlequote})",
		// The exporter-health row stays on Prometheus.
		`"group":"prometheus"`,
		"cnpg_pg_stat_ch_exported_events",
		"cnpg_pg_stat_ch_dropped_events",
		"cnpg_pg_stat_ch_send_failures",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing fragment: %s", needle)
		}
	}
}
