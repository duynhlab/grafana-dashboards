package platform

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	platformqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/platform"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// CertManager builds the cert-manager dashboard: certificate readiness and
// time-to-expiry, controller sync throughput and errors, client-go work
// queues, ACME client traffic, and per-component resource usage. Ported from
// the homelab GitOps repo's cert-manager.json (see the query file for the
// source commit).
func CertManager() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("cert-manager").
		Description("cert-manager health: certificate readiness and time-to-expiry, controller sync throughput and errors, client-go work queues, ACME client traffic (empty under a local CA issuer), and per-component resource usage.").
		Editable(true).
		Tags([]string{"cert-manager", "platform", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-24h", "now", "1m"))

	b = b.
		Panel("certs-ready", panels.StatValue("Certificates ready", "short", nil,
			platformqueries.CertManagerCertificatesReady, "")).
		Panel("certs-not-ready", panels.StatValue("Certificates NOT ready", "short", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("red", panels.Ptr(1)),
		}, platformqueries.CertManagerCertificatesNotReady, "")).
		Panel("certs-expiring-7d", panels.StatValue("Certificates expiring < 7d", "short", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("orange", panels.Ptr(1)),
		}, platformqueries.CertManagerCertsExpiringIn7d, "")).
		Panel("certs-expiring-24h", panels.StatValue("Certificates expiring < 24h", "short", []panels.Threshold{
			panels.Thr("green", nil), panels.Thr("red", panels.Ptr(1)),
		}, platformqueries.CertManagerCertsExpiringIn24h, "")).
		Panel("expiry-table", panels.Table("Time to expiry per certificate",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.QueryTable("A", platformqueries.CertManagerTimeToExpiry)},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{
						"Time": true, "__name__": true, "job": true, "instance": true,
						"endpoint": true, "container": true, "pod": true, "service": true, "prometheus": true,
					},
					map[string]string{"Value": "Time to expiry"},
				),
				panels.SortBy("Time to expiry", false),
			},
			panels.Column{Name: "Time to expiry", Unit: "s"},
		)).
		Panel("renewal-table", panels.Table("Next scheduled renewal per certificate",
			[]cog.Builder[dashboardv2.PanelQueryKind]{panels.QueryTable("A", platformqueries.CertManagerTimeToRenewal)},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{
						"Time": true, "__name__": true, "job": true, "instance": true,
						"endpoint": true, "container": true, "pod": true, "service": true, "prometheus": true,
					},
					map[string]string{"Value": "Time to renewal"},
				),
				panels.SortBy("Time to renewal", false),
			},
			panels.Column{Name: "Time to renewal", Unit: "s"},
		)).
		Panel("ready-status-by-condition", panels.SeriesExpr("Certificate ready status by condition", "short",
			platformqueries.CertManagerReadyStatusByCond, "{{condition}}")).
		Panel("not-ready-by-namespace", panels.SeriesExpr("Not-ready certificates by namespace", "short",
			platformqueries.CertManagerNotReadyByNamespace, "{{namespace}}")).
		Panel("controller-sync-rate", panels.SeriesExpr("Controller sync rate", "ops",
			platformqueries.CertManagerControllerSyncRate, "{{controller}}")).
		Panel("controller-sync-error-ratio", panels.SeriesExpr("Controller sync error ratio", "percentunit",
			platformqueries.CertManagerControllerErrorRate, "error ratio")).
		Panel("workqueue-depth", panels.SeriesExpr("Work queue depth", "short",
			platformqueries.CertManagerWorkqueueDepth, "{{name}}")).
		Panel("workqueue-add-rate", panels.SeriesExpr("Work queue add rate", "ops",
			platformqueries.CertManagerWorkqueueAddRate, "{{name}}")).
		Panel("acme-request-rate", panels.SeriesExpr("ACME client request rate", "ops",
			platformqueries.CertManagerACMERequestRate, "{{status}} {{path}}")).
		Panel("acme-avg-duration", panels.SeriesExpr("ACME client avg request duration", "s",
			platformqueries.CertManagerACMEAvgDuration, "avg")).
		Panel("cpu-by-component", panels.SeriesExpr("CPU by component", "percentunit",
			platformqueries.CertManagerCPUByComponent, "{{job}}")).
		Panel("memory-by-component", panels.SeriesExpr("Memory (RSS) by component", "bytes",
			platformqueries.CertManagerMemoryByComponent, "{{job}}"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Certificate health",
			panels.GridItem("certs-ready", 0, 0, 4, 4),
			panels.GridItem("certs-not-ready", 4, 0, 4, 4),
			panels.GridItem("certs-expiring-7d", 8, 0, 4, 4),
			panels.GridItem("certs-expiring-24h", 12, 0, 4, 4),
			panels.GridItem("expiry-table", 0, 4, 12, 8),
			panels.GridItem("renewal-table", 12, 4, 12, 8),
			panels.GridItem("ready-status-by-condition", 0, 12, 12, 8),
			panels.GridItem("not-ready-by-namespace", 12, 12, 12, 8),
		),
		panels.Row("Controller",
			panels.GridItem("controller-sync-rate", 0, 0, 12, 8),
			panels.GridItem("controller-sync-error-ratio", 12, 0, 12, 8),
			panels.GridItem("workqueue-depth", 0, 8, 12, 8),
			panels.GridItem("workqueue-add-rate", 12, 8, 12, 8),
		),
		panels.Row("ACME",
			panels.GridItem("acme-request-rate", 0, 0, 12, 8),
			panels.GridItem("acme-avg-duration", 12, 0, 12, 8),
		),
		panels.Row("Resources",
			panels.GridItem("cpu-by-component", 0, 0, 12, 8),
			panels.GridItem("memory-by-component", 12, 0, 12, 8),
		),
	))
}
