package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

// Series renders a timeseries panel with an explicit unit and one or more targets.
func Series(title, unit string, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	viz := timeseries.NewVisualizationV2Builder()
	if unit != "" {
		viz = viz.Unit(unit)
	}
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(viz)
}

// SeriesExpr renders a single-target timeseries panel with an explicit unit.
func SeriesExpr(title, unit, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	return Series(title, unit, PromQuery(expr, legend))
}

// SeriesQuantiles renders one histogram as three lines, p50 / p95 / p99. The
// legend prefix carries the grouping label ("{{op}} "), empty when ungrouped.
func SeriesQuantiles(title, unit, legend, p50, p95, p99 string) cog.Builder[dashboardv2.PanelKind] {
	return Series(title, unit,
		Query("A", p50, legend+"p50"),
		Query("B", p95, legend+"p95"),
		Query("C", p99, legend+"p99"),
	)
}

// SeriesThresholdLine renders a timeseries panel that draws a threshold line at
// the supplied steps. It is used when a warning boundary must be visible on the
// graph itself.
func SeriesThresholdLine(title, unit string, steps []Threshold, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	viz := timeseries.NewVisualizationV2Builder()
	if unit != "" {
		viz = viz.Unit(unit)
	}
	viz = viz.
		Thresholds(thresholds(steps)).
		ThresholdsStyle(common.NewGraphThresholdsStyleConfigBuilder().Mode(common.GraphThresholdsStyleModeLine))
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(viz)
}
