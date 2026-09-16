package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
)

// Threshold is a single colored step in a panel's threshold configuration.
// A nil Value represents the base (-Infinity) step.
type Threshold struct {
	Color string
	Value *float64
}

// Ptr returns a pointer to the supplied float64. It is a convenience helper for
// declaring threshold step values inline.
func Ptr(v float64) *float64 { return &v }

// Thr builds a threshold step. Pass a nil value for the base step.
func Thr(color string, value *float64) Threshold {
	return Threshold{Color: color, Value: value}
}

// thresholds converts a slice of steps into an absolute ThresholdsConfig builder.
func thresholds(steps []Threshold) cog.Builder[dashboardv2.ThresholdsConfig] {
	out := make([]dashboardv2.Threshold, 0, len(steps))
	for _, s := range steps {
		out = append(out, dashboardv2.Threshold{Color: s.Color, Value: s.Value})
	}
	return dashboardv2.NewThresholdsConfigBuilder().
		Mode(dashboardv2.ThresholdsModeAbsolute).
		Steps(out)
}
