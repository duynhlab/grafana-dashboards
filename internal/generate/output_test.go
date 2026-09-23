package generate

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/duynhlab/grafana-dashboards/internal/alerts"
	"github.com/duynhlab/grafana-dashboards/internal/registry"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func TestNewRejectsEmptyRoot(t *testing.T) {
	if _, err := New(" ", nil, nil); err == nil {
		t.Fatal("expected empty root error")
	}
}

func TestGeneratorProducesDeterministicBundleAndRemovesStaleFiles(t *testing.T) {
	root := t.TempDir()
	generator, err := New(root, registry.Dashboards, registry.Alerts)
	if err != nil {
		t.Fatal(err)
	}

	stale := filepath.Join(root, "deploy", "dashboards", "stale.yaml")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := generator.Run(); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale artifact was not removed: %v", err)
	}
	first := snapshot(t, root)

	if err := generator.Run(); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second := snapshot(t, root)
	if !equalSnapshot(first, second) {
		t.Fatal("generator output differs between runs")
	}

	wantFiles := []string{
		"deploy/kustomization.yaml",
		"deploy/folders/kustomization.yaml",
		"deploy/dashboards/kustomization.yaml",
		"deploy/folders/grafanamanifest-kubernetes-folder.yaml",
		"deploy/folders/grafanamanifest-databases-folder.yaml",
		"deploy/folders/grafanamanifest-observability-folder.yaml",
		"deploy/dashboards/grafanamanifest-kubernetes-cluster-overview.yaml",
		"deploy/dashboards/grafanamanifest-pg-io-waits.yaml",
		"deploy/dashboards/grafanamanifest-pg-maintenance.yaml",
		"deploy/dashboards/grafanamanifest-pg-query-performance.yaml",
		"deploy/dashboards/grafanamanifest-pg-exporter-instance.yaml",
		"deploy/dashboards/grafanamanifest-pgdog.yaml",
		"deploy/dashboards/grafanamanifest-temporal-worker.yaml",
		"deploy/dashboards/grafanaalertrulegroup-kubernetes.yaml",
		"deploy/dashboards/grafanaalertrulegroup-databases.yaml",
		"generated/alerts/kubernetes/kubernetes_crashlooping_pods.json",
		"generated/alerts/postgres/postgres_backends_waiting.json",
		"generated/dashboards/kubernetes/kubernetes-cluster-overview.spec.json",
		"generated/dashboards/postgres/pg-io-waits.manifest.json",
		"generated/dashboards/microservices/business-otel.spec.json",
	}
	for _, name := range wantFiles {
		data, ok := second[name]
		if !ok {
			t.Errorf("missing generated file %s", name)
			continue
		}
		if strings.HasSuffix(name, ".yaml") && !bytes.HasPrefix(data, []byte(generatedHeader)) {
			t.Errorf("%s has no generated header", name)
		}
	}

	assertDashboardManifest(t, second["deploy/dashboards/grafanamanifest-kubernetes-cluster-overview.yaml"], "kubernetes-cluster-overview", "kubernetes")
	assertFolderManifest(t, second["deploy/folders/grafanamanifest-databases-folder.yaml"], "databases", "Databases")
	assertAlertGroup(t, second["deploy/dashboards/grafanaalertrulegroup-databases.yaml"], "databases")
}

func TestRunUsesCurrentDirectory(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})

	if err := Run(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "deploy", "kustomization.yaml")); err != nil {
		t.Fatal(err)
	}
}

// A dashboard or alert without a domain would silently collapse its path back
// to the flat layout, so the generator refuses it.
func TestGeneratorRejectsMissingDomain(t *testing.T) {
	dashboard := registry.Dashboards[0]
	dashboard.Domain = ""
	generator, err := New(t.TempDir(), []registry.DashboardDefinition{dashboard}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.Run(); err == nil {
		t.Fatal("expected an error for a dashboard with no domain")
	}

	rule := registry.Alerts[0]
	rule.Domain = ""
	generator, err = New(t.TempDir(), nil, []alerts.Rule{rule})
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.Run(); err == nil {
		t.Fatal("expected an error for an alert with no domain")
	}
}

// prepareOutput removes the whole generated tree, so a file left over inside a
// domain directory must not survive a regeneration either.
func TestGeneratorRemovesStaleDomainFiles(t *testing.T) {
	root := t.TempDir()
	generator, err := New(root, registry.Dashboards, registry.Alerts)
	if err != nil {
		t.Fatal(err)
	}

	stale := filepath.Join(root, "generated", "dashboards", "kubernetes", "retired.spec.json")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := generator.Run(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale nested artifact was not removed: %v", err)
	}
}

func TestGeneratorWithNoResources(t *testing.T) {
	generator, err := New(t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratorReturnsFilesystemError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	if err := os.WriteFile(root, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	generator, err := New(root, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.Run(); err == nil {
		t.Fatal("expected filesystem error")
	}
}

func TestSerializationErrors(t *testing.T) {
	if err := writeJSON(filepath.Join(t.TempDir(), "bad.json"), make(chan int)); err == nil {
		t.Fatal("expected JSON marshal error")
	}
	missing := filepath.Join(t.TempDir(), "missing", "file")
	if err := writeJSON(missing+".json", map[string]string{"ok": "yes"}); err == nil {
		t.Fatal("expected JSON write error")
	}
	if err := writeYAML(missing+".yaml", map[string]string{"ok": "yes"}); err == nil {
		t.Fatal("expected YAML write error")
	}
}

func TestHelpers(t *testing.T) {
	cases := map[string]string{
		" Kubernetes ":    "kubernetes",
		"Cloud Databases": "cloud-databases",
		"A___B":           "a-b",
		"":                "",
	}
	for input, expected := range cases {
		if got := folderSlug(input); got != expected {
			t.Errorf("folderSlug(%q) = %q, want %q", input, got, expected)
		}
	}

	gotFolders := folders(
		[]registry.DashboardDefinition{{Folder: "Kubernetes"}, {Folder: ""}},
		[]alerts.Rule{{Folder: "Kubernetes"}, {Folder: "Databases"}},
	)
	if strings.Join(gotFolders, ",") != "Kubernetes,Databases" {
		t.Fatalf("unexpected folders: %v", gotFolders)
	}

	if metadata("example")["name"] != "example" {
		t.Fatal("metadata name missing")
	}
	if instanceSelector()["matchLabels"] == nil {
		t.Fatal("instance selector missing matchLabels")
	}
}

func TestGrafanaAlertRule(t *testing.T) {
	rule := alerts.Rule{
		UID:       "test_rule",
		Title:     "Test rule",
		Expr:      "up == 0",
		For:       "5m",
		Threshold: 1,
		Labels:    map[string]string{standards.LabelSeverity: standards.SeverityWarning},
	}
	rendered := grafanaAlertRule(rule)
	if rendered["uid"] != rule.UID || rendered["condition"] != "B" {
		t.Fatalf("unexpected rendered rule: %#v", rendered)
	}
	data := rendered["data"].([]map[string]any)
	if len(data) != 2 || data[0]["datasourceUid"] != "prometheus" || data[1]["datasourceUid"] != "__expr__" {
		t.Fatalf("unexpected alert queries: %#v", data)
	}
}

func snapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(name)] = errRead(path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func errRead(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

func equalSnapshot(left, right map[string][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for name, data := range left {
		if !bytes.Equal(data, right[name]) {
			return false
		}
	}
	return true
}

// assertDashboardManifest holds the wire contract that makes a v2 board
// reachable at all. Three of these four fields are silent when wrong: the
// outer namespace is the operator's and the inner one is Grafana's tenant, the
// inner metadata.name is the dashboard UID rather than a display name, and the
// folder annotation takes the folder's UID, which is the folder manifest's
// inner name and not the "-folder" name of the CR that wraps it.
func assertDashboardManifest(t *testing.T, data []byte, uid, folderUID string) {
	t.Helper()
	template := assertManifestCR(t, data, uid)

	if template["apiVersion"] != "dashboard.grafana.app/v2" {
		t.Fatalf("template.apiVersion = %v, want dashboard.grafana.app/v2", template["apiVersion"])
	}
	if template["kind"] != "Dashboard" {
		t.Fatalf("template.kind = %v, want Dashboard", template["kind"])
	}

	meta := template["metadata"].(map[string]any)
	if meta["name"] != uid {
		t.Fatalf("template.metadata.name = %v, want the dashboard UID %s", meta["name"], uid)
	}
	if meta["namespace"] != tenantNamespace {
		t.Fatalf("template.metadata.namespace = %v, want %s", meta["namespace"], tenantNamespace)
	}
	annotations := meta["annotations"].(map[string]any)
	if annotations[folderAnnotation] != folderUID {
		t.Fatalf("%s = %v, want the folder UID %s", folderAnnotation, annotations[folderAnnotation], folderUID)
	}

	spec := template["spec"].(map[string]any)
	if spec["title"] == nil || spec["elements"] == nil {
		t.Fatalf("template.spec is not a dashboard: %v", keys(spec))
	}
	if _, embedded := spec["panels"]; embedded {
		t.Fatal("template.spec carries v1 panels; the board was not built with dashboardv2")
	}
}

func assertFolderManifest(t *testing.T, data []byte, uid, title string) {
	t.Helper()
	template := assertManifestCR(t, data, uid+"-folder")

	if template["apiVersion"] != folderAPIVersion {
		t.Fatalf("template.apiVersion = %v, want %s", template["apiVersion"], folderAPIVersion)
	}
	meta := template["metadata"].(map[string]any)
	if meta["name"] != uid {
		t.Fatalf("template.metadata.name = %v, want the bare folder UID %s", meta["name"], uid)
	}
	if meta["namespace"] != tenantNamespace {
		t.Fatalf("template.metadata.namespace = %v, want %s", meta["namespace"], tenantNamespace)
	}
	if spec := template["spec"].(map[string]any); spec["title"] != title {
		t.Fatalf("template.spec.title = %v, want %s", spec["title"], title)
	}
}

func assertManifestCR(t *testing.T, data []byte, name string) map[string]any {
	t.Helper()
	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document["kind"] != "GrafanaManifest" {
		t.Fatalf("kind = %v, want GrafanaManifest; GrafanaDashboard cannot carry a v2 payload", document["kind"])
	}
	if document["apiVersion"] != operatorAPIVersion {
		t.Fatalf("apiVersion = %v, want %s", document["apiVersion"], operatorAPIVersion)
	}
	if meta := document["metadata"].(map[string]any); meta["name"] != name {
		t.Fatalf("metadata.name = %v, want %s", meta["name"], name)
	}

	spec := document["spec"].(map[string]any)
	if _, exists := spec["oci"]; exists {
		t.Fatal("spec.oci is set; GrafanaManifest has no content sources and must inline the object")
	}
	return spec["template"].(map[string]any)
}

// assertAlertGroup pins folderUID rather than folderRef. folderRef resolves by
// looking up a live GrafanaFolder object, and this repository no longer emits
// one, so a group left on folderRef would never find its folder.
func assertAlertGroup(t *testing.T, data []byte, folderUID string) {
	t.Helper()
	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document["kind"] != "GrafanaAlertRuleGroup" {
		t.Fatalf("kind = %v, want GrafanaAlertRuleGroup", document["kind"])
	}
	spec := document["spec"].(map[string]any)
	if _, exists := spec["folderRef"]; exists {
		t.Fatal("spec.folderRef is set; it resolves against a GrafanaFolder CR this repository no longer emits")
	}
	if spec["folderUID"] != folderUID {
		t.Fatalf("folderUID = %v, want %s", spec["folderUID"], folderUID)
	}
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
