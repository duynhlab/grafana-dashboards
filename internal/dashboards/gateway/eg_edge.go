// Package gateway holds the API Gateway folder's dashboards.
package gateway

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	gwqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/gateway"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

const egEdgeReadme = "**Control plane says** “traffic to `/payment` must go THERE, with THIS timeout and THIS TLS.” · **Data plane says** “OK — a `/payment` request just passed through me in 120 ms.” · **This dashboard says** “20k req/s, p99 is 1.2 s, and 5xx is rising.”\n\n" +
	"**What each row reads:** **Edge Overview** + **Data Plane** ← job `envoy-gateway` (Envoy proxy stats, gateway:19005 `/stats/prometheus`) · **Control Plane** ← job `envoy-gateway-controller` (EG controller, gateway:19001) · **Infrastructure** ← process/server metrics from both jobs.\n\n" +
	"Panels reading kube-state / cAdvisor metrics only exist on the Kubernetes cluster — they are empty locally by design; the affected panel carries a live local fallback query alongside them."

// EGEdge builds the "Envoy Gateway — Edge Overview" dashboard: edge SRE golden
// signals on the Envoy data plane, Envoy Gateway control-plane health, and
// process-level infrastructure for one Envoy Gateway install. Ported from the
// homelab GitOps repo; see internal/queries/prometheus/gateway/eg_edge.go for
// the source path and commit.
func EGEdge() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Envoy Gateway — Edge Overview").
		Description("Edge SRE view of the Envoy Gateway install: golden signals on the data plane (job=envoy-gateway), control-plane health (job=envoy-gateway-controller), and process-level infrastructure.").
		Editable(true).
		Tags([]string{"envoy-gateway", "edge", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-1h", "now", "30s"))

	b = b.
		Panel("readme", panels.Text("README — how to read this dashboard", egEdgeReadme)).

		// Edge Overview.
		Panel("edge-rps", panels.StatValue("Requests / sec (edge)", "reqps",
			[]panels.Threshold{panels.Thr("green", nil)},
			gwqueries.EGEdgeRequestsPerSec, "RPS")).
		Panel("edge-error-rate", panels.StatValue("Error rate (5xx)", "percentunit",
			[]panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("yellow", panels.Ptr(0.01)),
				panels.Thr("red", panels.Ptr(0.05)),
			},
			gwqueries.EGEdgeErrorRate5xx, "5xx ratio")).
		Panel("edge-p99", panels.StatValue("P99 latency (edge)", "ms",
			[]panels.Threshold{
				panels.Thr("green", nil),
				panels.Thr("yellow", panels.Ptr(500)),
				panels.Thr("red", panels.Ptr(1000)),
			},
			gwqueries.EGEdgeP99Latency, "p99")).
		Panel("edge-availability", panels.StatValue("Availability (dashboard window)", "percentunit",
			[]panels.Threshold{
				panels.Thr("red", nil),
				panels.Thr("yellow", panels.Ptr(0.99)),
				panels.Thr("green", panels.Ptr(0.999)),
			},
			gwqueries.EGEdgeAvailabilityWindow, "availability")).

		// Data Plane.
		Panel("route-request-rate", panels.SeriesExpr("Request rate by route (upstream cluster)", "reqps",
			gwqueries.EGEdgeRequestRateByRoute, "{{envoy_cluster_name}}")).
		Panel("edge-4xx-5xx", panels.SeriesExpr("4xx / 5xx response rate (edge)", "reqps",
			gwqueries.EGEdgeResponseRate4xx5xx, "{{envoy_response_code_class}}xx")).
		Panel("downstream-latency", panels.Series("Downstream latency p50 / p95 / p99 (edge)", "ms",
			panels.PromQuery(gwqueries.EGEdgeDownstreamLatencyP50, "p50"),
			panels.Query("B", gwqueries.EGEdgeDownstreamLatencyP95, "p95"),
			panels.Query("C", gwqueries.EGEdgeDownstreamLatencyP99, "p99"),
		)).
		Panel("upstream-latency-p95", panels.SeriesExpr("Upstream latency p95 by route", "ms",
			gwqueries.EGEdgeUpstreamLatencyP95ByRoute, "{{envoy_cluster_name}}")).
		Panel("upstream-retries-timeouts", panels.Series("Upstream retries & timeouts", "reqps",
			panels.PromQuery(gwqueries.EGEdgeUpstreamRetries, "retries"),
			panels.Query("B", gwqueries.EGEdgeUpstreamRetrySuccesses, "retry successes"),
			panels.Query("C", gwqueries.EGEdgeUpstreamTimeouts, "timeouts"),
			panels.Query("D", gwqueries.EGEdgeUpstreamPerTryTimeouts, "per-try timeouts"),
		)).
		Panel("active-connections", panels.Series("Active connections", "short",
			panels.PromQuery(gwqueries.EGEdgeActiveConnectionsDownstream, "downstream (edge HTTP)"),
			panels.Query("B", gwqueries.EGEdgeActiveConnectionsListener, "listener {{envoy_listener_address}}"),
			panels.Query("C", gwqueries.EGEdgeActiveConnectionsUpstream, "upstream (routes)"),
		)).

		// Control Plane (Envoy Gateway).
		Panel("watchable-events-publishes", panels.Series("Watchable events / publishes by runner", "ops",
			panels.PromQuery(gwqueries.EGEdgeWatchableEvents, "events {{runner}}"),
			panels.Query("B", gwqueries.EGEdgeWatchablePublishes, "publishes {{runner}}"),
		)).
		Panel("watchable-subscribe-duration", panels.SeriesExpr("Watchable subscribe duration p99 by runner", "s",
			gwqueries.EGEdgeWatchableSubscribeDurationP99, "{{runner}}")).
		Panel("watchable-queue-depth", panels.SeriesExpr("Watchable queue depth by runner", "short",
			gwqueries.EGEdgeWatchableQueueDepth, "{{runner}}")).
		Panel("xds-snapshot", panels.Series("xDS snapshot creates / updates", "short",
			panels.PromQuery(gwqueries.EGEdgeXDSSnapshotCreates, "create {{status}}"),
			panels.Query("B", gwqueries.EGEdgeXDSSnapshotUpdates, "update {{nodeID}} {{status}}"),
		)).
		Panel("gateway-httproute-status", panels.Series("Gateway / HTTPRoute status updates", "ops",
			panels.PromQuery(gwqueries.EGEdgeGatewayStatusUpdates, "status_update {{kind}}"),
			panels.Query("B", gwqueries.EGEdgeProviderStatusQueueDepth, "fallback: provider {{message}} queue depth"),
		)).
		Panel("control-plane-panics-certwatcher", panels.Series("Control-plane panics & cert watcher", "short",
			panels.PromQuery(gwqueries.EGEdgeWatchablePanicsRecovered, "watchable panics recovered"),
			panels.Query("B", gwqueries.EGEdgeCertWatcherReadErrors, "certwatcher read errors"),
			panels.Query("C", gwqueries.EGEdgeCertWatcherReadsTotal, "certwatcher reads"),
		)).

		// Infrastructure.
		Panel("control-plane-cpu", panels.SeriesExpr("Control-plane CPU (EG process)", "short",
			gwqueries.EGEdgeControlPlaneCPU, "EG CPU cores")).
		Panel("memory-eg-envoy", panels.Series("Memory — EG process & Envoy server", "bytes",
			panels.PromQuery(gwqueries.EGEdgeControlPlaneMemoryRSS, "EG RSS"),
			panels.Query("B", gwqueries.EGEdgeEnvoyMemoryAllocated, "Envoy allocated"),
			panels.Query("C", gwqueries.EGEdgeEnvoyMemoryHeapSize, "Envoy heap"),
			panels.Query("D", gwqueries.EGEdgeEnvoyMemoryPhysical, "Envoy physical"),
		)).
		Panel("edge-health", panels.StateTimeline("Edge health",
			[]panels.Threshold{
				panels.Thr("red", nil),
				panels.Thr("green", panels.Ptr(1)),
			},
			panels.PromQuery(gwqueries.EGEdgeEnvoyLive, "Envoy live"),
			panels.Query("B", gwqueries.EGEdgeEGScrapeUp, "EG scrape up"),
			panels.Query("C", gwqueries.EGEdgeEnvoyScrapeUp, "Envoy scrape up"),
		)).
		Panel("pod-restarts-container-cpu", panels.Series("Pod restarts / container CPU (kube-state + cAdvisor)", "short",
			panels.PromQuery(gwqueries.EGEdgePodRestarts, "restarts {{pod}}"),
			panels.Query("B", gwqueries.EGEdgeContainerCPU, "container CPU {{pod}}"),
			panels.Query("C", gwqueries.EGEdgeEnvoyUptimeFallback, "fallback: Envoy uptime (local)"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Edge Overview",
			panels.GridItem("readme", 0, 0, 24, 6),
			panels.GridItem("edge-rps", 0, 6, 6, 5),
			panels.GridItem("edge-error-rate", 6, 6, 6, 5),
			panels.GridItem("edge-p99", 12, 6, 6, 5),
			panels.GridItem("edge-availability", 18, 6, 6, 5),
		),
		panels.Row("Data Plane",
			panels.GridItem("route-request-rate", 0, 0, 12, 8),
			panels.GridItem("edge-4xx-5xx", 12, 0, 12, 8),
			panels.GridItem("downstream-latency", 0, 8, 12, 8),
			panels.GridItem("upstream-latency-p95", 12, 8, 12, 8),
			panels.GridItem("upstream-retries-timeouts", 0, 16, 12, 8),
			panels.GridItem("active-connections", 12, 16, 12, 8),
		),
		panels.Row("Control Plane (Envoy Gateway)",
			panels.GridItem("watchable-events-publishes", 0, 0, 12, 8),
			panels.GridItem("watchable-subscribe-duration", 12, 0, 12, 8),
			panels.GridItem("watchable-queue-depth", 0, 8, 12, 8),
			panels.GridItem("xds-snapshot", 12, 8, 12, 8),
			panels.GridItem("gateway-httproute-status", 0, 16, 12, 8),
			panels.GridItem("control-plane-panics-certwatcher", 12, 16, 12, 8),
		),
		panels.Row("Infrastructure",
			panels.GridItem("control-plane-cpu", 0, 0, 8, 8),
			panels.GridItem("memory-eg-envoy", 8, 0, 8, 8),
			panels.GridItem("edge-health", 16, 0, 8, 8),
			panels.GridItem("pod-restarts-container-cpu", 0, 8, 24, 6),
		),
	))
}
