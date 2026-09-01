package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// render draws a buffer through a real term.Backend into a byte buffer, so the
// escape sequences it emits can be inspected without a terminal.
func render(t *testing.T, cells []catatui.PositionedCell) string {
	t.Helper()
	var out bytes.Buffer
	b := NewBackend(&out)
	if err := b.Draw(cells); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	if err := b.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return out.String()
}

func cell(symbol string, style catatui.Style) catatui.Cell {
	c := catatui.NewCell(symbol)
	c.SetStyle(style)
	return c
}

func TestBackendWritesText(t *testing.T) {
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: catatui.NewCell("h")},
		{X: 1, Y: 0, Cell: catatui.NewCell("i")},
	})
	if !strings.Contains(got, "h") || !strings.Contains(got, "i") {
		t.Errorf("output %q does not contain the drawn text", got)
	}
	// Cursor positioning is 1-based in VT, so the origin is row 1 column 1.
	if !strings.HasPrefix(got, "\x1b[1;1H") {
		t.Errorf("output should start by homing the cursor, got %q", got)
	}
}

// TestBackendSkipsRedundantCursorMoves is what keeps a full redraw small: cells
// that follow on from the previous one need no positioning at all.
func TestBackendSkipsRedundantCursorMoves(t *testing.T) {
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: catatui.NewCell("a")},
		{X: 1, Y: 0, Cell: catatui.NewCell("b")},
		{X: 2, Y: 0, Cell: catatui.NewCell("c")},
	})
	if n := strings.Count(got, "H"); n != 1 {
		t.Errorf("three adjacent cells emitted %d cursor moves, want 1: %q", n, got)
	}
}

func TestBackendEmitsCursorMoveOnGap(t *testing.T) {
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: catatui.NewCell("a")},
		{X: 5, Y: 0, Cell: catatui.NewCell("b")},
		{X: 0, Y: 2, Cell: catatui.NewCell("c")},
	})
	if n := strings.Count(got, "H"); n != 3 {
		t.Errorf("non-adjacent cells emitted %d cursor moves, want 3: %q", n, got)
	}
	if !strings.Contains(got, "\x1b[1;6H") {
		t.Errorf("expected a move to row 1 column 6, got %q", got)
	}
	if !strings.Contains(got, "\x1b[3;1H") {
		t.Errorf("expected a move to row 3 column 1, got %q", got)
	}
}

// TestBackendAdvancesPastWideCells checks that printing a two-column glyph
// leaves the writer's idea of the cursor in the right place; if it did not, the
// next cell would emit a needless move, or worse, the wrong one.
func TestBackendAdvancesPastWideCells(t *testing.T) {
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: catatui.NewCell("あ")},
		{X: 2, Y: 0, Cell: catatui.NewCell("b")},
	})
	if n := strings.Count(got, "H"); n != 1 {
		t.Errorf("a wide cell followed by its neighbour emitted %d moves, want 1: %q", n, got)
	}
}

func TestBackendColors(t *testing.T) {
	cases := []struct {
		name  string
		style catatui.Style
		want  string
	}{
		{"named fg", catatui.NewStyle().Fg(catatui.ColorRed), "\x1b[31m"},
		{"named bg", catatui.NewStyle().Bg(catatui.ColorGreen), "\x1b[42m"},
		{"bright fg", catatui.NewStyle().Fg(catatui.ColorLightRed), "\x1b[91m"},
		{"bright bg", catatui.NewStyle().Bg(catatui.ColorWhite), "\x1b[107m"},
		{"indexed fg", catatui.NewStyle().Fg(catatui.Indexed(42)), "\x1b[38;5;42m"},
		{"indexed bg", catatui.NewStyle().Bg(catatui.Indexed(7)), "\x1b[48;5;7m"},
		{"rgb fg", catatui.NewStyle().Fg(catatui.Rgb(1, 2, 3)), "\x1b[38;2;1;2;3m"},
		{"rgb bg", catatui.NewStyle().Bg(catatui.Rgb(9, 8, 7)), "\x1b[48;2;9;8;7m"},
	}
	for _, c := range cases {
		got := render(t, []catatui.PositionedCell{{X: 0, Y: 0, Cell: cell("x", c.style)}})
		if !strings.Contains(got, c.want) {
			t.Errorf("%s: output %q does not contain %q", c.name, got, c.want)
		}
	}
}

func TestBackendModifiers(t *testing.T) {
	cases := []struct {
		mod  catatui.Modifier
		want string
	}{
		{catatui.ModifierBold, "\x1b[1m"},
		{catatui.ModifierDim, "\x1b[2m"},
		{catatui.ModifierItalic, "\x1b[3m"},
		{catatui.ModifierUnderlined, "\x1b[4m"},
		{catatui.ModifierReversed, "\x1b[7m"},
		{catatui.ModifierCrossedOut, "\x1b[9m"},
	}
	for _, c := range cases {
		got := render(t, []catatui.PositionedCell{
			{X: 0, Y: 0, Cell: cell("x", catatui.NewStyle().AddModifier(c.mod))},
		})
		if !strings.Contains(got, c.want) {
			t.Errorf("modifier %v: output %q does not contain %q", c.mod, got, c.want)
		}
	}
}

// TestBackendSkipsRedundantStyleChanges is the other half of keeping redraws
// small: a run of cells in one style should set that style once.
func TestBackendSkipsRedundantStyleChanges(t *testing.T) {
	red := catatui.NewStyle().Fg(catatui.ColorRed)
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: cell("a", red)},
		{X: 1, Y: 0, Cell: cell("b", red)},
		{X: 2, Y: 0, Cell: cell("c", red)},
	})
	if n := strings.Count(got, "\x1b[31m"); n != 1 {
		t.Errorf("three cells in one style emitted the color %d times, want 1: %q", n, got)
	}
}

// TestBackendResetsBeforeRemovingAModifier covers the awkward case: turning a
// modifier off is unreliable on many terminals, so the writer resets and
// rebuilds the style instead.
func TestBackendResetsBeforeRemovingAModifier(t *testing.T) {
	got := render(t, []catatui.PositionedCell{
		{X: 0, Y: 0, Cell: cell("a", catatui.NewStyle().AddModifier(catatui.ModifierBold))},
		{X: 1, Y: 0, Cell: cell("b", catatui.NewStyle())},
	})
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("dropping a modifier should emit a reset, got %q", got)
	}
}

func TestBackendClearRegion(t *testing.T) {
	cases := []struct {
		t    catatui.ClearType
		want string
	}{
		{catatui.ClearAll, "\x1b[2J"},
		{catatui.ClearAfterCursor, "\x1b[0J"},
		{catatui.ClearBeforeCursor, "\x1b[1J"},
		{catatui.ClearCurrentLine, "\x1b[2K"},
		{catatui.ClearUntilNewLine, "\x1b[0K"},
	}
	for _, c := range cases {
		var out bytes.Buffer
		b := NewBackend(&out)
		if err := b.ClearRegion(c.t); err != nil {
			t.Fatalf("ClearRegion(%v): %v", c.t, err)
		}
		_ = b.Flush()
		if got := out.String(); got != c.want {
			t.Errorf("ClearRegion(%v) = %q, want %q", c.t, got, c.want)
		}
	}
}

func TestBackendCursorVisibility(t *testing.T) {
	var out bytes.Buffer
	b := NewBackend(&out)
	_ = b.HideCursor()
	_ = b.ShowCursor()
	_ = b.Flush()
	if got := out.String(); got != "\x1b[?25l\x1b[?25h" {
		t.Errorf("cursor sequences = %q", got)
	}
}

// TestBackendDrivesATerminal is the end-to-end check: a full catatui.Terminal
// rendering through the real backend into a byte buffer.
func TestBackendDrivesATerminal(t *testing.T) {
	var out bytes.Buffer
	backend := NewBackend(&out)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}
	err = terminal.Draw(func(f *catatui.Frame) {
		f.Buffer().SetString(0, 0, "ok", catatui.NewStyle().Fg(catatui.ColorGreen))
	})
	if err != nil {
		t.Fatalf("Draw: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "\x1b[32m") {
		t.Errorf("expected a green foreground in %q", got)
	}
	if !strings.Contains(got, "o") || !strings.Contains(got, "k") {
		t.Errorf("expected the drawn text in %q", got)
	}
}

// TestBackendNonTerminalWriterFallsBackToADefaultSize keeps the backend usable
// when output is redirected to a file or a pipe, which is what happens when a
// test or a CI job runs a program that draws.
func TestBackendNonTerminalWriterFallsBackToADefaultSize(t *testing.T) {
	b := NewBackend(&bytes.Buffer{})
	size, err := b.Size()
	if err != nil {
		t.Fatalf("Size: %v", err)
	}
	if size.Width == 0 || size.Height == 0 {
		t.Errorf("fallback size = %+v, want something usable", size)
	}
}

func TestBackendCursorShape(t *testing.T) {
	cases := []struct {
		shape CursorShape
		want  string
	}{
		{CursorDefault, "\x1b[0 q"},
		{CursorBlinkingBlock, "\x1b[1 q"},
		{CursorSteadyBlock, "\x1b[2 q"},
		{CursorBlinkingUnderline, "\x1b[3 q"},
		{CursorSteadyUnderline, "\x1b[4 q"},
		{CursorBlinkingBar, "\x1b[5 q"},
		{CursorSteadyBar, "\x1b[6 q"},
	}
	for _, c := range cases {
		var out bytes.Buffer
		b := NewBackend(&out)
		if err := b.SetCursorShape(c.shape); err != nil {
			t.Fatalf("SetCursorShape(%v): %v", c.shape, err)
		}
		_ = b.Flush()
		if got := out.String(); got != c.want {
			t.Errorf("SetCursorShape(%v) = %q, want %q", c.shape, got, c.want)
		}
	}
}

// TestBackendCursorShapeDoesNotMoveTheCursor checks that changing the shape does
// not invalidate the writer's idea of where the cursor is; if it did, every
// shape change would cost a redundant cursor move on the next cell.
func TestBackendCursorShapeDoesNotMoveTheCursor(t *testing.T) {
	var out bytes.Buffer
	b := NewBackend(&out)
	_ = b.Draw([]catatui.PositionedCell{{X: 0, Y: 0, Cell: catatui.NewCell("a")}})
	_ = b.SetCursorShape(CursorSteadyBar)
	_ = b.Draw([]catatui.PositionedCell{{X: 1, Y: 0, Cell: catatui.NewCell("b")}})
	_ = b.Flush()

	if n := strings.Count(out.String(), "H"); n != 1 {
		t.Errorf("a shape change between adjacent cells caused %d cursor moves, want 1: %q", n, out.String())
	}
}

func TestCursorShapeString(t *testing.T) {
	cases := []struct {
		shape CursorShape
		want  string
	}{
		{CursorDefault, "Default"},
		{CursorSteadyBlock, "SteadyBlock"},
		{CursorBlinkingBar, "BlinkingBar"},
		{CursorShape(99), "Default"},
	}
	for _, c := range cases {
		if got := c.shape.String(); got != c.want {
			t.Errorf("CursorShape(%d).String() = %q, want %q", c.shape, got, c.want)
		}
	}
}

// TestBackendOf checks the escape hatch that lets a program reach the
// terminal-specific parts of the backend through a catatui.Terminal.
func TestBackendOf(t *testing.T) {
	backend := NewBackend(&bytes.Buffer{})
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}
	if got := BackendOf(terminal); got != backend {
		t.Errorf("BackendOf returned %p, want the backend %p", got, backend)
	}

	// A terminal on some other backend must report nil rather than panicking.
	other, err := catatui.NewTerminal(catatui.NewTestBackend(4, 2))
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}
	if got := BackendOf(other); got != nil {
		t.Errorf("BackendOf on a TestBackend returned %v, want nil", got)
	}
}
