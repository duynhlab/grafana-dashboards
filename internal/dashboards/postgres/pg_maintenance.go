package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// PGMaintenance builds the PostgreSQL Maintenance (CNPG) dashboard: locks,
// checkpointer, autovacuum/bloat and long-running transactions.
func PGMaintenance() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PostgreSQL — Maintenance (CNPG)").
		Description("CloudNativePG maintenance signals: locks & blocking, checkpointer, autovacuum & bloat, long transactions and VACUUM progress. Complements pg-io-waits and pg-query-performance. Tested on PostgreSQL 18.").
		Editable(true).
		Tags([]string{"postgresql", "cnpg", "database", "maintenance", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s")).
		QueryVariable(panels.QueryVar("cluster", "Cluster", pgqueries.ClusterLabelValues)).
		QueryVariable(panels.QueryVar("datname", "Database", pgqueries.DatnameLabelValues))

	b = b.
		Panel("locks-by-mode", panels.SeriesExpr("Locks by mode", "short", pgqueries.LocksByMode, "{{mode}}")).
		Panel("blocked", panels.StatValue("Blocked queries", "short", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("yellow", panels.Ptr(1)), panels.Thr("red", panels.Ptr(5)),
		}, pgqueries.BlockedQueries, "")).
		Panel("locks-by-db", panels.SeriesExpr("Locks by database", "short", pgqueries.LocksByDatabase, "{{datname}}")).
		Panel("checkpoints", panels.Series("Checkpoints / sec (timed vs requested)", "ops",
			panels.PromQuery(pgqueries.CheckpointsTimed, "timed"),
			panels.Query("B", pgqueries.CheckpointsRequested, "requested"),
		)).
		Panel("checkpoint-time", panels.Series("Checkpoint write / sync time", "ms",
			panels.PromQuery(pgqueries.CheckpointWriteTime, "write"),
			panels.Query("B", pgqueries.CheckpointSyncTime, "sync"),
		)).
		Panel("dead-tuple-ratio", panels.SeriesExpr("Dead-tuple ratio — top tables", "percentunit", pgqueries.DeadTupleRatioTopTables, "{{datname}}/{{relname}}")).
		Panel("total-dead", panels.StatValue("Total dead tuples", "short", nil, pgqueries.TotalDeadTuples, "")).
		Panel("autovacuum-runs", panels.SeriesExpr("Autovacuum runs / sec", "ops", pgqueries.AutovacuumRuns, "{{datname}}")).
		Panel("top-tables", panels.Table("Top tables by size",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.QueryTable("A", pgqueries.TopTablesBySize)},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{"Time": true, "job": true, "instance": true},
					map[string]string{"datname": "Database", "schemaname": "Schema", "tablename": "Table", "cnpg_io_cluster": "Cluster", "Value": "Size"},
				),
				panels.SortBy("Size", true),
			},
			panels.Column{Name: "Size", Unit: "bytes", Gauge: true},
		)).
		Panel("unused-indexes", panels.Table("Unused indexes (idx_scan = 0)",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.QueryTable("A", pgqueries.UnusedIndexes)},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{"Time": true, "job": true, "instance": true},
					map[string]string{"datname": "Database", "relname": "Table", "indexrelname": "Index", "Value": "Index size"},
				),
				panels.SortBy("Index size", true),
			},
			panels.Column{Name: "Index size", Unit: "bytes"},
		)).
		Panel("oldest-tx", panels.SeriesThresholdLine("Oldest transaction / idle-in-transaction age", "s",
			[]panels.Threshold{panels.Thr("green", nil), panels.Thr("red", panels.Ptr(300))},
			panels.PromQuery(pgqueries.OldestTransactionAge, "{{cnpg_io_cluster}} oldest tx"),
			panels.Query("B", pgqueries.OldestIdleInTxAge, "{{cnpg_io_cluster}} idle-in-tx"),
		)).
		Panel("idle-in-tx", panels.StatValue("Idle-in-tx sessions", "short", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("yellow", panels.Ptr(1)),
		}, pgqueries.IdleInTxSessions, "")).
		Panel("longest-query", panels.SeriesExpr("Longest active query", "s", pgqueries.LongestActiveQuery, "{{cnpg_io_cluster}}")).
		Panel("vacuum-progress", panels.SeriesExpr("Running VACUUM progress", "percentunit", pgqueries.VacuumProgress, "{{datname}}/{{relname}} ({{phase}})")).
		Panel("active-vacuums", panels.Table("Active vacuums",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.QueryTable("A", pgqueries.ActiveVacuumsSnapshot)},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{"Time": true, "job": true, "instance": true, "namespace": true, "pod": true, "container": true, "endpoint": true, "prometheus": true, "Value": true},
					map[string]string{"cnpg_io_cluster": "Cluster", "datname": "DB", "relname": "Table", "phase": "Phase", "pid": "PID"},
				),
			},
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Locks & Blocking",
			panels.GridItem("locks-by-mode", 0, 0, 12, 8),
			panels.GridItem("blocked", 12, 0, 6, 8),
			panels.GridItem("locks-by-db", 18, 0, 6, 8),
		),
		panels.Row("Checkpointer",
			panels.GridItem("checkpoints", 0, 0, 12, 8),
			panels.GridItem("checkpoint-time", 12, 0, 12, 8),
		),
		panels.Row("Autovacuum & Bloat",
			panels.GridItem("dead-tuple-ratio", 0, 0, 12, 8),
			panels.GridItem("total-dead", 12, 0, 6, 8),
			panels.GridItem("autovacuum-runs", 18, 0, 6, 8),
			panels.GridItem("top-tables", 0, 8, 12, 9),
			panels.GridItem("unused-indexes", 12, 8, 12, 9),
		),
		panels.Row("Transactions & Vacuum",
			panels.GridItem("oldest-tx", 0, 0, 12, 8),
			panels.GridItem("idle-in-tx", 12, 0, 6, 8),
			panels.GridItem("longest-query", 18, 0, 6, 8),
			panels.GridItem("vacuum-progress", 0, 8, 12, 8),
			panels.GridItem("active-vacuums", 12, 8, 12, 8),
		),
	))
}
