package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// PGQueryPerformance builds the PostgreSQL Query Performance dashboard sourced
// from pg_stat_statements (CNPG collector).
func PGQueryPerformance() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PostgreSQL — Query Performance (pg_stat_statements)").
		Description("CloudNativePG pg_stat_statements: throughput, latency, cache efficiency, block I/O and top statements. Complements pg-io-waits and pg-maintenance. Tested on PostgreSQL 18.").
		Editable(true).
		Tags([]string{"postgresql", "cnpg", "database", "query", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s")).
		QueryVariable(panels.QueryVar("cluster", "Cluster", pgqueries.ClusterLabelValues)).
		QueryVariable(panels.QueryVar("datname", "Database", pgqueries.DatnameLabelValues))

	b = b.
		Panel("calls", panels.StatValue("Calls / sec", "ops", nil, pgqueries.CallsPerSec, "")).
		Panel("mean-latency", panels.StatValue("Mean latency / call", "ms", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("yellow", panels.Ptr(50)), panels.Thr("red", panels.Ptr(200)),
		}, pgqueries.MeanLatency, "")).
		Panel("rows", panels.StatValue("Rows / sec", "ops", nil, pgqueries.RowsPerSec, "")).
		Panel("cache-hit", panels.StatValue("Cache hit ratio", "percentunit", []panels.Threshold{
			panels.Thr("red", nil), panels.Thr("yellow", panels.Ptr(0.9)), panels.Thr("green", panels.Ptr(0.99)),
		}, pgqueries.CacheHitRate, "")).
		Panel("top-exec", panels.SeriesExpr("Top by exec time (ms/s)", "ms", pgqueries.TopByExecTime, "{{query}}")).
		Panel("top-calls", panels.SeriesExpr("Top by calls / sec", "ops", pgqueries.TopByCalls, "{{query}}")).
		Panel("top-mean", panels.SeriesExpr("Top by mean latency / call", "ms", pgqueries.TopByMeanLatency, "{{query}}")).
		Panel("top-rows", panels.SeriesExpr("Top by rows / sec", "ops", pgqueries.TopByRows, "{{query}}")).
		Panel("shared-blocks", panels.Series("Shared blocks: hit vs read /s", "ops",
			panels.PromQuery(pgqueries.SharedBlocksHit, "hit"),
			panels.Query("B", pgqueries.SharedBlocksRead, "read (disk)"),
		)).
		Panel("temp-blocks", panels.Series("Temp blocks (spill) /s", "ops",
			panels.PromQuery(pgqueries.TempBlocksRead, "temp read"),
			panels.Query("B", pgqueries.TempBlocksWritten, "temp written"),
		)).
		Panel("stmt-table", panels.Table("Top statements (by exec time)",
			[]cog.Builder[dashboardv2.PanelQueryKind]{
				panels.QueryTable("A", pgqueries.StatementExecTime),
				panels.QueryTable("B", pgqueries.StatementCalls),
				panels.QueryTable("C", pgqueries.StatementMeanTime),
				panels.QueryTable("D", pgqueries.StatementRows),
			},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Merge(),
				panels.Organize(
					map[string]bool{"Time": true},
					map[string]string{
						"datname": "Database", "query": "Query", "queryid": "Query ID",
						"Value #A": "Exec ms/s", "Value #B": "Calls/s", "Value #C": "Mean ms", "Value #D": "Rows/s",
					},
				),
				panels.SortBy("Exec ms/s", true),
			},
			panels.Column{Name: "Exec ms/s", Unit: "ms", Gauge: true},
			panels.Column{Name: "Mean ms", Unit: "ms"},
			panels.Column{Name: "Query", Width: panels.Ptr(560)},
		)).
		Panel("io-time", panels.Series("Block I/O time /s (read vs write)", "ms",
			panels.PromQuery(pgqueries.BlockReadTime, "read"),
			panels.Query("B", pgqueries.BlockWriteTime, "write"),
		)).
		Panel("top-io", panels.SeriesExpr("Top queries by I/O time", "ms", pgqueries.TopByIOTime, "{{query}}")).
		Panel("top-rows-per-call", panels.SeriesExpr("Top queries by rows per call", "short", pgqueries.TopByRowsPerCall, "{{query}}"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview",
			panels.GridItem("calls", 0, 0, 6, 6),
			panels.GridItem("mean-latency", 6, 0, 6, 6),
			panels.GridItem("rows", 12, 0, 6, 6),
			panels.GridItem("cache-hit", 18, 0, 6, 6),
		),
		panels.Row("Top statements",
			panels.GridItem("top-exec", 0, 0, 12, 8),
			panels.GridItem("top-calls", 12, 0, 12, 8),
			panels.GridItem("top-mean", 0, 8, 12, 8),
			panels.GridItem("top-rows", 12, 8, 12, 8),
		),
		panels.Row("Block I/O",
			panels.GridItem("shared-blocks", 0, 0, 12, 8),
			panels.GridItem("temp-blocks", 12, 0, 12, 8),
		),
		panels.Row("Statement detail",
			panels.GridItem("stmt-table", 0, 0, 24, 10),
		),
		panels.Row("Block I/O time & row efficiency",
			panels.GridItem("io-time", 0, 0, 12, 8),
			panels.GridItem("top-io", 12, 0, 12, 8),
			panels.GridItem("top-rows-per-call", 0, 8, 12, 8),
		),
	))
}
