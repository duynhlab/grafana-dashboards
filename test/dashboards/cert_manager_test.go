package dashboards_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/platform"
)

func TestCertManager(t *testing.T) {
	dash, err := platform.CertManager().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "cert-manager" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(dash.Variables))
	}

	if len(dash.Elements) != 16 {
		t.Fatalf("expected 16 panels, got %d", len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("cert-manager", platform.CertManager()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "cert-manager" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, needle := range []string{
		"certmanager_certificate_ready_status",
		"certmanager_controller_sync_call_count",
		"workqueue_depth",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}
	for _, absent := range []string{"DS_PROMETHEUS", "[5m]"} {
		if strings.Contains(s, absent) {
			t.Fatalf("unexpected fragment present: %s", absent)
		}
	}
}
