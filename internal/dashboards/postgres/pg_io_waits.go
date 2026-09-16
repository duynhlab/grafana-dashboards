package postgres

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func PGIOWaits() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("PostgreSQL — IO & Waits (pg_stat_io)").
		Description("Instance-level IO attribution (PG18 pg_stat_io) and sampled wait classes. Complements pg-query-performance and pg-maintenance.").
		Editable(true).
		Tags([]string{"postgresql", "cnpg", "database", "io", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-6h", "now", "30s")).
		QueryVariable(dashboardv2.NewQueryVariableBuilder("cluster").
			Label("cluster").
			Definition(pgqueries.ClusterLabelValues).
			Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
			IncludeAll(true).
			Multi(true).
			Sort(dashboardv2.VariableSortAlphabeticalAsc),
		).
		QueryVariable(dashboardv2.NewQueryVariableBuilder("pod").
			Label("pod").
			Definition(pgqueries.PodLabelValues).
			Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
			IncludeAll(true).
			Multi(true).
			Sort(dashboardv2.VariableSortAlphabeticalAsc),
		)

	b = b.
		Panel("read-ops", panels.TimeSeriesExpr("Read ops /s by backend", pgqueries.ReadOpsByBackend, "{{backend_type}}")).
		Panel("write-ops", panels.TimeSeriesExpr("Write ops /s by backend", pgqueries.WriteOpsByBackend, "{{backend_type}}")).
		Panel("throughput", panels.TimeSeries("Throughput (bytes/s)",
			panels.PromQuery(pgqueries.ReadThroughput, "read"),
			panels.PromQuery(pgqueries.WriteThroughput, "write"),
		)).
		Panel("io-time", panels.TimeSeries("IO time (ms spent /s) by backend",
			panels.PromQuery(pgqueries.ReadIOTime, "read {{backend_type}}"),
			panels.PromQuery(pgqueries.WriteIOTime, "write {{backend_type}}"),
		)).
		Panel("buffers", panels.TimeSeries("Shared buffers: hits vs evictions vs reuses /s",
			panels.PromQuery(pgqueries.BufferHits, "hits"),
			panels.PromQuery(pgqueries.BufferEvictions, "evictions"),
			panels.PromQuery(pgqueries.BufferReuses, "reuses"),
		)).
		Panel("extends", panels.TimeSeriesExpr("Relation extends /s (growth)", pgqueries.RelationExtends, "{{backend_type}}")).
		Panel("fsyncs", panels.TimeSeriesExpr("Fsyncs /s by backend", pgqueries.FsyncsByBackend, "{{backend_type}}")).
		Panel("fsync-time", panels.TimeSeriesExpr("Fsync time (ms spent /s)", pgqueries.FsyncTime, "{{backend_type}}")).
		Panel("wait-class", panels.TimeSeriesExpr("Active backends by wait class (sampled)", pgqueries.ActiveBackendsByWaitClass, "{{wait_event_type}}")).
		Panel("backends-waiting", panels.TimeSeriesExpr("Backends waiting (built-in)", pgqueries.BackendsWaiting, "waiting"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Disk IO by backend (pg_stat_io)",
			panels.GridItem("read-ops", 0, 0, 12, 8),
			panels.GridItem("write-ops", 12, 0, 12, 8),
			panels.GridItem("throughput", 0, 8, 12, 8),
			panels.GridItem("io-time", 12, 8, 12, 8),
		),
		panels.Row("Buffers & durability",
			panels.GridItem("buffers", 0, 0, 12, 8),
			panels.GridItem("extends", 12, 0, 12, 8),
			panels.GridItem("fsyncs", 0, 8, 12, 8),
			panels.GridItem("fsync-time", 12, 8, 12, 8),
		),
		panels.Row("Waits (sampled from pg_stat_activity)",
			panels.GridItem("wait-class", 0, 0, 12, 8),
			panels.GridItem("backends-waiting", 12, 0, 12, 8),
		),
	))
}
