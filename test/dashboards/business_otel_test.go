package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

func TestBusinessOTel(t *testing.T) {
	dash, err := microservices.BusinessOTel().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Microservices — Business KPIs" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	// The source datasource picker and the custom `rate` window were both
	// dropped on the port, so this board carries no template variables.
	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 38 {
		t.Fatalf("expected 38 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("business-otel", microservices.BusinessOTel()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "business-otel" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"payment_authorization_total",
		"order_saga_outcome_total",
		"order_stock_reservation_total",
		"product_stock_reservations_total",
		"db_client_connections_usage",
		"reviews_rating_bucket",
		"checkout_confirm_duration_seconds_bucket",
		"$__rate_interval",
		"$__range",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	// Guard against the un-normalised rate window from the source board.
	if strings.Contains(s, "$rate") {
		t.Fatalf("un-normalised $rate window left in the board")
	}
}
