package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// Query builds a range Prometheus target with an explicit ref id. Distinct ref
// ids are required when a panel carries more than one target.
func Query(refID, expr, legend string) cog.Builder[dashboardv2.PanelQueryKind] {
	q := prometheus.NewQueryV2Builder().
		Expr(expr).
		Range(true).
		Datasource(standards.PrometheusDatasource())
	if legend != "" {
		q = q.LegendFormat(legend)
	}
	return dashboardv2.NewTargetBuilder().RefId(refID).Query(q)
}

// QueryHeatmap builds a range Prometheus target formatted as heatmap buckets. A
// heatmap panel fed with pre-bucketed histogram series needs this format so that
// Grafana reads the le label as the bucket boundary.
func QueryHeatmap(refID, expr, legend string) cog.Builder[dashboardv2.PanelQueryKind] {
	q := prometheus.NewQueryV2Builder().
		Expr(expr).
		Range(true).
		Format(prometheus.PromQueryFormatHeatmap).
		Datasource(standards.PrometheusDatasource())
	if legend != "" {
		q = q.LegendFormat(legend)
	}
	return dashboardv2.NewTargetBuilder().RefId(refID).Query(q)
}

// QueryTable builds an instant Prometheus target formatted as a table. It is used
// by table panels that snapshot current values across labels.
func QueryTable(refID, expr string) cog.Builder[dashboardv2.PanelQueryKind] {
	q := prometheus.NewQueryV2Builder().
		Expr(expr).
		Instant(true).
		Format(prometheus.PromQueryFormatTable).
		Datasource(standards.PrometheusDatasource())
	return dashboardv2.NewTargetBuilder().RefId(refID).Query(q)
}
