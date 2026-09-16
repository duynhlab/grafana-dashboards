package registry

import "github.com/duynhlab/grafana-dashboards/internal/alerts"

var Alerts = []alerts.Rule{
	alerts.CrashLoopingPods(),
	alerts.PendingPods(),
	alerts.PVCsAtRisk(),
	alerts.BackendsWaiting(),
}
