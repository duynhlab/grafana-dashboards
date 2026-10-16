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

	// db_name, application_name, scope (custom), query_id (textbox).
	if len(dash.Variables) != 4 {
		t.Fatalf("expected 4 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 34 {
		t.Fatalf("expected 34 panels, got %d", len(dash.Elements))
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
	// encoding/json escapes > and < inside the SQL strings; undo it so the
	// needles read as SQL.
	s := strings.NewReplacer(`\u003e`, ">", `\u003c`, "<").Replace(string(body))
	for _, needle := range []string{
		// ClickHouse targets are routed to the ClickHouse plugin and datasource.
		`"group":"grafana-clickhouse-datasource"`,
		`"name":"clickhouse"`,
		`"editorType":"sql"`,
		// The raw store and the rollups.
		"FROM pg_stat_ch.events_raw",
		"FROM pg_stat_ch.query_stats_5m",
		"FROM pg_stat_ch.errors_recent",
		"FROM pg_stat_ch.db_app_user_1m",
		"quantilesTDigestMerge(0.5, 0.95, 0.99)",
		// Both variables scope every events_raw query.
		"db_name IN (${db_name:singlequote})",
		"application_name IN (${application_name:singlequote})",
		"FROM pg_stat_ch.db_app_user_1m ORDER BY application_name",
		"application_name = ''",
		// Real errors only; platform sessions hidden by default.
		"err_elevel >= 21",
		"${scope:raw} = 1",
		// Drill-down, work_mem spills, per-statement worst case.
		"query_id = toInt64OrZero(${query_id:sqlstring})",
		"temp_blks_written > 0",
		"maxMerge(duration_max_state)",
		// The exporter-health row stays on Prometheus.
		`"group":"prometheus"`,
		"cnpg_pg_stat_ch_exported_events",
		"cnpg_pg_stat_ch_dropped_events",
		"cnpg_pg_stat_ch_send_failures",
		"cnpg_pg_stat_ch_last_success_age_seconds",
		"cnpg_pg_stat_ch_dsa_oom_count",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing fragment: %s", needle)
		}
	}
}
