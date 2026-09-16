package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/gauge"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

func Stat(title, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(PromQuery(expr, legend))).
		Visualization(stat.NewVisualizationV2Builder())
}

func Gauge(title, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(PromQuery(expr, legend))).
		Visualization(gauge.NewVisualizationV2Builder().Unit("percentunit"))
}

func TimeSeries(title string, queries ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.PanelKind] {
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(queryGroup(queries...)).
		Visualization(timeseries.NewVisualizationV2Builder())
}

func TimeSeriesExpr(title, expr, legend string) cog.Builder[dashboardv2.PanelKind] {
	return TimeSeries(title, PromQuery(expr, legend))
}

func GridItem(name string, x, y, w, h int64) cog.Builder[dashboardv2.GridLayoutItemKind] {
	return dashboardv2.GridItem(name).X(x).Y(y).Width(w).Height(h)
}

func Row(title string, items ...cog.Builder[dashboardv2.GridLayoutItemKind]) cog.Builder[dashboardv2.RowsLayoutRowKind] {
	grid := dashboardv2.Grid()
	for _, item := range items {
		grid = grid.Item(item)
	}
	return dashboardv2.Row(title).GridLayout(grid)
}
