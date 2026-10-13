package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	chqueries "github.com/duynhlab/grafana-dashboards/internal/queries/clickhouse/postgres"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// Thresholds shared by the latency stats and the table gauges: 50 ms is where a
// pooled OLTP statement stops being "fast", 1 s is the auto_explain threshold.
var (
	latencySteps = []panels.Threshold{
		panels.Thr("green", nil),
		panels.Thr("yellow", panels.Ptr(50)),
		panels.Thr("orange", panels.Ptr(250)),
		panels.Thr("red", panels.Ptr(1000)),
	}
	errorSteps = []panels.Threshold{
		panels.Thr("green", nil),
		panels.Thr("yellow", panels.Ptr(1)),
		panels.Thr("red", panels.Ptr(5)),
	}
	hitSteps = []panels.Threshold{
		panels.Thr("red", nil),
		panels.Thr("yellow", panels.Ptr(90)),
		panels.Thr("green", panels.Ptr(99)),
	}
	neutralSteps = []panels.Threshold{panels.Thr("blue", nil)}
)

const queryTelemetryIntro = "### PostgreSQL — Query telemetry (pg_stat_ch)\n" +
	"One event **per statement execution** from `platform-db` and `product-db`, shipped by the " +
	"`pg_stat_ch` extension into ClickHouse `pg_stat_ch.events_raw` (7 days hot, 90 days total). " +
	"This answers what `pg_stat_statements` cannot: **server-side p50/p95/p99 over time**, errors by " +
	"SQLSTATE, per-app/role load, and *what changed between two moments*. `db_name` is the cluster " +
	"key (every database name is unique across clusters). Under overload the ring **drops, never queues** " +
	"— check the exporter-health row before trusting a dip."

// PGQueryTelemetry is the per-execution companion to pg-query-performance:
// the same statements, but as raw events in ClickHouse rather than cumulative
// counters in VictoriaMetrics. Every panel but the last row queries ClickHouse;
// the exporter-health row reads the extension's own counters from Prometheus.
func PGQueryTelemetry() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PostgreSQL — Query telemetry (pg_stat_ch)").
		Description("Per-execution PostgreSQL statement events from the pg_stat_ch extension in ClickHouse: latency percentiles over time, top statements, errors, load attribution, and exporter health.").
		Editable(true).
		Tags([]string{"postgresql", "cloudnative-pg", "clickhouse", "pg_stat_ch", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-3h", "now", "30s")).
		Links([]cog.Builder[dashboardv2.DashboardLink]{
			dashboardv2.NewDashboardLinkBuilder().
				Title("CloudNativePG Cluster Overview").
				Type(dashboardv2.DashboardLinkTypeLink).
				Url("/d/cloudnative-pg").
				Icon("dashboard"),
			dashboardv2.NewDashboardLinkBuilder().
				Title("Runbook: plan regression").
				Type(dashboardv2.DashboardLinkTypeLink).
				Url("https://github.com/duynhlab/homelab/blob/main/docs/observability/runbooks/postgresql/plan-regression-investigation.md").
				Icon("doc").
				TargetBlank(true),
		}).
		QueryVariable(panels.ClickHouseQueryVar("db_name", "Database", chqueries.DBNameValues)).
		QueryVariable(panels.ClickHouseQueryVar("app", "Application", chqueries.AppValues))

	b = b.
		Panel("intro", panels.Text("", queryTelemetryIntro)).
		// Overview: range-wide figures, one row each.
		Panel("statements-per-s", panels.StatValues("Statements/s", "reqps", neutralSteps, panels.ClickHouseTable("A", chqueries.StatementsPerSec))).
		Panel("p50", panels.StatValues("p50 latency", "ms", latencySteps, panels.ClickHouseTable("A", chqueries.P50Ms))).
		Panel("p95", panels.StatValues("p95 latency", "ms", latencySteps, panels.ClickHouseTable("A", chqueries.P95Ms))).
		Panel("p99", panels.StatValues("p99 latency", "ms", latencySteps, panels.ClickHouseTable("A", chqueries.P99Ms))).
		Panel("error-rate", panels.StatValues("Error rate", "percent", errorSteps, panels.ClickHouseTable("A", chqueries.ErrorPct))).
		Panel("buffer-hit", panels.StatValues("Buffer hit ratio", "percent", hitSteps, panels.ClickHouseTable("A", chqueries.BufferHitPct))).
		// Traffic & errors.
		Panel("rate-by-operation", panels.Series("Statements/s by operation", "reqps", panels.ClickHouseSeries("A", chqueries.RateByOperation))).
		Panel("rate-by-database", panels.Series("Statements/s by database", "reqps", panels.ClickHouseSeries("A", chqueries.RateByDatabase))).
		Panel("errors-by-sqlstate", panels.Series("Errors/s by SQLSTATE", "reqps", panels.ClickHouseSeries("A", chqueries.ErrorsBySQLState))).
		Panel("rate-by-application", panels.Series("Statements/s by application", "reqps", panels.ClickHouseSeries("A", chqueries.RateByApplication))).
		// Latency.
		Panel("latency-percentiles", panels.Series("Statement latency percentiles", "ms", panels.ClickHouseSeries("A", chqueries.LatencyPercentiles))).
		Panel("p95-by-database", panels.Series("p95 latency by database", "ms", panels.ClickHouseSeries("A", chqueries.P95ByDatabase))).
		Panel("cpu-vs-io", panels.Series("Time split: CPU vs I/O wait", "ms", panels.ClickHouseSeries("A", chqueries.CPUVersusIOWait))).
		Panel("wal-and-temp", panels.Series("WAL bytes/s and temp bytes/s", "Bps", panels.ClickHouseSeries("A", chqueries.WALAndTempBytes))).
		// Top statements (by query_id fingerprint, shared with pg_stat_statements).
		Panel("top-by-total-time", panels.Table("Top statements by total time",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.ClickHouseTable("A", chqueries.TopStatementsByTotalTime)}, nil,
			panels.Column{Name: "p95_ms", Unit: "ms", Gauge: true},
			panels.Column{Name: "p99_ms", Unit: "ms", Gauge: true},
			panels.Column{Name: "hit_pct", Unit: "percent", Gauge: true},
			panels.Column{Name: "query_text", Width: panels.Ptr(520)},
		)).
		Panel("top-by-physical-reads", panels.Table("Top statements by physical reads",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.ClickHouseTable("A", chqueries.TopStatementsByPhysicalReads)}, nil,
			panels.Column{Name: "avg_ms", Unit: "ms"},
			panels.Column{Name: "query_text", Width: panels.Ptr(360)},
		)).
		Panel("slowest-executions", panels.Table("Slowest individual executions",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.ClickHouseTable("A", chqueries.SlowestExecutions)}, nil,
			panels.Column{Name: "duration_ms", Unit: "ms", Gauge: true},
			panels.Column{Name: "query_text", Width: panels.Ptr(300)},
		)).
		// Errors.
		Panel("recent-errors", panels.Table("Recent errors",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.ClickHouseTable("A", chqueries.RecentErrors)}, nil,
			panels.Column{Name: "err_message", Width: panels.Ptr(380)},
			panels.Column{Name: "query_text", Width: panels.Ptr(380)},
		)).
		// Exporter health (Prometheus): is the picture complete?
		Panel("events-flow", panels.Series("Events: enqueued / exported / dropped per second", "reqps",
			panels.Query("A", pgqueries.PgStatChEnqueuedRate, "{{cnpg_io_cluster}} enqueued"),
			panels.Query("B", pgqueries.PgStatChExportedRate, "{{cnpg_io_cluster}} exported"),
			panels.Query("C", pgqueries.PgStatChDroppedRate, "{{cnpg_io_cluster}} dropped"),
		)).
		Panel("send-failures", panels.Series("Send failures (5m) and ring usage %", "short",
			panels.Query("A", pgqueries.PgStatChSendFailures, "{{cnpg_io_cluster}} send failures / 5m"),
			panels.Query("B", pgqueries.PgStatChQueueUsage, "{{pod}} ring %"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview — health at a glance",
			panels.GridItem("intro", 0, 0, 24, 3),
			panels.GridItem("statements-per-s", 0, 3, 4, 5),
			panels.GridItem("p50", 4, 3, 4, 5),
			panels.GridItem("p95", 8, 3, 4, 5),
			panels.GridItem("p99", 12, 3, 4, 5),
			panels.GridItem("error-rate", 16, 3, 4, 5),
			panels.GridItem("buffer-hit", 20, 3, 4, 5),
		),
		panels.Row("Traffic & errors",
			panels.GridItem("rate-by-operation", 0, 0, 12, 8),
			panels.GridItem("rate-by-database", 12, 0, 12, 8),
			panels.GridItem("errors-by-sqlstate", 0, 8, 12, 8),
			panels.GridItem("rate-by-application", 12, 8, 12, 8),
		),
		panels.Row("Latency",
			panels.GridItem("latency-percentiles", 0, 0, 12, 8),
			panels.GridItem("p95-by-database", 12, 0, 12, 8),
			panels.GridItem("cpu-vs-io", 0, 8, 12, 8),
			panels.GridItem("wal-and-temp", 12, 8, 12, 8),
		),
		panels.Row("Top statements (by fingerprint)",
			panels.GridItem("top-by-total-time", 0, 0, 24, 11),
			panels.GridItem("top-by-physical-reads", 0, 11, 12, 9),
			panels.GridItem("slowest-executions", 12, 11, 12, 9),
		),
		panels.Row("Errors",
			panels.GridItem("recent-errors", 0, 0, 24, 10),
		),
		panels.Row("Exporter health — is the picture complete? (cnpg_pg_stat_ch_*)",
			panels.GridItem("events-flow", 0, 0, 12, 8),
			panels.GridItem("send-failures", 12, 0, 12, 8),
		),
	))
}
