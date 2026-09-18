package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

func RowsLayout(rows ...cog.Builder[dashboardv2.RowsLayoutRowKind]) cog.Builder[dashboardv2.RowsLayoutKind] {
	b := dashboardv2.Rows()
	for _, row := range rows {
		b = b.Row(row)
	}
	return b
}

// CollapsedRow is Row with the section folded on load. Boards that carry one
// section per service domain use it so the page opens on the sections that
// matter and the rest stay one click away.
func CollapsedRow(title string, items ...cog.Builder[dashboardv2.GridLayoutItemKind]) cog.Builder[dashboardv2.RowsLayoutRowKind] {
	grid := dashboardv2.Grid()
	for _, item := range items {
		grid = grid.Item(item)
	}
	return dashboardv2.Row(title).GridLayout(grid).Collapse(true)
}
