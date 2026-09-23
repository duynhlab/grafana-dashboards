package standards

const (
	FolderKubernetes    = "Kubernetes"
	FolderDatabases     = "Databases"
	FolderObservability = "Observability"
	FolderMicroservices = "Microservices"

	// FolderPlatform holds the cluster-support boards: the things every workload
	// depends on but no workload owns.
	FolderPlatform = "Platform"

	// FolderGateway is the north-south edge. The vendored Envoy Gateway boards
	// live in the same Grafana folder, so the two delivery paths meet here.
	FolderGateway = "API Gateway"
)
