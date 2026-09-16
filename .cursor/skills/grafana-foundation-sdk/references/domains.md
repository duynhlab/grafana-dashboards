# Domains

Grafana folder title/UID equals the domain display name. CR metadata name is the lowercase slug.

| Folder | Typical packages |
|--------|------------------|
| Kubernetes | `dashboards/kubernetes`, `alerts/kubernetes`, `queries/prometheus/kubernetes` |
| Databases | `dashboards/postgres` (and later mongodb), `alerts/postgres` |
| Observability | Tempo, SLO, Pyrra — when ported |

Adding a domain should not change `cmd/generate`. Register the dashboard/alert with `Folder: "NewDomain"`.

## Kubernetes sections

Overview, Compute, Memory, Network, Storage, Reliability, Cost — only the ones that earn a place on that board.

Example family: cluster overview, workload, nodes.

## Java

JVM Overview, CPU, Heap, GC, Threads, HTTP, Reliability.

## Karpenter

Capacity, NodeClaims, Provisioning, Disruption, Consolidation, Spot, Cost.

## Traefik

Overview, Traffic, Latency, Errors, Backend Health.

## Scaling

Grow by adding domain + dashboard + section + query + alert. Do not grow by enlarging the generator or one 3000-line dashboard file.

## Porting from legacy JSON

1. Read `dashboard/` for titles, PromQL, variables (do not edit those files)
2. Extract queries
3. Rewrite as `dashboardv2` — no schema v1 parity
4. Add matching alerts when the board has operational thresholds
