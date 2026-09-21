package registry

import (
	"github.com/duynhlab/grafana-dashboards/internal/dashboards/kubernetes"
	"github.com/duynhlab/grafana-dashboards/internal/dashboards/microservices"
	"github.com/duynhlab/grafana-dashboards/internal/dashboards/observability"
	"github.com/duynhlab/grafana-dashboards/internal/dashboards/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

var Dashboards = []DashboardDefinition{
	{
		UID:    "kubernetes-cluster-overview",
		Domain: standards.DomainKubernetes,
		Folder: standards.FolderKubernetes,
		Build:  kubernetes.ClusterOverview,
	},
	{
		UID:    "pg-io-waits",
		Domain: standards.DomainPostgres,
		Folder: standards.FolderDatabases,
		Build:  postgres.PGIOWaits,
	},
	{
		UID:    "pg-maintenance",
		Domain: standards.DomainPostgres,
		Folder: standards.FolderDatabases,
		Build:  postgres.PGMaintenance,
	},
	{
		UID:    "pg-query-performance",
		Domain: standards.DomainPostgres,
		Folder: standards.FolderDatabases,
		Build:  postgres.PGQueryPerformance,
	},
	{
		UID:    "pg-exporter-instance",
		Domain: standards.DomainPostgres,
		Folder: standards.FolderDatabases,
		Build:  postgres.PGExporterInstance,
	},
	{
		UID:    "pgdog",
		Domain: standards.DomainPostgres,
		Folder: standards.FolderDatabases,
		Build:  postgres.PGDog,
	},
	{
		UID:    "temporal-worker",
		Domain: standards.DomainObservability,
		Folder: standards.FolderObservability,
		Build:  observability.TemporalWorker,
	},
	{
		UID:    "microservices-monitoring-001-otel",
		Domain: standards.DomainMicroservices,
		Folder: standards.FolderMicroservices,
		Build:  microservices.MicroservicesOTel,
	},
	{
		UID:    "business-otel",
		Domain: standards.DomainMicroservices,
		Folder: standards.FolderMicroservices,
		Build:  microservices.BusinessOTel,
	},
}
