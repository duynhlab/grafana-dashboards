package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/platform"
)

func TestKeycloakIdentity(t *testing.T) {
	dash, err := platform.KeycloakIdentity().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Keycloak — Identity" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 18 {
		t.Fatalf("expected 18 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("keycloak-identity", platform.KeycloakIdentity()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "keycloak-identity" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)

	for _, needle := range []string{
		"keycloak_user_events_total",
		"http_server_requests_seconds_bucket",
		"agroal_available_count",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	for _, absent := range []string{
		"DS_PROMETHEUS",
		"[5m]",
	} {
		if strings.Contains(s, absent) {
			t.Fatalf("unexpected fragment present: %s", absent)
		}
	}
}
