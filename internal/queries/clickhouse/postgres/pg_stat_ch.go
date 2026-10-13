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

// Scope shared by every events_raw query: the dashboard range and the two
// variables. Both variables are include-all; Grafana expands $__all to every
// value, which is why `IN (...)` and not `=` is used.
const whereEvents = `$__timeFilter(ts) AND db_name IN (${db_name:singlequote}) AND app IN (${app:singlequote})`

// whereRollup is the same scope for the 5-minute rollup, which has no app column.
const whereRollup = `$__timeFilter(bucket) AND db_name IN (${db_name:singlequote})`

// Variables.
const (
	DBNameValues = `SELECT DISTINCT db_name FROM pg_stat_ch.events_raw WHERE ts > now() - INTERVAL 7 DAY ORDER BY db_name`
	AppValues    = `SELECT DISTINCT app FROM pg_stat_ch.events_raw WHERE ts > now() - INTERVAL 7 DAY ORDER BY app`
)

// Overview — one row over the whole range, so a stat reduces to a figure that
// means "over this window" rather than "in the partial last bucket".
const (
	StatementsPerSec = `SELECT count() / greatest(dateDiff('second', $__fromTime, $__toTime), 1) AS statements_per_s FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P50Ms            = `SELECT quantile(0.5)(duration_us) / 1000 AS p50_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P95Ms            = `SELECT quantile(0.95)(duration_us) / 1000 AS p95_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	P99Ms            = `SELECT quantile(0.99)(duration_us) / 1000 AS p99_ms FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	ErrorPct         = `SELECT 100.0 * countIf(err_elevel > 0) / greatest(count(), 1) AS error_pct FROM pg_stat_ch.events_raw WHERE ` + whereEvents
	BufferHitPct     = `SELECT 100.0 * sum(shared_blks_hit) / greatest(sum(shared_blks_hit) + sum(shared_blks_read), 1) AS hit_pct FROM pg_stat_ch.events_raw WHERE ` + whereEvents
)

// Traffic & errors.
const (
	RateByOperation   = `SELECT $__timeInterval(ts) AS time, db_operation, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, db_operation ORDER BY time`
	RateByDatabase    = `SELECT $__timeInterval(ts) AS time, db_name, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, db_name ORDER BY time`
	RateByApplication = `SELECT $__timeInterval(ts) AS time, app, count() / $__interval_s AS rps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` GROUP BY time, app ORDER BY time`
	ErrorsBySQLState  = `SELECT $__timeInterval(ts) AS time, err_sqlstate, count() / $__interval_s AS eps FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND err_elevel > 0 GROUP BY time, err_sqlstate ORDER BY time`
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

// Top statements. The total-time table reads the 5-minute rollup so a 30-day
// window stays cheap; the two drill-downs read events_raw.
const (
	TopStatementsByTotalTime = `SELECT query_id, db_name, db_operation, countMerge(calls_state) AS calls, round(sumMerge(duration_sum_state) / 1e6, 2) AS total_s, ` +
		`round(100 * sumMerge(duration_sum_state) / (SELECT sumMerge(duration_sum_state) FROM pg_stat_ch.query_stats_5m WHERE ` + whereRollup + `), 2) AS pct_runtime, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 1) / 1000, 2) AS p50_ms, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 2) / 1000, 2) AS p95_ms, ` +
		`round(arrayElement(quantilesTDigestMerge(0.5, 0.95, 0.99)(duration_q_state), 3) / 1000, 2) AS p99_ms, ` +
		`sumMerge(errors_sum_state) AS errors, ` +
		`round(100.0 * sumMerge(shared_hit_sum_state) / greatest(sumMerge(shared_hit_sum_state) + sumMerge(shared_read_sum_state), 1), 1) AS hit_pct, ` +
		`anyMerge(query_text_state) AS query_text ` +
		`FROM pg_stat_ch.query_stats_5m WHERE ` + whereRollup + ` GROUP BY query_id, db_name, db_operation ORDER BY total_s DESC LIMIT 25`
	TopStatementsByPhysicalReads = `SELECT query_id, db_name, count() AS calls, sum(shared_blks_read) AS blks_read, round(avg(shared_blks_read), 1) AS avg_blks_read, round(avg(duration_us) / 1000, 2) AS avg_ms, any(query_text) AS query_text ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` AND shared_blks_read > 0 GROUP BY query_id, db_name ORDER BY blks_read DESC LIMIT 15`
	SlowestExecutions = `SELECT ts, db_name, app, db_user, pid, query_id, round(duration_us / 1000, 2) AS duration_ms, rows, shared_blks_read, query_text ` +
		`FROM pg_stat_ch.events_raw WHERE ` + whereEvents + ` ORDER BY duration_us DESC LIMIT 25`
)

// Errors — the 30-day errors_recent rollup, newest first.
const RecentErrors = `SELECT ts, db_name, app, db_user, err_sqlstate, err_elevel, err_message, query_text, client_addr ` +
	`FROM pg_stat_ch.errors_recent WHERE ` + whereEvents + ` ORDER BY ts DESC LIMIT 100`
