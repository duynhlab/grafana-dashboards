package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/statetimeline"
)

// StateTimeline renders discrete state over time as coloured bands. Use it for
// a signal that is a state rather than a measurement - a scaler being active, a
// probe passing - where a line chart of 0 and 1 reads worse than a band.
func StateTimeline(title string, steps []Threshold, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	viz := statetimeline.NewVisualizationV2Builder().
		ShowValue(common.VisibilityModeNever).
		MergeValues(true)
	if len(steps) > 0 {
		viz = viz.Thresholds(thresholds(steps))
	}
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(viz)
}
