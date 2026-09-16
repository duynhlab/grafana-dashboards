package registry

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

type DashboardDefinition struct {
	UID    string
	Folder string
	Build  func() cog.Builder[dashboardv2.Dashboard]
}

func Folders() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(folder string) {
		if folder == "" {
			return
		}
		if _, ok := seen[folder]; ok {
			return
		}
		seen[folder] = struct{}{}
		out = append(out, folder)
	}
	for _, d := range Dashboards {
		add(d.Folder)
	}
	for _, a := range Alerts {
		add(a.Folder)
	}
	return out
}
