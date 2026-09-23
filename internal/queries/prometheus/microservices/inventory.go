package microservices

// PromQL for the "Inventory Service — Stock Authority" board (UID
// inventory-overview).
//
// Ported from the homelab GitOps repo,
// kubernetes/infra/configs/observability/grafana/dashboards/inventory.json
// at commit 64765bfd1cece46c152e2d4245e6f99f7845cda2.
//
// Six of the board's seven panels read `inventory:*` recording rules defined in
// homelab at
// kubernetes/infra/configs/observability/metrics/prometheusrules/microservices/inventory.yaml.
// This board is USELESS without that PrometheusRule applied to the cluster: the
// panels reference the recording-rule names directly, not the raw counters and
// histograms those rules aggregate, so with the rules absent every one of those
// panels shows no data.
//
// The one exception is "gRPC request rate by method"
// (InventoryGRPCRequestRateByMethod), which reads the raw
// rpc_server_call_duration_seconds_count metric directly rather than a recording
// rule. The source board's literal `[5m]` window on that query was rewritten to
// `$__rate_interval`, per the usual "never a literal rate window" convention.
// That convention does NOT apply to the other five constants below: their
// `rate5m` / `p95_5m` suffixes are part of the recording-rule NAME, not a
// window baked into the query text, so there is nothing to rewrite.
const (
	// InventoryReservationRate5m is the reservation FSM throughput by
	// operation (reserve|release|commit) and outcome. Business outcomes
	// (insufficient/conflict/concurrency/invalid_transition/not_found) and the
	// infra outcome=error share this series.
	InventoryReservationRate5m = `inventory:reservation:rate5m`

	// InventoryCheckRate5m is the availability-check throughput by outcome
	// (fulfillable|shortage|unknown_sku|error).
	InventoryCheckRate5m = `inventory:check:rate5m`

	// InventoryGRPCRequestRateByMethod is the one raw query on this board: the
	// otelgrpc server call rate for InventoryService, not backed by a
	// recording rule. app="inventory" names the service (OTLP push path has no
	// job label; vmagent maps service_name->app), not an environment.
	InventoryGRPCRequestRateByMethod = `sum by (rpc_method) (rate(rpc_server_call_duration_seconds_count{app="inventory", rpc_method=~"inventory.v1.InventoryService/.+"}[$__rate_interval]))`

	// InventoryRPCErrorRatio5m is the gRPC non-OK / all ratio per RPC method.
	// Business rejections (FailedPrecondition on shortage) count as non-OK
	// here; read it alongside InventoryReservationRate5m to separate business
	// shortage from an infra fault.
	InventoryRPCErrorRatio5m = `inventory:rpc_error:ratio5m`

	// InventoryRPCDurationP95_5m is the otelgrpc server call duration p95 per
	// RPC method.
	InventoryRPCDurationP95_5m = `inventory:rpc_duration:p95_5m`

	// InventoryDBOperationDurationP95_5m is the otelpgx statement latency p95
	// (pgx_operation_type="query" isolates statement time from
	// connect/prepare/acquire).
	InventoryDBOperationDurationP95_5m = `inventory:db_operation_duration:p95_5m`
)

// InventoryReadme is the source board's own text panel, kept verbatim on port:
// it explains why a flat rate is a finding (not the expected phase-1 reading)
// now that inventory-service has three live callers since RFC-0021 phase 4.
const InventoryReadme = `**inventory-service is the sole stock authority and has three live callers** since RFC-0021 phase 4: the order saga (reserve/commit/release), checkout (availability), and product ` + "`/details`" + `. gRPC-only -- no HTTP edge route -- so ` + "`:8080`" + ` carries only ` + "`/health`" + ` and ` + "`/ready`" + ` and every panel here reads the rpc_* RED metrics.

A flat rate is therefore a **finding**, not the expected phase-1 reading it used to be. On a freshly rebuilt cluster it just means no traffic has run yet.`
