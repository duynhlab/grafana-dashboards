package panels

import (
	"errors"
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// A Prometheus variable query is not PromQL and it does not belong in `expr`.
// Put it in `expr` and the datasource answers `422: unsupported function
// "label_values"`; leave it only in `definition` and nothing reaches the
// datasource at all, so the picker renders with an empty dropdown.
//
// Dashboard v1 accepted `query` as a bare string and parsed `label_values()`
// itself. QueryVariableSpec.Query in v2 is a DataQueryKind with no string
// variant, so the classic form cannot be expressed and the datasource's own
// variable-query object has to be built instead: qryType 1 is "Label values"
// and refId is the constant Grafana's Prometheus variable editor writes.
const (
	promVariableLabelValues = 1
	promVariableRefID       = "PrometheusVariableQueryEditor-VariableQuery"
)

// variableQuery renders the datasource variable-query object for a
// label_values() expression. The SDK exposes no builder for DataQueryKind, so
// this implements cog.Builder itself.
type variableQuery struct {
	expr string
}

func (q variableQuery) Build() (dashboardv2.DataQueryKind, error) {
	if strings.TrimSpace(q.expr) == "" {
		return dashboardv2.DataQueryKind{}, errors.New("variable query: expression is empty")
	}

	datasource, err := standards.PrometheusDatasource().Build()
	if err != nil {
		return dashboardv2.DataQueryKind{}, err
	}

	return dashboardv2.DataQueryKind{
		Kind:       "DataQuery",
		Group:      standards.PrometheusPlugin,
		Version:    "v0",
		Datasource: &datasource,
		Spec: map[string]any{
			"qryType": promVariableLabelValues,
			"query":   q.expr,
			"refId":   promVariableRefID,
		},
	}, nil
}

// allSelected is the value a multi-value variable must carry before a reader
// picks anything. Dashboard v2 serialises `current` with no omitempty, so an
// unset one becomes {"text":"","value":""} and Grafana reads that as a real
// selection: `$var` expands to "" and every `label=~"$var"` matcher returns
// nothing.
func allSelected() dashboardv2.VariableOption {
	text := "All"
	value := "$__all"
	return dashboardv2.VariableOption{
		Text:  dashboardv2.StringOrArrayOfString{String: &text},
		Value: dashboardv2.StringOrArrayOfString{String: &value},
	}
}

// QueryVar builds a multi-value, include-all query variable driven by a
// label_values definition. Chained variables encode their dependency inside the
// definition (for example datname referencing $cluster).
//
// AllowCustomValue stays false on purpose: a reader who types a value that does
// not exist gets a silently empty panel rather than an error.
func QueryVar(name, label, definition string) *dashboardv2.QueryVariableBuilder {
	return dashboardv2.NewQueryVariableBuilder(name).
		Label(label).
		Definition(definition).
		Query(variableQuery{expr: definition}).
		Current(allSelected()).
		AllowCustomValue(false).
		Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
		IncludeAll(true).
		Multi(true).
		Sort(dashboardv2.VariableSortAlphabeticalAsc)
}

var _ cog.Builder[dashboardv2.DataQueryKind] = variableQuery{}

// SingleQueryVar builds a single-select query variable. Use it when the queries
// match the variable exactly (`label="$var"`), where include-all cannot work
// because $__all is not a label value.
//
// current is deliberately left unset here: with options loading correctly there
// is no All to preselect, and Grafana picks the first option on dashboard load.
// The include-all case is different and QueryVar handles it.
func SingleQueryVar(name, label, definition string) *dashboardv2.QueryVariableBuilder {
	return dashboardv2.NewQueryVariableBuilder(name).
		Label(label).
		Definition(definition).
		Query(variableQuery{expr: definition}).
		AllowCustomValue(false).
		Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
		Sort(dashboardv2.VariableSortAlphabeticalAsc)
}
