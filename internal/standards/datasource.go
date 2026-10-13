package standards

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

// PrometheusPlugin is the datasource plugin id that owns every query this
// repository emits. Dashboard v2 records it as the `group` of a DataQueryKind,
// so a query whose group is empty is not routed to any datasource at all.
const PrometheusPlugin = "prometheus"

// PrometheusDatasource returns a logical datasource reference resolved at deploy time.
func PrometheusDatasource() cog.Builder[dashboardv2.Dashboardv2DataQueryKindDatasource] {
	return dashboardv2.NewDashboardv2DataQueryKindDatasourceBuilder().Name(PrometheusPlugin)
}

// ClickHousePlugin is the datasource plugin id for SQL queries against the
// ClickHouse OLAP store (per-execution PostgreSQL events, OTel logs and traces).
// The Foundation SDK has no typed query builder for it; panels.ClickHouse*
// assemble the DataQueryKind by hand with this group.
const ClickHousePlugin = "grafana-clickhouse-datasource"

// ClickHouseDatasourceName is the logical datasource reference. Like
// `prometheus`, it is a name the consumer resolves at deploy time: homelab
// provisions its ClickHouse datasource with uid `clickhouse`.
const ClickHouseDatasourceName = "clickhouse"

// ClickHouseDatasource returns the logical ClickHouse datasource reference.
func ClickHouseDatasource() cog.Builder[dashboardv2.Dashboardv2DataQueryKindDatasource] {
	return dashboardv2.NewDashboardv2DataQueryKindDatasourceBuilder().Name(ClickHouseDatasourceName)
}
