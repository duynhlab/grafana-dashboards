package dashboards_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/duynhlab/grafana-dashboards/internal/registry"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func TestAlertRegistry(t *testing.T) {
	dashUIDs := map[string]struct{}{}
	for _, d := range registry.Dashboards {
		dashUIDs[d.UID] = struct{}{}
	}

	seen := map[string]struct{}{}
	if len(registry.Alerts) == 0 {
		t.Fatal("expected seed alerts")
	}

	for _, a := range registry.Alerts {
		if a.UID == "" || a.Title == "" || a.Folder == "" || a.Expr == "" || a.For == "" {
			t.Fatalf("incomplete alert: %+v", a)
		}
		if _, dup := seen[a.UID]; dup {
			t.Fatalf("duplicate alert UID %s", a.UID)
		}
		seen[a.UID] = struct{}{}

		for _, key := range []string{standards.LabelSeverity, standards.LabelDomain, standards.LabelComponent} {
			if a.Labels[key] == "" {
				t.Fatalf("%s missing label %s", a.UID, key)
			}
		}
		for _, key := range []string{standards.AnnotationSummary, standards.AnnotationDescription, standards.AnnotationDashboardUID} {
			if a.Annotations[key] == "" {
				t.Fatalf("%s missing annotation %s", a.UID, key)
			}
		}
		dashUID := a.Annotations[standards.AnnotationDashboardUID]
		if _, ok := dashUIDs[dashUID]; !ok {
			t.Fatalf("%s dashboard_uid %s is not in dashboard registry", a.UID, dashUID)
		}
	}
}

func TestDomainFolders(t *testing.T) {
	want := map[string]struct{}{
		standards.FolderKubernetes:    {},
		standards.FolderDatabases:     {},
		standards.FolderObservability: {},
		standards.FolderMicroservices: {},
		standards.FolderPlatform:      {},
		standards.FolderGateway:       {},
	}
	for _, folder := range registry.Folders() {
		delete(want, folder)
	}
	if len(want) != 0 {
		t.Fatalf("missing folders: %v", want)
	}

	for _, d := range registry.Dashboards {
		if d.Folder == "as-code" {
			t.Fatalf("dashboard %s still uses catch-all folder as-code", d.UID)
		}
	}
}

// Alert queries are evaluated outside any dashboard, so a template variable
// such as $cluster stays literal and matches nothing: the rule sits in NoData.
// Only Grafana's own $__ macros are allowed.
func TestAlertExpressionsHaveNoDashboardVariables(t *testing.T) {
	variable := regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)`)
	for _, a := range registry.Alerts {
		for _, m := range variable.FindAllStringSubmatch(a.Expr, -1) {
			if !strings.HasPrefix(m[1], "__") {
				t.Errorf("%s: expression uses dashboard variable $%s: %s", a.UID, m[1], a.Expr)
			}
		}
	}
}

// count() over an empty vector returns no series, which Grafana evaluates as
// NoData. Every count() rule must fall back to 0 so a healthy cluster is Normal.
func TestCountAlertsFallBackToZero(t *testing.T) {
	for _, a := range registry.Alerts {
		if strings.HasPrefix(strings.TrimSpace(a.Expr), "count(") && !strings.HasSuffix(a.Expr, "or vector(0)") {
			t.Errorf("%s: count() expression without `or vector(0)`: %s", a.UID, a.Expr)
		}
	}
}
