package alerts

type Rule struct {
	UID   string
	Title string
	// Domain is the owning Go package, a standards.Domain* constant, and the
	// directory segment under generated/alerts. It is required.
	Domain      string
	Folder      string
	Group       string
	Expr        string
	For         string
	Threshold   float64
	Labels      map[string]string
	Annotations map[string]string
}

// orZero makes a count() alert read 0 when nothing matches. count() over an
// empty vector returns no series, which Grafana evaluates as NoData rather
// than Normal, so a healthy cluster kept every such rule in NoData.
func orZero(expr string) string {
	return expr + " or vector(0)"
}
