package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// PGExporterInstance builds the PG Exporter Instance dashboard (Pigsty PGRDS).
//
// This board is a faithful port of the Pigsty postgres_exporter instance
// dashboard and therefore uses the Pigsty metric model (pg_* raw metrics plus
// pg:ins:* / pg:db:* / pg:cls:* recording rules with ins/cls/datname labels),
// not the CloudNativePG (cnpg_*) model used by the other PostgreSQL dashboards.
// The Pigsty cross-dashboard navigation header (dropdown links and $primary
// summary tables that depend on sibling Pigsty dashboards not present in this
// repository) is intentionally omitted; every analytical panel is preserved.
func PGExporterInstance() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PG Exporter Instance").
		Description("PostgreSQL monitoring for remote/RDS instances via the Pigsty postgres_exporter (pg_* metrics and pg:* recording rules). Faithful port of the Pigsty PGRDS instance dashboard by Ruohang Feng. Tested on PostgreSQL 18.").
		Editable(true).
		Tags([]string{"postgresql", "pigsty", "pgrds", "instance", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-1h", "now", "")).
		QueryVariable(panels.SingleQueryVar("ins", "Instance", pgqueries.PigstyInsValues)).
		QueryVariable(panels.SingleQueryVar("cls", "Cluster", pgqueries.PigstyClsValues).
			Hide(dashboardv2.VariableHideHideVariable))

	b = exporterOverviewPanels(b)
	b = exporterActivityPanels(b)
	b = exporterSessionPanels(b)
	b = exporterPersistPanels(b)
	b = exporterDatabasePanels(b)
	b = exporterTableQueryPanels(b)

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview",
			panels.GridItem("kpi-commit", 0, 0, 2, 4),
			panels.GridItem("kpi-rollback", 2, 0, 2, 4),
			panels.GridItem("kpi-rt", 4, 0, 2, 4),
			panels.GridItem("kpi-conn-pct", 6, 0, 2, 4),
			panels.GridItem("kpi-backend", 8, 0, 2, 4),
			panels.GridItem("kpi-active-conn", 10, 0, 2, 4),
			panels.GridItem("kpi-ixact-conn", 12, 0, 2, 4),
			panels.GridItem("kpi-row-fetch", 14, 0, 2, 4),
			panels.GridItem("kpi-row-change", 16, 0, 2, 4),
			panels.GridItem("kpi-blks-read", 18, 0, 2, 4),
			panels.GridItem("kpi-age", 20, 0, 2, 4),
			panels.GridItem("kpi-size", 22, 0, 2, 4),
			panels.GridItem("cluster-load", 0, 4, 6, 6),
			panels.GridItem("instance-load", 6, 4, 9, 6),
			panels.GridItem("alerts-count", 15, 4, 3, 6),
			panels.GridItem("alerts-list", 18, 4, 6, 6),
		),
		panels.Row("Activity",
			panels.GridItem("act-commits-rollbacks", 0, 0, 12, 6),
			panels.GridItem("act-txn-rt", 12, 0, 12, 6),
			panels.GridItem("act-tps", 0, 6, 12, 6),
			panels.GridItem("act-load", 12, 6, 12, 6),
			panels.GridItem("act-row-fetched", 0, 12, 12, 6),
			panels.GridItem("act-row-modified", 12, 12, 12, 6),
			panels.GridItem("act-locks-category", 0, 18, 12, 6),
			panels.GridItem("act-locks", 12, 18, 12, 12),
			panels.GridItem("act-sage", 0, 24, 12, 6),
		),
		panels.Row("Session",
			panels.GridItem("sess-conn-usage", 0, 0, 12, 6),
			panels.GridItem("sess-idle-in-tx", 12, 0, 12, 6),
			panels.GridItem("sess-backends", 0, 6, 12, 6),
			panels.GridItem("sess-new-sessions", 12, 6, 12, 6),
			panels.GridItem("sess-backends-state", 0, 12, 12, 6),
			panels.GridItem("sess-backends-type", 12, 12, 12, 6),
			panels.GridItem("sess-max-conn-lifespan", 0, 18, 12, 6),
			panels.GridItem("sess-backends-wait", 12, 18, 12, 6),
			panels.GridItem("sess-active-pct", 0, 24, 12, 6),
			panels.GridItem("sess-failure", 12, 24, 12, 6),
		),
		panels.Row("Persist",
			panels.GridItem("persist-lsn", 0, 0, 12, 6),
			panels.GridItem("persist-age-usage", 12, 0, 12, 6),
			panels.GridItem("persist-cluster-size", 0, 6, 12, 6),
			panels.GridItem("persist-wal-log-size", 12, 6, 12, 6),
			panels.GridItem("persist-checkpoint-sched", 0, 12, 12, 6),
			panels.GridItem("persist-checkpoint-time", 12, 12, 12, 6),
			panels.GridItem("persist-bgwriter-flush", 0, 18, 12, 6),
			panels.GridItem("persist-bgwriter-alloc", 12, 18, 12, 6),
			panels.GridItem("persist-blocks-access", 0, 24, 12, 6),
			panels.GridItem("persist-blocks-hit-ratio", 12, 24, 12, 6),
			panels.GridItem("persist-blocks-read", 0, 30, 12, 6),
			panels.GridItem("persist-blocks-rw-time", 12, 30, 12, 6),
		),
		panels.Row("Database",
			panels.GridItem("db-size", 0, 0, 12, 6),
			panels.GridItem("db-size-delta", 12, 0, 12, 6),
			panels.GridItem("db-tps", 0, 6, 12, 6),
			panels.GridItem("db-session", 12, 6, 12, 6),
			panels.GridItem("db-blocks-hit", 0, 12, 12, 6),
			panels.GridItem("db-idle-in-tx", 12, 12, 12, 6),
			panels.GridItem("db-conn-usage", 0, 18, 12, 6),
			panels.GridItem("db-new-sessions", 12, 18, 12, 6),
			panels.GridItem("db-row-fetched", 0, 24, 12, 6),
			panels.GridItem("db-row-modified", 12, 24, 12, 6),
		),
		panels.Row("Table & Query",
			panels.GridItem("tq-table-scan", 0, 0, 12, 7),
			panels.GridItem("tq-tuple-read", 12, 0, 12, 7),
			panels.GridItem("tq-query-call", 0, 7, 12, 11),
			panels.GridItem("tq-query-time", 12, 7, 12, 11),
		),
	))
}
