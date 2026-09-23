package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

func TestCutoverBaseline(t *testing.T) {
	dash, err := microservices.CutoverBaseline().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Order Saga & Payment — Cutover Baseline" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 8 {
		t.Fatalf("expected 8 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("rfc0021-baseline", microservices.CutoverBaseline()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "rfc0021-baseline" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"rfc0021:checkout_confirm_success:rate5m",
		"rfc0021:order_saga_outcome:rate5m",
		"rfc0021:payment_provider_duration:p95_5m",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, needle := range []string{
		"DS_PROMETHEUS",
		"product_stock_reservations",
		"checkout_shadow_compare",
		"checkout_shadow_divergence",
	} {
		if strings.Contains(s, needle) {
			t.Fatalf("unexpected fragment present (retired panel or legacy datasource var): %s", needle)
		}
	}
}
