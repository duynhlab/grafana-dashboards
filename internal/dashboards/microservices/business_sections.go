package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
)

// The business-KPI board spans ten service domains; panel registration is split
// by domain to keep each function readable. Every function registers its panels
// on the shared builder and returns it for chaining. Layout is assembled in
// business.go.

func businessPayments(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("pay-decline-rate", panels.StatValue("Card decline rate", "percentunit", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("yellow", panels.Ptr(0.1)),
			panels.Thr("red", panels.Ptr(0.3)),
		}, msqueries.PaymentCardDeclineRatio, "Card decline rate")).
		Panel("pay-authorization", panels.SeriesExpr("Authorization outcomes", "reqps", msqueries.PaymentAuthorizationOutcomes, "{{result}}")).
		Panel("pay-operations", panels.SeriesExpr("Money operations by op & result", "reqps", msqueries.PaymentMoneyOperations, "{{op}} · {{result}}")).
		Panel("pay-provider-latency", panels.SeriesQuantiles("Provider-hop latency p50 / p95 / p99 (mockpay)", "s", "{{op}} ", msqueries.PaymentProviderLatencyP50, msqueries.PaymentProviderLatencyP95, msqueries.PaymentProviderLatencyP99)).
		Panel("pay-reconciliation", panels.SeriesExpr("Reconciliation discrepancies", "ops", msqueries.PaymentReconciliationDiscrepancies, "{{kind}}"))
}

func businessOrders(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("order-saga-outcomes", panels.SeriesExpr("Saga outcomes", "reqps", msqueries.OrderSagaOutcomes, "{{outcome}}")).
		// Base green with a red step at 0.0001: any non-zero count of failed
		// compensations means money is stuck, so the tile goes red immediately.
		Panel("order-failed-compensations", panels.StatValue("Failed compensations (stuck money)", "short", []panels.Threshold{
			panels.Thr("green", nil),
			panels.Thr("red", panels.Ptr(0.0001)),
		}, msqueries.OrderFailedCompensations, "Failed compensations (stuck money)")).
		Panel("order-average-value", panels.StatValues("Order value — average, p50 / p95 / p99", "currencyUSD", []panels.Threshold{
			panels.Thr("green", nil),
		},
			panels.Query("A", msqueries.OrderAverageValueUSD, "Average"),
			panels.Query("B", msqueries.OrderValueP50USD, "p50"),
			panels.Query("C", msqueries.OrderValueP95USD, "p95"),
			panels.Query("D", msqueries.OrderValueP99USD, "p99"),
		)).
		Panel("order-payment-activity", panels.SeriesExpr("Payment activity by op & result", "reqps", msqueries.OrderPaymentActivity, "{{op}} · {{result}}")).
		Panel("order-stock-reservation", panels.SeriesExpr("Stock reservation outcomes", "reqps", msqueries.OrderSagaStockReservationOutcomes, "{{result}}")).
		Panel("order-compensation-steps", panels.SeriesExpr("Compensation steps", "reqps", msqueries.OrderCompensationSteps, "{{step}} · {{result}}"))
}

func businessAuth(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("auth-registrations", panels.SeriesExpr("Registrations by outcome", "reqps", msqueries.AuthRegistrations, "{{result}}")).
		Panel("auth-refresh-rotations", panels.SeriesExpr("Refresh rotations by outcome", "reqps", msqueries.AuthRefreshRotations, "{{result}}")).
		Panel("auth-family-revocations", panels.SeriesExpr("Token-family revocations by reason", "reqps", msqueries.AuthFamilyRevocations, "{{reason}}")).
		Panel("auth-bcrypt-latency", panels.SeriesExpr("bcrypt latency p95", "s", msqueries.AuthPasswordHashLatencyP95, "{{op}}"))
}

func businessProduct(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		// Inverted thresholds: a LOW hit ratio is the bad state, so the base
		// step is red and green only starts at 0.8.
		Panel("product-cache-hit-ratio", panels.StatValue("Cache hit ratio", "percentunit", []panels.Threshold{
			panels.Thr("red", nil),
			panels.Thr("yellow", panels.Ptr(0.5)),
			panels.Thr("green", panels.Ptr(0.8)),
		}, msqueries.ProductCacheHitRatio, "Cache hit ratio")).
		Panel("product-cache-gets", panels.SeriesExpr("Cache gets by result", "reqps", msqueries.ProductCacheGets, "{{result}}")).
		Panel("product-stock-reservation", panels.SeriesExpr("Stock reservation outcomes", "reqps", msqueries.ProductStockReservationOutcomes, "{{result}}")).
		Panel("product-cache-pool", panels.SeriesExpr("Cache pool usage (Valkey)", "short", msqueries.ProductCachePoolConns, "{{state}}")).
		Panel("product-cache-use-time", panels.SeriesExpr("Cache conn use-time p95 (Valkey)", "ms", msqueries.ProductCacheConnUseTimeP95, "p95"))
}

func businessCart(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("cart-items-added", panels.SeriesExpr("Items added by result", "reqps", msqueries.CartItemsAdded, "{{result}}")).
		Panel("cart-cleared", panels.SeriesExpr("Carts cleared by source", "reqps", msqueries.CartCleared, "{{source}}")).
		Panel("cart-snapshot-requests", panels.SeriesExpr("Snapshot requests by result", "reqps", msqueries.CartSnapshotRequests, "{{result}}"))
}

func businessShipping(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("shipping-created", panels.SeriesExpr("Shipments created by outcome", "reqps", msqueries.ShipmentsCreated, "{{outcome}}")).
		Panel("shipping-cancelled", panels.SeriesExpr("Shipments cancelled by outcome", "reqps", msqueries.ShipmentsCancelled, "{{outcome}}")).
		Panel("shipping-lookups", panels.SeriesExpr("Lookups by kind & found", "reqps", msqueries.ShipmentLookups, "{{kind}} · found={{found}}"))
}

func businessUser(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("user-profile-updates", panels.SeriesExpr("Profile updates by result", "reqps", msqueries.UserProfileUpdates, "{{result}}")).
		Panel("user-profile-lookups", panels.SeriesExpr("Profile lookups by audience & found", "reqps", msqueries.UserProfileLookups, "{{audience}} · found={{found}}"))
}

func businessReview(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("review-rating-distribution", panels.Heatmap("Rating distribution", "short", "Oranges",
			panels.QueryHeatmap("A", msqueries.ReviewsRatingBuckets, "{{le}}"),
		)).
		Panel("review-average-rating", panels.StatValue("Average rating", "short", []panels.Threshold{
			panels.Thr("green", nil),
		}, msqueries.ReviewsAverageRating, "Average rating")).
		Panel("review-quality-signals", panels.Series("Quality signals", "ops",
			panels.PromQuery(msqueries.ReviewsDuplicateRejected, "duplicate rejected"),
			panels.Query("B", msqueries.ReviewsGRPCTruncated, "gRPC truncated"),
		))
}

func businessNotification(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("notification-reads", panels.SeriesExpr("Reads by mode", "reqps", msqueries.NotificationReads, "{{mode}}")).
		Panel("notification-send-latency", panels.SeriesQuantiles("Send latency p50 / p95 / p99 by channel", "s", "{{channel}} ", msqueries.NotificationSendLatencyP50, msqueries.NotificationSendLatencyP95, msqueries.NotificationSendLatencyP99))
}

func businessCheckout(b *dashboardv2.DashboardBuilder) *dashboardv2.DashboardBuilder {
	return b.
		Panel("checkout-sessions-confirmed", panels.StatValue("Sessions confirmed", "short", []panels.Threshold{
			panels.Thr("green", nil),
		}, msqueries.CheckoutSessionsConfirmed, "Confirmed rate")).
		Panel("checkout-sessions-expired", panels.SeriesExpr("Expirations by reason", "ops", msqueries.CheckoutSessionsExpired, "{{reason}}")).
		Panel("checkout-promo", panels.Series("Promo redeemed vs rejected", "ops",
			panels.PromQuery(msqueries.CheckoutPromoRedeemed, "redeemed"),
			panels.Query("B", msqueries.CheckoutPromoRejected, "rejected · {{reason}}"),
		)).
		Panel("checkout-price-changed", panels.SeriesExpr("Price-changed on confirm", "ops", msqueries.CheckoutPriceChanged, "price changed")).
		Panel("checkout-confirm-latency", panels.SeriesQuantiles("Confirm latency p50 / p95 / p99", "s", "", msqueries.CheckoutConfirmLatencyP50, msqueries.CheckoutConfirmLatencyP95, msqueries.CheckoutConfirmLatencyP99))
}
