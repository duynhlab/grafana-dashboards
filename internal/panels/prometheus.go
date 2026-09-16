package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func promQueryBuilder(expr, legend string) cog.Builder[dashboardv2.DataQueryKind] {
	b := prometheus.NewQueryV2Builder().
		Expr(expr).
		Datasource(standards.PrometheusDatasource())
	if legend != "" {
		b = b.LegendFormat(legend)
	}
	return b
}

func PromQuery(expr, legend string) cog.Builder[dashboardv2.PanelQueryKind] {
	return dashboardv2.NewTargetBuilder().RefId("A").Query(promQueryBuilder(expr, legend))
}

func queryGroup(targets ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.QueryGroupKind] {
	b := dashboardv2.NewQueryGroupBuilder()
	for _, t := range targets {
		b = b.Target(t)
	}
	return b
}
