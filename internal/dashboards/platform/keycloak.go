package platform

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	platformqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/platform"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// keycloakReadme is the source board's own explanation of how failures are
// modeled (an `error` label, not a dedicated event) and why the cluster twin
// of this board needs no query remapping. Kept verbatim on port.
const keycloakReadme = `**What this board answers:** is Keycloak issuing tokens (logins, refreshes, code exchanges), how many auth events are **failing**, is the auth path **fast enough** (p95 / 250 ms SLO bucket), and is the process itself healthy (Agroal DB pool, JVM heap, GC, password-hash cost).

**Failures are an ` + "`error`" + ` LABEL, not an event.** Keycloak emits the same ` + "`event`" + ` (` + "`login`" + ` / ` + "`refresh_token`" + ` / ` + "`code_to_token`" + `) with ` + "`error=\"invalid_user_credentials\"|\"user_not_found\"|\"invalid_token\"`" + ` set on failure — there is **no** ` + "`login_error`" + ` event. Failure ratio = ` + "`sum(rate(keycloak_user_events_total{error!=\"\"}))`" + ` / ` + "`sum(rate(keycloak_user_events_total))`" + ` over the auth events.

**Both stacks share job ` + "`keycloak`" + `** (management interface ` + "`:9000`" + ` ` + "`/metrics`" + `) with identical metric and label names — the cluster twin of this board needs no query remapping, only the datasource input differs. HTTP ` + "`uri`" + ` labels are templated (` + "`/realms/{realm}/protocol/{protocol}/auth`" + `, ` + "`.../token`" + `, ` + "`/admin/...`" + `).

**Context:** closes the Keycloak row of the observability gap map (` + "`docs/observability/`" + ` — RFC-0022 identity observability, ADR-041).`

// KeycloakIdentity builds the "Keycloak — Identity" dashboard, ported from the
// homelab observability stack. It covers auth throughput and failure ratio,
// realm/token endpoint latency against a 250 ms SLO bucket, auth event
// breakdowns by type/error/realm/client, and process infrastructure (Agroal
// DB pool, JVM heap, GC, password-hash cost).
func KeycloakIdentity() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Keycloak — Identity").
		Description("Identity-provider health for the platform Keycloak: auth event rates and failure ratio (error label on login/refresh_token/code_to_token), realm/token endpoint latency and the 250 ms SLO bucket, and process infrastructure (Agroal DB pool, JVM heap, GC, Argon2id hash cost). Job `keycloak` scrapes the management interface :9000; the cluster twin uses the same queries with only the datasource input changed.").
		Editable(true).
		Tags([]string{"keycloak", "identity", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-1h", "now", "30s"))

	b = b.
		Panel("readme", panels.Text("README — how to read this dashboard", keycloakReadme)).
		Panel("logins-per-min", panels.StatValue("Successful logins / min", "opm", []panels.Threshold{
			panels.Thr("green", nil),
		}, platformqueries.KeycloakSuccessfulLoginsPerMin, "logins/min")).
		Panel("auth-failure-ratio", panels.StatValue("Auth failure ratio", "percentunit", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.05)),
			panels.Thr("red", panels.Ptr(0.2)),
		}, platformqueries.KeycloakAuthFailureRatio, "failure ratio")).
		Panel("auth-p95", panels.StatValue("Auth p95 latency", "s", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.25)),
			panels.Thr("red", panels.Ptr(1)),
		}, platformqueries.KeycloakAuthP95Latency, "p95")).
		Panel("slo-compliance", panels.StatValue("SLO compliance (< 250 ms)", "percentunit", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("yellow", panels.Ptr(0.95)),
			panels.Thr("green", panels.Ptr(0.99)),
		}, platformqueries.KeycloakSLOCompliance, "compliance")).
		Panel("scrape-up", panels.StatValue("Keycloak scrape up", "short", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("green", panels.Ptr(1)),
		}, platformqueries.KeycloakScrapeUp, "up")).
		Panel("realm-latency", panels.Series("Realm endpoints — latency quantiles", "s",
			panels.PromQuery(platformqueries.KeycloakRealmLatencyP50, "p50"),
			panels.Query("B", platformqueries.KeycloakRealmLatencyP95, "p95"),
			panels.Query("C", platformqueries.KeycloakRealmLatencyP99, "p99"),
		)).
		Panel("token-latency", panels.Series("Token endpoint — p95 / p99", "s",
			panels.PromQuery(platformqueries.KeycloakTokenLatencyP95, "p95"),
			panels.Query("B", platformqueries.KeycloakTokenLatencyP99, "p99"),
		)).
		Panel("slowest-endpoints", panels.Table("Slowest endpoints (p95, dashboard window)",
			[]cog.Builder[dashboardv2.PanelQueryKind]{
				panels.QueryTable("A", platformqueries.KeycloakSlowestEndpoints),
			},
			[]cog.Builder[dashboardv2.TransformationKind]{
				panels.Organize(
					map[string]bool{"Time": true},
					map[string]string{"Value": "p95"},
				),
				panels.SortBy("p95", true),
			},
			panels.Column{Name: "p95", Unit: "s"},
		)).
		Panel("events-by-type", panels.SeriesExpr("User events by type", "ops", platformqueries.KeycloakUserEventsByType, "{{event}}")).
		Panel("failed-events", panels.SeriesExpr("Failed events by error", "ops", platformqueries.KeycloakFailedEventsByError, "{{event}} / {{error}}")).
		Panel("events-by-realm-client", panels.SeriesExpr("Auth events by realm / client", "ops", platformqueries.KeycloakEventsByRealmClient, "{{realm}} / {{client_id}}")).
		Panel("token-grants", panels.SeriesExpr("Token grants — refresh vs code exchange", "ops", platformqueries.KeycloakTokenGrants, "{{event}}")).
		Panel("agroal-pool", panels.Series("Agroal DB pool (datasource=default)", "short",
			panels.PromQuery(platformqueries.KeycloakAgroalAvailable, "available"),
			panels.Query("B", platformqueries.KeycloakAgroalActive, "active"),
			panels.Query("C", platformqueries.KeycloakAgroalAwaiting, "awaiting"),
		)).
		Panel("jvm-heap", panels.Series("JVM heap", "bytes",
			panels.PromQuery(platformqueries.KeycloakJVMHeapUsed, "used"),
			panels.Query("B", platformqueries.KeycloakJVMHeapCommitted, "committed"),
			panels.Query("C", platformqueries.KeycloakJVMHeapMax, "max"),
		)).
		Panel("gc-pause", panels.Series("GC pause", "s",
			panels.PromQuery(platformqueries.KeycloakGCPauseRate, "pause time /s"),
			panels.Query("B", platformqueries.KeycloakGCPauseMax, "max pause"),
		)).
		Panel("password-hash", panels.SeriesExpr("Password hash validations", "ops", platformqueries.KeycloakPasswordHashValidations, "{{realm}} / {{outcome}}")).
		Panel("uptime", panels.StatValue("Process uptime", "dtdurations", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("green", panels.Ptr(300)),
		}, platformqueries.KeycloakProcessUptime, "uptime"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Overview",
			panels.GridItem("readme", 0, 0, 24, 6),
			panels.GridItem("logins-per-min", 0, 6, 5, 5),
			panels.GridItem("auth-failure-ratio", 5, 6, 5, 5),
			panels.GridItem("auth-p95", 10, 6, 5, 5),
			panels.GridItem("slo-compliance", 15, 6, 5, 5),
			panels.GridItem("scrape-up", 20, 6, 4, 5),
		),
		panels.Row("Latency",
			panels.GridItem("realm-latency", 0, 0, 9, 8),
			panels.GridItem("token-latency", 9, 0, 9, 8),
			panels.GridItem("slowest-endpoints", 18, 0, 6, 8),
		),
		panels.Row("Auth events",
			panels.GridItem("events-by-type", 0, 0, 12, 8),
			panels.GridItem("failed-events", 12, 0, 12, 8),
			panels.GridItem("events-by-realm-client", 0, 8, 12, 8),
			panels.GridItem("token-grants", 12, 8, 12, 8),
		),
		panels.Row("Infrastructure",
			panels.GridItem("agroal-pool", 0, 0, 8, 8),
			panels.GridItem("jvm-heap", 8, 0, 8, 8),
			panels.GridItem("gc-pause", 16, 0, 8, 8),
			panels.GridItem("password-hash", 0, 8, 16, 7),
			panels.GridItem("uptime", 16, 8, 8, 7),
		),
	))
}
