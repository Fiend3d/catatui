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
	cells [][]string
	w, h  uint16
	x, y  uint16
	// advance is how far the cursor moves after a cluster is printed.
	advance func(string) uint16
	// spread is what the terminal leaves in the columns the cluster covers: one
	// entry per column, an empty entry meaning the cell is left as it was. A
	// terminal that keeps a cluster together puts the whole of it in the first
	// cell; the Windows console spreads it one code point per cell, which is
	// what makes a write into the middle of a glyph destroy it. nil means the
	// former.
	spread         func(string) []string
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
	if s.spread == nil {
		// A terminal that keeps a cluster together: the whole of it goes in the
		// first cell and it covers the rest of its columns itself, so those are
		// left exactly as they were.
		s.cells[s.y][s.x] = cluster
	} else {
		for i, p := range s.spread(cluster) {
			if s.x+uint16(i) < s.w {
				s.cells[s.y][s.x+uint16(i)] = p
			}
		}
	}
	s.x += max(s.advance(cluster), 1)
}

// textAt reads back the n columns starting at (x, y), which is the glyph a
// terminal shows there.
func (s *vtScreen) textAt(x, y, n uint16) string {
	var sb strings.Builder
	for i := uint16(0); i < n && x+i < s.w; i++ {
		sb.WriteString(s.cells[y][x+i])
	}
	return sb.String()
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

// renderOnConsole drives a Backend over the given cells and returns the screen
// the Windows console would end up showing.
//
// The console is modelled from measurement: it has no notion of a grapheme
// cluster, so it stores one code point per cell and advances by every one of
// them, never by less than a column. Anything written into the columns a
// cluster covers therefore lands on part of that cluster and destroys it.
func renderOnConsole(t *testing.T, w, h uint16, frames ...[]catatui.PositionedCell) *vtScreen {
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
	screen := newVTScreen(w, h, consoleAdvance)
	screen.spread = consoleSpread
	screen.feed(out.String())
	return screen
}

// consoleSpread lays a cluster out one code point per cell. The second column
// of a wide code point is emptied: the code point to its left covers it, and
// nothing of its own is shown there.
func consoleSpread(cluster string) []string {
	var out []string
	for _, r := range cluster {
		out = append(out, string(r))
		for i := uint16(1); i < consoleRuneWidth(r); i++ {
			out = append(out, "")
		}
	}
	return out
}

func consoleAdvance(cluster string) uint16 {
	var w uint16
	for _, r := range cluster {
		w += consoleRuneWidth(r)
	}
	return w
}

// consoleRuneWidth is a code point's own width, floored at one column.
func consoleRuneWidth(r rune) uint16 {
	return uint16(max(catatui.StringWidth(string(r)), 1))
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

// TestBackendSurvivesATerminalThatSplitsClustersIntoCells is the bug the user
// reported: Devanagari, Bengali, Tamil and Telugu came out with whole clusters
// missing, हिन्दी drawn as न्दी and भारत as रत.
//
// The cause was arithmetic, not shaping. The Windows console spends a cell on
// every code point, so a consonant carrying a spacing vowel sign covers two
// columns; measuring it as the one glyph it is drawn as put the next cell on
// top of its second half and the console dropped the pair. Every cluster here
// is placed at the column the width policy predicts, and every one of them has
// to read back whole.
func TestBackendSurvivesATerminalThatSplitsClustersIntoCells(t *testing.T) {
	lines := []string{
		"हिन्दी परीक्षण पाठ।",
		"भारत एक महान देश है।",
		"বাংলা পরীক্ষা লেখা।",
		"தமிழ் சோதனை உரை.",
		"తెలుగు పరీక్ష వచనం.",
		"ภาษาไทยทดสอบ",
		"العربية اختبار",
		"日本語 mixed ascii",
	}
	for _, line := range lines {
		cells := cellsOf(0, 0, line)
		screen := renderOnConsole(t, 80, 2, cells)
		if screen.wroteOffScreen {
			t.Errorf("%q: the backend wrote outside the screen", line)
		}
		for _, pc := range cells {
			symbol := pc.Cell.GetSymbol()
			w := uint16(catatui.StringWidth(symbol))
			if got := screen.textAt(pc.X, pc.Y, w); got != symbol {
				t.Errorf("%q: the glyph at column %d reads back as %q, want %q",
					line, pc.X, got, symbol)
			}
		}
	}
}
