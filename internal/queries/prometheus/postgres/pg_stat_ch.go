package postgres

// Exporter health of the pg_stat_ch extension, from the CNPG custom query
// `pg_stat_ch` (pg_stat_ch_stats(), one series per instance). The extension
// drops events instead of back-pressuring PostgreSQL, so these counters are
// the only evidence that the ClickHouse picture is complete.
const (
	PgStatChEnqueuedRate = `sum by (cnpg_io_cluster) (rate(cnpg_pg_stat_ch_enqueued_events[5m]))`
	PgStatChExportedRate = `sum by (cnpg_io_cluster) (rate(cnpg_pg_stat_ch_exported_events[5m]))`
	PgStatChDroppedRate  = `sum by (cnpg_io_cluster) (rate(cnpg_pg_stat_ch_dropped_events[5m]))`
	PgStatChSendFailures = `sum by (cnpg_io_cluster) (increase(cnpg_pg_stat_ch_send_failures[5m]))`
	PgStatChQueueUsage   = `max by (cnpg_io_cluster, pod) (cnpg_pg_stat_ch_queue_usage_pct)`
	PgStatChDSAOOMRate   = `sum by (cnpg_io_cluster) (rate(cnpg_pg_stat_ch_dsa_oom_count[5m]))`
	// Seconds since the last batch ClickHouse accepted; it climbs without
	// bound while ClickHouse is unreachable.
	PgStatChLastSuccessAge = `max by (cnpg_io_cluster, pod) (cnpg_pg_stat_ch_last_success_age_seconds)`
	// enqueued_events counts only events that entered the ring (a dropped event
	// never did), so enqueued - exported - still-in-ring is what was lost in a
	// batch the server rejected.
	PgStatChLostInExport = `sum by (cnpg_io_cluster) (cnpg_pg_stat_ch_enqueued_events - cnpg_pg_stat_ch_exported_events - cnpg_pg_stat_ch_queue_size)`
	PgStatChInRing       = `sum by (cnpg_io_cluster) (cnpg_pg_stat_ch_queue_size)`
)
