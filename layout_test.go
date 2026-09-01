// Harness for the generated ratatui layout case tables in layout_cases_test.go,
// plus hand-written tests for the parts of Layout that have no table.

package catatui

import (
	"fmt"
	"strings"
	"testing"
)

// rng is a left..right span, matching the Range<u16> ratatui's tests use.
type rng struct{ start, end uint16 }

// pair is an (x, width) or (y, height) pair.
type pair struct{ a, b uint16 }

func constraintsString(cs []Constraint) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = c.String()
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// TestLayoutLetters is ratatui's `letters` harness: split a one-row area, fill
// each segment with a repeated letter, and compare the rendered row.
//
// Rendering the result rather than comparing Rects is what makes these readable
// — "aaabbb" says at a glance what "[Rect{0,0,3,1} Rect{3,0,3,1}]" does not.
func TestLayoutLetters(t *testing.T) {
	for _, c := range lettersCases {
		area := NewRect(0, 0, c.width, 1)
		segments := HorizontalLayout(c.constraints...).Flex(c.flex).Split(area)
		if len(segments) != len(c.constraints) {
			t.Errorf("%s/%s %s width=%d: got %d segments, want %d",
				c.name, c.flex, constraintsString(c.constraints), c.width,
				len(segments), len(c.constraints))
			continue
		}

		buf := NewBuffer(area)
		for i, seg := range segments {
			letter := string(rune('a' + i))
			buf.SetString(seg.X, seg.Y, strings.Repeat(letter, int(seg.Width)), NewStyle())
		}
		if got := buf.String(); got != c.expected {
			t.Errorf("%s/%s %s width=%d:\n  got  %q\n  want %q",
				c.name, c.flex, constraintsString(c.constraints), c.width, got, c.expected)
		}
	}
}

func TestLayoutRanges(t *testing.T) {
	for _, c := range rangeCases {
		got := splitRanges(HorizontalLayout(c.constraints...).Flex(c.flex), NewRect(0, 0, c.width, 1))
		if !equalRanges(got, c.expected) {
			t.Errorf("%s/%s %s width=%d:\n  got  %v\n  want %v",
				c.name, c.flex, constraintsString(c.constraints), c.width, got, c.expected)
		}
	}
}

func TestLayoutFlexRanges(t *testing.T) {
	for _, c := range flexRangeCases {
		got := splitRanges(HorizontalLayout(c.constraints...).Flex(c.flex), NewRect(0, 0, 100, 1))
		if !equalRanges(got, c.expected) {
			t.Errorf("%s/%s %s:\n  got  %v\n  want %v",
				c.name, c.flex, constraintsString(c.constraints), got, c.expected)
		}
	}
}

func TestLayoutPairs(t *testing.T) {
	for _, c := range pairCases {
		l := HorizontalLayout(c.constraints...).Flex(c.flex).Spacing(spacingOf(c.spacing))
		got := splitPairs(l, NewRect(0, 0, 100, 1))
		if !equalPairs(got, c.expected) {
			t.Errorf("%s/%s %s spacing=%d:\n  got  %v\n  want %v",
				c.name, c.flex, constraintsString(c.constraints), c.spacing, got, c.expected)
		}
	}
}

// TestLayoutSpacers checks the gaps between segments, which widgets that
// collapse borders draw into. There is always one more spacer than segment.
func TestLayoutSpacers(t *testing.T) {
	for _, c := range spacerCases {
		l := HorizontalLayout(c.constraints...).Flex(c.flex).Spacing(spacingOf(c.spacing))
		_, spacers := l.SplitWithSpacers(NewRect(0, 0, 100, 1))
		if len(spacers) != len(c.constraints)+1 {
			t.Errorf("%s: got %d spacers, want %d", c.name, len(spacers), len(c.constraints)+1)
			continue
		}
		got := make([]pair, len(spacers))
		for i, r := range spacers {
			got[i] = pair{r.X, r.Width}
		}
		if !equalPairs(got, c.expected) {
			t.Errorf("%s/%s %s spacing=%d:\n  got  %v\n  want %v",
				c.name, c.flex, constraintsString(c.constraints), c.spacing, got, c.expected)
		}
	}
}

// TestLayoutWidthsAcrossFlex checks that Length keeps its priority over Min and
// Max in every non-Legacy flex mode, which is what ratatui's
// length_is_higher_priority_in_flex asserts.
func TestLayoutWidthsAcrossFlex(t *testing.T) {
	flexes := []Flex{FlexStart, FlexEnd, FlexCenter, FlexSpaceAround, FlexSpaceEvenly, FlexSpaceBetween}
	for _, c := range widthCases {
		for _, flex := range flexes {
			segments := HorizontalLayout(c.constraints...).Flex(flex).Split(NewRect(0, 0, 100, 1))
			got := make([]uint16, len(segments))
			for i, r := range segments {
				got[i] = r.Width
			}
			if fmt.Sprint(got) != fmt.Sprint(c.expected) {
				t.Errorf("%s/%s %s:\n  got  %v\n  want %v",
					c.name, flex, constraintsString(c.constraints), got, c.expected)
			}
		}
	}
}

func spacingOf(n int) Spacing {
	if n < 0 {
		return Overlap(uint16(-n))
	}
	return Space(uint16(n))
}

func splitRanges(l Layout, area Rect) []rng {
	segments := l.Split(area)
	out := make([]rng, len(segments))
	for i, r := range segments {
		out[i] = rng{r.Left(), r.Right()}
	}
	return out
}

func splitPairs(l Layout, area Rect) []pair {
	segments := l.Split(area)
	out := make([]pair, len(segments))
	for i, r := range segments {
		out[i] = pair{r.X, r.Width}
	}
	return out
}

func equalRanges(a, b []rng) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalPairs(a, b []pair) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- hand-written tests ---------------------------------------------------

func TestLayoutDefault(t *testing.T) {
	l := NewLayout()
	if l.GetDirection() != Vertical {
		t.Errorf("default direction = %v, want Vertical", l.GetDirection())
	}
	if l.GetFlex() != FlexStart {
		t.Errorf("default flex = %v, want Start", l.GetFlex())
	}
	if l.GetSpacing() != Space(0) {
		t.Errorf("default spacing = %v, want Space(0)", l.GetSpacing())
	}
	if l.GetMargin() != (Margin{}) {
		t.Errorf("default margin = %+v, want zero", l.GetMargin())
	}
}

// TestLayoutVerticalSplitByHeight is ratatui's vertical_split_by_height: the
// segments must tile the area exactly, with no gaps and no overlap.
func TestLayoutVerticalSplitByHeight(t *testing.T) {
	target := NewRect(2, 2, 10, 10)
	segments := VerticalLayout(Percentage(10), Max(5), Min(1)).
		Flex(FlexLegacy).Split(target)

	var totalHeight uint16
	for _, r := range segments {
		totalHeight += r.Height
		if r.X != target.X || r.Width != target.Width {
			t.Errorf("a vertical split must preserve x and width: got %+v, want x=%d width=%d",
				r, target.X, target.Width)
		}
	}
	if totalHeight != target.Height {
		t.Errorf("segment heights sum to %d, want %d", totalHeight, target.Height)
	}
	// And they must be contiguous.
	for i := 1; i < len(segments); i++ {
		if segments[i].Y != segments[i-1].Bottom() {
			t.Errorf("segment %d starts at y=%d but the previous ends at %d",
				i, segments[i].Y, segments[i-1].Bottom())
		}
	}
}

func TestLayoutMargin(t *testing.T) {
	area := NewRect(0, 0, 10, 10)
	segments := VerticalLayout(Fill(1)).Margin(2).Split(area)
	if got, want := segments[0], NewRect(2, 2, 6, 6); got != want {
		t.Errorf("Margin(2) gave %+v, want %+v", got, want)
	}

	segments = VerticalLayout(Fill(1)).HorizontalMargin(3).Split(area)
	if got, want := segments[0], NewRect(3, 0, 4, 10); got != want {
		t.Errorf("HorizontalMargin(3) gave %+v, want %+v", got, want)
	}

	segments = VerticalLayout(Fill(1)).VerticalMargin(3).Split(area)
	if got, want := segments[0], NewRect(0, 3, 10, 4); got != want {
		t.Errorf("VerticalMargin(3) gave %+v, want %+v", got, want)
	}
}

// TestLayoutIsDeterministic guards the ordered-iteration change in the solver at
// the level users actually see. A layout that shifted between runs of the same
// program would be a maddening bug to chase.
func TestLayoutIsDeterministic(t *testing.T) {
	l := HorizontalLayout(Fill(1), Fill(1), Fill(1), Min(5), Percentage(20))
	area := NewRect(0, 0, 37, 1)
	first := l.Split(area)
	for i := range 200 {
		got := l.Split(area)
		for j := range got {
			if got[j] != first[j] {
				t.Fatalf("run %d segment %d = %+v, but the first run gave %+v",
					i, j, got[j], first[j])
			}
		}
	}
}

// TestLayoutSplitReturnsOnePerConstraint holds even in areas too small to
// satisfy anything, so that indexing the result is always safe.
func TestLayoutSplitReturnsOnePerConstraint(t *testing.T) {
	constraints := []Constraint{Length(10), Min(20), Percentage(50), Fill(1)}
	for _, size := range []uint16{0, 1, 2, 5, 40, 200} {
		got := HorizontalLayout(constraints...).Split(NewRect(0, 0, size, 1))
		if len(got) != len(constraints) {
			t.Errorf("width %d: got %d segments, want %d", size, len(got), len(constraints))
		}
	}
}

func TestLayoutNoConstraints(t *testing.T) {
	got := HorizontalLayout().Split(NewRect(0, 0, 10, 1))
	if len(got) != 0 {
		t.Errorf("a layout with no constraints should split into nothing, got %v", got)
	}
}
