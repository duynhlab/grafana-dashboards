package panels

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	"github.com/grafana/grafana-foundation-sdk/go/text"
)

// Text renders a static markdown panel. Several ported boards open with one
// explaining how to read the panels below it, usually which flat lines are
// expected and which are a signal; that prose is part of the board and is kept
// rather than dropped on port.
//
// It carries no query, so it is the one panel helper that does not take an
// expression.
func Text(title, markdown string) cog.Builder[dashboardv2.PanelKind] {
	return dashboardv2.NewPanelBuilder().
		Title(title).
		Visualization(text.NewVisualizationV2Builder().
			Mode(text.TextModeMarkdown).
			Content(markdown))
}
