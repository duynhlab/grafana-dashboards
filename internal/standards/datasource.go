package standards

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

// PrometheusDatasource returns a logical datasource reference resolved at deploy time.
func PrometheusDatasource() cog.Builder[dashboardv2.Dashboardv2DataQueryKindDatasource] {
	return dashboardv2.NewDashboardv2DataQueryKindDatasourceBuilder().Name("prometheus")
}
