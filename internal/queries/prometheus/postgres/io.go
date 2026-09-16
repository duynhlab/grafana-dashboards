package postgres

const (
	ReadOpsByBackend  = `sum by (backend_type) (rate(cnpg_pg_stat_io_reads{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	WriteOpsByBackend = `sum by (backend_type) (rate(cnpg_pg_stat_io_writes{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`

	ReadThroughput  = `sum (rate(cnpg_pg_stat_io_read_bytes{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	WriteThroughput = `sum (rate(cnpg_pg_stat_io_write_bytes{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`

	ReadIOTime  = `sum by (backend_type) (rate(cnpg_pg_stat_io_read_time{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	WriteIOTime = `sum by (backend_type) (rate(cnpg_pg_stat_io_write_time{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`

	BufferHits      = `sum (rate(cnpg_pg_stat_io_hits{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	BufferEvictions = `sum (rate(cnpg_pg_stat_io_evictions{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	BufferReuses    = `sum (rate(cnpg_pg_stat_io_reuses{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	RelationExtends = `sum by (backend_type) (rate(cnpg_pg_stat_io_extends{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	FsyncsByBackend = `sum by (backend_type) (rate(cnpg_pg_stat_io_fsyncs{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`
	FsyncTime       = `sum by (backend_type) (rate(cnpg_pg_stat_io_fsync_time{cnpg_io_cluster=~"$cluster", pod=~"$pod"}[$__rate_interval]))`

	ActiveBackendsByWaitClass = `sum by (wait_event_type) (cnpg_pg_wait_events_active_backends{cnpg_io_cluster=~"$cluster", pod=~"$pod"})`
	BackendsWaiting           = `sum (cnpg_backends_waiting_total{cnpg_io_cluster=~"$cluster", pod=~"$pod"})`
)
