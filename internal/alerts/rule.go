package alerts

type Rule struct {
	UID         string
	Title       string
	Folder      string
	Group       string
	Expr        string
	For         string
	Threshold   float64
	Labels      map[string]string
	Annotations map[string]string
}
