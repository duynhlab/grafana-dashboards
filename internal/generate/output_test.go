package generate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

	stale := filepath.Join(root, "deploy", "manifests", "stale.yaml")
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
		"deploy/manifests/grafanafolder-kubernetes.yaml",
		"deploy/manifests/grafanafolder-databases.yaml",
		"deploy/manifests/grafanadashboard-kubernetes-cluster-overview.yaml",
		"deploy/manifests/grafanadashboard-pg-io-waits.yaml",
		"deploy/manifests/grafanaalertrulegroup-kubernetes.yaml",
		"deploy/manifests/grafanaalertrulegroup-databases.yaml",
		"generated/alerts/kubernetes_crashlooping_pods.json",
		"generated/alerts/postgres_backends_waiting.json",
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

	assertManifest(t, second["deploy/manifests/grafanadashboard-kubernetes-cluster-overview.yaml"], "GrafanaDashboard", "kubernetes")
	assertManifest(t, second["deploy/manifests/grafanaalertrulegroup-databases.yaml"], "GrafanaAlertRuleGroup", "databases")
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

func assertManifest(t *testing.T, data []byte, kind, folderRef string) {
	t.Helper()
	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document["kind"] != kind {
		t.Fatalf("kind = %v, want %s", document["kind"], kind)
	}
	spec := document["spec"].(map[string]any)
	if spec["folderRef"] != folderRef {
		t.Fatalf("folderRef = %v, want %s", spec["folderRef"], folderRef)
	}
	if kind == "GrafanaDashboard" {
		var dashboard map[string]any
		if err := json.Unmarshal([]byte(spec["json"].(string)), &dashboard); err != nil {
			t.Fatalf("invalid dashboard JSON: %v", err)
		}
	}
}
