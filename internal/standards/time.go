package standards

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

func TimeSettings(from, to, refresh string) cog.Builder[dashboardv2.TimeSettingsSpec] {
	b := dashboardv2.NewTimeSettingsBuilder().
		From(from).
		To(to).
		Timezone("browser")
	if refresh != "" {
		b = b.AutoRefresh(refresh)
	}
	return b
}
