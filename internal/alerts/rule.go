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
