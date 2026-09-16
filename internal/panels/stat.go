package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
)

// StatValue renders a single-value stat panel with an explicit unit and optional
// threshold coloring. Values reduce to the last non-null sample and color the
// displayed value, matching the source dashboards.
func StatValue(title, unit string, steps []Threshold, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	viz := stat.NewVisualizationV2Builder().
		ColorMode(common.BigValueColorModeValue).
		ReduceOptions(common.NewReduceDataOptionsBuilder().Calcs([]string{"lastNotNull"}))
	if unit != "" {
		viz = viz.Unit(unit)
	}
	if len(steps) > 0 {
		viz = viz.Thresholds(thresholds(steps))
	}
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(PromQuery(expr, legend))).
		Visualization(viz)
}
