package registry

import (
	"github.com/duynhlab/grafana-dashboards/internal/dashboards/kubernetes"
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
}
