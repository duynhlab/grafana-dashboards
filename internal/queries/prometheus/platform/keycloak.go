package platform

// PromQL for the "Keycloak — Identity" board (UID keycloak-identity), ported
// from the homelab repo at
// kubernetes/infra/configs/observability/grafana/dashboards/keycloak-identity.json
// (commit 2d9f32f2865f525eff4645efce5b610174ffb5ca).
//
// Metric model: Keycloak on Quarkus. HTTP server latency comes from
// `http_server_requests_seconds_bucket`/`_count` (labels `uri`, `outcome`,
// `status`, `method`); auth events come from `keycloak_user_events_total`
// (labels `event`, `realm`, `client_id`, `error` — failures are the same
// event with a non-empty `error` label, there is no dedicated failure event);
// the DB pool is Agroal (`agroal_available_count`, `agroal_active_count`,
// `agroal_awaiting_count`); the process is a standard JVM
// (`jvm_memory_used_bytes`, `jvm_gc_pause_seconds_sum/max`,
// `process_uptime_seconds`). All series carry `job="keycloak"`, the scrape
// target on the management interface (:9000).
//
// Normalisation applied while porting: `${DS_PROMETHEUS}` is dropped in favor
// of the repository's logical prometheus datasource; every `$__rate_interval`
// window is kept as-is. The "Slowest endpoints" table keeps its source
// `$__range` instant aggregation — it is a dashboard-window snapshot, not a
// rate window.
const (
	// --- Overview ---------------------------------------------------------

	// KeycloakSuccessfulLoginsPerMin is the rate of `login` events without an
	// `error` label, scaled to per-minute.
	KeycloakSuccessfulLoginsPerMin = `sum(rate(keycloak_user_events_total{job="keycloak",event="login",error=""}[$__rate_interval])) * 60`

	// KeycloakAuthFailureRatio is the share of auth events (login /
	// refresh_token / code_to_token) carrying a non-empty `error` label.
	KeycloakAuthFailureRatio = `sum(rate(keycloak_user_events_total{job="keycloak",event=~"login|refresh_token|code_to_token",error!=""}[$__rate_interval])) / sum(rate(keycloak_user_events_total{job="keycloak",event=~"login|refresh_token|code_to_token"}[$__rate_interval]))`

	// KeycloakAuthP95Latency is p95 HTTP server latency across realm
	// endpoints — the token/auth path a user feels.
	KeycloakAuthP95Latency = `histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~"/realms/.*"}[$__rate_interval])) by (le))`

	// KeycloakSLOCompliance is the share of realm-endpoint requests
	// completing under the 250 ms histogram bucket (le="0.25").
	KeycloakSLOCompliance = `sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~"/realms/.*",le="0.25"}[$__rate_interval])) / sum(rate(http_server_requests_seconds_count{job="keycloak",uri=~"/realms/.*"}[$__rate_interval]))`

	// KeycloakScrapeUp is Prometheus `up` for the keycloak job.
	KeycloakScrapeUp = `up{job="keycloak"}`

	// --- Latency ------------------------------------------------------

	// KeycloakRealmLatencyP50/P95/P99 are latency quantiles across all
	// `/realms/...` endpoints (auth, token, certs, login-actions).
	KeycloakRealmLatencyP50 = `histogram_quantile(0.50, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~"/realms/.*"}[$__rate_interval])) by (le))`
	KeycloakRealmLatencyP95 = `histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~"/realms/.*"}[$__rate_interval])) by (le))`
	KeycloakRealmLatencyP99 = `histogram_quantile(0.99, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~"/realms/.*"}[$__rate_interval])) by (le))`

	// KeycloakTokenLatencyP95/P99 are latency quantiles for the token
	// endpoint only (`uri=~".*/token"`) — code exchange and refresh grants.
	KeycloakTokenLatencyP95 = `histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~".*/token"}[$__rate_interval])) by (le))`
	KeycloakTokenLatencyP99 = `histogram_quantile(0.99, sum(rate(http_server_requests_seconds_bucket{job="keycloak",uri=~".*/token"}[$__rate_interval])) by (le))`

	// KeycloakSlowestEndpoints ranks templated URIs by p95 latency over the
	// dashboard time range. $__range is intentional: this is a snapshot
	// instant query, not a rate window.
	KeycloakSlowestEndpoints = `topk(7, histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{job="keycloak"}[$__range])) by (le, uri)))`

	// --- Auth events --------------------------------------------------

	// KeycloakUserEventsByType is all Keycloak user events by `event` label
	// (successes and failures together).
	KeycloakUserEventsByType = `sum(rate(keycloak_user_events_total{job="keycloak"}[$__rate_interval])) by (event)`

	// KeycloakFailedEventsByError is auth events carrying a non-empty
	// `error` label, split by event and error.
	KeycloakFailedEventsByError = `sum(rate(keycloak_user_events_total{job="keycloak",error!=""}[$__rate_interval])) by (event, error)`

	// KeycloakEventsByRealmClient splits login / refresh_token /
	// code_to_token per realm and client_id.
	KeycloakEventsByRealmClient = `sum(rate(keycloak_user_events_total{job="keycloak",event=~"login|refresh_token|code_to_token"}[$__rate_interval])) by (realm, client_id)`

	// KeycloakTokenGrants compares refresh_token vs code_to_token rates.
	KeycloakTokenGrants = `sum(rate(keycloak_user_events_total{job="keycloak",event=~"refresh_token|code_to_token"}[$__rate_interval])) by (event)`

	// --- Infrastructure -------------------------------------------------

	// KeycloakAgroalAvailable/Active/Awaiting are the Agroal DB connection
	// pool gauges for the default datasource.
	KeycloakAgroalAvailable = `agroal_available_count{job="keycloak",datasource="default"}`
	KeycloakAgroalActive    = `agroal_active_count{job="keycloak",datasource="default"}`
	KeycloakAgroalAwaiting  = `agroal_awaiting_count{job="keycloak",datasource="default"}`

	// KeycloakJVMHeapUsed/Committed/Max are heap memory gauges summed across
	// all heap pools.
	KeycloakJVMHeapUsed      = `sum(jvm_memory_used_bytes{job="keycloak",area="heap"})`
	KeycloakJVMHeapCommitted = `sum(jvm_memory_committed_bytes{job="keycloak",area="heap"})`
	KeycloakJVMHeapMax       = `sum(jvm_memory_max_bytes{job="keycloak",area="heap"})`

	// KeycloakGCPauseRate is the fraction of wall clock spent paused for GC.
	KeycloakGCPauseRate = `sum(rate(jvm_gc_pause_seconds_sum{job="keycloak"}[$__rate_interval]))`

	// KeycloakGCPauseMax is the largest single GC pause in the scrape
	// interval.
	KeycloakGCPauseMax = `max(jvm_gc_pause_seconds_max{job="keycloak"})`

	// KeycloakPasswordHashValidations is Argon2id password verifications per
	// second by realm and outcome.
	KeycloakPasswordHashValidations = `sum(rate(keycloak_credentials_password_hashing_validations_total{job="keycloak"}[$__rate_interval])) by (realm, outcome)`

	// KeycloakProcessUptime is process_uptime_seconds; a reset to ~0 marks a
	// restart.
	KeycloakProcessUptime = `process_uptime_seconds{job="keycloak"}`
)
