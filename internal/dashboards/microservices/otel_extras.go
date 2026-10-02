package microservices

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/table"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// Additions to the OTel board that the helm-charts source does not have: deploy
// markers, the running versions table, and links from a service's series to its
// traces and logs. They live apart from the 1:1 port so the port stays easy to
// compare with its source.

// Datasource uids provisioned by homelab (GrafanaDatasource CRs). The logs and
// traces carry the same service.name as the metrics' service_name.
const (
	tracesDatasourceUID  = "victoriatraces"
	tracesDatasourceType = "jaeger"
	logsDatasourceUID    = "victorialogs"
	logsDatasourceType   = "victoriametrics-logs-datasource"
)

// Placeholders replaced after URL-escaping, so Grafana sees its own variables
// unescaped and interpolates them when the link is clicked.
const (
	serviceToken = "SERVICETOKEN"
	fromToken    = "FROMTOKEN"
	toToken      = "TOTOKEN"
)

// exploreURL builds an Explore link that opens one query over the panel's
// time range. service is the Grafana expression that yields the service name:
// a series label on a timeseries, a row field on a table.
func exploreURL(datasourceUID string, query map[string]any, service string) string {
	pane := map[string]any{
		"a": map[string]any{
			"datasource": datasourceUID,
			"queries":    []any{query},
			"range":      map[string]any{"from": fromToken, "to": toToken},
		},
	}
	raw, err := json.Marshal(pane)
	if err != nil {
		panic(err) // a map of strings always marshals
	}
	return strings.NewReplacer(
		serviceToken, service,
		fromToken, "${__from}",
		toToken, "${__to}",
	).Replace("/explore?schemaVersion=1&panes=" + url.QueryEscape(string(raw)))
}

// otelServiceLinks opens the traces and the logs of the service under the
// cursor in Explore.
func otelServiceLinks(service string) []any {
	traces := exploreURL(tracesDatasourceUID, map[string]any{
		"refId":      "A",
		"datasource": map[string]any{"type": tracesDatasourceType, "uid": tracesDatasourceUID},
		"queryType":  "search",
		"service":    serviceToken,
	}, service)
	logs := exploreURL(logsDatasourceUID, map[string]any{
		"refId":      "A",
		"datasource": map[string]any{"type": logsDatasourceType, "uid": logsDatasourceUID},
		"expr":       `service.name:="` + serviceToken + `"`,
	}, service)
	return []any{
		dashboardv2.DataLink{Title: "Traces of " + service, Url: traces},
		dashboardv2.DataLink{Title: "Logs of " + service, Url: logs},
	}
}

// seriesService and rowService are the two ways a panel names the service: a
// label on a timeseries, a field on a table row.
const (
	seriesService = "${__field.labels.service_name}"
	rowService    = "${__data.fields.service_name}"
)

// otelDeployMarkers marks the first sample of every new service_version on
// each panel's time axis.
func otelDeployMarkers() cog.Builder[dashboardv2.AnnotationQueryKind] {
	return dashboardv2.NewAnnotationQueryBuilder().
		Name("Deploys").
		Enable(true).
		IconColor("blue").
		Query(annotationQuery{expr: msqueries.OTelVersionFirstSeen}).
		// The Prometheus datasource reads these from the annotation itself.
		LegacyOptions(map[string]any{
			"expr":        msqueries.OTelVersionFirstSeen,
			"step":        "60s",
			"titleFormat": "{{service_name}} {{service_version}}",
			"textFormat":  "first sample of this version",
		})
}

type annotationQuery struct {
	expr string
}

func (q annotationQuery) Build() (dashboardv2.DataQueryKind, error) {
	datasource, err := standards.PrometheusDatasource().Build()
	if err != nil {
		return dashboardv2.DataQueryKind{}, err
	}
	return dashboardv2.DataQueryKind{
		Kind:       "DataQuery",
		Group:      standards.PrometheusPlugin,
		Version:    "v0",
		Datasource: &datasource,
		Spec: map[string]any{
			"expr":  q.expr,
			"refId": "Anno",
		},
	}, nil
}

// otelRunningVersions lists every service with the version its instances
// report now, and links each row to that service's traces and logs.
func otelRunningVersions() cog.Builder[dashboardv2.PanelKind] {
	data := dashboardv2.NewQueryGroupBuilder().
		Target(panels.QueryTable("A", msqueries.OTelRunningVersions)).
		Transformation(panels.Organize(
			map[string]bool{"Time": true},
			map[string]string{"service_version": "version", "Value": "instances"},
		)).
		Transformation(panels.SortBy("service_name", false))
	viz := table.NewVisualizationV2Builder().
		CellHeight(common.TableCellHeightSm).
		DataLinks(otelServiceLinks(rowService))
	return dashboardv2.NewPanelBuilder().
		Title("Running versions").
		Description("The service_version each service's instances report now, from the OTel resource attributes on go_goroutine_count; a rollout in progress shows two rows for one service. Click a row to open that service's traces or logs. Blue markers on the time series are the first sample of a new version.").
		Data(data).
		Visualization(viz)
}
