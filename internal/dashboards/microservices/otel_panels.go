package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/piechart"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
)

// Panel builders for the Microservices (OTel) board. The board is a 1:1 port of
// the helm-charts JSON, whose panels carry display options (table legends on
// the right, fill 10, tooltips over every series) that the shared helpers in
// internal/panels deliberately do not set. Changing those helpers would restyle
// every other board, so the source's look lives here instead.

// otelPanel wraps a visualization and its targets into a described panel. Every
// caller passes explicit, distinct ref ids (PromQuery is A, Query is B onward).
func otelPanel(title, description string, viz cog.Builder[dashboardv2.VizConfigKind], queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	data := dashboardv2.NewQueryGroupBuilder()
	for _, q := range queries {
		data = data.Target(q)
	}
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Description(description).
		Data(data).
		Visualization(viz)
}

func otelThresholds(steps []panels.Threshold) cog.Builder[dashboardv2.ThresholdsConfig] {
	out := make([]dashboardv2.Threshold, 0, len(steps))
	for _, s := range steps {
		out = append(out, dashboardv2.Threshold{Color: s.Color, Value: s.Value})
	}
	return dashboardv2.NewThresholdsConfigBuilder().
		Mode(dashboardv2.ThresholdsModeAbsolute).
		Steps(out)
}

func otelColor(mode dashboardv2.FieldColorModeId) cog.Builder[dashboardv2.FieldColor] {
	return dashboardv2.NewFieldColorBuilder().Mode(mode)
}

// otelStat is the overview tile: the value coloured by its thresholds over an
// area sparkline, reduced to the last non-null sample.
func otelStat(title, description, unit string, steps []panels.Threshold, expr string) cog.Builder[dashboardv2.PanelKind] {
	viz := stat.NewVisualizationV2Builder().
		Unit(unit).
		ColorScheme(otelColor(dashboardv2.FieldColorModeIdThresholds)).
		Thresholds(otelThresholds(steps)).
		ColorMode(common.BigValueColorModeValue).
		GraphMode(common.BigValueGraphModeArea).
		JustifyMode(common.BigValueJustifyModeAuto).
		Orientation(common.VizOrientationAuto).
		ReduceOptions(common.NewReduceDataOptionsBuilder().
			Calcs([]string{"lastNotNull"}).
			Fields("").
			Values(false))
	return otelPanel(title, description, viz, panels.PromQuery(expr, ""))
}

// otelPie is a categorical breakdown with a table legend showing both the value
// and its share of the total.
func otelPie(title, description, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	viz := piechart.NewVisualizationV2Builder().
		ColorScheme(otelColor(dashboardv2.FieldColorModeIdPaletteClassic)).
		HideFrom(common.NewHideSeriesConfigBuilder().Legend(false).Tooltip(false).Viz(false)).
		PieType(piechart.PieChartTypePie).
		// ShowLegend must be explicit: the SDK serializes an unset value as false,
		// while the source JSON omitted the key and so got Grafana's default true.
		Legend(piechart.NewPieChartLegendOptionsBuilder().
			ShowLegend(true).
			DisplayMode(common.LegendDisplayModeTable).
			Placement(common.LegendPlacementRight).
			Values([]piechart.PieChartLegendValues{
				piechart.PieChartLegendValuesValue,
				piechart.PieChartLegendValuesPercent,
			})).
		ReduceOptions(common.NewReduceDataOptionsBuilder().
			Calcs([]string{"lastNotNull"}).
			Fields("").
			Values(false))
	return otelPanel(title, description, viz, panels.PromQuery(expr, legend))
}

// otelSeriesOptions holds the per-panel departures from the board's timeseries
// defaults. The zero value is the common case.
type otelSeriesOptions struct {
	// LineWidth defaults to 1.
	LineWidth float64
	// Calcs and SortBy shape the table legend; they default to mean, max sorted
	// by Max.
	Calcs  []string
	SortBy string
	// Steps defaults to a single green base step.
	Steps []panels.Threshold
	// ThresholdLine draws Steps on the graph.
	ThresholdLine bool
	// ColorByThresholds colours the series by Steps instead of the classic
	// palette.
	ColorByThresholds bool
	// ServiceLinks links each series to its service's traces and logs; the
	// series must carry a service_name label.
	ServiceLinks bool
}

// otelSeries is the board's timeseries: thin filled lines, no points, gaps left
// as gaps, a tooltip over every series and a table legend on the right sorted
// by its largest value, so the heaviest series reads first.
func otelSeries(title, description, unit string, opts otelSeriesOptions, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	if opts.LineWidth == 0 {
		opts.LineWidth = 1
	}
	if opts.Calcs == nil {
		opts.Calcs = []string{"mean", "max"}
	}
	if opts.SortBy == "" {
		opts.SortBy = "Max"
	}
	if opts.Steps == nil {
		opts.Steps = []panels.Threshold{panels.Thr("green", nil)}
	}
	color := dashboardv2.FieldColorModeIdPaletteClassic
	if opts.ColorByThresholds {
		color = dashboardv2.FieldColorModeIdThresholds
	}
	thresholdsMode := common.GraphThresholdsStyleModeOff
	if opts.ThresholdLine {
		thresholdsMode = common.GraphThresholdsStyleModeLine
	}
	spanNulls := false

	viz := timeseries.NewVisualizationV2Builder().
		Unit(unit).
		ColorScheme(otelColor(color)).
		Thresholds(otelThresholds(opts.Steps)).
		ThresholdsStyle(common.NewGraphThresholdsStyleConfigBuilder().Mode(thresholdsMode)).
		FillOpacity(10).
		LineWidth(opts.LineWidth).
		ShowPoints(common.VisibilityModeNever).
		SpanNulls(common.BoolOrFloat64{Bool: &spanNulls}).
		Tooltip(common.NewVizTooltipOptionsBuilder().
			Mode(common.TooltipDisplayModeMulti).
			Sort(common.SortOrderNone)).
		Legend(common.NewVizLegendOptionsBuilder().
			ShowLegend(true).
			DisplayMode(common.LegendDisplayModeTable).
			Placement(common.LegendPlacementRight).
			Calcs(opts.Calcs).
			SortBy(opts.SortBy).
			SortDesc(true))
	if opts.ServiceLinks {
		viz = viz.DataLinks(otelServiceLinks(seriesService))
	}
	return otelPanel(title, description, viz, queries...)
}

// otelDBViz is the Database row's timeseries, which the source styles apart from
// the rest of the board: a slightly heavier fill, a table legend underneath
// with the current and peak value, and the tooltip sorted largest first. It
// returns the visualization so a panel can add its axis bounds or overrides.
func otelDBViz(unit string) *timeseries.VisualizationV2Builder {
	return timeseries.NewVisualizationV2Builder().
		Unit(unit).
		FillOpacity(12).
		LineWidth(1).
		ShowPoints(common.VisibilityModeNever).
		Tooltip(common.NewVizTooltipOptionsBuilder().
			Mode(common.TooltipDisplayModeMulti).
			Sort(common.SortOrderDescending)).
		Legend(common.NewVizLegendOptionsBuilder().
			ShowLegend(true).
			DisplayMode(common.LegendDisplayModeTable).
			Placement(common.LegendPlacementBottom).
			Calcs([]string{"lastNotNull", "max"}))
}
