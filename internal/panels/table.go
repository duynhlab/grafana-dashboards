package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/table"
)

// Column describes a per-field override applied to a table panel.
type Column struct {
	// Name is the field name after any organize/rename transformation.
	Name string
	// Unit is an optional Grafana unit applied to the column.
	Unit string
	// Gauge renders the cell as a gradient gauge when true.
	Gauge bool
	// Width, when non-nil, pins the column width in pixels.
	Width *float64
}

// Table renders a table panel from instant targets, data transformations and
// per-column overrides. It is used for the snapshot tables in the PostgreSQL
// dashboards (top tables, unused indexes, statement detail, active vacuums).
func Table(title string, queries []cog.Builder[dashboardv2.PanelQueryKind], transforms []cog.Builder[dashboardv2.TransformationKind], columns ...Column) cog.Builder[dashboardv2.PanelKind] {
	group := dashboardv2.NewQueryGroupBuilder()
	for _, q := range queries {
		group = group.Target(q)
	}
	for _, tr := range transforms {
		group = group.Transformation(tr)
	}

	viz := table.NewVisualizationV2Builder()
	for _, c := range columns {
		var props []dashboardv2.DynamicConfigValue
		if c.Unit != "" {
			props = append(props, dashboardv2.DynamicConfigValue{Id: "unit", Value: c.Unit})
		}
		if c.Gauge {
			props = append(props, dashboardv2.DynamicConfigValue{
				Id:    "custom.cellOptions",
				Value: map[string]any{"type": "gauge", "mode": "gradient"},
			})
		}
		if c.Width != nil {
			props = append(props, dashboardv2.DynamicConfigValue{Id: "custom.width", Value: *c.Width})
		}
		if len(props) > 0 {
			viz = viz.OverrideByName(c.Name, props)
		}
	}

	return dashboardv2.NewPanelBuilder().
		Title(title).
		Data(group).
		Visualization(viz)
}

// Merge builds a merge transformation that combines table frames sharing labels.
func Merge() cog.Builder[dashboardv2.TransformationKind] {
	return dashboardv2.NewTransformationBuilder().
		Group("merge").
		Options(map[string]any{})
}

// Organize builds an organize transformation that hides and renames fields.
func Organize(exclude map[string]bool, rename map[string]string) cog.Builder[dashboardv2.TransformationKind] {
	return dashboardv2.NewTransformationBuilder().
		Group("organize").
		Options(map[string]any{
			"excludeByName": exclude,
			"renameByName":  rename,
		})
}

// SortBy builds a sortBy transformation on a single field.
func SortBy(field string, desc bool) cog.Builder[dashboardv2.TransformationKind] {
	return dashboardv2.NewTransformationBuilder().
		Group("sortBy").
		Options(map[string]any{
			"fields": map[string]any{},
			"sort":   []map[string]any{{"field": field, "desc": desc}},
		})
}
