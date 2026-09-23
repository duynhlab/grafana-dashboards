package microservices

// PromQL for the "Order Saga & Payment — Cutover Baseline" board (UID
// rfc0021-baseline).
//
// Ported from the homelab GitOps repo,
// kubernetes/infra/configs/observability/grafana/dashboards/cutover-baseline.json
// at commit 64765bfd1cece46c152e2d4245e6f99f7845cda2.
//
// Metric model: none. Every query here is a bare selector on a `rfc0021:*`
// recording rule — there is no raw metric or $__rate_interval to write, the
// averaging window is already baked into the rule name (:rate5m, :p95_5m,
// :ratio5m). The rules live in the homelab GitOps repo at
//
//	kubernetes/infra/configs/observability/metrics/prometheusrules/microservices/rfc0021-baseline.yaml
//	kubernetes/infra/configs/observability/metrics/prometheusrules/microservices/rfc0021-write-migration.yaml
//
// This board is USELESS without those PrometheusRules applied to the
// cluster: every panel selects a recorded series by name, and a series that
// is never recorded renders as "No data" rather than falling back to a raw
// query.
//
// Panels dropped (3 of the source's 11): the source board titled these
// "(RETIRED — no data is expected)" and its own panel descriptions say the
// underlying recording rules were removed in RFC-0021 phase 4. Confirmed
// against homelab: none of the three rules below exist in
// rfc0021-baseline.yaml or rfc0021-write-migration.yaml any more.
//
//   - "Product ReserveStock outcomes" (rfc0021:product_stock_reservations:rate5m)
//   - "Shadow compare outcomes (per second)" (rfc0021:checkout_shadow_compare:rate5m)
//   - "Shadow divergence ratio" (rfc0021:checkout_shadow_divergence:ratio5m)
//
// Dropping the first two of those emptied the row "RFC-0021 read migration —
// shadow compare (P2-7)" entirely, so that row is dropped too. The board
// ends at 8 panels across 3 rows instead of the source's 11 panels across 4
// rows.
//
// The legacy board's `${DS_PROMETHEUS}` datasource variable and `__inputs`
// entry are dropped: standards.PrometheusDatasource() already resolves the
// logical prometheus datasource for every generated board, and the source
// carried no other template variables.
const (
	// CutoverConfirmSuccessRate is the program-level checkout-confirm success
	// signal: confirmed sessions/sec.
	CutoverConfirmSuccessRate = `rfc0021:checkout_confirm_success:rate5m`

	// CutoverConfirmBounceRate is confirms bounced by re-validation
	// (PRICE_CHANGED / STOCK_UNAVAILABLE) — the population the
	// availability-source cutover must not grow.
	CutoverConfirmBounceRate = `rfc0021:checkout_confirm_bounce:rate5m`

	// CutoverConfirmDurationP95 is end-to-end checkout-confirm handler
	// latency p95, the baseline the extra Inventory hop budgets against.
	CutoverConfirmDurationP95 = `rfc0021:checkout_confirm_duration:p95_5m`

	// CutoverOrderStockReservationRate is the saga-side ReserveStock view by
	// result — the series the phase-3 flip rehomes onto inventory.
	CutoverOrderStockReservationRate = `rfc0021:order_stock_reservation:rate5m`

	// CutoverProductRPCDurationP95 is product gRPC latency p95 by method
	// (product.v1.ProductService/*).
	CutoverProductRPCDurationP95 = `rfc0021:product_rpc_duration:p95_5m`

	// CutoverProductRPCErrorRatio is product gRPC error ratio (non-OK / all)
	// by method.
	CutoverProductRPCErrorRatio = `rfc0021:product_rpc_error:ratio5m`

	// CutoverOrderSagaOutcomeRate is order_saga_outcome_total by outcome:
	// confirmed | failed (pre-capture, voided) | compensated (post-capture,
	// refunded).
	CutoverOrderSagaOutcomeRate = `rfc0021:order_saga_outcome:rate5m`

	// CutoverOrderSagaCompensationRate is compensation runs by step and
	// result — result="failed" is the stuck-money population, whose baseline
	// should be flat zero.
	CutoverOrderSagaCompensationRate = `rfc0021:order_saga_compensation:rate5m`

	// CutoverPaymentProviderDurationP95 is the money-hop SLI: payment
	// provider latency p95 by op (charge/capture/void/refund).
	CutoverPaymentProviderDurationP95 = `rfc0021:payment_provider_duration:p95_5m`

	// CutoverPaymentAuthorizationRate is authorization outcomes/sec
	// (authorized | declined | error).
	CutoverPaymentAuthorizationRate = `rfc0021:payment_authorization:rate5m`
)
