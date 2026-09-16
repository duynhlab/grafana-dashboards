package postgres

// PromQL for the PostgreSQL Query Performance dashboard, sourced from
// pg_stat_statements exposed by the CloudNativePG collector. Scoped by the
// $cluster / $datname template variables.
const (
	// Overview stats.
	CallsPerSec  = `sum(rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	MeanLatency  = `sum(rate(cnpg_pg_stat_statements_time_milliseconds{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) / clamp_min(sum(rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])), 1)`
	RowsPerSec   = `sum(rate(cnpg_pg_stat_statements_rows{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	CacheHitRate = `sum(rate(cnpg_pg_stat_statements_shared_blks_hit{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) / clamp_min(sum(rate(cnpg_pg_stat_statements_shared_blks_hit{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) + sum(rate(cnpg_pg_stat_statements_shared_blks_read{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])), 1)`

	// Top statements (timeseries).
	TopByExecTime    = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_time_milliseconds{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])))`
	TopByCalls       = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])))`
	TopByMeanLatency = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_time_milliseconds{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) / clamp_min(sum by (queryid, query) (rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])), 1))`
	TopByRows        = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_rows{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])))`

	// Block I/O.
	SharedBlocksHit   = `sum(rate(cnpg_pg_stat_statements_shared_blks_hit{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	SharedBlocksRead  = `sum(rate(cnpg_pg_stat_statements_shared_blks_read{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	TempBlocksRead    = `sum(rate(cnpg_pg_stat_statements_temp_blks_read{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	TempBlocksWritten = `sum(rate(cnpg_pg_stat_statements_temp_blks_written{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	BlockReadTime     = `sum(rate(cnpg_pg_stat_statements_blk_read_time{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	BlockWriteTime    = `sum(rate(cnpg_pg_stat_statements_blk_write_time{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	TopByIOTime       = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_blk_read_time{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]) + rate(cnpg_pg_stat_statements_blk_write_time{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])))`
	TopByRowsPerCall  = `topk(10, sum by (queryid, query) (rate(cnpg_pg_stat_statements_rows{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) / clamp_min(sum by (queryid, query) (rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])), 1))`

	// Statement detail table (instant, merged by queryid/query).
	StatementExecTime = `topk(25, sum by (datname, queryid, query) (rate(cnpg_pg_stat_statements_time_milliseconds{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])))`
	StatementCalls    = `sum by (datname, queryid, query) (rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
	StatementMeanTime = `sum by (datname, queryid, query) (rate(cnpg_pg_stat_statements_time_milliseconds{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])) / clamp_min(sum by (datname, queryid, query) (rate(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m])), 1)`
	StatementRows     = `sum by (datname, queryid, query) (rate(cnpg_pg_stat_statements_rows{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[5m]))`
)
