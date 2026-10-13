package panels

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func TestClickHouseQueryBuildsPluginSpec(t *testing.T) {
	q, err := clickhouseQuery{sql: "SELECT 1", queryType: clickhouseQueryTypeTable, format: clickhouseFormatTable}.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if q.Kind != "DataQuery" || q.Version != "v0" {
		t.Fatalf("unexpected kind/version: %s/%s", q.Kind, q.Version)
	}
	if q.Group != standards.ClickHousePlugin {
		t.Fatalf("group = %q, want %q", q.Group, standards.ClickHousePlugin)
	}
	if q.Datasource == nil || q.Datasource.Name == nil || *q.Datasource.Name != standards.ClickHouseDatasourceName {
		t.Fatalf("datasource = %+v, want name %q", q.Datasource, standards.ClickHouseDatasourceName)
	}
	spec, ok := q.Spec.(map[string]any)
	if !ok {
		t.Fatalf("spec is %T, want map", q.Spec)
	}
	if spec["rawSql"] != "SELECT 1" || spec["editorType"] != "sql" || spec["queryType"] != "table" || spec["format"] != clickhouseFormatTable {
		t.Fatalf("unexpected spec: %v", spec)
	}
	if _, has := spec["refId"]; has {
		t.Fatalf("panel target spec must not carry refId: %v", spec)
	}
}

func TestClickHouseQueryRejectsEmptySQL(t *testing.T) {
	if _, err := (clickhouseQuery{sql: "  "}).Build(); err == nil {
		t.Fatal("expected an error for empty sql")
	}
}

func TestClickHouseTargetsCarryRefIDAndShape(t *testing.T) {
	series, err := ClickHouseSeries("B", "SELECT $__timeInterval(ts) AS time, count() FROM t GROUP BY time").Build()
	if err != nil {
		t.Fatalf("series: %v", err)
	}
	if series.Spec.RefId != "B" {
		t.Fatalf("series refId = %q, want B", series.Spec.RefId)
	}
	if spec := series.Spec.Query.Spec.(map[string]any); spec["queryType"] != "timeseries" || spec["format"] != clickhouseFormatTimeSeries {
		t.Fatalf("series spec: %v", spec)
	}

	table, err := ClickHouseTable("A", "SELECT 1").Build()
	if err != nil {
		t.Fatalf("table: %v", err)
	}
	if spec := table.Spec.Query.Spec.(map[string]any); spec["queryType"] != "table" || spec["format"] != clickhouseFormatTable {
		t.Fatalf("table spec: %v", spec)
	}
}

func TestClickHouseQueryVar(t *testing.T) {
	v, err := ClickHouseQueryVar("db_name", "Database", "SELECT DISTINCT db_name FROM pg_stat_ch.events_raw").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		`"name":"db_name"`,
		`"group":"grafana-clickhouse-datasource"`,
		`"rawSql":"SELECT DISTINCT db_name FROM pg_stat_ch.events_raw"`,
		`"refId":"A"`,
		`"includeAll":true`,
		`"multi":true`,
		`"value":"$__all"`,
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("variable JSON lacks %s: %s", needle, s)
		}
	}
	if strings.Contains(s, `"allowCustomValue":true`) {
		t.Fatalf("variable must not allow custom values: %s", s)
	}
}
