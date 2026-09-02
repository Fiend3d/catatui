package term

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// vtScreen is a minimal terminal: enough of VT to interpret what Backend emits.
//
// Its point is the advance function. A real terminal decides for itself how far
// the cursor moves after a cluster, and its answer need not match Unicode's
// tables — that disagreement is what corrupted rows of Devanagari and Thai. The
// tests below drive this screen with a deliberately wrong advance, so a backend
// that trusts its own prediction fails and one that re-anchors does not.
type vtScreen struct {
	cells          [][]string
	w, h           uint16
	x, y           uint16
	advance        func(string) uint16
	wroteOffScreen bool
}

func newVTScreen(w, h uint16, advance func(string) uint16) *vtScreen {
	cells := make([][]string, h)
	for i := range cells {
		cells[i] = make([]string, w)
		for j := range cells[i] {
			cells[i][j] = " "
		}
	}
	return &vtScreen{cells: cells, w: w, h: h, advance: advance}
}

// feed interprets a stream of output: CUP moves, SGR (ignored, it does not move
// the cursor), and printable text.
func (s *vtScreen) feed(out string) {
	for len(out) > 0 {
		if out[0] == 0x1b {
			n := s.escape(out)
			out = out[n:]
			continue
		}
		// One grapheme cluster of printable text.
		var cluster string
		for g := range catatui.AllGraphemes(out) {
			cluster = g.Symbol
			break
		}
		if cluster == "" {
			out = out[1:]
			continue
		}
		s.put(cluster)
		out = out[len(cluster):]
	}
}

// escape consumes one escape sequence and returns its length in bytes.
func (s *vtScreen) escape(out string) int {
	if len(out) < 2 || out[1] != '[' {
		return 1
	}
	i := 2
	for i < len(out) && (out[i] == ';' || (out[i] >= '0' && out[i] <= '9') || out[i] == '?') {
		i++
	}
	if i >= len(out) {
		return i
	}
	final := out[i]
	body := out[2:i]
	if final == 'H' {
		row, col := 1, 1
		if a, b, ok := strings.Cut(body, ";"); ok {
			row, _ = strconv.Atoi(a)
			col, _ = strconv.Atoi(b)
		}
		s.x, s.y = uint16(max(col-1, 0)), uint16(max(row-1, 0))
	}
	return i + 1
}

func (s *vtScreen) put(cluster string) {
	if s.y >= s.h || s.x >= s.w {
		// A real terminal would clamp or wrap; for the test, writing outside
		// the screen means the cursor model had already gone wrong.
		s.wroteOffScreen = true
		return
	}
	s.cells[s.y][s.x] = cluster
	s.x += max(s.advance(cluster), 1)
}

// render drives a Backend over the given cells and returns the screen a
// terminal with the given advance behaviour would end up showing.
func renderOnVT(t *testing.T, w, h uint16, advance func(string) uint16,
	frames ...[]catatui.PositionedCell) *vtScreen {
	t.Helper()
	var out strings.Builder
	b := NewBackend(&out)
	for _, cells := range frames {
		if err := b.Draw(cells); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	screen := newVTScreen(w, h, advance)
	screen.feed(out.String())
	return screen
}

// cellsOf lays a string out as one row of cells starting at x, one cell per
// grapheme cluster, the way a Buffer would.
func cellsOf(x, y uint16, s string) []catatui.PositionedCell {
	var out []catatui.PositionedCell
	for g := range catatui.AllGraphemes(s) {
		out = append(out, catatui.PositionedCell{X: x, Y: y, Cell: catatui.NewCell(g.Symbol)})
		x += g.Width
	}
	return out
}

// TestBackendSurvivesATerminalThatDisagreesAboutWidth is the regression test for
// the corruption seen with Devanagari, Bengali, Tamil, Telugu, Thai and Arabic
// in Windows Terminal.
//
// The simulated terminal advances one column for every cluster no matter what
// Unicode says. A backend that carries its predicted position forward writes
// every later cell in the row to the wrong column; one that re-anchors after
// non-ASCII lands them all correctly.
func TestBackendSurvivesATerminalThatDisagreesAboutWidth(t *testing.T) {
	lines := []string{
		"हिन्दी परीक्षण",
		"ภาษาไทยทดสอบ",
		"தமிழ் சோதனை",
		"日本語 mixed ascii",
		"العربية اختبار",
	}
	// Every cluster takes exactly one column, whatever its Unicode width.
	stubborn := func(string) uint16 { return 1 }

	for _, line := range lines {
		cells := cellsOf(0, 0, line)
		screen := renderOnVT(t, 60, 2, stubborn, cells)
		if screen.wroteOffScreen {
			t.Errorf("%q: the backend wrote outside the screen", line)
		}
		for _, pc := range cells {
			got := screen.cells[pc.Y][pc.X]
			if want := pc.Cell.GetSymbol(); got != want {
				t.Errorf("%q: cell (%d,%d) shows %q, want %q",
					line, pc.X, pc.Y, got, want)
			}
		}
	}
}

// TestBackendOverwritesAPreviousFrameOnADisagreeingTerminal is the artifact as
// the user saw it: scrolling replaced a long line of complex script with a
// shorter one, and fragments of the old line stayed on screen.
func TestBackendOverwritesAPreviousFrameOnADisagreeingTerminal(t *testing.T) {
	stubborn := func(string) uint16 { return 1 }

	first := cellsOf(0, 0, "हिन्दी परीक्षण मूल पाठ")
	// The second frame blanks the whole row, as a redraw of a shorter line does.
	var second []catatui.PositionedCell
	for x := uint16(0); x < 40; x++ {
		second = append(second, catatui.PositionedCell{X: x, Y: 0, Cell: catatui.NewCell(" ")})
	}

	screen := renderOnVT(t, 60, 2, stubborn, first, second)
	for x := uint16(0); x < 40; x++ {
		if got := screen.cells[0][x]; got != " " {
			t.Fatalf("column %d still shows %q after the row was blanked", x, got)
		}
	}
}

// TestBackendIsExactOnAWellBehavedTerminal checks the same rendering against a
// terminal that does agree with Unicode, so the fix cannot have broken the
// ordinary case.
func TestBackendIsExactOnAWellBehavedTerminal(t *testing.T) {
	agreeable := func(s string) uint16 { return uint16(max(catatui.StringWidth(s), 1)) }

	cells := cellsOf(0, 0, "日本語 mixed ascii ﾊﾞ é")
	screen := renderOnVT(t, 60, 2, agreeable, cells)
	for _, pc := range cells {
		if got, want := screen.cells[pc.Y][pc.X], pc.Cell.GetSymbol(); got != want {
			t.Errorf("cell (%d,%d) shows %q, want %q", pc.X, pc.Y, got, want)
		}
	}
}
