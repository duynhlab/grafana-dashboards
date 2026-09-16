package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

// QueryVar builds a multi-value, include-all query variable driven by a
// label_values definition. Chained variables encode their dependency inside the
// definition (for example datname referencing $cluster).
func QueryVar(name, label, definition string) *dashboardv2.QueryVariableBuilder {
	return dashboardv2.NewQueryVariableBuilder(name).
		Label(label).
		Definition(definition).
		Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
		IncludeAll(true).
		Multi(true).
		Sort(dashboardv2.VariableSortAlphabeticalAsc)
}
