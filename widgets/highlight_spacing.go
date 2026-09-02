// Port of ratatui-widgets/src/table/highlight_spacing.rs @ ratatui-v0.30.2

package widgets

// HighlightSpacing decides when a List or Table reserves a column for the
// highlight symbol, which is what keeps the content from shifting sideways as
// the selection appears and disappears.
type HighlightSpacing uint8

const (
	// HighlightSpacingWhenSelected reserves the column only while something is
	// selected. This is the default.
	HighlightSpacingWhenSelected HighlightSpacing = iota
	// HighlightSpacingAlways reserves the column whether or not anything is
	// selected, so the content never moves.
	HighlightSpacingAlways
	// HighlightSpacingNever reserves no column, even when something is selected.
	HighlightSpacingNever
)

// ShouldAdd reports whether the highlight column is reserved given whether
// there is a selection.
func (h HighlightSpacing) ShouldAdd(hasSelection bool) bool {
	switch h {
	case HighlightSpacingAlways:
		return true
	case HighlightSpacingNever:
		return false
	default:
		return hasSelection
	}
}

// String returns the variant's ratatui name: "Always", "WhenSelected" or
// "Never".
func (h HighlightSpacing) String() string {
	switch h {
	case HighlightSpacingAlways:
		return "Always"
	case HighlightSpacingNever:
		return "Never"
	default:
		return "WhenSelected"
	}
}
