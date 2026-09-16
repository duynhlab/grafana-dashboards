package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
)

// The PG Exporter Instance dashboard is large; panel registration is split by
// section to keep each function readable. Every function registers its panels
// on the shared builder and returns it for chaining. Layout is assembled in
// pg_exporter_instance.go.

func exporterOverviewPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("kpi-commit", panels.StatValue("Commit", "none", nil, pgqueries.InsCommitRate, "Commit")).
		Panel("kpi-rollback", panels.StatValue("Rollback", "none", nil, pgqueries.InsRollbackKPI, "Rollback")).
		Panel("kpi-rt", panels.StatValue("RT", "s", nil, pgqueries.InsRT, "RT")).
		Panel("kpi-conn-pct", panels.StatValue("Conn%", "percentunit", nil, pgqueries.InsConnPct, "Conn%")).
		Panel("kpi-backend", panels.StatValue("Backend", "none", nil, pgqueries.InsBackend, "Backend")).
		Panel("kpi-active-conn", panels.StatValue("Active Conn", "none", nil, pgqueries.InsActiveConn, "Active Conn")).
		Panel("kpi-ixact-conn", panels.StatValue("iXact Conn", "none", nil, pgqueries.InsIxactConn, "iXact Conn")).
		Panel("kpi-row-fetch", panels.StatValue("Row Fetch", "rowsps", nil, pgqueries.InsRowFetch, "Row Fetch")).
		Panel("kpi-row-change", panels.StatValue("Row Change", "rowsps", nil, pgqueries.InsRowChange, "Row Change")).
		Panel("kpi-blks-read", panels.StatValue("Blks Read", "Bps", nil, pgqueries.InsBlksRead, "Blks Read")).
		Panel("kpi-age", panels.StatValue("Age", "percentunit", nil, pgqueries.InsAgePct, "Age")).
		Panel("kpi-size", panels.StatValue("Size", "decbytes", nil, pgqueries.InsSize, "Size")).
		Panel("cluster-load", panels.SeriesExpr("Cluster Load", "short", pgqueries.ClusterLoad, "Load")).
		Panel("instance-load", panels.SeriesExpr("Instance Load", "short", pgqueries.InstanceDBLoad, "{{datname}}")).
		Panel("alerts-count", panels.StatValue("Firing Alerts", "none", nil, pgqueries.FiringAlertsCount, "Alert")).
		Panel("alerts-list", panels.Series("Active Alerts", "none",
			panels.PromQuery(pgqueries.FiringAlerts, "[{{severity}}🔥] {{alertname}}"),
			panels.Query("B", pgqueries.PendingAlerts, "[{{severity}}⏰] {{alertname}}"),
		))
}

func exporterActivityPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("act-commits-rollbacks", panels.Series("Transaction Commits / Rollbacks (rate5m)", "short",
			panels.PromQuery(pgqueries.XactCommitRate5m, "Commits"),
			panels.Query("B", pgqueries.XactRollbackRate5m, "Rollbacks"),
		)).
		Panel("act-txn-rt", panels.Series("Transaction RT (1m)", "s",
			panels.PromQuery(pgqueries.TxnRTInstance, "Instance"),
			panels.Query("B", pgqueries.TxnRTDatabase, "{{datname}}"),
		)).
		Panel("act-tps", panels.SeriesExpr("TPS by Database (rate1m)", "short", pgqueries.TPSByDatabaseAct, "{{datname}}")).
		Panel("act-load", panels.SeriesExpr("Postgres Load (rate1m)", "short", pgqueries.PostgresLoad, "{{datname}}")).
		Panel("act-row-fetched", panels.SeriesExpr("Row Fetched (rate1m)", "cps", pgqueries.RowFetchedAct, "{{datname}}")).
		Panel("act-row-modified", panels.Series("Row Modified (rate1m)", "cps",
			panels.PromQuery(pgqueries.RowInsertedAct, "INSERTED"),
			panels.Query("C", pgqueries.RowUpdatedAct, "UPDATED"),
			panels.Query("D", pgqueries.RowDeletedAct, "DELETED"),
		)).
		Panel("act-locks-category", panels.Series("Locks by Category", "short",
			panels.PromQuery(pgqueries.LocksExclusive, "Exclusive"),
			panels.Query("B", pgqueries.LocksWrite, "Write"),
			panels.Query("C", pgqueries.LocksRead, "Read"),
		)).
		Panel("act-locks", panels.SeriesExpr("Locks", "short", pgqueries.PigstyLocksByMode, "{{mode}}")).
		Panel("act-sage", panels.SeriesExpr("SAGE", "s", pgqueries.MaxTxDuration, "{{state}}"))
}

func exporterSessionPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("sess-conn-usage", panels.SeriesExpr("Connection Usage", "percentunit", pgqueries.ConnUsageSession, "{{datname}}")).
		Panel("sess-idle-in-tx", panels.SeriesExpr("Idle in Transaction Backends", "", pgqueries.IdleInTxBackends, "{{datname}}")).
		Panel("sess-backends", panels.SeriesExpr("Backends", "", pgqueries.BackendsByDatabase, "{{datname}}")).
		Panel("sess-new-sessions", panels.SeriesExpr("New Sessions (increase1m)", "", pgqueries.NewSessions1m, "{{ins}}")).
		Panel("sess-backends-state", panels.SeriesExpr("Backends by State", "", pgqueries.BackendsByState, "{{state}}")).
		Panel("sess-backends-type", panels.SeriesExpr("Backends by Type", "", pgqueries.BackendsByType, "{{type}}")).
		Panel("sess-max-conn-lifespan", panels.SeriesExpr("Max Conn Lifespan", "s", pgqueries.MaxConnLifespan, "{{state}}")).
		Panel("sess-backends-wait", panels.SeriesExpr("Backends by Wait Event", "", pgqueries.BackendsByWaitEvent, "{{event}}")).
		Panel("sess-active-pct", panels.SeriesExpr("Active% (of Session Time)", "percentunit", pgqueries.ActivePctSession, "{{datname}}")).
		Panel("sess-failure", panels.Series("Sessions Failure in 1m", "",
			panels.PromQuery(pgqueries.SessionsAbandoned, "[Abandoned] : {{datname}}"),
			panels.Query("B", pgqueries.SessionsFatal, "[Fatal] : {{datname}}"),
			panels.Query("C", pgqueries.SessionsKilled, "[Killed] : {{datname}}"),
		))
}

func exporterPersistPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("persist-lsn", panels.SeriesExpr("LSN Progress (rate1m)", "Bps", pgqueries.LSNProgress, "{{ins}}")).
		Panel("persist-age-usage", panels.SeriesExpr("Age Usage", "percentunit", pgqueries.AgeUsage, "{{datname}}")).
		Panel("persist-cluster-size", panels.SeriesExpr("Database Cluster Size", "decbytes", pgqueries.DatabaseClusterSize, "{{datname}}")).
		Panel("persist-wal-log-size", panels.SeriesExpr("Database WAL/Log Size", "decbytes", pgqueries.DatabaseWALLogSize, "{{datname}}")).
		Panel("persist-checkpoint-sched", panels.Series("Checkpoint Scheduled / Requested", "",
			panels.PromQuery(pgqueries.CheckpointScheduled, "Scheduled"),
			panels.Query("B", pgqueries.CheckpointRequested, "Requested"),
		)).
		Panel("persist-checkpoint-time", panels.Series("Checkpoint Time : Sync/Write", "ms",
			panels.PromQuery(pgqueries.CheckpointSyncTimeIns, "Sync"),
			panels.Query("B", pgqueries.CheckpointWriteTimeIns, "Write"),
		)).
		Panel("persist-bgwriter-flush", panels.Series("BGWriter Buffer Flush", "decbytes",
			panels.PromQuery(pgqueries.BGWriterClean, "Clean"),
			panels.Query("B", pgqueries.BGWriterCheckpoint, "Checkpoint"),
			panels.Query("C", pgqueries.BGWriterBackend, "Backend"),
		)).
		Panel("persist-bgwriter-alloc", panels.SeriesExpr("BGWriter Buffer Alloc", "decbytes", pgqueries.BGWriterAlloc, "Alloc")).
		Panel("persist-blocks-access", panels.SeriesExpr("Blocks Access 1m", "decbytes", pgqueries.BlocksAccess1m, "{{datname}}")).
		Panel("persist-blocks-hit-ratio", panels.SeriesExpr("Blocks Hit Ratio (1m)", "percentunit", pgqueries.BlocksHitRatio1m, "{{datname}}")).
		Panel("persist-blocks-read", panels.SeriesExpr("Blocks Read (1m)", "decbytes", pgqueries.BlocksRead1m, "{{datname}}")).
		Panel("persist-blocks-rw-time", panels.Series("Blocks Read/Write Time Spent Rate1m", "percentunit",
			panels.PromQuery(pgqueries.BlockReadTimeRate1m, "▲ {{datname}}"),
			panels.Query("B", pgqueries.BlockWriteTimeRate1m, "▼ {{datname}}"),
		))
}

func exporterDatabasePanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("db-size", panels.SeriesExpr("Database Size", "decbytes", pgqueries.DatabaseSize, "{{datname}}")).
		Panel("db-size-delta", panels.SeriesExpr("Database Size Delta 10m", "decbytes", pgqueries.DatabaseSizeDelta, "{{datname}}")).
		Panel("db-tps", panels.SeriesExpr("TPS by Database (rate1m)", "short", pgqueries.TPSByDatabaseDB, "{{datname}}")).
		Panel("db-session", panels.SeriesExpr("Session by Database", "", pgqueries.SessionByDatabase, "{{datname}}")).
		Panel("db-blocks-hit", panels.SeriesExpr("Database Blocks Hit Ratio (1m)", "percentunit", pgqueries.BlocksHitRatioDB, "{{datname}}")).
		Panel("db-idle-in-tx", panels.SeriesExpr("Idle in Transaction Backends", "", pgqueries.IdleInTxBackendsDB, "{{datname}}")).
		Panel("db-conn-usage", panels.SeriesExpr("Connection Usage", "percentunit", pgqueries.ConnUsageDB, "{{datname}}")).
		Panel("db-new-sessions", panels.SeriesExpr("New Sessions (incr1m)", "", pgqueries.NewSessionsDB, "{{datname}}")).
		Panel("db-row-fetched", panels.SeriesExpr("Row Fetched", "cps", pgqueries.RowFetchedDB, "{{datname}}")).
		Panel("db-row-modified", panels.Series("Row Modified", "cps",
			panels.PromQuery(pgqueries.RowInsertedDB, "INSERT.{{datname}}"),
			panels.Query("C", pgqueries.RowUpdatedDB, "UPDATE.{{datname}}"),
			panels.Query("D", pgqueries.RowDeletedDB, "DELETE.{{datname}}"),
		))
}

func exporterTableQueryPanels(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("tq-table-scan", panels.SeriesExpr("Table Scan (scan/s)", "short", pgqueries.TableScanRate1m, "{{datname}}.{{relname}}")).
		Panel("tq-tuple-read", panels.SeriesExpr("Tuple Read (rows/s)", "short", pgqueries.TupleReadRate1m, "{{datname}}.{{relname}}")).
		Panel("tq-query-call", panels.SeriesExpr("Query Call", "ops", pgqueries.QueryCallRate1m, "{{datname}}.{{query}}")).
		Panel("tq-query-time", panels.SeriesExpr("Query Time", "s", pgqueries.QueryExecTime1m, "{{datname}}.{{query}}"))
}
