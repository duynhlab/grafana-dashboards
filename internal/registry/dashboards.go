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
		Folder: standards.FolderKubernetes,
		Build:  kubernetes.ClusterOverview,
	},
	{
		UID:    "pg-io-waits",
		Folder: standards.FolderDatabases,
		Build:  postgres.PGIOWaits,
	},
	{
		UID:    "pg-maintenance",
		Folder: standards.FolderDatabases,
		Build:  postgres.PGMaintenance,
	},
	{
		UID:    "pg-query-performance",
		Folder: standards.FolderDatabases,
		Build:  postgres.PGQueryPerformance,
	},
	{
		UID:    "pg-exporter-instance",
		Folder: standards.FolderDatabases,
		Build:  postgres.PGExporterInstance,
	},
	{
		UID:    "pgdog",
		Folder: standards.FolderDatabases,
		Build:  postgres.PGDog,
	},
	{
		UID:    "temporal-worker",
		Folder: standards.FolderObservability,
		Build:  observability.TemporalWorker,
	},
	{
		UID:    "microservices-monitoring-001-otel",
		Folder: standards.FolderMicroservices,
		Build:  microservices.MicroservicesOTel,
	},
	{
		UID:    "business-otel",
		Folder: standards.FolderMicroservices,
		Build:  microservices.BusinessOTel,
	},
}
