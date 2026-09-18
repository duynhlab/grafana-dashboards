package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// BusinessOTel builds the RFC-0017 business-KPI board (UID business-otel),
// ported from the helm-charts grafana-dashboards chart. One row per service
// domain, in the order a purchase travels through the platform.
//
// The board carries no template variables. The source had two — a datasource
// picker and a custom `rate` window — and both were dropped on the port: the
// helpers bind the logical prometheus datasource, and every rate window is now
// $__rate_interval. The two empty text panels the source used as layout spacers
// are dropped as well, and the rows are re-laid so the 24-column grid is filled.
//
// Only the first two rows (Payments, Orders / Saga) open expanded; the eight
// per-service rows below them are collapsed, matching the source.
func BusinessOTel() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Microservices — Business KPIs").
		Description("RFC-0017 business KPIs, one row per service domain, ported from the helm-charts grafana-dashboards chart.").
		Editable(true).
		Tags([]string{"microservices", "business", "otel", "kpi", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-30m", "now", "1m"))

	b = businessPayments(b)
	b = businessOrders(b)
	b = businessAuth(b)
	b = businessProduct(b)
	b = businessCart(b)
	b = businessShipping(b)
	b = businessUser(b)
	b = businessReview(b)
	b = businessNotification(b)
	b = businessCheckout(b)

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Payments",
			panels.GridItem("pay-decline-rate", 0, 0, 6, 8),
			panels.GridItem("pay-authorization", 6, 0, 9, 8),
			panels.GridItem("pay-operations", 15, 0, 9, 8),
			panels.GridItem("pay-provider-latency", 0, 8, 12, 8),
			panels.GridItem("pay-reconciliation", 12, 8, 12, 8),
		),
		panels.Row("Orders / Saga",
			panels.GridItem("order-saga-outcomes", 0, 0, 12, 8),
			panels.GridItem("order-failed-compensations", 12, 0, 6, 8),
			panels.GridItem("order-average-value", 18, 0, 6, 8),
			panels.GridItem("order-payment-activity", 0, 8, 12, 8),
			panels.GridItem("order-stock-reservation", 12, 8, 12, 8),
			panels.GridItem("order-compensation-steps", 0, 16, 24, 8),
		),
		panels.CollapsedRow("Auth",
			panels.GridItem("auth-registrations", 0, 0, 12, 8),
			panels.GridItem("auth-refresh-rotations", 12, 0, 12, 8),
			panels.GridItem("auth-family-revocations", 0, 8, 12, 8),
			panels.GridItem("auth-bcrypt-latency", 12, 8, 12, 8),
		),
		panels.CollapsedRow("Product",
			panels.GridItem("product-cache-hit-ratio", 0, 0, 6, 8),
			panels.GridItem("product-cache-gets", 6, 0, 9, 8),
			panels.GridItem("product-stock-reservation", 15, 0, 9, 8),
			panels.GridItem("product-cache-pool", 0, 8, 12, 8),
			panels.GridItem("product-cache-use-time", 12, 8, 12, 8),
		),
		panels.CollapsedRow("Cart",
			panels.GridItem("cart-items-added", 0, 0, 12, 8),
			panels.GridItem("cart-cleared", 12, 0, 12, 8),
			panels.GridItem("cart-snapshot-requests", 0, 8, 24, 8),
		),
		panels.CollapsedRow("Shipping",
			panels.GridItem("shipping-created", 0, 0, 12, 8),
			panels.GridItem("shipping-cancelled", 12, 0, 12, 8),
			panels.GridItem("shipping-lookups", 0, 8, 24, 8),
		),
		panels.CollapsedRow("User",
			panels.GridItem("user-profile-updates", 0, 0, 12, 8),
			panels.GridItem("user-profile-lookups", 12, 0, 12, 8),
		),
		panels.CollapsedRow("Review",
			panels.GridItem("review-rating-distribution", 0, 0, 18, 8),
			panels.GridItem("review-average-rating", 18, 0, 6, 8),
			panels.GridItem("review-quality-signals", 0, 8, 24, 8),
		),
		panels.CollapsedRow("Notification",
			panels.GridItem("notification-reads", 0, 0, 12, 8),
			panels.GridItem("notification-send-latency", 12, 0, 12, 8),
		),
		panels.CollapsedRow("Checkout",
			panels.GridItem("checkout-sessions-confirmed", 0, 0, 6, 8),
			panels.GridItem("checkout-sessions-expired", 6, 0, 9, 8),
			panels.GridItem("checkout-promo", 15, 0, 9, 8),
			panels.GridItem("checkout-price-changed", 0, 8, 12, 8),
			panels.GridItem("checkout-confirm-latency", 12, 8, 12, 8),
		),
	))
}
