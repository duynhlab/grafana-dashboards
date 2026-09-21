package standards

// Domains name the Go packages that own a family of resources. They double as
// the directory segment under generated/dashboards and generated/alerts, so the
// generated tree mirrors internal/dashboards and internal/alerts.
//
// A domain is not a Grafana folder. The five PostgreSQL boards live in the
// postgres domain but in the Databases folder, so the two names are declared
// separately and neither is derived from the other.
const (
	DomainKubernetes    = "kubernetes"
	DomainPostgres      = "postgres"
	DomainObservability = "observability"
	DomainMicroservices = "microservices"
)
