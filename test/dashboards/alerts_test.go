package dashboards_test

import (
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
