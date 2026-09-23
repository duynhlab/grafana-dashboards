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
