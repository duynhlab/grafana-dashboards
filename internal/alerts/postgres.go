package alerts

import (
	pgqueries "github.com/duynhlab/grafana-dashboards/internal/queries/prometheus/postgres"
	"github.com/duynhlab/grafana-dashboards/internal/standards"
)

func BackendsWaiting() Rule {
	return Rule{
		UID:       "postgres_backends_waiting",
		Title:     "PostgreSQL backends waiting",
		Folder:    standards.FolderDatabases,
		Group:     "databases",
		Expr:      pgqueries.BackendsWaiting,
		For:       "10m",
		Threshold: 0,
		Labels: map[string]string{
			standards.LabelSeverity:  standards.SeverityWarning,
			standards.LabelDomain:    "postgres",
			standards.LabelComponent: "io-waits",
		},
		Annotations: map[string]string{
			standards.AnnotationSummary:      "PostgreSQL backends are waiting",
			standards.AnnotationDescription:  "cnpg_backends_waiting_total is above zero for the evaluation window.",
			standards.AnnotationDashboardUID: "pg-io-waits",
		},
	}
}
