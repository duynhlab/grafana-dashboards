package microservices

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"

	"github.com/duynhlab/grafana-dashboards/internal/panels"
	msqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

// InventoryOverview builds the "Inventory Service — Stock Authority" board
// (UID inventory-overview), ported from the homelab GitOps repo's
// inventory.json. See the query file for the source path, commit, and the
// PrometheusRule this board depends on.
//
// inventory-service is the platform's sole stock authority (gRPC-only, no
// HTTP edge route) with three live callers since RFC-0021 phase 4: the order
// saga (reserve/commit/release), checkout (availability), and product
// /details. The board has no template variables — there is exactly one
// inventory-service deployment to look at.
func InventoryOverview() cog.Builder[dashboardv2.Dashboard] {
	b := dashboardv2.NewDashboardBuilder("Inventory Service — Stock Authority").
		Description("inventory-service, the platform's sole stock authority. Reservation FSM outcomes, availability checks, gRPC RED, and DB latency. Reads inventory:* recording rules where available. Landed at RFC-0021 phase 1; live since phase 4 with three callers -- the order saga, checkout, and product /details.").
		Editable(true).
		Tags([]string{"rfc-0021", "inventory", "generated", "foundation-sdk"}).
		TimeSettings(standards.TimeSettings("now-24h", "now", "1m"))

	b = b.
		Panel("inventory-readme", panels.Text("Read this panel before the flat lines", msqueries.InventoryReadme)).
		Panel("inventory-reservation-outcomes", panels.SeriesExpr("Reservation outcomes by operation", "ops",
			msqueries.InventoryReservationRate5m, "{{operation}} {{outcome}}")).
		Panel("inventory-check-outcomes", panels.SeriesExpr("Availability check outcomes", "ops",
			msqueries.InventoryCheckRate5m, "{{outcome}}")).
		Panel("inventory-grpc-rate", panels.SeriesExpr("gRPC request rate by method", "ops",
			msqueries.InventoryGRPCRequestRateByMethod, "{{rpc_method}}")).
		Panel("inventory-grpc-error-ratio", panels.SeriesExpr("gRPC error ratio by method", "percentunit",
			msqueries.InventoryRPCErrorRatio5m, "{{rpc_method}}")).
		Panel("inventory-grpc-p95", panels.SeriesExpr("gRPC p95 by method", "s",
			msqueries.InventoryRPCDurationP95_5m, "{{rpc_method}}")).
		Panel("inventory-db-p95", panels.SeriesExpr("DB query p95", "s",
			msqueries.InventoryDBOperationDurationP95_5m, "p95"))

	return b.RowsLayout(panels.RowsLayout(
		panels.Row("Reservation FSM — stock authority",
			panels.GridItem("inventory-readme", 0, 0, 24, 3),
			panels.GridItem("inventory-reservation-outcomes", 0, 3, 12, 8),
			panels.GridItem("inventory-check-outcomes", 12, 3, 12, 8),
		),
		panels.Row("gRPC RED — InventoryService",
			panels.GridItem("inventory-grpc-rate", 0, 0, 8, 8),
			panels.GridItem("inventory-grpc-error-ratio", 8, 0, 8, 8),
			panels.GridItem("inventory-grpc-p95", 16, 0, 8, 8),
		),
		panels.Row("Database (product-db via PgDog)",
			panels.GridItem("inventory-db-p95", 0, 0, 12, 8),
		),
	))
}
