# grafana-dashboards

## Description
This repository contains a modern set of Grafana dashboards for Kubernetes.
They are inspired by many other dashboards from `kubernetes-mixin` and `grafana.com`.

You can also download more on [Grafana.com](https://grafana.com/grafana/dashboards/).

## Dashboards
| File name                       | Description         |
|:--------------------------------|:--------------------|
| cloudnative-pg-cluster.json     | CloudNativePG cluster overview: connections, replication, storage, and WAL metrics. |
| high-level-sloth-slos.json      | Dashboard for High level Sloth SLOs. |
| kong-dashboard.json             | Kong API Gateway: request rate, latency, status codes, and upstream health. |
| kubernetes-cluster-overview.json | Kubernetes cluster overview: nodes, pods, CPU/memory, and workload health. |
| kubernetes-nginx-ingress.json   | Dashboard for Nginx Ingress Controller. |
| microservices-dashboard.json    | Microservices Observability Platform: RED metrics (rate/errors/duration), Go runtime, and per-service drill-down for the duynhlab platform. |
| pg-exporter-instance.json       | postgres_exporter per-instance metrics: sessions, transactions, locks, and I/O. |
| pg-exporter-self.json           | postgres_exporter self-monitoring: scrape duration, errors, and collector health. |
| pg-monitoring.json              | PostgreSQL monitoring: throughput, cache hit ratio, connections, and bloat. |
| pg-query-drilldown.json         | PostgreSQL per-query drill-down via pg_stat_statements. |
| pg-query-overview.json          | PostgreSQL query overview: top statements by time, calls, and rows. |
| pgbouncer.json                  | PgBouncer pooler: client/server connections, pool saturation, and wait times. |
| pgdog.json                      | PgDog sharding/pooling proxy metrics. |
| postgres-replication-lag.json   | PostgreSQL streaming replication lag across replicas. |
| redis.json                      | Redis/Valkey: memory, throughput, hit ratio, and connected clients. |
| redis-exporter.json             | Dashboard for Redis |
| slo-detail.json                 | Dashboard for SLO / Detail. |
| tempo-observability-dashboard.json | Grafana Tempo observability: distributed tracing throughput, latency, and errors. |
| valut.json                 | Dashboard for Vault |
