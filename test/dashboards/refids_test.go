package dashboards_test

import (
	"encoding/json"
	"testing"

	"github.com/duynhlab/grafana-dashboards/internal/registry"
)

// TestPanelQueryRefIDsAreUnique holds the defect that broke pg-io-waits and
// kubernetes-cluster-overview on Grafana 13: panels.PromQuery always writes
// refId "A", so a panel built from two or three of them sends duplicates and
// Grafana answers "Multiple queries using the same RefId is not allowed".
// It runs over the whole registry so a new board cannot reintroduce it.
func TestPanelQueryRefIDsAreUnique(t *testing.T) {
	for _, definition := range registry.Dashboards {
		dash, err := definition.Build().Build()
		if err != nil {
			t.Fatalf("%s: build: %v", definition.UID, err)
		}
		body, err := json.Marshal(dash.Elements)
		if err != nil {
			t.Fatalf("%s: marshal elements: %v", definition.UID, err)
		}
		var elements map[string]struct {
			Spec struct {
				Title string `json:"title"`
				Data  struct {
					Spec struct {
						Queries []struct {
							Spec struct {
								RefID string `json:"refId"`
							} `json:"spec"`
						} `json:"queries"`
					} `json:"spec"`
				} `json:"data"`
			} `json:"spec"`
		}
		if err := json.Unmarshal(body, &elements); err != nil {
			t.Fatalf("%s: unmarshal elements: %v", definition.UID, err)
		}
		for key, el := range elements {
			seen := map[string]bool{}
			for _, q := range el.Spec.Data.Spec.Queries {
				if q.Spec.RefID == "" {
					t.Errorf("%s/%s (%q): empty refId", definition.UID, key, el.Spec.Title)
				}
				if seen[q.Spec.RefID] {
					t.Errorf("%s/%s (%q): refId %q used twice", definition.UID, key, el.Spec.Title, q.Spec.RefID)
				}
				seen[q.Spec.RefID] = true
			}
		}
	}
}
