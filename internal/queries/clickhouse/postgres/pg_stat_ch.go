// Package postgres holds the ClickHouse SQL for the PostgreSQL query-telemetry
// board. The rows come from the pg_stat_ch extension: one event per statement
// execution in `pg_stat_ch.events_raw` (7 days hot, 90 days total) and the
// rollups `query_stats_5m`, `db_app_user_1m`, `errors_recent` that homelab's
// clickhouse-ddl image creates. `db_name` is the cluster key on that platform:
// every database name is unique across platform-db and product-db, and the
// native export path leaves the extension's instance columns empty.
//
// Macros ($__timeFilter, $__timeInterval, $__fromTime, $__toTime,
// $__interval_s) belong to grafana-clickhouse-datasource v4; the variable
// formats (${db_name:singlequote}) to Grafana.
package postgres

// scopeEvents hides the platform's own sessions unless the `scope` variable is
// 1 ("Include platform sessions"). The names were measured on Kind once every
// workload set PGAPPNAME: CNPG's instance manager and metrics exporter poll every
// database, PgDog's health check (`pg_is_in_recovery()`) runs in each pooled
// database, Barman's backup holds pg_backup_start open in database `postgres`,
// and `db-authz` / `pgroles` are the names RFC-0029 gives the grant jobs.
// Keycloak, Temporal and psql stay visible: they are real load or a person.
const scopeEvents = `(${scope:raw} = 1 OR (application_name NOT IN ('cnpg-instance-manager', 'cnpg_metrics_exporter', 'PgDog', 'barman_cloud_backup', 'db-authz', 'pgroles') AND db_name != 'postgres'))`

// scopeRollup is scopeEvents for query_stats_5m, which has no application_name
// column: only
// the `postgres` database can be excluded there.
const scopeRollup = `(${scope:raw} = 1 OR db_name != 'postgres')`

// Scope shared by every events_raw and errors_recent query: the dashboard
// range, the db_name and application_name variables, and the platform-session
// scope. Both
// list variables are include-all; Grafana expands $__all to every value, which
// is why `IN (...)` and not `=` is used.
const whereEvents = `$__timeFilter(ts) AND db_name IN (${db_name:singlequote}) AND application_name IN (${application_name:singlequote}) AND ` + scopeEvents

// whereRollup is the same scope for the 5-minute rollup, which has no
// application_name column.
const whereRollup = `$__timeFilter(bucket) AND db_name IN (${db_name:singlequote}) AND ` + scopeRollup

// whereAppUser is the same scope for the 1-minute db_app_user_1m rollup.
const whereAppUser = `$__timeFilter(bucket) AND db_name IN (${db_name:singlequote}) AND application_name IN (${application_name:singlequote}) AND ` + scopeEvents

// Error severity. err_elevel carries the PostgreSQL elevel: 0 is success, 19
// WARNING, 21 ERROR, 22 FATAL, 23 PANIC. A real error is >= 21. The rollups'
// errors_sum_state is sum(err_elevel > 0) and errors_recent keeps every row
// with err_elevel > 0, so both include warnings.
const (
	isError   = `err_elevel >= 21`
	isWarning = `err_elevel = 19`
)

// Variables. Both read the 1-minute rollup, which keeps the same 90 days as
// events_raw at a fraction of the rows. "All" expands to exactly this list, so
// a list shorter than the retention would silently drop older rows from any
// wider range.
const (
	DBNameValues          = `SELECT DISTINCT db_name FROM pg_stat_ch.db_app_user_1m ORDER BY db_name`
	ApplicationNameValues = `SELECT DISTINCT application_name FROM pg_stat_ch.db_app_user_1m ORDER BY application_name`
)

// Overview — one row over the whole range, so a stat reduces to a figure that
// means "over this window" rather than "in the partial last bucket".
const (
	StatementsPerSec = `SELECT count() / greatest(dateDiff('second', $__fromTime, $__toTime), 1) AS statements_per_s FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P50Ms            = `SELECT quantile(0.5)(duration_us) / 1000 AS p50_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P95Ms            = `SELECT quantile(0.95)(duration_us) / 1000 AS p95_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P99Ms            = `SELECT quantile(0.99)(duration_us) / 1000 AS p99_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	ErrorPct         = `SELECT 100.0 * countIf(` + isError + `) / greatest(count(), 1) AS error_pct FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	BufferHitPct     = `SELECT 100.0 * sum(shared_blks_hit) / greatest(sum(shared_blks_hit) + sum(shared_blks_read), 1) AS hit_pct FROM pg_stat_ch.events_raw WHERE ` + whereEvents
)

// Traffic & errors.
const (
	RateByOperation    = `SELECT $__timeInterval(ts) AS time, db_operation, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, db_operation ORDER BY time`
	RateByDatabase     = `SELECT $__timeInterval(ts) AS time, db_name, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, db_name ORDER BY time`
	RateByApplication  = `SELECT $__timeInterval(ts) AS time, application_name, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, application_name ORDER BY time`
	ErrorsBySQLState   = `SELECT $__timeInterval(ts) AS time, err_sqlstate, count() / $__interval_s AS eps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND ` + isError + ` GROUP BY time, err_sqlstate ORDER BY time`
	WarningsBySQLState = `SELECT $__timeInterval(ts) AS time, err_sqlstate, count() / $__interval_s AS wps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND ` + isWarning + ` GROUP BY time, err_sqlstate ORDER BY time`
)

// Latency.
const (
	LatencyPercentiles = `SELECT $__timeInterval(ts) AS time, quantile(0.5)(duration_us) / 1000 AS p50_ms, quantile(0.95)(duration_us) / 1000 AS p95_ms, quantile(0.99)(duration_us) / 1000 AS p99_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time ORDER BY time`
	P95ByDatabase      = `SELECT $__timeInterval(ts) AS time, db_name, quantile(0.95)(duration_us) / 1000 AS p95_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, db_name ORDER BY time`
	// CPU (getrusage delta) against block I/O wait (track_io_timing); the gap
	// below the duration line is lock or network wait.
	CPUVersusIOWait = `SELECT $__timeInterval(ts) AS time, avg(cpu_user_time_us + cpu_sys_time_us) / 1000 AS cpu_ms, avg(shared_blk_read_time_us + shared_blk_write_time_us + temp_blk_read_time_us + temp_blk_write_time_us) / 1000 AS io_wait_ms, avg(duration_us) / 1000 AS duration_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time ORDER BY time`
	WALAndTempBytes = `SELECT $__timeInterval(ts) AS time, sum(wal_bytes) / $__interval_s AS wal_bytes_per_s, sum(temp_blks_written) * 8192 / $__interval_s AS temp_bytes_per_s FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time ORDER BY time`
)

// Writes. Full-page images follow every checkpoint: a burst of wal_fpi right
// after one is expected, a steady high ratio to wal_records points at
// checkpoint_timeout or full_page_writes on a write-heavy table.
const (
	WALRecordsAndFPI = `SELECT $__timeInterval(ts) AS time, sum(wal_records) / $__interval_s AS wal_records_per_s, sum(wal_fpi) / $__interval_s AS wal_fpi_per_s FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time ORDER BY time`
	BlocksDirtied    = `SELECT $__timeInterval(ts) AS time, sum(shared_blks_dirtied) / $__interval_s AS shared_dirtied_per_s, sum(local_blks_dirtied) / $__interval_s AS local_dirtied_per_s, sum(temp_blks_written) / $__interval_s AS temp_written_per_s FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time ORDER BY time`
)

// Top statements. The total-time table reads the 5-minute rollup so a 30-day
// window stays cheap; the others read events_raw. query_stats_5m has no
// application_name column, so that variable and the application_name half of the
// scope do not apply to it.
const (
	TopStatementsByTotalTime = `SELECT query_id, db_name, db_operation, countMerge(calls_state) AS calls, round(sumMerge(duration_sum_state) / 1e6, 2) AS total_s, ` +
		`round(100 * sumMerge(duration_sum_state) / (SELECT sumMerge(duration_sum_state) FROM pg_stat_ch.query_stats_5m WHERE ` + whereRollup + `), 2) AS pct_runtime, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 1) / 1000, 2) AS p50_ms, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 2) / 1000, 2) AS p95_ms, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 3) / 1000, 2) AS p99_ms, ` +
		`round(sumMerge(duration_sum_state) / greatest(countMerge(calls_state), 1) / 1000, 2) AS avg_ms, ` +
		`round(maxMerge(duration_max_state) / 1000, 2) AS max_ms, ` +
		// errors_sum_state counts err_elevel > 0, so warnings are included.
		`sumMerge(errors_sum_state) AS errors, ` +
		`round(100.0 * sumMerge(shared_hit_sum_state) / greatest(sumMerge(shared_hit_sum_state) + sumMerge(shared_read_sum_state), 1), 1) AS hit_pct, ` +
		`anyMerge(query_text_state) AS query_text ` +
		`FROM pg_stat_ch.query_stats_5m WHERE ` + whereRollup + ` GROUP BY query_id, db_name, db_operation ORDER BY total_s DESC LIMIT 25`
	TopStatementsByPhysicalReads = `SELECT query_id, db_name, count() AS calls, sum(shared_blks_read) AS blks_read, round(avg(shared_blks_read), 1) AS avg_blks_read, round(avg(duration_us) / 1000, 2) AS avg_ms, any(query_text) AS query_text ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND shared_blks_read > 0 GROUP BY query_id, db_name ORDER BY blks_read DESC LIMIT 15`
	SlowestExecutions = `SELECT ts, db_name, application_name, db_user, pid, query_id, round(duration_us / 1000, 2) AS duration_ms, rows, shared_blks_read, query_text ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` ORDER BY duration_us DESC LIMIT 25`
	// A statement that writes temp blocks sorted or hashed more than work_mem.
	TempSpillStatements = `SELECT query_id, db_name, count() AS executions, round(avg(temp_blks_written), 1) AS avg_temp_blks, round(avg(duration_us) / 1000, 2) AS avg_ms, any(query_text) AS query_text ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND temp_blks_written > 0 GROUP BY query_id, db_name ORDER BY sum(temp_blks_written) DESC LIMIT 15`
)

// Statement drill-down: one fingerprint typed into the query_id textbox, to see
// what changed between two moments. Only the range and db_name scope it, so a
// fingerprint from a platform session still shows. An empty textbox matches
// nothing (query_id 0 is a real fingerprint for utility statements, so the
// empty string is excluded explicitly). sqlstring quotes and escapes the value.
const (
	queryIDFilter    = `${query_id:sqlstring} != '' AND query_id = toInt64OrZero(${query_id:sqlstring})`
	StatementLatency = `SELECT $__timeInterval(ts) AS time, count() / $__interval_s AS calls_per_s, avg(duration_us) / 1000 AS avg_ms, quantile(0.95)(duration_us) / 1000 AS p95_ms FROM pg_stat_ch.events_raw ` +
		`WHERE $__timeFilter(ts) AND db_name IN (${db_name:singlequote}) AND ` + queryIDFilter + ` GROUP BY time ORDER BY time`
	StatementErrors = `SELECT ts, db_name, application_name, db_user, err_sqlstate, err_elevel, err_message, query_text FROM pg_stat_ch.errors_recent ` +
		`WHERE $__timeFilter(ts) AND db_name IN (${db_name:singlequote}) AND ` + queryIDFilter + ` ORDER BY ts DESC LIMIT 100`
)

// Load attribution from the 1-minute db_app_user_1m rollup. errors_sum_state
// counts err_elevel > 0, so these error figures include warnings.
const (
	LoadByAppAndRole = `SELECT application_name, db_user, countMerge(calls_state) AS calls, round(sumMerge(duration_sum_state) / 1e6, 2) AS total_s, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 3) / 1000, 2) AS p99_ms, sumMerge(errors_sum_state) AS errors ` +
		`FROM pg_stat_ch.db_app_user_1m WHERE ` + whereAppUser + ` GROUP BY application_name, db_user ORDER BY total_s DESC LIMIT 25`
	// Every session must set application_name (PGAPPNAME). Not scoped by the
	// Sessions variable: a platform session without a name is a finding too.
	MissingApplicationName = `SELECT db_name, db_user, countMerge(calls_state) AS statements, max(bucket) AS last_seen ` +
		`FROM pg_stat_ch.db_app_user_1m WHERE $__timeFilter(bucket) AND db_name IN (${db_name:singlequote}) AND application_name = '' ` +
		`GROUP BY db_name, db_user ORDER BY statements DESC LIMIT 25`
	ErrorRateByDatabaseAndRole = `SELECT db_name, db_user, countMerge(calls_state) AS queries, sumMerge(errors_sum_state) AS errors, round(100.0 * errors / greatest(queries, 1), 2) AS error_pct ` +
		`FROM pg_stat_ch.db_app_user_1m WHERE ` + whereAppUser + ` GROUP BY db_name, db_user HAVING errors > 0 ORDER BY error_pct DESC LIMIT 25`
)

// JIT and parallel query (PostgreSQL 18 fields). JIT time is generation +
// inlining + optimization + emission; deform runs inside execution.
const (
	JITOverhead = `SELECT query_id, count() AS executions, round(avg(duration_us) / 1000, 2) AS avg_total_ms, ` +
		`round(avg(jit_generation_time_us + jit_inlining_time_us + jit_optimization_time_us + jit_emission_time_us) / 1000, 2) AS avg_jit_ms, ` +
		`round(100 * avg_jit_ms / greatest(avg_total_ms, 0.001), 1) AS jit_pct_of_total ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND jit_functions > 0 GROUP BY query_id HAVING avg_jit_ms > 10 ORDER BY avg_jit_ms DESC LIMIT 15`
	ParallelWorkersShort = `SELECT query_id, db_operation, count() AS executions, round(avg(parallel_workers_planned), 1) AS avg_planned, round(avg(parallel_workers_launched), 1) AS avg_launched, ` +
		`round(avg(parallel_workers_planned - parallel_workers_launched), 1) AS avg_missed ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND parallel_workers_planned > parallel_workers_launched GROUP BY query_id, db_operation ORDER BY executions * avg_missed DESC LIMIT 15`
)

// Errors — the 30-day errors_recent rollup, newest first.
const RecentErrors = `SELECT ts, db_name, application_name, db_user, err_sqlstate, err_elevel, err_message, query_text, client_addr ` +
	`FROM pg_stat_ch.errors_recent WHERE ` + whereEvents + ` ORDER BY ts DESC LIMIT 100`
