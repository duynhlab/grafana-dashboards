package panels

import "testing"

const labelValues = `label_values(cnpg_collector_up, cnpg_io_cluster)`

func TestQueryVarCarriesADatasourceQuery(t *testing.T) {
	variable, err := QueryVar("cluster", "Cluster", labelValues).Build()
	if err != nil {
		t.Fatal(err)
	}

	spec, ok := variable.Spec.Query.Spec.(map[string]any)
	if !ok {
		t.Fatalf("query.spec is %T, want the datasource variable-query object", variable.Spec.Query.Spec)
	}
	if spec["query"] != labelValues {
		t.Errorf("query.spec.query = %v, want the label_values expression", spec["query"])
	}
	if spec["qryType"] != promVariableLabelValues {
		t.Errorf("query.spec.qryType = %v, want %d", spec["qryType"], promVariableLabelValues)
	}
	if spec["refId"] != promVariableRefID {
		t.Errorf("query.spec.refId = %v, want %s", spec["refId"], promVariableRefID)
	}
	if variable.Spec.Query.Group == "" || variable.Spec.Query.Datasource == nil {
		t.Errorf("query is routed to no datasource: group=%q datasource=%v",
			variable.Spec.Query.Group, variable.Spec.Query.Datasource)
	}

	// An unset current serialises as an empty string, which Grafana reads as a
	// real selection, so $cluster would expand to "".
	if got := variable.Spec.Current.Value.String; got == nil || *got != "$__all" {
		t.Errorf("current.value = %v, want $__all", got)
	}
	if variable.Spec.AllowCustomValue {
		t.Error("allowCustomValue is true; a typed value that does not exist renders an empty panel")
	}
}

func TestSingleQueryVarIsSingleSelect(t *testing.T) {
	variable, err := SingleQueryVar("ins", "Instance", labelValues).Build()
	if err != nil {
		t.Fatal(err)
	}
	if variable.Spec.Query.Spec == nil {
		t.Fatal("query.spec is nil")
	}
	// ins="$ins" matches exactly, so $__all is not a value the query can use.
	if variable.Spec.IncludeAll || variable.Spec.Multi {
		t.Errorf("includeAll=%v multi=%v, want both false", variable.Spec.IncludeAll, variable.Spec.Multi)
	}
}

func TestVariableQueryRejectsAnEmptyExpression(t *testing.T) {
	if _, err := (variableQuery{expr: "  "}).Build(); err == nil {
		t.Fatal("expected an empty variable expression to be rejected")
	}
}

func TestTextVarStartsEmpty(t *testing.T) {
	variable, err := TextVar("query_id", "query_id").Build()
	if err != nil {
		t.Fatal(err)
	}
	if variable.Spec.Name != "query_id" || variable.Spec.Label == nil || *variable.Spec.Label != "query_id" {
		t.Errorf("name=%q label=%v", variable.Spec.Name, variable.Spec.Label)
	}
	if got := variable.Spec.Current.Value.String; got == nil || *got != "" {
		t.Errorf("current.value = %v, want empty", got)
	}
}

func TestCustomVarPreselectsTheFirstChoice(t *testing.T) {
	variable, err := CustomVar("scope", "Sessions",
		Choice{Text: "Workload only", Value: "0"},
		Choice{Text: "Include platform sessions", Value: "1"},
	).Build()
	if err != nil {
		t.Fatal(err)
	}
	if variable.Spec.Query != "Workload only : 0, Include platform sessions : 1" {
		t.Errorf("query = %q", variable.Spec.Query)
	}
	if len(variable.Spec.Options) != 2 {
		t.Fatalf("options = %d, want 2", len(variable.Spec.Options))
	}
	if got := variable.Spec.Current.Value.String; got == nil || *got != "0" {
		t.Errorf("current.value = %v, want 0", got)
	}
	if variable.Spec.Multi || variable.Spec.IncludeAll || variable.Spec.AllowCustomValue {
		t.Error("custom variable must be single-select with no custom values")
	}

	empty, err := CustomVar("none", "None").Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(empty.Spec.Options) != 0 {
		t.Errorf("options = %d, want 0", len(empty.Spec.Options))
	}
}
