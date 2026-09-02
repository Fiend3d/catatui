// Tests ported from ratatui-widgets/src/table/highlight_spacing.rs @ ratatui-v0.30.2

package widgets

import "testing"

func TestHighlightSpacingToString(t *testing.T) {
	cases := []struct {
		h    HighlightSpacing
		want string
	}{
		{HighlightSpacingAlways, "Always"},
		{HighlightSpacingWhenSelected, "WhenSelected"},
		{HighlightSpacingNever, "Never"},
	}
	for _, c := range cases {
		if got := c.h.String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
	}
}

func TestHighlightSpacingShouldAdd(t *testing.T) {
	cases := []struct {
		h            HighlightSpacing
		hasSelection bool
		want         bool
	}{
		{HighlightSpacingAlways, false, true},
		{HighlightSpacingAlways, true, true},
		{HighlightSpacingWhenSelected, false, false},
		{HighlightSpacingWhenSelected, true, true},
		{HighlightSpacingNever, false, false},
		{HighlightSpacingNever, true, false},
	}
	for _, c := range cases {
		if got := c.h.ShouldAdd(c.hasSelection); got != c.want {
			t.Errorf("%v.ShouldAdd(%v) = %v, want %v", c.h, c.hasSelection, got, c.want)
		}
	}
}

func TestHighlightSpacingDefaultIsWhenSelected(t *testing.T) {
	var h HighlightSpacing
	if h != HighlightSpacingWhenSelected {
		t.Errorf("the zero value should be WhenSelected, got %v", h)
	}
}
