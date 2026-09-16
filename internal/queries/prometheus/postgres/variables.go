package postgres

// Shared CNPG template-variable definitions. The cluster/datname pair is reused
// across the CNPG dashboards (io-waits, maintenance, query-performance).
const (
	ClusterLabelValues = `label_values(cnpg_collector_up, cnpg_io_cluster)`
	PodLabelValues     = `label_values(cnpg_collector_up{cnpg_io_cluster=~"$cluster"}, pod)`
	DatnameLabelValues = `label_values(cnpg_pg_stat_statements_calls{cnpg_io_cluster=~"$cluster"}, datname)`
)
