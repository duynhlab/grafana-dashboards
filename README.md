# grafana-dashboards

> **⚠️ DEPRECATED — this repo is being retired.** The RFC-0017 platform boards
> (microservices OTel + Business KPIs) now live in
> [`duynhlab/helm-charts`](https://github.com/duynhlab/helm-charts) as the
> `grafana-dashboards` chart (dashboards-as-code, consumed by homelab via
> `GrafanaDashboard.configMapRef`). The remaining legacy boards here migrate in
> a later wave; do not add new dashboards to this repo.


## Description
This repository contains a modern set of Grafana dashboards for Kubernetes.
They are inspired by many other dashboards from `kubernetes-mixin` and `grafana.com`.

You can also download more on [Grafana.com](https://grafana.com/grafana/dashboards/).

## Dashboards

Organized by category for easier navigation:

### Kubernetes & API Gateway (`kubernetes/`)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| kubernetes-cluster-overview.json | Kubernetes cluster overview: nodes, pods, CPU/memory, and workload health. |
| kubernetes-nginx-ingress.json   | Nginx Ingress Controller metrics: request rate, latency, status codes. |
| kong-dashboard.json             | Kong API Gateway: request rate, latency, status codes, and upstream health. |

### PostgreSQL & Databases (`postgresql/`)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| cloudnative-pg-cluster.json     | CloudNativePG cluster overview: connections, replication, storage, and WAL metrics. |
| pg-monitoring.json              | PostgreSQL monitoring: throughput, cache hit ratio, connections, and bloat. |
| pg-query-overview.json          | PostgreSQL query overview: top statements by time, calls, and rows. |
| pg-query-drilldown.json         | PostgreSQL per-query drill-down via pg_stat_statements. |
| pg-exporter-instance.json       | postgres_exporter per-instance metrics: sessions, transactions, locks, and I/O. |
| pg-exporter-self.json           | postgres_exporter self-monitoring: scrape duration, errors, and collector health. |
| postgres-replication-lag.json   | PostgreSQL streaming replication lag across replicas. |
| pgbouncer.json                  | PgBouncer pooler: client/server connections, pool saturation, and wait times. |
| pgdog.json                      | PgDog sharding/pooling proxy metrics. |

### Redis (`redis/`)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| redis.json                      | Redis/Valkey: memory, throughput, hit ratio, and connected clients. |
| redis-exporter.json             | redis_exporter self-monitoring: metrics collection and exporter health. |

### Messaging (`messaging/`)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| amazonmq/amazonmq-rabbitmq.json | AmazonMQ RabbitMQ broker and per-queue metrics via YACE. Source: CloudWatch AWS/AmazonMQ namespace. |

### Observability & Monitoring (`observability/`)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| high-level-sloth-slos.json      | High-level SLO dashboard using Sloth SLO framework. |
| slo-detail.json                 | SLO detail view with error budgets and availability tracking. |
| pyrra_list.json                 | Pyrra SLO list view: overview of all SLOs and their status. |
| pyrra_detail.json               | Pyrra SLO detail view: drill-down into individual SLO metrics. |
| tempo-observability-dashboard.json | Grafana Tempo observability: distributed tracing throughput, latency, and errors. |

### Platform & Multi-Purpose (root)
| File name                       | Description         |
|:--------------------------------|:--------------------|
| microservices-dashboard.json    | Microservices Observability Platform: RED metrics (rate/errors/duration), Go runtime, and per-service drill-down for the duynhlab platform. |
| vault.json                      | Vault secret management and audit logging dashboard. |
