package postgres

// PromQL for the PG Exporter Instance dashboard (Pigsty PGRDS, gnet author
// Ruohang Feng). These metrics use the Pigsty postgres_exporter model
// (pg_* raw metrics and pg:ins:* / pg:db:* / pg:cls:* recording rules with
// ins/cls/datname labels) and are intentionally kept separate from the CNPG
// (cnpg_*) query set. Scoped by the $ins / $cls template variables.
const (
	// Overview KPI strip.
	InsCommitRate  = `pg:ins:xact_commit_rate1m{ins="$ins"}`
	InsRollbackKPI = `pg:ins:xact_rollback_rate1m{ins="$ins"}`
	InsRT          = `pg:ins:active_time_rate1m{ins="$ins"} / pg:ins:xact_total_rate1m{ins="$ins"}`
	InsConnPct     = `max(pg:db:conn_usage{cls="$cls"})`
	InsBackend     = `sum by (cls) (pg_db_numbackends{cls="$cls"})`
	InsActiveConn  = `pg:ins:active_backends{ins="$ins"}`
	InsIxactConn   = `max(pg:ins:ixact_backends{ins="$ins"})`
	InsRowFetch    = `pg:ins:tup_fetched_rate1m{ins="$ins"}`
	InsRowChange   = `pg:ins:tup_modified_rate1m{ins="$ins"}`
	InsBlksRead    = `sum(pg:db:blks_read_1m{ins="$ins"}) * 4096`
	InsAgePct      = `pg:ins:age{ins="$ins"} / 2147483647`
	InsSize        = `sum(pg_size_bytes{ins="$ins"})`

	// Overview load & alerts.
	ClusterLoad       = `pg:cls:active_time_rate1m{cls="$cls"}`
	InstanceDBLoad    = `pg:db:active_time_rate1m{ins="$ins", datname!~'template\d'}`
	FiringAlertsCount = `count(ALERTS{ins="$ins", alertstate="firing"}) or on() vector(0)`
	FiringAlerts      = `ALERTS{ins="$ins", alertstate="firing"}`
	PendingAlerts     = `ALERTS{ins="$ins", alertstate="pending"}`

	// Activity.
	XactCommitRate5m   = `sum by (ins) (pg:ins:xact_commit_rate5m{ins="$ins"})`
	XactRollbackRate5m = `sum by (ins) (pg:ins:xact_rollback_rate5m{ins="$ins"})`
	TxnRTInstance      = `pg:ins:active_time_rate1m{ins="$ins"} / pg:ins:xact_total_rate1m{ins="$ins"}`
	TxnRTDatabase      = `pg:db:active_time_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'} / pg:db:xact_total_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}`
	TPSByDatabaseAct   = `pg:db:xact_commit_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}`
	PostgresLoad       = `pg:db:active_time_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}`
	RowFetchedAct      = `sum by (datname) (rate(pg_db_tup_fetched{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}[1m]))`
	RowInsertedAct     = `sum(rate(pg_db_tup_inserted{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}[1m]))`
	RowUpdatedAct      = `sum(rate(pg_db_tup_updated{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}[1m]))`
	RowDeletedAct      = `sum(rate(pg_db_tup_deleted{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}[1m]))`
	LocksExclusive     = `pg:ins:xlock_count{ins="$ins"}`
	LocksWrite         = `pg:ins:wlock_count{ins="$ins"}`
	LocksRead          = `pg:ins:rlock_count{ins="$ins"}`
	PigstyLocksByMode  = `sum by (mode) (pg_lock_count{ins="$ins"})`
	MaxTxDuration      = `sum by (state) (pg_activity_max_tx_duration{ins="$ins", state!~"(idle.*|disabled)"})`

	// Session.
	ConnUsageSession    = `pg:db:conn_usage{ins="$ins", datname!~'template\d'}`
	IdleInTxBackends    = `sum by (datname) (pg_activity_count{ins="$ins", state=~"idle in.*"})`
	BackendsByDatabase  = `sum by (datname) (pg_activity_count{ins="$ins"})`
	NewSessions1m       = `sum by (datname) (increase(pg_db_sessions{ins="$ins"}[1m]))`
	BackendsByState     = `sum by (state) (pg_activity_count{ins="$ins"})`
	BackendsByType      = `sum by (type) (pg_backend_count{ins="$ins"})`
	MaxConnLifespan     = `sum by (state) (pg_activity_max_conn_duration{ins="$ins", state!="idle|disabled"})`
	BackendsByWaitEvent = `sum by (event) (pg_wait_count{ins="$ins"})`
	ActivePctSession    = `sum by (datname) (rate(pg_db_active_time{ins="$ins"}[1m])) / sum by (datname) (rate(pg_db_session_time{ins="$ins"}[1m]))`
	SessionsAbandoned   = `sum by (datname) (increase(pg_db_sessions_abandoned{ins="$ins"}[1m]))`
	SessionsFatal       = `sum by (datname) (increase(pg_db_sessions_fatal{ins="$ins"}[1m]))`
	SessionsKilled      = `sum by (datname) (increase(pg_db_sessions_killed{ins="$ins"}[1m]))`

	// Persist.
	LSNProgress            = `pg:ins:lsn_rate1m{ins="$ins"}`
	AgeUsage               = `max by (datname) (pg:db:age{cls="$cls"})`
	DatabaseClusterSize    = `sum by (datname) (pg_size_bytes{ins="$ins"})`
	DatabaseWALLogSize     = `sum by (datname) (pg_size_bytes{ins="$ins", datname=~"(log|wal)"})`
	CheckpointScheduled    = `pg_bgwriter_checkpoints_timed{ins="$ins"}`
	CheckpointRequested    = `pg_bgwriter_checkpoints_req{ins="$ins"}`
	CheckpointSyncTimeIns  = `rate(pg_bgwriter_checkpoint_sync_time{ins="$ins"}[1m])`
	CheckpointWriteTimeIns = `- rate(pg_bgwriter_checkpoint_write_time{ins="$ins"}[1m])`
	BGWriterClean          = `rate(pg_bgwriter_buffers_clean{ins="$ins"}[1m]) * 8192`
	BGWriterCheckpoint     = `rate(pg_bgwriter_buffers_checkpoint{ins="$ins"}[1m]) * 8192`
	BGWriterBackend        = `rate(pg_bgwriter_buffers_backend{ins="$ins"}[1m]) * 8192`
	BGWriterAlloc          = `rate(pg_bgwriter_buffers_alloc{ins="$ins"}[1m]) * 8192`
	BlocksAccess1m         = `sum by (datname) (pg:db:blks_access_1m{ins="$ins"}) * 4096`
	BlocksHitRatio1m       = `pg:db:blks_hit_ratio1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}`
	BlocksRead1m           = `sum by (datname) (pg:db:blks_read_1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'}) * 4096`
	BlockReadTimeRate1m    = `sum by (datname) (pg:db:blk_read_time_seconds_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'})`
	BlockWriteTimeRate1m   = `- sum by (datname) (pg:db:blk_write_time_seconds_rate1m{ins="$ins", datname!~'rdsadmin|polardb_admin|template\d'})`

	// Database.
	DatabaseSize       = `pg_size_bytes{ins="$ins", datname!~'wal|log|template\d'}`
	DatabaseSizeDelta  = `delta(pg_size_bytes{ins="$ins", datname!~'wal|log|template\d'}[10m])`
	TPSByDatabaseDB    = `sum by (datname) (pg:db:xact_commit_rate1m{ins="$ins", datname!~'template\d'})`
	SessionByDatabase  = `sum by (datname) (pg_activity_count{ins="$ins"})`
	BlocksHitRatioDB   = `pg:db:blks_hit_ratio1m{ins="$ins", datname!~'template\d'}`
	IdleInTxBackendsDB = `sum by (datname) (pg_activity_count{ins="$ins", state=~"idle in.*"})`
	ConnUsageDB        = `pg:db:conn_usage{ins="$ins", datname!~'template\d'}`
	NewSessionsDB      = `increase(pg_db_sessions{ins="$ins", datname!~'template\d'}[1m])`
	RowFetchedDB       = `sum by (datname) (rate(pg_db_tup_fetched{ins="$ins", datname!~'template\d'}[1m]))`
	RowInsertedDB      = `sum(rate(pg_db_tup_inserted{ins="$ins", datname!~'template\d'}[1m])) by (datname)`
	RowUpdatedDB       = `sum(rate(pg_db_tup_updated{ins="$ins", datname!~'template\d'}[1m])) by (datname)`
	RowDeletedDB       = `sum(rate(pg_db_tup_deleted{ins="$ins", datname!~'template\d'}[1m])) by (datname)`

	// Table & Query.
	TableScanRate1m = `sum by (datname,relname) (pg:table:scan_rate1m{ins="$ins", datname!~'template\d'})`
	TupleReadRate1m = `sum by (datname,relname) (rate(pg_table_tup_read{ins="$ins", datname!~'template\d'}[1m]))`
	QueryCallRate1m = `pg:query:call_rate1m{ins="$ins", datname!~'template\d'}`
	QueryExecTime1m = `rate(pg_query_exec_time{ins="$ins", datname!~'template\d'}[1m])`
)

// PG Exporter Instance template-variable definitions.
const (
	PigstyInsValues = `label_values(pg_up, ins)`
	PigstyClsValues = `label_values(pg_up{ins="$ins"}, cls)`
)
