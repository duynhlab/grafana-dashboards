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
