package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/heatmap"
)

// Heatmap renders a bucket heatmap from a query that already returns histogram
// buckets keyed by le. Calculation is disabled because the series arrive
// pre-bucketed; letting Grafana re-bucket them would distort the distribution.
func Heatmap(title, yUnit, scheme string, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	viz := heatmap.NewVisualizationV2Builder().
		Calculate(false).
		CellGap(2).
		Color(heatmap.NewHeatmapColorOptionsBuilder().
			Mode(heatmap.HeatmapColorModeScheme).
			Scheme(scheme).
			// Scheme mode requires at least two colour steps; the schema
			// rejects the zero value. 64 is Grafana's own default.
			Steps(64)).
		Legend(heatmap.NewHeatmapLegendBuilder().Show(true))
	if yUnit != "" {
		viz = viz.YAxis(heatmap.NewYAxisConfigBuilder().Unit(yUnit))
	}

	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(viz)
}
