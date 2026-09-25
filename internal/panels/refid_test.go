package panels

import (
	"errors"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

func TestQueryGroupGivesEveryTargetADistinctRefID(t *testing.T) {
	group, err := queryGroup(
		PromQuery("up", ""),   // A
		PromQuery("up", ""),   // A again → B
		Query("B", "up", ""),  // explicit B taken → C
		QueryTable("Z", "up"), // explicit, free → kept
		PromQuery("up", ""),   // A taken → D
	).Build()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, q := range group.Spec.Queries {
		got = append(got, q.Spec.RefId)
	}
	want := []string{"A", "B", "C", "Z", "D"}
	if len(got) != len(want) {
		t.Fatalf("refIds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("refIds = %v, want %v", got, want)
		}
	}
}

type failingTarget struct{}

func (failingTarget) Build() (dashboardv2.PanelQueryKind, error) {
	return dashboardv2.PanelQueryKind{}, errors.New("boom")
}

func TestUniqueRefIDPropagatesBuildErrors(t *testing.T) {
	_, err := uniqueRefID{target: failingTarget{}, used: map[string]bool{}}.Build()
	if err == nil {
		t.Fatal("want the target's build error, got nil")
	}
}
