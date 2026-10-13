package panels

import (
	"errors"
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// ClickHouse targets. The Foundation SDK carries no typed builder for the
// grafana-clickhouse-datasource plugin, so the DataQueryKind is assembled by
// hand the way variableQuery is for Prometheus variables. The spec is what the
// plugin's SQL editor (v4) writes: editorType sql, rawSql, a queryType and a
// format. The SQL is sent as written — the plugin expands its own macros
// ($__timeFilter, $__timeInterval, $__fromTime, $__toTime, $__interval_s) and
// Grafana expands template variables such as ${db_name:singlequote}.
const (
	clickhouseQueryTypeTimeSeries = "timeseries"
	clickhouseQueryTypeTable      = "table"
	// format is the plugin's result shape: 0 = time series (one series per
	// non-time column, or per string column value), 1 = table.
	clickhouseFormatTimeSeries = 0
	clickhouseFormatTable      = 1
	clickhouseVariableRefID    = "A"
)

// clickhouseQuery renders a ClickHouse SQL DataQueryKind. refID is set only on
// variable queries, where the plugin expects it inside the spec; panel targets
// carry theirs on the PanelQuery.
type clickhouseQuery struct {
	sql       string
	queryType string
	format    int
	refID     string
}

func (q clickhouseQuery) Build() (dashboardv2.DataQueryKind, error) {
	if strings.TrimSpace(q.sql) == "" {
		return dashboardv2.DataQueryKind{}, errors.New("clickhouse query: sql is empty")
	}

	datasource, err := standards.ClickHouseDatasource().Build()
	if err != nil {
		return dashboardv2.DataQueryKind{}, err
	}

	spec := map[string]any{
		"editorType": "sql",
		"rawSql":     q.sql,
		"queryType":  q.queryType,
		"format":     q.format,
	}
	if q.refID != "" {
		spec["refId"] = q.refID
	}

	return dashboardv2.DataQueryKind{
		Kind:       "DataQuery",
		Group:      standards.ClickHousePlugin,
		Version:    "v0",
		Datasource: &datasource,
		Spec:       spec,
	}, nil
}

var _ cog.Builder[dashboardv2.DataQueryKind] = clickhouseQuery{}

// ClickHouseSeries builds a time-series target from SQL whose first column is
// the time bucket (`$__timeInterval(ts) AS time`). Every other numeric column
// is a series; a string column splits the rows into one series per value.
func ClickHouseSeries(refID, sql string) cog.Builder[dashboardv2.PanelQueryKind] {
	return dashboardv2.NewTargetBuilder().RefId(refID).Query(clickhouseQuery{
		sql:       sql,
		queryType: clickhouseQueryTypeTimeSeries,
		format:    clickhouseFormatTimeSeries,
	})
}

// ClickHouseTable builds a table target: rows as returned. A stat panel fed
// with a one-row table reduces it like any other frame, which is how the
// range-wide figures (p95 over the whole window, error share) are produced
// without a partial last bucket skewing them.
func ClickHouseTable(refID, sql string) cog.Builder[dashboardv2.PanelQueryKind] {
	return dashboardv2.NewTargetBuilder().RefId(refID).Query(clickhouseQuery{
		sql:       sql,
		queryType: clickhouseQueryTypeTable,
		format:    clickhouseFormatTable,
	})
}

// ClickHouseQueryVar builds a multi-value, include-all variable from a SQL
// statement returning one column of values (`SELECT DISTINCT db_name ...`).
// Same defaults as QueryVar: no custom values, refreshed on load, sorted.
func ClickHouseQueryVar(name, label, sql string) *dashboardv2.QueryVariableBuilder {
	return dashboardv2.NewQueryVariableBuilder(name).
		Label(label).
		Definition(sql).
		Query(clickhouseQuery{
			sql:       sql,
			queryType: clickhouseQueryTypeTable,
			format:    clickhouseFormatTable,
			refID:     clickhouseVariableRefID,
		}).
		Current(allSelected()).
		AllowCustomValue(false).
		Refresh(dashboardv2.VariableRefreshOnDashboardLoad).
		IncludeAll(true).
		Multi(true).
		Sort(dashboardv2.VariableSortAlphabeticalAsc)
}
