package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// pgdogLegend is the shared per-series legend for PGDog panels filtered by the
// seven-dimension pooler/host/port/shard/role/database/user selector.
const pgdogLegend = "{{host}}:{{port}} shard:{{shard}} role:{{role}} db:{{database}} user:{{user}}"

// PGDog builds the PGDog connection-pooler dashboard (upstream Grafana.com
// dashboard 24583), normalized onto the repository's logical prometheus
// datasource and the stable UID "pgdog".
func PGDog() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PGDog").
		Description("PGDog connection pooler metrics: clients, servers, connections, transactions, queries, traffic, prepared statements and mirroring. Ported from Grafana.com dashboard 24583.").
		Editable(true).
		Tags([]string{"postgresql", "pgdog", "pooler", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-1h", "now", "1m")).
		QueryVariable(panels.QueryVar("pooler", "Pooler", pgqueries.PGDogPoolerValues)).
		QueryVariable(panels.QueryVar("host", "host", pgqueries.PGDogHostValues)).
		QueryVariable(panels.QueryVar("port", "port", pgqueries.PGDogPortValues)).
		QueryVariable(panels.QueryVar("shard", "shard", pgqueries.PGDogShardValues)).
		QueryVariable(panels.QueryVar("role", "role", pgqueries.PGDogRoleValues)).
		QueryVariable(panels.QueryVar("database", "database", pgqueries.PGDogDatabaseValues)).
		QueryVariable(panels.QueryVar("user", "user", pgqueries.PGDogUserValues))

	b = b.
		Panel("clients-total", panels.SeriesExpr("Clients Total", "short", pgqueries.PGDogClientsTotal, "")).
		Panel("clients-waiting", panels.SeriesExpr("Clients Waiting", "short", pgqueries.PGDogClientsWaiting, pgdogLegend)).
		Panel("sv-active", panels.SeriesExpr("Servers Active", "short", pgqueries.PGDogServersActive, pgdogLegend)).
		Panel("sv-idle", panels.SeriesExpr("Servers Idle", "short", pgqueries.PGDogServersIdle, pgdogLegend)).
		Panel("sv-idle-xact", panels.SeriesExpr("Servers Idle In Transaction", "short", pgqueries.PGDogServersIdleXact, pgdogLegend)).
		Panel("maxwait", panels.SeriesExpr("Connection Max Wait Time", "s", pgqueries.PGDogMaxWait, pgdogLegend)).
		Panel("errors", panels.SeriesExpr("Connection Errors Rate", "short", pgqueries.PGDogErrors, pgdogLegend)).
		Panel("out-of-sync", panels.SeriesExpr("Connection Out Of Sync Rate", "short", pgqueries.PGDogOutOfSync, pgdogLegend)).
		Panel("xact-rate", panels.SeriesExpr("Transactions Rate", "short", pgqueries.PGDogXactRate, pgdogLegend)).
		Panel("2pc-rate", panels.SeriesExpr("Two-Phase Commit Transactions Rate", "short", pgqueries.PGDog2PCRate, pgdogLegend)).
		Panel("xact-time", panels.SeriesExpr("Transactions Executing Time Rate", "dtdurationms", pgqueries.PGDogXactTime, pgdogLegend)).
		Panel("avg-xact", panels.SeriesExpr("Average Transactions Count", "short", pgqueries.PGDogAvgXact, pgdogLegend)).
		Panel("avg-2pc", panels.SeriesExpr("Average Two-Phase Commit Transactions Count", "short", pgqueries.PGDogAvg2PC, pgdogLegend)).
		Panel("avg-xact-time", panels.SeriesExpr("Average Time Spent Executing Transactions", "dtdurationms", pgqueries.PGDogAvgXactTime, pgdogLegend)).
		Panel("query-rate", panels.SeriesExpr("Query Rate", "short", pgqueries.PGDogQueryRate, pgdogLegend)).
		Panel("query-time", panels.SeriesExpr("Query Executing Time Rate", "dtdurationms", pgqueries.PGDogQueryTime, pgdogLegend)).
		Panel("avg-query", panels.SeriesExpr("Average Query Count", "short", pgqueries.PGDogAvgQuery, pgdogLegend)).
		Panel("avg-query-time", panels.SeriesExpr("Average Query Executing Time", "dtdurationms", pgqueries.PGDogAvgQueryTime, pgdogLegend)).
		Panel("cache-size", panels.SeriesExpr("Query Cache Size", "short", pgqueries.PGDogQueryCacheSize, "")).
		Panel("cache-rate", panels.Series("Query Cache Rate", "short",
			panels.PromQuery(pgqueries.PGDogQueryCacheHits, "HIT {{instance}}"),
			panels.Query("B", pgqueries.PGDogQueryCacheMisses, "MISS {{instance}}"),
			panels.Query("C", pgqueries.PGDogQueryCacheDirect, "DIRECT {{instance}}"),
			panels.Query("D", pgqueries.PGDogQueryCacheCross, "CROSS {{instance}}"),
		)).
		Panel("received-rate", panels.SeriesExpr("Traffic Received Rate", "decbytes", pgqueries.PGDogReceivedRate, pgdogLegend)).
		Panel("sent-rate", panels.SeriesExpr("Traffic Sent Rate", "decbytes", pgqueries.PGDogSentRate, pgdogLegend)).
		Panel("avg-received", panels.SeriesExpr("Average Traffic Received", "decbytes", pgqueries.PGDogAvgReceived, pgdogLegend)).
		Panel("avg-sent", panels.SeriesExpr("Average Traffic Sent", "decbytes", pgqueries.PGDogAvgSent, pgdogLegend)).
		Panel("prep-memory", panels.SeriesExpr("Prepared Statements Memory Used", "decbytes", pgqueries.PGDogPreparedMemory, "")).
		Panel("prep-evict-rate", panels.SeriesExpr("Prepared Statements Eviction Rate", "short", pgqueries.PGDogPreparedEvictRate, pgdogLegend)).
		Panel("prep-count", panels.SeriesExpr("Prepared Statements Count", "short", pgqueries.PGDogPreparedCount, "")).
		Panel("avg-prep-evict", panels.SeriesExpr("Average Prepared Statements Eviction Count", "short", pgqueries.PGDogAvgPreparedEvict, pgdogLegend)).
		Panel("mirror-rate", panels.Series("Mirror Rate", "short",
			panels.PromQuery(pgqueries.PGDogMirrorTotal, "TOTAL {{instance}} db:{{database}} user:{{user}}"),
			panels.Query("B", pgqueries.PGDogMirrorMirrored, "MIRRORED {{instance}} db:{{database}} user:{{user}}"),
			panels.Query("C", pgqueries.PGDogMirrorDropped, "DROPPED {{instance}} db:{{database}} user:{{user}}"),
			panels.Query("D", pgqueries.PGDogMirrorError, "ERROR {{instance}} db:{{database}} user:{{user}}"),
		)).
		Panel("mirror-queue", panels.SeriesExpr("Mirror Queue Size", "short", pgqueries.PGDogMirrorQueue, "{{instance}} db:{{database}} user:{{user}}"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Client",
			panels.GridItem("clients-total", 0, 0, 12, 8),
			panels.GridItem("clients-waiting", 12, 0, 12, 8),
		),
		panels.Row("Server",
			panels.GridItem("sv-active", 0, 0, 8, 9),
			panels.GridItem("sv-idle", 8, 0, 8, 9),
			panels.GridItem("sv-idle-xact", 16, 0, 8, 9),
		),
		panels.Row("Connection",
			panels.GridItem("maxwait", 0, 0, 8, 9),
			panels.GridItem("errors", 8, 0, 8, 9),
			panels.GridItem("out-of-sync", 16, 0, 8, 9),
		),
		panels.Row("Transaction",
			panels.GridItem("xact-rate", 0, 0, 8, 9),
			panels.GridItem("2pc-rate", 8, 0, 8, 9),
			panels.GridItem("xact-time", 16, 0, 8, 9),
			panels.GridItem("avg-xact", 0, 9, 8, 9),
			panels.GridItem("avg-2pc", 8, 9, 8, 9),
			panels.GridItem("avg-xact-time", 16, 9, 8, 9),
		),
		panels.Row("Query",
			panels.GridItem("query-rate", 0, 0, 12, 8),
			panels.GridItem("query-time", 12, 0, 12, 8),
			panels.GridItem("avg-query", 0, 8, 12, 8),
			panels.GridItem("avg-query-time", 12, 8, 12, 8),
			panels.GridItem("cache-size", 0, 16, 12, 8),
			panels.GridItem("cache-rate", 12, 16, 12, 8),
		),
		panels.Row("Traffic",
			panels.GridItem("received-rate", 0, 0, 12, 9),
			panels.GridItem("sent-rate", 12, 0, 12, 9),
			panels.GridItem("avg-received", 0, 9, 12, 9),
			panels.GridItem("avg-sent", 12, 9, 12, 9),
		),
		panels.Row("Prepared Statement",
			panels.GridItem("prep-memory", 0, 0, 12, 8),
			panels.GridItem("prep-evict-rate", 12, 0, 12, 8),
			panels.GridItem("prep-count", 0, 8, 12, 8),
			panels.GridItem("avg-prep-evict", 12, 8, 12, 8),
		),
		panels.Row("Mirror",
			panels.GridItem("mirror-rate", 0, 0, 12, 8),
			panels.GridItem("mirror-queue", 12, 0, 12, 8),
		),
	))
}
