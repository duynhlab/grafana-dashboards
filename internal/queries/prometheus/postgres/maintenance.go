package postgres

// PromQL for the PostgreSQL Maintenance (CNPG) dashboard. Queries target the
// CloudNativePG collector metrics and are scoped by the $cluster / $datname
// template variables.
const (
	// Locks & blocking.
	LocksByMode     = `sum by (mode) (cnpg_pg_locks_count_count{cnpg_io_cluster=~"$cluster", datname=~"$datname"})`
	BlockedQueries  = `sum(cnpg_pg_blocking_queries_blocked_queries{cnpg_io_cluster=~"$cluster"})`
	LocksByDatabase = `sum by (datname) (cnpg_pg_locks_count_count{cnpg_io_cluster=~"$cluster", datname=~"$datname"})`

	// Checkpointer.
	CheckpointsTimed     = `sum(rate(cnpg_pg_stat_checkpointer_checkpoints_timed{cnpg_io_cluster=~"$cluster"}[5m]))`
	CheckpointsRequested = `sum(rate(cnpg_pg_stat_checkpointer_checkpoints_req{cnpg_io_cluster=~"$cluster"}[5m]))`
	CheckpointWriteTime  = `sum(rate(cnpg_pg_stat_checkpointer_checkpoint_write_time{cnpg_io_cluster=~"$cluster"}[5m]))`
	CheckpointSyncTime   = `sum(rate(cnpg_pg_stat_checkpointer_checkpoint_sync_time{cnpg_io_cluster=~"$cluster"}[5m]))`

	// Autovacuum & bloat.
	DeadTupleRatioTopTables = `topk(10, sum by (datname, relname) (cnpg_pg_stat_user_tables_autovacuum_n_dead_tup{cnpg_io_cluster=~"$cluster", datname=~"$datname"}) / clamp_min(sum by (datname, relname) (cnpg_pg_stat_user_tables_autovacuum_n_dead_tup{cnpg_io_cluster=~"$cluster", datname=~"$datname"}) + sum by (datname, relname) (cnpg_pg_stat_user_tables_autovacuum_n_live_tup{cnpg_io_cluster=~"$cluster", datname=~"$datname"}), 1))`
	TotalDeadTuples         = `sum(cnpg_pg_stat_user_tables_autovacuum_n_dead_tup{cnpg_io_cluster=~"$cluster", datname=~"$datname"})`
	AutovacuumRuns          = `sum by (datname) (rate(cnpg_pg_stat_user_tables_autovacuum_autovacuum_count{cnpg_io_cluster=~"$cluster", datname=~"$datname"}[15m]))`
	TopTablesBySize         = `topk(20, sum by (cnpg_io_cluster, datname, schemaname, tablename) (cnpg_pg_table_size_total_bytes{cnpg_io_cluster=~"$cluster", datname=~"$datname"}))`
	UnusedIndexes           = `sum by (datname, relname, indexrelname) (cnpg_pg_stat_user_indexes_index_bytes{cnpg_io_cluster=~"$cluster", datname=~"$datname"}) and (sum by (datname, relname, indexrelname) (cnpg_pg_stat_user_indexes_idx_scan{cnpg_io_cluster=~"$cluster", datname=~"$datname"}) == 0)`

	// Transactions & vacuum.
	OldestTransactionAge  = `max by (cnpg_io_cluster) (cnpg_pg_long_running_transactions_oldest_transaction_seconds{cnpg_io_cluster=~"$cluster"})`
	OldestIdleInTxAge     = `max by (cnpg_io_cluster) (cnpg_pg_long_running_transactions_oldest_idle_in_transaction_seconds{cnpg_io_cluster=~"$cluster"})`
	IdleInTxSessions      = `max by (cnpg_io_cluster) (cnpg_pg_long_running_transactions_idle_in_transaction_count{cnpg_io_cluster=~"$cluster"})`
	LongestActiveQuery    = `max by (cnpg_io_cluster) (cnpg_pg_long_running_transactions_longest_active_query_seconds{cnpg_io_cluster=~"$cluster"})`
	VacuumProgress        = `cnpg_pg_stat_progress_vacuum_heap_blks_scanned{cnpg_io_cluster=~"$cluster", datname=~"$datname"} / clamp_min(cnpg_pg_stat_progress_vacuum_heap_blks_total{cnpg_io_cluster=~"$cluster", datname=~"$datname"}, 1)`
	ActiveVacuumsSnapshot = `cnpg_pg_stat_progress_vacuum_heap_blks_scanned{cnpg_io_cluster=~"$cluster", datname=~"$datname"}`
)
