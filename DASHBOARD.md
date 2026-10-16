# Dashboard & alert catalog

The inventory this repository generates: 19 dashboards across 6 folders, and 4
alert rules. For what the repository is and how to build a board, see
[`README.md`](README.md). This file is the catalog; keep it in step with
`internal/registry/` whenever a dashboard, alert, or folder is added, removed, or
renamed.

## Dashboards

Every query uses a logical datasource resolved at deploy time: `prometheus`
(homelab maps it to VictoriaMetrics) or, for the pg_stat_ch board, `clickhouse`
(homelab's ClickHouse datasource uid). *Elements* is the number of panels and
other elements in the generated spec.

| Folder | UID | Title | Elements | What it shows |
|---|---|---|---:|---|
| Kubernetes | `kubernetes-cluster-overview` | Kubernetes Cluster Overview | 24 | Cluster health by the USE method: node and pod counts, workload health, utilization, persistent-volume storage |
| Kubernetes | `kubernetes-workloads` | Kubernetes Workloads | 11 | Per-workload and per-pod CPU, memory, network and reliability, filtered by namespace and owner |
| Kubernetes | `keda` | KEDA — Worker Autoscaling | 17 | What the Temporal scaler computed per worker version, what the HPA did, the backlog it scales on, KEDA's own health |
| Databases | `pg-io-waits` | PostgreSQL — IO & Waits (pg_stat_io) | 10 | IO attribution from PostgreSQL 18 `pg_stat_io` and sampled wait classes |
| Databases | `pg-maintenance` | PostgreSQL — Maintenance (CNPG) | 15 | Locks and blocking, checkpointer, autovacuum and bloat, long transactions, VACUUM progress |
| Databases | `pg-query-performance` | PostgreSQL — Query Performance (pg_stat_statements) | 14 | Throughput, latency, cache efficiency, block I/O, top statements |
| Databases | `pg-exporter-instance` | PG Exporter Instance | 61 | Remote/RDS instances via the Pigsty `postgres_exporter` model (`pg_*` metrics, `pg:*` rules, `ins`/`cls` labels — not `cnpg_*`) |
| Databases | `pgdog` | PGDog | 30 | PGDog pooler: clients, servers, transactions, queries, traffic, prepared statements, mirroring |
| Databases | `postgresql-query-telemetry` | PostgreSQL — Query telemetry (pg_stat_ch) | 34 | Per-execution statement events from the pg_stat_ch extension, read from ClickHouse (logical datasource `clickhouse`): p50/p95/p99 over time, statements/s by operation, database and `application_name`, errors (ERROR and above) and warnings by SQLSTATE, WAL and dirtied-block writes, top statements, work_mem spills, a per-`query_id` drill-down (textbox), load and error rate by `application_name`/role from the 1-minute rollup, the roles whose statements arrive without an `application_name`, JIT and parallel-worker shortfall (collapsed), plus the exporter's own health from Prometheus. A *Sessions* variable hides CNPG's own sessions by default |
| Observability | `temporal-worker` | Temporal — Workflows & Activities | 21 | Both halves of Temporal: SDK/worker metrics over OTLP and server metrics (gRPC by role, persistence, task-queue backlog) |
| Observability | `otel-collector-health` | OTel Collector Health | 10 | Collector pipeline throughput, exporter fan-out, self resource usage |
| Microservices | `microservices-monitoring-001-otel` | Microservices (OTel) | 41 | 1:1 port of the helm-charts board: Go services RED, Go runtime and HTTP I/O, east-west gRPC by method and per callee, and the `otelpgx` DB client and pool, from OpenTelemetry semantic-convention metrics keyed on `service_name`. Adds deploy markers, a running-versions table, and links from per-service series to their traces and logs |
| Microservices | `business-otel` | Microservices — Business KPIs | 38 | Business KPIs, one row per domain: payments, orders and saga, auth, product, cart, shipping, user, review, notification, checkout |
| Microservices | `red-spanmetrics` | Microservices — RED Span Metrics | 8 | RED per service from the Collector's spanmetrics connector — distinct from the board above, which reads `http_server_*`/`rpc_server_*` directly |
| Microservices | `rfc0021-baseline` | Order Saga & Payment — Cutover Baseline | 8 | The order-saga and payment signals every cutover gate is judged against |
| Microservices | `inventory-overview` | Inventory Service — Stock Authority | 7 | Reservation FSM outcomes, availability checks, gRPC RED, DB latency |
| Platform | `cert-manager` | cert-manager | 16 | Certificate readiness and time to expiry, controller sync, work queues, ACME traffic |
| Platform | `keycloak-identity` | Keycloak — Identity | 18 | Auth event rates and failure ratio, realm and token endpoint latency against a 250 ms SLO, JVM and DB pool |
| API Gateway | `eg-edge` | Envoy Gateway — Edge Overview | 21 | Edge golden signals on the data plane, control-plane health, process-level resources |

Notes:

- **Recording rules required.** `rfc0021-baseline` and `inventory-overview` read
  `rfc0021:*` and `inventory:*` recording rules, not raw metrics. They live in
  homelab under
  `kubernetes/infra/configs/observability/metrics/prometheusrules/microservices/`;
  without them the two boards are empty.
- **PostgreSQL version.** The PostgreSQL boards are tested on PostgreSQL 18.
- **Provenance.** Ported boards record the source path and the commit they were
  read at in their query file. `microservices-monitoring-001-otel` and
  `business-otel` came from the `duynhlab/helm-charts` `grafana-dashboards` chart,
  `pgdog` from Grafana.com dashboard 24583, and the rest from JSON that homelab
  used to vendor.

## Folders

| Title | UID | Dashboards | Alert groups |
|---|---|---|---|
| Kubernetes | `kubernetes` | `kubernetes-cluster-overview`, `kubernetes-workloads`, `keda` | `kubernetes` |
| Databases | `databases` | `pg-io-waits`, `pg-maintenance`, `pg-query-performance`, `pg-exporter-instance`, `pgdog`, `postgresql-query-telemetry` | `databases` |
| Observability | `observability` | `temporal-worker`, `otel-collector-health` | — |
| Microservices | `microservices` | `microservices-monitoring-001-otel`, `business-otel`, `red-spanmetrics`, `rfc0021-baseline`, `inventory-overview` | — |
| Platform | `platform` | `cert-manager`, `keycloak-identity` | — |
| API Gateway | `api-gateway` | `eg-edge` | — |

The folder UID is the slug of its title. Other dashboards can share these folders
by UID; homelab files its Envoy Gateway, service-graph and Vector boards into
`api-gateway`, `microservices` and `observability`.

## Alert rules

| UID | Title | Group / folder | Severity | For | Linked dashboard | Fires when |
|---|---|---|---|---|---|---|
| `kubernetes_crashlooping_pods` | Kubernetes crashlooping pods | `kubernetes` / Kubernetes | warning | 5m | `kubernetes-cluster-overview` | any container is waiting in `CrashLoopBackOff` |
| `kubernetes_pending_pods` | Kubernetes pending pods | `kubernetes` / Kubernetes | warning | 15m | `kubernetes-cluster-overview` | any pod is `Pending` |
| `kubernetes_pvcs_at_risk` | Kubernetes PVCs above 80% used | `kubernetes` / Kubernetes | warning | 15m | `kubernetes-cluster-overview` | any PVC is more than 80% full |
| `postgres_backends_waiting` | PostgreSQL backends waiting | `databases` / Databases | warning | 10m | `pg-io-waits` | CNPG reports waiting backends |

Every rule shares one shape: a 1-minute group interval, an instant query over the
last 10 minutes against datasource UID `prometheus`, a threshold of `> 0`, and
`NoData` / `Error` as the no-data and error states. Each rule carries the
`dashboard_uid` annotation of the board that explains it.
