package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func promQueryBuilder(expr, legend string) cog.Builder[dashboardv2.DataQueryKind] {
	b := prometheus.NewQueryV2Builder().
		Expr(expr).
		Datasource(standards.PrometheusDatasource())
	if legend != "" {
		b = b.LegendFormat(legend)
	}
	return b
}

func PromQuery(expr, legend string) cog.Builder[dashboardv2.PanelQueryKind] {
	return dashboardv2.NewTargetBuilder().RefId("A").Query(promQueryBuilder(expr, legend))
}

// queryGroup collects a panel's targets and gives each a distinct refId.
// PromQuery always writes "A", so a panel built from several of them would
// send duplicates, which Grafana rejects ("Multiple queries using the same
// RefId is not allowed"). An explicit refId from Query/QueryTable/QueryHeatmap
// is kept; a duplicate is moved to the next free letter, in target order.
func queryGroup(targets ...cog.Builder[dashboardv2.PanelQueryKind]) cog.Builder[dashboardv2.QueryGroupKind] {
	b := dashboardv2.NewQueryGroupBuilder()
	used := map[string]bool{}
	for _, t := range targets {
		b = b.Target(uniqueRefID{target: t, used: used})
	}
	return b
}

// uniqueRefID renames a target's refId when an earlier target in the same
// panel already holds it. The query group builder builds targets in order, so
// the shared set sees them in the order the panel lists them.
type uniqueRefID struct {
	target cog.Builder[dashboardv2.PanelQueryKind]
	used   map[string]bool
}

func (u uniqueRefID) Build() (dashboardv2.PanelQueryKind, error) {
	q, err := u.target.Build()
	if err != nil {
		return q, err
	}
	if u.used[q.Spec.RefId] {
		for c := 'A'; c <= 'Z'; c++ {
			if !u.used[string(c)] {
				q.Spec.RefId = string(c)
				break
			}
		}
	}
	u.used[q.Spec.RefId] = true
	return q, nil
}
