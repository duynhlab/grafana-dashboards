package dashboards_test

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
)

// otelSourcePanelTitles is every panel of the helm-charts source board, in its
// order. The port is 1:1, so the generated board must carry exactly these.
var otelSourcePanelTitles = []string{
	// Overview & Key Metrics
	"99th Percentile Response Success",
	"95th Percentile Response Success",
	"50th Percentile Response Success",
	"Total RPS (All Requests)",
	"Success RPS (2xx)",
	"Error RPS (4xx/5xx)",
	"Success Rate % (non-5xx)",
	"Error Rate % (5xx)",
	"Apdex Score",
	"Total Request",
	// Traffic & Requests
	"Status Code Distribution",
	"Total Requests by Endpoint",
	"Request Rate by Endpoint",
	// Errors & Performance
	"Request Rate by Method and Endpoint",
	"Server Error Rate by Method and Endpoint (5xx)",
	"Client Errors (4xx)",
	"Server Errors (5xx)",
	"Response time 95th percentile",
	"Response time 50th percentile",
	"Response time 99th percentile",
	// Go Runtime & HTTP I/O (OTel)
	"Memory In-Use",
	"Memory Allocation Rate",
	"Goroutines",
	"GC Pacing Pressure (used / goal)",
	"Total Network Traffic per Service",
	// gRPC East-West (RED)
	"gRPC Server RPS by Method",
	"gRPC Client RPS by Method",
	"gRPC Server Error Rate (non-OK)",
	"gRPC Client Error Rate (non-OK)",
	"gRPC Server P95 Latency",
	"gRPC Client P95 Latency",
	// gRPC East-West (RED) — Per Callee
	"gRPC Server RPS per Callee",
	"gRPC Server Error Ratio per Callee",
	"gRPC Server P95 Latency per Callee",
	// Database (client — otelpgx)
	"DB query p95 by service",
	"DB p95 by operation type",
	"DB operation errors",
	"Pool in-flight (acquired conns)",
	"Pool saturation (acquired / max)",
	"Pool contention (waiting acquires)",
}

// otelAddedPanels are the panels the board adds to the port, keyed by the
// source title they follow.
var otelAddedPanels = map[string]string{
	"Total Request": "Running versions",
}

// otelLinkedPanels link every series (or table row) to the service's traces and
// logs, so each must carry a service_name; the value is how the link names it.
var otelLinkedPanels = map[string]string{
	"Running versions":                   "${__data.fields.service_name}",
	"Client Errors (4xx)":                "${__field.labels.service_name}",
	"Server Errors (5xx)":                "${__field.labels.service_name}",
	"gRPC Server RPS per Callee":         "${__field.labels.service_name}",
	"gRPC Server Error Ratio per Callee": "${__field.labels.service_name}",
	"gRPC Server P95 Latency per Callee": "${__field.labels.service_name}",
	"DB query p95 by service":            "${__field.labels.service_name}",
	"DB operation errors":                "${__field.labels.service_name}",
}

var otelSourceRowTitles = []string{
	"Overview & Key Metrics",
	"Traffic & Requests",
	"Errors & Performance",
	"Go Runtime & HTTP I/O (OTel)",
	"gRPC East-West (RED)",
	"gRPC East-West (RED) — Per Callee",
	"Database (client — otelpgx)",
}

// otelPanel is the slice of a rendered panel the contract below inspects.
type otelPanel struct {
	Spec struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Data        struct {
			Spec struct {
				Queries []struct {
					Spec struct {
						Query struct {
							Spec struct {
								Expr string `json:"expr"`
							} `json:"spec"`
						} `json:"query"`
					} `json:"spec"`
				} `json:"queries"`
			} `json:"spec"`
		} `json:"data"`
		VizConfig struct {
			Spec struct {
				Options     json.RawMessage `json:"options"`
				FieldConfig struct {
					Defaults struct {
						Min   *float64 `json:"min"`
						Max   *float64 `json:"max"`
						Links []struct {
							Title string `json:"title"`
							URL   string `json:"url"`
						} `json:"links"`
					} `json:"defaults"`
					Overrides []struct {
						Matcher struct {
							ID      string `json:"id"`
							Options string `json:"options"`
						} `json:"matcher"`
						Properties []struct {
							ID    string `json:"id"`
							Value any    `json:"value"`
						} `json:"properties"`
					} `json:"overrides"`
				} `json:"fieldConfig"`
			} `json:"spec"`
		} `json:"vizConfig"`
	} `json:"spec"`
}

func TestMicroservicesOTel(t *testing.T) {
	dash, err := microservices.MicroservicesOTel().Build()
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if dash.Title != "Microservices (OTel)" {
		t.Fatalf("unexpected title: %s", dash.Title)
	}

	if len(dash.Variables) != 1 || dash.Variables[0].QueryVariableKind == nil ||
		dash.Variables[0].QueryVariableKind.Spec.Name != "app" {
		t.Fatalf("expected the single app variable, got %d variables", len(dash.Variables))
	}

	var wantTitles []string
	for _, title := range otelSourcePanelTitles {
		wantTitles = append(wantTitles, title)
		if added, ok := otelAddedPanels[title]; ok {
			wantTitles = append(wantTitles, added)
		}
	}
	if len(dash.Elements) != len(wantTitles) {
		t.Fatalf("expected %d panels, got %d", len(wantTitles), len(dash.Elements))
	}

	manifest, err := dashboardv2.Manifest("microservices-monitoring-001-otel", microservices.MicroservicesOTel()).Build()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Metadata.Name != "microservices-monitoring-001-otel" {
		t.Fatalf("unexpected manifest name: %s", manifest.Metadata.Name)
	}

	body, err := json.Marshal(dash.Elements)
	if err != nil {
		t.Fatalf("marshal elements: %v", err)
	}
	var elements map[string]otelPanel
	if err := json.Unmarshal(body, &elements); err != nil {
		t.Fatalf("unmarshal elements: %v", err)
	}

	// Walk the layout so the order checked is the order a reader sees.
	rows := dash.Layout.RowsLayoutKind
	if rows == nil {
		t.Fatal("expected a rows layout")
	}
	if len(rows.Spec.Rows) != len(otelSourceRowTitles) {
		t.Fatalf("expected %d rows, got %d", len(otelSourceRowTitles), len(rows.Spec.Rows))
	}
	var titles []string
	byTitle := map[string]otelPanel{}
	for i, row := range rows.Spec.Rows {
		if row.Spec.Title == nil || *row.Spec.Title != otelSourceRowTitles[i] {
			t.Errorf("row %d: title %v, want %q", i, row.Spec.Title, otelSourceRowTitles[i])
		}
		if row.Spec.Collapse != nil && *row.Spec.Collapse {
			t.Errorf("row %q is collapsed; the source opens every row", otelSourceRowTitles[i])
		}
		grid := row.Spec.Layout.GridLayoutKind
		if grid == nil {
			t.Fatalf("row %q: expected a grid layout", otelSourceRowTitles[i])
		}
		for _, item := range grid.Spec.Items {
			panel, ok := elements[item.Spec.Element.Name]
			if !ok {
				t.Fatalf("layout references unknown element %q", item.Spec.Element.Name)
			}
			titles = append(titles, panel.Spec.Title)
			byTitle[panel.Spec.Title] = panel
		}
	}
	if strings.Join(titles, "\n") != strings.Join(wantTitles, "\n") {
		t.Fatalf("panel titles out of source order:\n got: %q\nwant: %q", titles, wantTitles)
	}

	for title, panel := range byTitle {
		if strings.TrimSpace(panel.Spec.Description) == "" {
			t.Errorf("%q: empty description", title)
		}
		for _, q := range panel.Spec.Data.Spec.Queries {
			if strings.Contains(q.Spec.Query.Spec.Expr, "[$rate]") {
				t.Errorf("%q: still uses the legacy $rate window", title)
			}
		}
	}

	// Route latency keeps the status code in its grouping, as the source does.
	for _, title := range []string{
		"Response time 95th percentile",
		"Response time 50th percentile",
		"Response time 99th percentile",
	} {
		expr := byTitle[title].Spec.Data.Spec.Queries[0].Spec.Query.Spec.Expr
		if !strings.Contains(expr, "by (le, http_route, http_response_status_code)") {
			t.Errorf("%q: grouping lost http_response_status_code: %s", title, expr)
		}
	}

	// Both pies show their table legend (the SDK serializes an unset showLegend
	// as false), count over the selected range, and leave out the /health probes.
	for _, title := range []string{"Status Code Distribution", "Total Requests by Endpoint"} {
		panel := byTitle[title]
		options := panel.Spec.VizConfig.Spec.Options
		if !strings.Contains(string(options), `"showLegend":true`) {
			t.Errorf("%q: legend hidden: %s", title, options)
		}
		expr := panel.Spec.Data.Spec.Queries[0].Spec.Query.Spec.Expr
		if !strings.Contains(expr, "[$__range]") || !strings.Contains(expr, `http_route!="/health"`) {
			t.Errorf("%q: expected an increase over $__range without /health: %s", title, expr)
		}
	}

	for title, panel := range byTitle {
		links := panel.Spec.VizConfig.Spec.FieldConfig.Defaults.Links
		service, linked := otelLinkedPanels[title]
		if !linked {
			if len(links) != 0 {
				t.Errorf("%q: unexpected links %+v", title, links)
			}
			continue
		}
		checkServiceLinks(t, title, service, links)
	}

	if len(dash.Annotations) != 1 {
		t.Fatalf("expected the deploy-marker annotation, got %d", len(dash.Annotations))
	}
	deploys := dash.Annotations[0].Spec
	expr, _ := deploys.Query.Spec.(map[string]any)["expr"].(string)
	if deploys.Name != "Deploys" || !deploys.Enable ||
		!strings.Contains(expr, "unless") || !strings.Contains(expr, `service_name=~"$app"`) ||
		deploys.LegacyOptions["expr"] != expr ||
		deploys.LegacyOptions["titleFormat"] != "{{service_name}} {{service_version}}" {
		t.Errorf("deploy markers: %+v", deploys)
	}

	saturation := byTitle["Pool saturation (acquired / max)"].Spec.VizConfig.Spec.FieldConfig.Defaults
	if saturation.Min == nil || *saturation.Min != 0 || saturation.Max == nil || *saturation.Max != 1 {
		t.Errorf("pool saturation axis is not pinned to [0, 1]: min=%v max=%v", saturation.Min, saturation.Max)
	}

	contention := byTitle["Pool contention (waiting acquires)"].Spec.VizConfig.Spec.FieldConfig.Overrides
	if len(contention) != 1 || contention[0].Matcher.ID != "byFrameRefID" || contention[0].Matcher.Options != "B" {
		t.Fatalf("pool contention: expected one byFrameRefID B override, got %+v", contention)
	}
	props := map[string]any{}
	for _, p := range contention[0].Properties {
		props[p.ID] = p.Value
	}
	if props["custom.axisPlacement"] != "right" || props["custom.axisLabel"] != "wait-time s/s" {
		t.Errorf("pool contention override: got %v", props)
	}

	full, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(full)
	for _, needle := range []string{
		"http_server_request_duration_seconds_bucket",
		"http_server_response_body_size_bytes_sum",
		"rpc_server_call_duration_seconds_count",
		"rpc_client_call_duration_seconds_count",
		"go_memory_gc_goal_bytes",
		"db_client_operation_errors_total",
		"pgxpool_acquired_connections",
		"$__rate_interval",
		"$__range",
		"label_values(go_goroutine_count, service_name)",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing query fragment: %s", needle)
		}
	}

	// The legacy $rate custom interval, the dead $namespace variable and the
	// datasource picker must not survive the port.
	for _, forbidden := range []string{"$rate", "$namespace", "DS_PROMETHEUS"} {
		if strings.Contains(s, forbidden) {
			t.Fatalf("dashboard still references %s", forbidden)
		}
	}
}

// checkServiceLinks asserts a panel's traces and logs links open Explore on the
// right datasource for the service the link names, over the panel's range.
func checkServiceLinks(t *testing.T, title, service string, links []struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}) {
	t.Helper()
	if len(links) != 2 {
		t.Fatalf("%q: expected traces and logs links, got %+v", title, links)
	}
	for i, want := range []struct{ title, datasource, field, value string }{
		{"Traces of " + service, "victoriatraces", "service", service},
		{"Logs of " + service, "victorialogs", "expr", `service.name:="` + service + `"`},
	} {
		link := links[i]
		if link.Title != want.title {
			t.Errorf("%q: link %d title %q, want %q", title, i, link.Title, want.title)
		}
		const prefix = "/explore?schemaVersion=1&panes="
		if !strings.HasPrefix(link.URL, prefix) {
			t.Fatalf("%q: link %d is not an Explore link: %s", title, i, link.URL)
		}
		// Grafana interpolates the variables before navigating, so they stay
		// literal; everything else is escaped JSON.
		raw, err := url.QueryUnescape(strings.TrimPrefix(link.URL, prefix))
		if err != nil {
			t.Fatalf("%q: link %d: %v", title, i, err)
		}
		var panes map[string]struct {
			Datasource string           `json:"datasource"`
			Queries    []map[string]any `json:"queries"`
			Range      struct {
				From string `json:"from"`
				To   string `json:"to"`
			} `json:"range"`
		}
		if err := json.Unmarshal([]byte(raw), &panes); err != nil {
			t.Fatalf("%q: link %d pane is not JSON: %v\n%s", title, i, err, raw)
		}
		pane := panes["a"]
		if pane.Datasource != want.datasource || len(pane.Queries) != 1 ||
			pane.Queries[0][want.field] != want.value ||
			pane.Range.From != "${__from}" || pane.Range.To != "${__to}" {
			t.Errorf("%q: link %d pane %+v", title, i, pane)
		}
	}
}
