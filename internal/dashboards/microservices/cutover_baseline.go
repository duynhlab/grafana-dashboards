package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// CutoverBaseline builds the "Order Saga & Payment — Cutover Baseline" board
// (UID rfc0021-baseline), ported from the homelab GitOps repo's
// cutover-baseline.json. See the query file for the source path, commit, the
// PrometheusRule dependency, and which panels were dropped.
func CutoverBaseline() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Order Saga & Payment — Cutover Baseline").
		Description("RFC-0021 overhaul baseline: the pre-refactor signals every cutover gate is judged against (CP-0 requires 7 days of data). All queries read rfc0021:* recording rules.").
		Editable(true).
		Tags([]string{"rfc-0021", "baseline", "microservices", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-24h", "now", "1m"))

	b = b.
		// Checkout confirm — read-path cutover gate.
		Panel("cutover-confirm-success-bounce", panels.Series("Confirm success vs bounce (per second)", "ops",
			panels.PromQuery(msqueries.CutoverConfirmSuccessRate, "confirmed"),
			panels.Query("B", msqueries.CutoverConfirmBounceRate, "bounced (requoted)"),
		)).
		Panel("cutover-confirm-p95", panels.SeriesExpr("Confirm handler p95", "s",
			msqueries.CutoverConfirmDurationP95, "p95")).
		Panel("cutover-saga-reserve-stock", panels.SeriesExpr("Saga-side ReserveStock outcomes", "ops",
			msqueries.CutoverOrderStockReservationRate, "{{result}}")).

		// Product stock surface — what phases 2-4 replace.
		Panel("cutover-product-rpc-p95", panels.SeriesExpr("Product gRPC p95 by method", "s",
			msqueries.CutoverProductRPCDurationP95, "{{rpc_method}}")).
		Panel("cutover-product-rpc-error", panels.SeriesExpr("Product gRPC error ratio by method", "percentunit",
			msqueries.CutoverProductRPCErrorRatio, "{{rpc_method}}")).

		// Order saga + payment money hop.
		Panel("cutover-saga-outcomes", panels.SeriesExpr("Saga terminal outcomes", "ops",
			msqueries.CutoverOrderSagaOutcomeRate, "{{outcome}}")).
		Panel("cutover-saga-compensation", panels.SeriesExpr("Compensation runs by step and result", "ops",
			msqueries.CutoverOrderSagaCompensationRate, "{{step}} {{result}}")).
		Panel("cutover-payment-provider", panels.Series("Payment provider p95 by op + authorization outcomes", "s",
			panels.PromQuery(msqueries.CutoverPaymentProviderDurationP95, "p95 {{op}}"),
			panels.Query("B", msqueries.CutoverPaymentAuthorizationRate, "auth {{result}}"),
		))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Checkout confirm — read-path cutover gate",
			panels.GridItem("cutover-confirm-success-bounce", 0, 0, 8, 8),
			panels.GridItem("cutover-confirm-p95", 8, 0, 8, 8),
			panels.GridItem("cutover-saga-reserve-stock", 16, 0, 8, 8),
		),
		panels.Row("Product stock surface — what phases 2-4 replace",
			panels.GridItem("cutover-product-rpc-p95", 0, 0, 12, 8),
			panels.GridItem("cutover-product-rpc-error", 12, 0, 12, 8),
		),
		panels.Row("Order saga + payment money hop",
			panels.GridItem("cutover-saga-outcomes", 0, 0, 8, 8),
			panels.GridItem("cutover-saga-compensation", 8, 0, 8, 8),
			panels.GridItem("cutover-payment-provider", 16, 0, 8, 8),
		),
	))
}
