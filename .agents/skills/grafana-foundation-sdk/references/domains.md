# Domains

A domain is a Grafana folder plus the Go packages that feed it. Folder title equals the
`standards.Folder*` constant; the CR name is the lowercase slug.

| Folder | Packages | Dashboards (UID) | Alert group |
|---|---|---|---|
| `Kubernetes` | `queries/prometheus/kubernetes`, `dashboards/kubernetes`, `alerts/kubernetes.go` | `kubernetes-cluster-overview`, `keda` | `kubernetes` |
| `Databases` | `queries/prometheus/postgres`, `dashboards/postgres`, `alerts/postgres.go` | `pg-io-waits`, `pg-maintenance`, `pg-query-performance`, `pg-exporter-instance`, `pgdog` | `databases` |
| `Observability` | `queries/prometheus/observability`, `dashboards/observability` | `temporal-worker`, `otel-collector-health` | none yet |
| `Microservices` | `queries/prometheus/microservices`, `dashboards/microservices` | `microservices-monitoring-001-otel`, `business-otel`, `red-spanmetrics`, `rfc0021-baseline`, `inventory-overview` | none yet |
| `Platform` | `queries/prometheus/platform`, `dashboards/platform` | `cert-manager`, `keycloak-identity` | none yet |
| `API Gateway` | `queries/prometheus/gateway`, `dashboards/gateway` | `eg-edge` | none yet |

The domain slug and the folder title differ wherever the folder reads better to a human
than the package does: `postgres` -> `Databases`, `gateway` -> `API Gateway`.

Adding a domain: add a `Domain*` constant in `internal/standards/domains.go` and a
`Folder*` constant in `internal/standards/folders.go`, create the packages, register
resources with both, extend the `want` map in `TestDomainFolders`. `cmd/generate` does
not change.

## Metric models, check before writing PromQL

The Databases folder mixes two exporter models. Do not assume one for the other.

| Model | Metrics | Key labels | Used by |
|---|---|---|---|
| CloudNativePG (CNPG) | `cnpg_*` (`cnpg_pg_stat_statements_*`, `cnpg_backends_waiting_total`, `cnpg_pg_stat_io_*`) | `cnpg_io_cluster`, `pod`, `datname` | `pg-io-waits`, `pg-maintenance`, `pg-query-performance` |
| Pigsty `postgres_exporter` | `pg_*` raw plus `pg:*` recording rules | `ins`, `cls`, `datname` | `pg-exporter-instance` |
| PGDog | `pgdog_*` | `host`, `port`, `shard`, `role`, `database`, `user` | `pgdog` |
| kube-state-metrics + cAdvisor | `kube_*`, `container_*`, `kubelet_*` | `namespace`, `pod`, `node`, `persistentvolumeclaim` | `kubernetes-cluster-overview` |
| Temporal SDK | `temporal_workflow_*_total`, `temporal_activity_*` | `namespace`, `service_name`, `task_queue`, `workflow_type`, `activity_type` | `temporal-worker` |
| Temporal server | `service_requests`, `service_error_with_type`, `persistence_requests`, `persistence_error_with_type` | `job=~".*temporal.*"`, `service_name` | `temporal-worker` (Server row) |
| KEDA | `keda_scaler_*`, `keda_scaled_object_errors_total`, plus kube-state-metrics HPA | `exported_namespace`, `scaledObject`, `scaler` | `keda` |
| cert-manager | `certmanager_certificate_*`, `certmanager_controller_sync_*`, `workqueue_*` | `name`, `namespace`, `condition`, `controller` | `cert-manager` |
| Keycloak on Quarkus | `http_server_requests_seconds_*`, `keycloak_user_events_total`, `agroal_*`, `jvm_*` | `uri`, `outcome`, `event`, `realm`, `client_id` | `keycloak-identity` |
| Envoy + Envoy Gateway | `envoy_http_downstream_*`, `envoy_cluster_upstream_*`, `watchable_*` | `envoy_cluster_name`, `envoy_response_code_class`, `runner` | `eg-edge` |
| OTel Collector internals | `otelcol_receiver_*`, `otelcol_exporter_*`, `otelcol_processor_batch_*` | `receiver`, `exporter`, `data_type` | `otel-collector-health` |
| spanmetrics connector | `spanmetrics_calls_total`, `spanmetrics_duration_milliseconds_bucket` | `service_name`, `span_kind`, `status_code` | `red-spanmetrics` |
| RFC-0021 recording rules | `rfc0021:*`, `inventory:*` | varies per rule | `rfc0021-baseline`, `inventory-overview` |
| OTel semconv (Go SDK) | `http_server_*`, `rpc_server_*`, `rpc_client_*`, `go_memory_*`, `go_goroutine_count`, `db_client_*`, `pgxpool_*` | `service_name`, `http_route`, `http_request_method`, `http_response_status_code`, `rpc_method`, `rpc_response_status_code`, `server_address` | `microservices-monitoring-001-otel` |
| Per-service business instruments | `payment_*`, `order_*`, `auth_*`, `product_*`, `cart_*`, `shipment_*`, `user_*`, `reviews_*`, `notification_*`, `checkout_*` | `result`, `outcome`, `op`, `reason`, `step`, `kind`, `found`, `channel`, `mode` | `business-otel` |

Shared CNPG template-variable definitions live in
`internal/queries/prometheus/postgres/variables.go`. Reuse them for a new CNPG board.

PostgreSQL boards are validated against PostgreSQL 18 (`pg_stat_io` exists).

## Section vocabulary

Pick only the sections that earn a place on the board.

- Kubernetes: Overview, Workload Health, Resource Utilization, Storage, Network, Reliability
- PostgreSQL: Throughput, Latency, Connections, Locks, Checkpointer, Autovacuum/Bloat, IO, Top statements
- Pooler: Clients, Servers, Transactions, Queries, Traffic, Cache, Mirroring
- Worker / RED: Rate, Errors, Duration per workflow or activity, task-queue backlog

## Porting from legacy JSON

Legacy boards under `dashboard/` are read-only input. A source board may also live in
another repo, for example the helm-charts `grafana-dashboards` chart. Treat it the same
way, never edit it, and record its path and commit in the Go doc comment of the query
file so the next reader can diff against the original.

1. Read the JSON for title, template variables, panel titles, PromQL, units, and grid
   positions. Do not edit the file.
2. Extract PromQL into constants under `internal/queries/prometheus/<domain>/`. Normalise
   datasource variables to the logical `prometheus` datasource. Keep the metric model.
3. Rewrite the board with the `panels.*` helpers. Do not aim for v1 schema parity; drop
   panels that only existed because of v1 limitations.
4. Keep the legacy UID when the board replaces the legacy one, otherwise choose a new slug.
5. Add the contract test and register. Add alerts when the board has operational thresholds.

Already ported: `kubernetes-cluster-overview`, `pg-io-waits`, `pg-maintenance`,
`pg-query-performance`, `pg-exporter-instance`, `pgdog`, `temporal`, and from the
helm-charts chart, `microservices-dashboard-otel` and `business-otel`.

Porting candidates still in `dashboard/`: `kubernetes/kubernetes-nginx-ingress`,
`kubernetes/kong-dashboard`, `redis/redis`, `redis/redis-exporter`,
`postgresql/pgbouncer`, `postgresql/cloudnative-pg-cluster`, `postgresql/pg-exporter-self`,
`observability/tempo-observability-dashboard`, `observability/pyrra_list`,
`observability/pyrra_detail`, `observability/slo-detail`, `observability/high-level-sloth-slos`,
`messaging/amazonmq/amazonmq-rabbitmq`. Redis, Kong, ingress, and RabbitMQ need a new
folder decision first (for example `Databases` for Redis, a new `Ingress` or `Messaging`).
