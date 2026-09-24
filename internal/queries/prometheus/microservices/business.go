package microservices

// PromQL for the "Microservices — Business KPIs" board (UID business-otel),
// ported from the helm-charts grafana-dashboards chart. The metric model is the
// per-service custom business instruments defined by RFC-0017 (payment_*,
// order_*, auth_*, product_*, cart_*, shipment_*, user_*, reviews_*,
// notification_*, checkout_*), plus two legacy OpenTelemetry DB-semconv
// metrics emitted by the product service's cache client.
//
// Normalisation applied while porting: the source board carried a `rate` custom
// variable that parameterised every rate window; it is replaced here by
// $__rate_interval so the window always tracks the panel's own resolution. The
// $__range aggregations below are deliberate and are NOT rate windows — see the
// note on the ratio and total stats.
const (
	// --- Payments -------------------------------------------------------

	// PaymentCardDeclineRatio uses $__range rather than a rate window on
	// purpose: authorization attempts are sparse business events, and a short
	// rate window makes the stat flicker to "No data" between bursts. The
	// `or vector(0)` keeps the numerator defined when nothing was declined.
	PaymentCardDeclineRatio = `(sum(increase(payment_authorization_total{result="declined"}[$__range])) or vector(0)) / sum(increase(payment_authorization_total[$__range]))`

	// PaymentAuthorizationOutcomes splits authorize attempts into
	// authorized / declined / error.
	PaymentAuthorizationOutcomes = `sum by (result) (rate(payment_authorization_total[$__rate_interval]))`

	// PaymentMoneyOperations covers the capture / void / refund lifecycle.
	PaymentMoneyOperations = `sum by (op,result) (rate(payment_operation_total[$__rate_interval]))`

	// PaymentProviderLatencyP50/P95/P99 are the RFC-0017 W2 money-hop SLI:
	// latency of the outbound call to the mockpay provider, per operation.
	// Every business histogram is shown at p50 (the typical call), p95 (the
	// tail an SLO watches) and p99 (the tail an incident shows first).
	PaymentProviderLatencyP50 = `histogram_quantile(0.5, sum by (le,op) (rate(payment_provider_request_duration_seconds_bucket[$__rate_interval])))`
	PaymentProviderLatencyP95 = `histogram_quantile(0.95, sum by (le,op) (rate(payment_provider_request_duration_seconds_bucket[$__rate_interval])))`
	PaymentProviderLatencyP99 = `histogram_quantile(0.99, sum by (le,op) (rate(payment_provider_request_duration_seconds_bucket[$__rate_interval])))`

	// PaymentReconciliationDiscrepancies counts ledger-vs-provider mismatches
	// found by the reconciler. Healthy systems sit at ~0.
	PaymentReconciliationDiscrepancies = `sum by (kind) (rate(payment_reconciliation_discrepancies_total[$__rate_interval]))`

	// --- Orders / Saga --------------------------------------------------

	// OrderSagaOutcomes is the terminal outcome of the order saga per run.
	OrderSagaOutcomes = `sum by (outcome) (rate(order_saga_outcome_total[$__rate_interval]))`

	// OrderFailedCompensations is the money-stuck signal: compensations that
	// themselves failed, so a debit was never reversed. Target 0, which is why
	// the panel colors red at anything above zero. $__range is used because
	// these events are rare; a rate window would show nothing most of the time.
	OrderFailedCompensations = `sum(increase(order_saga_compensation_total{result="failed"}[$__range])) or vector(0)`

	// OrderAverageValueUSD is the average order value over the selected range.
	// order_value_minor_sum / order_value_minor_count yields the mean in MINOR
	// currency units (cents); the trailing / 100 converts to whole dollars,
	// which is why the panel is rendered with the currencyUSD unit.
	OrderAverageValueUSD = `sum(increase(order_value_minor_sum[$__range])) / sum(increase(order_value_minor_count[$__range])) / 100`

	// OrderValueP50USD/P95USD/P99USD place the average: a mean pulled up by a
	// few large baskets reads very differently from one the median agrees
	// with. Same minor-unit division as the average.
	OrderValueP50USD = `histogram_quantile(0.5, sum by (le) (increase(order_value_minor_bucket[$__range]))) / 100`
	OrderValueP95USD = `histogram_quantile(0.95, sum by (le) (increase(order_value_minor_bucket[$__range]))) / 100`
	OrderValueP99USD = `histogram_quantile(0.99, sum by (le) (increase(order_value_minor_bucket[$__range]))) / 100`

	// OrderPaymentActivity is the order -> payment saga activity calls.
	OrderPaymentActivity = `sum by (op,result) (rate(order_payment_activity_total[$__rate_interval]))`

	// OrderSagaStockReservationOutcomes is the reserve-stock STEP OF THE SAGA,
	// recorded by the order service on the SINGULAR metric
	// order_stock_reservation_total. It is a different instrument from the
	// product service's ProductStockReservationOutcomes below, even though both
	// panels are titled "Stock reservation outcomes".
	OrderSagaStockReservationOutcomes = `sum by (result) (rate(order_stock_reservation_total[$__rate_interval]))`

	// OrderCompensationSteps shows which saga step compensated and whether the
	// compensation itself succeeded.
	OrderCompensationSteps = `sum by (step,result) (rate(order_saga_compensation_total[$__rate_interval]))`

	// --- Auth -----------------------------------------------------------

	// AuthRegistrations splits registrations into success / conflict / error.
	AuthRegistrations = `sum by (result) (rate(auth_registrations_total[$__rate_interval]))`

	// AuthRefreshRotations covers rotated / invalid / expired / reuse_detected.
	// The reuse_detected series is the stolen-token replay signal.
	AuthRefreshRotations = `sum by (result) (rate(auth_refresh_operations_total[$__rate_interval]))`

	// AuthFamilyRevocations separates logout-driven from reuse-driven teardown
	// of a refresh-token family.
	AuthFamilyRevocations = `sum by (reason) (rate(auth_family_revocations_total[$__rate_interval]))`

	// AuthPasswordHashLatencyP95 is the bcrypt cost inside auth latency:
	// hash on register, compare on login.
	AuthPasswordHashLatencyP95 = `histogram_quantile(0.95, sum by (le,op) (rate(auth_password_hash_duration_seconds_bucket[$__rate_interval])))`

	// --- Product --------------------------------------------------------

	// ProductCacheHitRatio measures Valkey cache-aside effectiveness
	// (RFC-0017 W2). It is inverted against the usual threshold direction:
	// low is bad. $__range is used because cache reads on this board are
	// sparse enough that a short rate window flickers to "No data".
	ProductCacheHitRatio = `sum(increase(product_cache_gets_total{result="hit"}[$__range])) / sum(increase(product_cache_gets_total{result=~"hit|miss"}[$__range]))`

	// ProductCacheGets splits cache reads into hit / miss / error.
	ProductCacheGets = `sum by (result) (rate(product_cache_gets_total[$__rate_interval]))`

	// ProductStockReservationOutcomes is the PRODUCT-side reservation result on
	// the PLURAL metric product_stock_reservations_total. Do not confuse it with
	// OrderSagaStockReservationOutcomes, which reads the singular
	// order_stock_reservation_total from the order service.
	ProductStockReservationOutcomes = `sum by (result) (rate(product_stock_reservations_total[$__rate_interval]))`

	// ProductCachePoolConns reports the redisotel (go-redis) connection pool by
	// state. db_client_connections_usage is a PRE-1.26 OpenTelemetry DB-semconv
	// metric name; despite the generic "db_client" prefix this describes the
	// VALKEY CACHE pool of the product service, not a PostgreSQL pool. A newer
	// collector/SDK renames it (db.client.connection.count), so this constant
	// must be revisited on upgrade.
	ProductCachePoolConns = `sum by (state) (db_client_connections_usage{service_name="product"})`

	// ProductCacheConnUseTimeP95 is how long a checked-out cache connection is
	// held per operation. db_client_connections_use_time_milliseconds_bucket is
	// the matching pre-1.26 OTel DB-semconv histogram, again the VALKEY CACHE
	// pool and not Postgres, and it is reported in milliseconds.
	ProductCacheConnUseTimeP95 = `histogram_quantile(0.95, sum by (le) (rate(db_client_connections_use_time_milliseconds_bucket{service_name="product"}[$__rate_interval])))`

	// --- Cart -----------------------------------------------------------

	// CartItemsAdded is the top of the purchase funnel: add-to-cart attempts
	// by outcome (added vs rejected_invalid_qty).
	CartItemsAdded = `sum by (result) (rate(cart_items_added_total[$__rate_interval]))`

	// CartCleared separates user-initiated from checkout-driven clears.
	CartCleared = `sum by (source) (rate(cart_cleared_total[$__rate_interval]))`

	// CartSnapshotRequests are cart snapshot reads on the checkout east-west path.
	CartSnapshotRequests = `sum by (result) (rate(cart_snapshot_requests_total[$__rate_interval]))`

	// --- Shipping -------------------------------------------------------

	// ShipmentsCreated is CreateShipment by outcome, step 2 of the fulfillment
	// saga. Idempotent replays also report ok.
	ShipmentsCreated = `sum by (outcome) (rate(shipment_created_total[$__rate_interval]))`

	// ShipmentsCancelled is CancelShipment by outcome, i.e. how often the saga
	// compensates.
	ShipmentsCancelled = `sum by (outcome) (rate(shipment_cancelled_total[$__rate_interval]))`

	// ShipmentLookups counts shipment lookups and whether the record was found.
	ShipmentLookups = `sum by (kind,found) (rate(shipment_lookup_total[$__rate_interval]))`

	// --- User -----------------------------------------------------------

	// UserProfileUpdates separates successful writes from authz rejections.
	UserProfileUpdates = `sum by (result) (rate(user_profile_updated_total[$__rate_interval]))`

	// UserProfileLookups splits profile reads by audience (public / private /
	// internal) and hit/miss.
	UserProfileLookups = `sum by (audience,found) (rate(user_profile_lookup_total[$__rate_interval]))`

	// --- Review ---------------------------------------------------------

	// ReviewsRatingBuckets feeds the rating-distribution heatmap. The series
	// arrive pre-bucketed by le (stars 1-5), so the panel must consume them
	// with a heatmap-format target and calculation disabled.
	ReviewsRatingBuckets = `sum by (le) (increase(reviews_rating_bucket[$__rate_interval]))`

	// ReviewsAverageRating is the mean star rating of reviews created in the
	// selected range. $__range rather than a rate window: review creation is
	// sparse and the stat would otherwise read "No data" between submissions.
	ReviewsAverageRating = `sum(increase(reviews_rating_sum[$__range])) / sum(increase(reviews_rating_count[$__range]))`

	// ReviewsDuplicateRejected counts reviews rejected as duplicates.
	ReviewsDuplicateRejected = `sum(rate(reviews_duplicate_rejected_total[$__rate_interval]))`

	// ReviewsGRPCTruncated counts gRPC review reads that were silently
	// truncated, which is a correctness smell rather than a load signal.
	ReviewsGRPCTruncated = `sum(rate(grpc_reviews_truncated_total[$__rate_interval]))`

	// --- Notification ---------------------------------------------------

	// NotificationReads splits single from bulk mark-as-read.
	NotificationReads = `sum by (mode) (rate(notification_read_total[$__rate_interval]))`

	// NotificationSendLatencyP50/P95/P99 are delivery latency per channel.
	NotificationSendLatencyP50 = `histogram_quantile(0.5, sum by (le,channel) (rate(notification_send_duration_seconds_bucket[$__rate_interval])))`
	NotificationSendLatencyP95 = `histogram_quantile(0.95, sum by (le,channel) (rate(notification_send_duration_seconds_bucket[$__rate_interval])))`
	NotificationSendLatencyP99 = `histogram_quantile(0.99, sum by (le,channel) (rate(notification_send_duration_seconds_bucket[$__rate_interval])))`

	// --- Checkout -------------------------------------------------------

	// CheckoutSessionsConfirmed is the funnel exit: sessions confirmed into an
	// order. $__range keeps the count stable over sparse confirmations, and the
	// `or vector(0)` renders 0 instead of "No data" on a quiet range.
	CheckoutSessionsConfirmed = `sum(increase(checkout_sessions_confirmed_total[$__range])) or vector(0)`

	// CheckoutSessionsExpired explains why sessions expired (idle / TTL).
	CheckoutSessionsExpired = `sum by (reason) (rate(checkout_sessions_expired_total[$__rate_interval]))`

	// CheckoutPromoRedeemed counts promo redemptions at the authoritative
	// confirm gate.
	CheckoutPromoRedeemed = `sum(rate(checkout_promo_redeemed_total[$__rate_interval]))`

	// CheckoutPromoRejected is the other half of the promo funnel, by reason.
	CheckoutPromoRejected = `sum by (reason) (rate(checkout_promo_rejected_total[$__rate_interval]))`

	// CheckoutPriceChanged counts quotes that drifted between session start and
	// confirm.
	CheckoutPriceChanged = `sum(rate(checkout_price_changed_total[$__rate_interval]))`

	// CheckoutConfirmLatencyP50/P95/P99 are end-to-end checkout confirm
	// duration (the RFC-0015 exemplar path).
	CheckoutConfirmLatencyP50 = `histogram_quantile(0.5, sum by (le) (rate(checkout_confirm_duration_seconds_bucket[$__rate_interval])))`
	CheckoutConfirmLatencyP95 = `histogram_quantile(0.95, sum by (le) (rate(checkout_confirm_duration_seconds_bucket[$__rate_interval])))`
	CheckoutConfirmLatencyP99 = `histogram_quantile(0.99, sum by (le) (rate(checkout_confirm_duration_seconds_bucket[$__rate_interval])))`
)
