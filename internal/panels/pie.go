package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/piechart"
)

// Pie renders a pie chart with a table legend showing both the raw value and
// the share of the total. It suits categorical breakdowns such as a status-code
// distribution, where the proportions matter more than the time dimension.
func Pie(title string, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	viz := piechart.NewVisualizationV2Builder().
		PieType(piechart.PieChartTypePie).
		Legend(piechart.NewPieChartLegendOptionsBuilder().
			DisplayMode(common.LegendDisplayModeTable).
			Placement(common.LegendPlacementRight).
			Values([]piechart.PieChartLegendValues{
				piechart.PieChartLegendValuesValue,
				piechart.PieChartLegendValuesPercent,
			}))

	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(viz)
}
