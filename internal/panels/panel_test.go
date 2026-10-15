package panels

import (
	"errors"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

type failingPanel struct{}

func (failingPanel) Build() (dashboardv2.PanelKind, error) {
	return dashboardv2.PanelKind{}, errors.New("boom")
}

func TestDescribeSetsThePanelDescription(t *testing.T) {
	p, err := Describe("empty until set", Text("t", "body")).Build()
	if err != nil {
		t.Fatal(err)
	}
	if p.Spec.Description != "empty until set" || p.Spec.Title != "t" {
		t.Errorf("description=%q title=%q", p.Spec.Description, p.Spec.Title)
	}
	if _, err := Describe("x", failingPanel{}).Build(); err == nil {
		t.Fatal("expected the inner build error to propagate")
	}
}
