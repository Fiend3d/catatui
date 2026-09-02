// Tests ported from ratatui-widgets/src/scrollbar.rs @ ratatui-v0.30.2

package widgets

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// scrollbarNoArrows is ratatui's scrollbar_no_arrows fixture.
func scrollbarNoArrows() Scrollbar {
	return NewScrollbar(ScrollbarHorizontalTop).
		BeginSymbolNone().
		EndSymbolNone().
		TrackSymbol("-").
		ThumbSymbol("#")
}

// scrollbarCase is one rstest case: the expected single-line rendering, the
// position and the content length.
type scrollbarCase struct {
	name          string
	expected      string
	position      int
	contentLength int
}

// renderScrollbarLine renders a scrollbar into a one-row buffer as wide as the
// expected string and compares.
func renderScrollbarLine(t *testing.T, sb Scrollbar, state ScrollbarState, expected string) {
	t.Helper()
	width := uint16(catatui.StringWidth(expected))
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, 1))
	sb.RenderStateful(buf.Area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(expected))
}

func TestScrollDirectionToString(t *testing.T) {
	if got := ScrollForward.String(); got != "Forward" {
		t.Errorf("got %q, want Forward", got)
	}
	if got := ScrollBackward.String(); got != "Backward" {
		t.Errorf("got %q, want Backward", got)
	}
}

func TestScrollbarOrientationToString(t *testing.T) {
	cases := map[ScrollbarOrientation]string{
		ScrollbarVerticalRight:    "VerticalRight",
		ScrollbarVerticalLeft:     "VerticalLeft",
		ScrollbarHorizontalBottom: "HorizontalBottom",
		ScrollbarHorizontalTop:    "HorizontalTop",
	}
	for o, want := range cases {
		if got := o.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestScrollbarRenderSimplest(t *testing.T) {
	cases := []scrollbarCase{
		{"area 2 position 0", "#-", 0, 2},
		{"area 2 position 1", "-#", 1, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

func TestScrollbarRenderSimple(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "#####-----", 0, 10},
		{"position 1", "-#####----", 1, 10},
		{"position 2", "-#####----", 2, 10},
		{"position 3", "--#####---", 3, 10},
		{"position 4", "--#####---", 4, 10},
		{"position 5", "---#####--", 5, 10},
		{"position 6", "---#####--", 6, 10},
		{"position 7", "----#####-", 7, 10},
		{"position 8", "----#####-", 8, 10},
		{"position 9", "-----#####", 9, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

func TestScrollbarRenderNobar(t *testing.T) {
	state := NewScrollbarState(0).Position(0)
	renderScrollbarLine(t, scrollbarNoArrows(), state, "          ")
}

func TestScrollbarRenderFullbar(t *testing.T) {
	cases := []scrollbarCase{
		{"fullbar position 0", "##########", 0, 1},
		{"almost fullbar position 0", "#########-", 0, 2},
		{"almost fullbar position 1", "-#########", 1, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

func TestScrollbarRenderAlmostFullbar(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "#########-", 0, 2},
		{"position 1", "-#########", 1, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

// doubleHorizontalCases are shared by the tests that render the default
// double-line horizontal set without arrows.
var doubleHorizontalCases = []scrollbarCase{
	{"position 0", "█████═════", 0, 10},
	{"position 1", "═█████════", 1, 10},
	{"position 2", "═█████════", 2, 10},
	{"position 3", "══█████═══", 3, 10},
	{"position 4", "══█████═══", 4, 10},
	{"position 5", "═══█████══", 5, 10},
	{"position 6", "═══█████══", 6, 10},
	{"position 7", "════█████═", 7, 10},
	{"position 8", "════█████═", 8, 10},
	{"position 9", "═════█████", 9, 10},
	{"position out of bounds", "═════█████", 100, 10},
}

func TestScrollbarRenderWithoutSymbols(t *testing.T) {
	for _, c := range doubleHorizontalCases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			sb := NewScrollbar(ScrollbarHorizontalBottom).BeginSymbolNone().EndSymbolNone()
			renderScrollbarLine(t, sb, state, c.expected)
		})
	}
}

func TestScrollbarRenderWithoutTrackSymbols(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "█████     ", 0, 10},
		{"position 1", " █████    ", 1, 10},
		{"position 2", " █████    ", 2, 10},
		{"position 3", "  █████   ", 3, 10},
		{"position 4", "  █████   ", 4, 10},
		{"position 5", "   █████  ", 5, 10},
		{"position 6", "   █████  ", 6, 10},
		{"position 7", "    █████ ", 7, 10},
		{"position 8", "    █████ ", 8, 10},
		{"position 9", "     █████", 9, 10},
		{"position out of bounds", "     █████", 100, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			sb := NewScrollbar(ScrollbarHorizontalBottom).
				TrackSymbolNone().
				BeginSymbolNone().
				EndSymbolNone()
			renderScrollbarLine(t, sb, state, c.expected)
		})
	}
}

// TestScrollbarRenderWithoutTrackSymbolsOverContent pins that a missing track
// symbol leaves the cells underneath untouched rather than blanking them.
func TestScrollbarRenderWithoutTrackSymbolsOverContent(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "█████-----", 0, 10},
		{"position 1", "-█████----", 1, 10},
		{"position 2", "-█████----", 2, 10},
		{"position 3", "--█████---", 3, 10},
		{"position 4", "--█████---", 4, 10},
		{"position 5", "---█████--", 5, 10},
		{"position 6", "---█████--", 6, 10},
		{"position 7", "----█████-", 7, 10},
		{"position 8", "----█████-", 8, 10},
		{"position 9", "-----█████", 9, 10},
		{"position out of bounds", "-----█████", 100, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			width := uint16(catatui.StringWidth(c.expected))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, 1))
			catatui.TextFromString(strings.Repeat("-", int(width))).Render(buf.Area, buf)
			state := NewScrollbarState(c.contentLength).Position(c.position)
			NewScrollbar(ScrollbarHorizontalBottom).
				TrackSymbolNone().
				BeginSymbolNone().
				EndSymbolNone().
				RenderStateful(buf.Area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.expected))
		})
	}
}

// arrowCases are shared by the tests that render "<", "-", "#", ">" symbols.
var arrowCases = []scrollbarCase{
	{"position 0", "<####---->", 0, 10},
	{"position 1", "<####---->", 1, 10},
	{"position 2", "<-####--->", 2, 10},
	{"position 3", "<-####--->", 3, 10},
	{"position 4", "<--####-->", 4, 10},
	{"position 5", "<--####-->", 5, 10},
	{"position 6", "<---####->", 6, 10},
	{"position 7", "<---####->", 7, 10},
	{"position 8", "<---####->", 8, 10},
	{"position 9", "<----####>", 9, 10},
	{"position one out of bounds", "<----####>", 10, 10},
	{"position few out of bounds", "<----####>", 15, 10},
	{"position very many out of bounds", "<----####>", 500, 10},
}

func scrollbarWithArrows(o ScrollbarOrientation) Scrollbar {
	return NewScrollbar(o).
		BeginSymbol("<").
		EndSymbol(">").
		TrackSymbol("-").
		ThumbSymbol("#")
}

func TestScrollbarRenderWithSymbols(t *testing.T) {
	for _, c := range arrowCases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).Position(c.position)
			renderScrollbarLine(t, scrollbarWithArrows(ScrollbarHorizontalTop), state, c.expected)
		})
	}
}

func TestScrollbarRenderHorizontalBottom(t *testing.T) {
	for _, c := range doubleHorizontalCases {
		t.Run(c.name, func(t *testing.T) {
			width := uint16(catatui.StringWidth(c.expected))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, 2))
			state := NewScrollbarState(c.contentLength).Position(c.position)
			NewScrollbar(ScrollbarHorizontalBottom).
				BeginSymbolNone().
				EndSymbolNone().
				RenderStateful(buf.Area, buf, &state)
			empty := strings.Repeat(" ", int(width))
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(empty, c.expected))
		})
	}
}

func TestScrollbarRenderHorizontalTop(t *testing.T) {
	for _, c := range doubleHorizontalCases {
		t.Run(c.name, func(t *testing.T) {
			width := uint16(catatui.StringWidth(c.expected))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, 2))
			state := NewScrollbarState(c.contentLength).Position(c.position)
			NewScrollbar(ScrollbarHorizontalTop).
				BeginSymbolNone().
				EndSymbolNone().
				RenderStateful(buf.Area, buf, &state)
			empty := strings.Repeat(" ", int(width))
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.expected, empty))
		})
	}
}

// verticalLines turns a horizontal expectation into one line per character,
// with the character placed in the given column of a five-column row.
func verticalLines(expected string, column int) []string {
	var lines []string
	for _, r := range expected {
		pad := strings.Repeat(" ", column)
		rest := strings.Repeat(" ", 4-column)
		lines = append(lines, pad+string(r)+rest)
	}
	return lines
}

func TestScrollbarRenderVerticalLeft(t *testing.T) {
	for _, c := range arrowCases {
		t.Run(c.name, func(t *testing.T) {
			size := uint16(catatui.StringWidth(c.expected))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, size))
			state := NewScrollbarState(c.contentLength).Position(c.position)
			scrollbarWithArrows(ScrollbarVerticalLeft).RenderStateful(buf.Area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(verticalLines(c.expected, 0)...))
		})
	}
}

func TestScrollbarRenderVerticalRight(t *testing.T) {
	for _, c := range arrowCases {
		t.Run(c.name, func(t *testing.T) {
			size := uint16(catatui.StringWidth(c.expected))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, size))
			state := NewScrollbarState(c.contentLength).Position(c.position)
			scrollbarWithArrows(ScrollbarVerticalRight).RenderStateful(buf.Area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(verticalLines(c.expected, 4)...))
		})
	}
}

func TestScrollbarCustomViewportLength(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "##--------", 0, 10},
		{"position 1", "-##-------", 1, 10},
		{"position 2", "--##------", 2, 10},
		{"position 3", "---##-----", 3, 10},
		{"position 4", "----##----", 4, 10},
		{"position 5", "-----##---", 5, 10},
		{"position 6", "-----##---", 6, 10},
		{"position 7", "------##--", 7, 10},
		{"position 8", "-------##-", 8, 10},
		{"position 9", "--------##", 9, 10},
		{"position one out of bounds", "--------##", 10, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).
				Position(c.position).
				ViewportContentLength(2)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

// TestScrollbarThumbVisibleOnVerySmallTrack is ratatui's regression test for
// PR #959: the thumb must still show when the viewport is tiny next to the
// content.
func TestScrollbarThumbVisibleOnVerySmallTrack(t *testing.T) {
	cases := []scrollbarCase{
		{"position 0", "#----", 0, 100},
		{"position 10", "#----", 10, 100},
		{"position 20", "-#---", 20, 100},
		{"position 30", "-#---", 30, 100},
		{"position 40", "--#--", 40, 100},
		{"position 50", "--#--", 50, 100},
		{"position 60", "---#-", 60, 100},
		{"position 70", "---#-", 70, 100},
		{"position 80", "----#", 80, 100},
		{"position 90", "----#", 90, 100},
		{"position one out of bounds", "----#", 100, 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewScrollbarState(c.contentLength).
				Position(c.position).
				ViewportContentLength(2)
			renderScrollbarLine(t, scrollbarNoArrows(), state, c.expected)
		})
	}
}

func TestScrollbarDoNotRenderWithEmptyArea(t *testing.T) {
	cases := []struct {
		name          string
		width, height uint16
	}{
		{"scrollbar height 0", 10, 0},
		{"scrollbar width 0", 0, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sb := scrollbarWithArrows(ScrollbarVerticalRight)
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10))
			state := NewScrollbarState(10)
			sb.RenderStateful(catatui.NewRect(0, 0, c.width, c.height), buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10)))
		})
	}
}

var allOrientations = []ScrollbarOrientation{
	ScrollbarVerticalLeft,
	ScrollbarVerticalRight,
	ScrollbarHorizontalTop,
	ScrollbarHorizontalBottom,
}

func TestScrollbarRenderInMinimalBuffer(t *testing.T) {
	for _, o := range allOrientations {
		t.Run(o.String(), func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
			state := NewScrollbarState(10).Position(5)
			// This must not panic, even though the buffer is too small for the bar.
			NewScrollbar(o).RenderStateful(buf.Area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
		})
	}
}

func TestScrollbarRenderInZeroSizeBuffer(t *testing.T) {
	for _, o := range allOrientations {
		t.Run(o.String(), func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.ZeroRect)
			state := NewScrollbarState(10).Position(5)
			// This must not panic, even though the buffer has zero size.
			NewScrollbar(o).RenderStateful(buf.Area, buf, &state)
		})
	}
}

func TestScrollbarPartLengthsReturnsZerosWhenTrackLenIsZero(t *testing.T) {
	cases := []struct {
		name string
		o    ScrollbarOrientation
		area catatui.Rect
	}{
		{"horizontal width eq arrows", ScrollbarHorizontalTop, catatui.NewRect(0, 0, 2, 1)},
		{"horizontal width lt arrows", ScrollbarHorizontalTop, catatui.NewRect(0, 0, 1, 1)},
		{"vertical height eq arrows", ScrollbarVerticalLeft, catatui.NewRect(0, 0, 1, 2)},
		{"vertical height lt arrows", ScrollbarVerticalLeft, catatui.NewRect(0, 0, 1, 1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sb := scrollbarWithArrows(c.o)
			state := NewScrollbarState(10).Position(5).ViewportContentLength(2)
			start, thumb, end := sb.partLengths(c.area, &state)
			if start != 0 || thumb != 0 || end != 0 {
				t.Errorf("got (%d, %d, %d), want (0, 0, 0)", start, thumb, end)
			}
		})
	}
}

func TestScrollbarPartLengthsReturnsZerosWhenAreaDimensionIsZeroEvenWithoutArrows(t *testing.T) {
	sb := scrollbarNoArrows()
	state := NewScrollbarState(10).Position(3).ViewportContentLength(2)
	start, thumb, end := sb.partLengths(catatui.NewRect(0, 0, 0, 1), &state)
	if start != 0 || thumb != 0 || end != 0 {
		t.Errorf("got (%d, %d, %d), want (0, 0, 0)", start, thumb, end)
	}
}

// TestScrollbarThumbStaysWithinTrackForLargeThumbAtEnd is the regression test
// for ratatui issue #2582: a thumb large relative to the track must not overrun
// it at the end, or it pushes the end symbol out of the area.
func TestScrollbarThumbStaysWithinTrackForLargeThumbAtEnd(t *testing.T) {
	sb := NewScrollbar(ScrollbarVerticalRight)
	// Height 24 minus the two arrow heads leaves a track of 22.
	area := catatui.NewRect(0, 0, 1, 24)
	state := NewScrollbarState(9).Position(8)

	start, thumb, end := sb.partLengths(area, &state)
	if start+thumb > 22 {
		t.Errorf("thumb overruns the track: start=%d + thumb=%d > 22", start, thumb)
	}
	if start+thumb+end != 22 {
		t.Errorf("parts must sum to the track length: got %d", start+thumb+end)
	}
}

// TestScrollbarRenderKeepsEndSymbolForLargeThumb is the visual form of the
// test above.
func TestScrollbarRenderKeepsEndSymbolForLargeThumb(t *testing.T) {
	state := NewScrollbarState(9).Position(8)
	renderScrollbarLine(t, scrollbarWithArrows(ScrollbarHorizontalTop), state, "<-----#################>")
}

// TestScrollbarStateNavigation covers the state's movement methods, which
// ratatui exercises only through its doc examples.
func TestScrollbarStateNavigation(t *testing.T) {
	state := NewScrollbarState(3)
	state.Next()
	state.Next()
	state.Next()
	if got := state.GetPosition(); got != 2 {
		t.Errorf("after Next past the end: got %d, want 2", got)
	}
	state.Prev()
	if got := state.GetPosition(); got != 1 {
		t.Errorf("after Prev: got %d, want 1", got)
	}
	state.First()
	state.Prev()
	if got := state.GetPosition(); got != 0 {
		t.Errorf("after First and Prev: got %d, want 0", got)
	}
	state.Last()
	if got := state.GetPosition(); got != 2 {
		t.Errorf("after Last: got %d, want 2", got)
	}
	state.Scroll(ScrollBackward)
	if got := state.GetPosition(); got != 1 {
		t.Errorf("after Scroll backward: got %d, want 1", got)
	}
	state.Scroll(ScrollForward)
	if got := state.GetPosition(); got != 2 {
		t.Errorf("after Scroll forward: got %d, want 2", got)
	}

	empty := NewScrollbarState(0)
	empty.Next()
	empty.Last()
	if got := empty.GetPosition(); got != 0 {
		t.Errorf("empty content must stay at 0, got %d", got)
	}
}
