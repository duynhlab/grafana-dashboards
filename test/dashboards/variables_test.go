package dashboards_test

import (
	"encoding/json"
	"testing"

	"github.com/duynhlab/grafana-dashboards/internal/registry"
)

// variableSpec mirrors the fields of a rendered v2 QueryVariable that decide
// whether Grafana's apiserver accepts the board and whether the picker works.
// The assertions live on the marshalled JSON on purpose: the defects this test
// exists for were all invisible in the Go builders and only showed up on the
// wire.
type variableSpec struct {
	Kind string `json:"kind"`
	Spec struct {
		Name    string `json:"name"`
		Current struct {
			Value any `json:"value"`
		} `json:"current"`
		Query struct {
			Group string `json:"group"`
			Spec  *struct {
				QryType int    `json:"qryType"`
				Query   string `json:"query"`
				RefID   string `json:"refId"`
			} `json:"spec"`
		} `json:"query"`
		IncludeAll       bool `json:"includeAll"`
		AllowCustomValue bool `json:"allowCustomValue"`
	} `json:"spec"`
}

// TestQueryVariablesAreResolvable holds the three variable defects that made
// Grafana reject every board carrying a variable:
//
//   - a null query.spec, which leaves the label_values() expression in
//     `definition` where no datasource ever sees it;
//   - an empty query.group, which routes the variable query nowhere;
//   - an empty current.value on an include-all variable, which Grafana reads as
//     a real selection so `label=~"$var"` matches nothing.
//
// It runs over the whole registry so a new board cannot reintroduce them.
func TestQueryVariablesAreResolvable(t *testing.T) {
	for _, definition := range registry.Dashboards {
		dash, err := definition.Build().Build()
		if err != nil {
			t.Fatalf("%s: build: %v", definition.UID, err)
		}

		body, err := json.Marshal(dash.Variables)
		if err != nil {
			t.Fatalf("%s: marshal variables: %v", definition.UID, err)
		}

		var variables []variableSpec
		if err := json.Unmarshal(body, &variables); err != nil {
			t.Fatalf("%s: unmarshal variables: %v", definition.UID, err)
		}

		for _, v := range variables {
			if v.Kind != "QueryVariable" {
				continue
			}
			name := definition.UID + "/" + v.Spec.Name

			if v.Spec.Query.Spec == nil {
				t.Errorf("%s: query.spec is null; the label_values() expression never reaches the datasource", name)
				continue
			}
			if v.Spec.Query.Spec.Query == "" {
				t.Errorf("%s: query.spec.query is empty", name)
			}
			if v.Spec.Query.Spec.QryType != 1 {
				t.Errorf("%s: query.spec.qryType = %d, want 1 (label values)", name, v.Spec.Query.Spec.QryType)
			}
			if v.Spec.Query.Spec.RefID == "" {
				t.Errorf("%s: query.spec.refId is empty", name)
			}
			if v.Spec.Query.Group == "" {
				t.Errorf("%s: query.group is empty; the variable query is routed to no datasource", name)
			}
			if v.Spec.AllowCustomValue {
				t.Errorf("%s: allowCustomValue is true; a typed value that does not exist renders an empty panel", name)
			}
			if v.Spec.IncludeAll && v.Spec.Current.Value == "" {
				t.Errorf("%s: include-all variable has an empty current.value; $%s expands to \"\"", name, v.Spec.Name)
			}
		}
	}
}
