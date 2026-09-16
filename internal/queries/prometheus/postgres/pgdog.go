package postgres

// PromQL for the PGDog connection-pooler dashboard (upstream Grafana.com
// dashboard 24583). Metrics use the pgdog_* namespace with a seven-dimension
// filter ($pooler/$host/$port/$shard/$role/$database/$user). These labels are
// specific to PGDog and are intentionally kept local to this dashboard.
const (
	// Client.
	PGDogClientsTotal   = `sum (pgdog_clients{job=~"$pooler"}) by (instance)`
	PGDogClientsWaiting = `sum (pgdog_cl_waiting{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`

	// Server.
	PGDogServersActive   = `sum (pgdog_sv_active{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogServersIdle     = `sum (pgdog_sv_idle{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogServersIdleXact = `sum (pgdog_sv_idle_xact{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`

	// Connection.
	PGDogMaxWait   = `max (pgdog_maxwait{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogErrors    = `sum (rate(pgdog_errors{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogOutOfSync = `sum (rate(pgdog_out_of_sync{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`

	// Transaction.
	PGDogXactRate    = `sum (rate(pgdog_total_xact_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDog2PCRate     = `sum (rate(pgdog_total_xact_2pc_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogXactTime    = `avg (rate(pgdog_total_xact_time{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogAvgXact     = `sum (pgdog_avg_xact_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogAvg2PC      = `sum (pgdog_avg_xact_2pc_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogAvgXactTime = `avg (pgdog_avg_xact_time{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`

	// Query.
	PGDogQueryRate        = `sum (rate(pgdog_total_query_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogQueryTime        = `avg (rate(pgdog_total_query_time{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogAvgQuery         = `sum (pgdog_avg_query_count{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogAvgQueryTime     = `avg (pgdog_avg_query_time{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogQueryCacheSize   = `sum (pgdog_query_cache_size{job=~"$pooler"}) by (instance)`
	PGDogQueryCacheHits   = `sum (rate(pgdog_query_cache_hits{job=~"$pooler"}[$__rate_interval])) by (instance)`
	PGDogQueryCacheMisses = `sum (rate(pgdog_query_cache_misses{job=~"$pooler"}[$__rate_interval])) by (instance)`
	PGDogQueryCacheDirect = `sum (rate(pgdog_query_cache_direct{job=~"$pooler"}[$__rate_interval])) by (instance)`
	PGDogQueryCacheCross  = `sum (rate(pgdog_query_cache_cross{job=~"$pooler"}[$__rate_interval])) by (instance)`

	// Traffic.
	PGDogReceivedRate = `sum (rate(pgdog_total_received{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogSentRate     = `sum (rate(pgdog_total_sent{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogAvgReceived  = `sum (pgdog_avg_received{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`
	PGDogAvgSent      = `sum (pgdog_avg_sent{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`

	// Prepared statements.
	PGDogPreparedMemory    = `sum (pgdog_prepared_statements_memory_used{job=~"$pooler"}) by (instance)`
	PGDogPreparedEvictRate = `sum (rate(pgdog_total_prepared_evictions{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}[$__rate_interval])) by (host, port, shard, role, database, user)`
	PGDogPreparedCount     = `sum (pgdog_prepared_statements{job=~"$pooler"}) by (instance)`
	PGDogAvgPreparedEvict  = `sum (pgdog_avg_prepared_evictions{job=~"$pooler", host=~"$host", port=~"$port", shard="$shard", role=~"$role", database=~"$database", user=~"$user"}) by (host, port, shard, role, database, user)`

	// Mirror.
	PGDogMirrorTotal    = `sum (rate(pgdog_mirror_total_count{job=~"$pooler", database=~"$database", user=~"$user"}[$__rate_interval])) by (instance, database, user)`
	PGDogMirrorMirrored = `sum (rate(pgdog_mirror_mirrored_count{job=~"$pooler", database=~"$database", user=~"$user"}[$__rate_interval])) by (instance, database, user)`
	PGDogMirrorDropped  = `sum (rate(pgdog_mirror_dropped_count{job=~"$pooler", database=~"$database", user=~"$user"}[$__rate_interval])) by (instance, database, user)`
	PGDogMirrorError    = `sum (rate(pgdog_mirror_error_count{job=~"$pooler", database=~"$database", user=~"$user"}[$__rate_interval])) by (instance, database, user)`
	PGDogMirrorQueue    = `sum (pgdog_mirror_queue_length{job=~"$pooler", database=~"$database", user=~"$user"}) by (instance, database, user)`
)

// PGDog template-variable definitions.
const (
	PGDogPoolerValues   = `label_values(pgdog_total_sent, job)`
	PGDogHostValues     = `label_values(pgdog_total_sent{job=~"$pooler"}, host)`
	PGDogPortValues     = `label_values(pgdog_total_sent{job=~"$pooler"}, port)`
	PGDogShardValues    = `label_values(pgdog_total_sent{job=~"$pooler"}, shard)`
	PGDogRoleValues     = `label_values(pgdog_total_sent{job=~"$pooler"}, role)`
	PGDogDatabaseValues = `label_values(pgdog_total_sent{job=~"$pooler"}, database)`
	PGDogUserValues     = `label_values(pgdog_total_sent{job=~"$pooler"}, user)`
)
