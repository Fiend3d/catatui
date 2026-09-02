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
//
// It also models background colour erase: erasing fills with the background
// colour in effect at the time, which is how a stale SGR state turns a resize
// into a screenful of the wrong colour. bgs holds each cell's background as the
// SGR parameter that set it, with "" for the default.
type vtScreen struct {
	cells          [][]string
	bgs            [][]string
	curBg          string
	w, h           uint16
	x, y           uint16
	advance        func(string) uint16
	wroteOffScreen bool
}

func newVTScreen(w, h uint16, advance func(string) uint16) *vtScreen {
	cells := make([][]string, h)
	bgs := make([][]string, h)
	for i := range cells {
		cells[i] = make([]string, w)
		bgs[i] = make([]string, w)
		for j := range cells[i] {
			cells[i][j] = " "
		}
	}
	return &vtScreen{cells: cells, bgs: bgs, w: w, h: h, advance: advance}
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
	switch final {
	case 'H':
		row, col := 1, 1
		if a, b, ok := strings.Cut(body, ";"); ok {
			row, _ = strconv.Atoi(a)
			col, _ = strconv.Atoi(b)
		}
		s.x, s.y = uint16(max(col-1, 0)), uint16(max(row-1, 0))
	case 'm':
		s.sgr(body)
	case 'J':
		s.erase(body)
	}
	return i + 1
}

// sgr tracks the background colour, which is all this screen needs of SGR.
func (s *vtScreen) sgr(body string) {
	params := strings.Split(body, ";")
	if body == "" {
		params = []string{"0"}
	}
	for i := 0; i < len(params); i++ {
		p := params[i]
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		switch {
		case n == 0, n == 49:
			s.curBg = ""
		case n >= 40 && n <= 47, n >= 100 && n <= 107:
			s.curBg = p
		case n == 48:
			// An extended colour swallows its own parameters.
			s.curBg = strings.Join(params[i:], ";")
			i = len(params)
		}
	}
}

// erase blanks part of the screen, filling with the current background the way
// a real terminal does.
func (s *vtScreen) erase(body string) {
	if body != "2" {
		return // Only the full clear is modelled; it is all the backend emits.
	}
	for y := range s.cells {
		for x := range s.cells[y] {
			s.cells[y][x] = " "
			s.bgs[y][x] = s.curBg
		}
	}
}

func (s *vtScreen) put(cluster string) {
	if s.y >= s.h || s.x >= s.w {
		// A real terminal would clamp or wrap; for the test, writing outside
		// the screen means the cursor model had already gone wrong.
		s.wroteOffScreen = true
		return
	}
	s.cells[s.y][s.x] = cluster
	s.bgs[s.y][s.x] = s.curBg
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

// TestTerminalResizeDoesNotRepaintTheScreenWithTheLastBackground is the
// artifact as the user saw it: examples/hello draws fine, then resizing the
// window leaves the background wrong.
//
// A resize clears the screen before redrawing, and an erase fills with the
// background colour in effect. The last thing the previous frame wrote was a
// coloured status line, so the clear painted every cell that colour, and the
// diff — which only rewrites cells that differ from a blank buffer — left every
// blank cell of the new frame showing it.
func TestTerminalResizeDoesNotRepaintTheScreenWithTheLastBackground(t *testing.T) {
	var out strings.Builder
	backend := NewBackend(&out)
	size := catatui.Size{Width: 20, Height: 6}
	backend.sizer = func() (catatui.Size, error) { return size, nil }

	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}
	// A title bar, an untouched body, and a status line, as the example has.
	draw := func(f *catatui.Frame) {
		area := f.Area()
		buf := f.Buffer()
		buf.SetString(0, 0, strings.Repeat(" ", int(area.Width)),
			catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorWhite))
		buf.SetString(0, area.Bottom()-1, strings.Repeat(" ", int(area.Width)),
			catatui.NewStyle().Bg(catatui.ColorDarkGray).Fg(catatui.ColorWhite))
	}
	if err := terminal.Draw(draw); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	size = catatui.Size{Width: 24, Height: 8}
	if err := terminal.Draw(draw); err != nil {
		t.Fatalf("Draw after resize: %v", err)
	}

	screen := newVTScreen(size.Width, size.Height, func(string) uint16 { return 1 })
	screen.feed(out.String())

	for y := uint16(1); y < size.Height-1; y++ {
		for x := uint16(0); x < size.Width; x++ {
			if bg := screen.bgs[y][x]; bg != "" {
				t.Fatalf("after the resize, cell (%d,%d) has background %q, want the default",
					x, y, bg)
			}
		}
	}
}

// TestBackendLeavesTheTerminalOnDefaultAttributes covers the same rule from the
// other side: whatever happens between frames — a scroll, an erase, the
// terminal reflowing as the window is dragged — must not inherit the last
// cell's colour.
func TestBackendLeavesTheTerminalOnDefaultAttributes(t *testing.T) {
	var out strings.Builder
	b := NewBackend(&out)
	blue := catatui.NewStyle().Bg(catatui.ColorBlue)
	if err := b.Draw([]catatui.PositionedCell{{X: 0, Y: 0, Cell: cell("x", blue)}}); err != nil {
		t.Fatal(err)
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("a frame should end on a reset, got %q", got)
	}

	// So an erase after a frame needs no reset of its own.
	out.Reset()
	if err := b.ClearRegion(catatui.ClearAll); err != nil {
		t.Fatal(err)
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "\x1b[2J" {
		t.Errorf("erase after a frame = %q, want a bare clear", got)
	}

	// And when a colour is in effect for any other reason, the erase resets
	// first rather than filling the screen with it.
	out.Reset()
	b.w.setStyle(catatui.ColorReset, catatui.ColorBlue, catatui.ColorReset, 0)
	if err := b.ClearRegion(catatui.ClearAll); err != nil {
		t.Fatal(err)
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.HasSuffix(got, "\x1b[0m\x1b[2J") {
		t.Errorf("an erase with a colour set should reset first, got %q", got)
	}
}
