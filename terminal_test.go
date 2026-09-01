package catatui

import "testing"

// textWidget draws a string at the top left of its area, for tests that need
// something simple to render.
type textWidget string

func (w textWidget) Render(area Rect, buf *Buffer) {
	buf.SetStringn(area.X, area.Y, string(w), area.Width, NewStyle())
}

func TestTerminalDraw(t *testing.T) {
	backend := NewTestBackend(10, 2)
	term, err := NewTerminal(backend)
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}
	if err := term.Draw(func(f *Frame) {
		f.RenderWidget(textWidget("hello"), f.Area())
	}); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	AssertBuffer(t, backend.Buffer(), NewBufferWithStrings(
		"hello     ",
		"          ",
	))
}

// TestTerminalDrawOnlyWritesChangedCells is the reason for the double buffer:
// a frame that changes one cell must send one cell to the backend, not the
// whole screen. This is what keeps redraws cheap over a slow terminal.
func TestTerminalDrawOnlyWritesChangedCells(t *testing.T) {
	backend := &countingBackend{TestBackend: NewTestBackend(20, 3)}
	term, err := NewTerminal(backend)
	if err != nil {
		t.Fatalf("NewTerminal: %v", err)
	}

	draw := func(s string) {
		if err := term.Draw(func(f *Frame) {
			f.RenderWidget(textWidget(s), f.Area())
		}); err != nil {
			t.Fatalf("Draw: %v", err)
		}
	}

	draw("hello")
	backend.drawn = 0

	// Redrawing the identical frame should write nothing at all.
	draw("hello")
	if backend.drawn != 0 {
		t.Errorf("redrawing an identical frame wrote %d cells, want 0", backend.drawn)
	}

	// Changing one character should write one cell.
	backend.drawn = 0
	draw("hallo")
	if backend.drawn != 1 {
		t.Errorf("changing one character wrote %d cells, want 1", backend.drawn)
	}
}

type countingBackend struct {
	*TestBackend
	drawn int
}

func (b *countingBackend) Draw(cells []PositionedCell) error {
	b.drawn += len(cells)
	return b.TestBackend.Draw(cells)
}

func TestTerminalFrameArea(t *testing.T) {
	backend := NewTestBackend(7, 4)
	term, _ := NewTerminal(backend)
	var got Rect
	_ = term.Draw(func(f *Frame) { got = f.Area() })
	if want := NewRect(0, 0, 7, 4); got != want {
		t.Errorf("frame area = %+v, want %+v", got, want)
	}
}

func TestTerminalFrameCount(t *testing.T) {
	backend := NewTestBackend(4, 1)
	term, _ := NewTerminal(backend)
	for i := range uint64(3) {
		var got uint64
		_ = term.Draw(func(f *Frame) { got = f.Count() })
		if got != i {
			t.Errorf("frame %d reported count %d", i, got)
		}
	}
	if got := term.FrameCount(); got != 3 {
		t.Errorf("FrameCount() = %d, want 3", got)
	}
}

// TestTerminalCursorFollowsTheFrame checks the contract that a frame which does
// not ask for a cursor gets it hidden, which is what stops a stray cursor
// blinking in the middle of a rendered UI.
func TestTerminalCursorFollowsTheFrame(t *testing.T) {
	backend := NewTestBackend(10, 2)
	term, _ := NewTerminal(backend)

	_ = term.Draw(func(f *Frame) {})
	if !backend.CursorHidden() {
		t.Error("a frame that sets no cursor should leave it hidden")
	}

	_ = term.Draw(func(f *Frame) { f.SetCursor(3, 1) })
	if backend.CursorHidden() {
		t.Error("a frame that sets a cursor should show it")
	}
	if got, _ := backend.GetCursorPosition(); got != (Position{3, 1}) {
		t.Errorf("cursor at %+v, want (3, 1)", got)
	}
}

func TestTerminalAutoresize(t *testing.T) {
	backend := NewTestBackend(10, 3)
	term, _ := NewTerminal(backend)
	_ = term.Draw(func(f *Frame) { f.RenderWidget(textWidget("abc"), f.Area()) })

	backend.Resize(5, 2)
	var area Rect
	if err := term.Draw(func(f *Frame) {
		area = f.Area()
		f.RenderWidget(textWidget("xy"), f.Area())
	}); err != nil {
		t.Fatalf("Draw after resize: %v", err)
	}
	if want := NewRect(0, 0, 5, 2); area != want {
		t.Errorf("frame area after resize = %+v, want %+v", area, want)
	}
	AssertBuffer(t, backend.Buffer(), NewBufferWithStrings("xy   ", "     "))
}

// TestTerminalDrawIntoBufferDirectly covers the way nezumi uses ratatui: ignore
// the widget layer and write cells straight into the frame's buffer.
func TestTerminalDrawIntoBufferDirectly(t *testing.T) {
	backend := NewTestBackend(5, 2)
	term, _ := NewTerminal(backend)
	_ = term.Draw(func(f *Frame) {
		buf := f.Buffer()
		buf.SetSpan(0, 0, NewStyledSpan("ab", NewStyle().Fg(ColorRed)), 5)
		buf.SetSpan(0, 1, NewSpan("cd"), 5)
	})
	if got := backend.Buffer().Get(0, 0).Fg; got != ColorRed {
		t.Errorf("styled span lost its color: %v", got)
	}
	if got, want := backend.Buffer().String(), "ab   \ncd   "; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestTerminalTryDrawPropagatesError(t *testing.T) {
	backend := NewTestBackend(4, 1)
	term, _ := NewTerminal(backend)
	sentinel := errTest("boom")
	if err := term.TryDraw(func(f *Frame) error { return sentinel }); err != sentinel {
		t.Errorf("TryDraw returned %v, want the callback's error", err)
	}
	// A failed frame must not reach the backend.
	if got := backend.Buffer().String(); got != "    " {
		t.Errorf("a failed draw wrote %q to the backend", got)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestRenderStatefulWidget(t *testing.T) {
	buf := NewBuffer(NewRect(0, 0, 6, 1))
	state := 2
	RenderStatefulWidget[int](selectorWidget{}, buf.Area, buf, &state)
	if got, want := buf.String(), "  ^   "; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
	if state != 3 {
		t.Errorf("the widget should have advanced the state to 3, got %d", state)
	}
}

// selectorWidget marks the selected column and advances the selection, so that
// both directions of the state contract are covered.
type selectorWidget struct{}

func (selectorWidget) RenderStateful(area Rect, buf *Buffer, state *int) {
	if uint16(*state) < area.Width {
		buf.SetString(area.X+uint16(*state), area.Y, "^", NewStyle())
	}
	*state++
}

func TestTestBackendAppendLines(t *testing.T) {
	backend := NewTestBackend(3, 3)
	_ = backend.Draw([]PositionedCell{
		{0, 0, NewCell("a")}, {0, 1, NewCell("b")}, {0, 2, NewCell("c")},
	})
	if err := backend.AppendLines(1); err != nil {
		t.Fatalf("AppendLines: %v", err)
	}
	if got, want := backend.Buffer().String(), "b  \nc  \n   "; got != want {
		t.Errorf("after scrolling: %q, want %q", got, want)
	}
	if got := backend.Scrollback(); len(got) != 1 || got[0] != "a  " {
		t.Errorf("scrollback = %q, want [\"a  \"]", got)
	}
}

func TestTestBackendClearRegion(t *testing.T) {
	newFilled := func() *TestBackend {
		b := NewTestBackend(3, 2)
		for y := range uint16(2) {
			for x := range uint16(3) {
				*b.Buffer().Get(x, y) = NewCell("x")
			}
		}
		return b
	}

	b := newFilled()
	_ = b.ClearRegion(ClearAll)
	if got, want := b.Buffer().String(), "   \n   "; got != want {
		t.Errorf("ClearAll gave %q, want %q", got, want)
	}

	b = newFilled()
	_ = b.SetCursorPosition(Position{0, 1})
	_ = b.ClearRegion(ClearCurrentLine)
	if got, want := b.Buffer().String(), "xxx\n   "; got != want {
		t.Errorf("ClearCurrentLine gave %q, want %q", got, want)
	}

	b = newFilled()
	_ = b.SetCursorPosition(Position{1, 0})
	_ = b.ClearRegion(ClearUntilNewLine)
	if got, want := b.Buffer().String(), "x  \nxxx"; got != want {
		t.Errorf("ClearUntilNewLine gave %q, want %q", got, want)
	}
}

func TestTerminalWithFixedViewport(t *testing.T) {
	backend := NewTestBackend(10, 5)
	term, err := NewTerminalWithViewport(backend, FixedViewport(NewRect(2, 1, 5, 2)))
	if err != nil {
		t.Fatalf("NewTerminalWithViewport: %v", err)
	}
	_ = term.Draw(func(f *Frame) { f.RenderWidget(textWidget("hi"), f.Area()) })
	AssertBuffer(t, backend.Buffer(), NewBufferWithStrings(
		"          ",
		"  hi      ",
		"          ",
		"          ",
		"          ",
	))

	// A fixed viewport must ignore terminal resizes.
	backend.Resize(20, 20)
	var area Rect
	_ = term.Draw(func(f *Frame) { area = f.Area() })
	if want := NewRect(2, 1, 5, 2); area != want {
		t.Errorf("a fixed viewport resized to %+v, want %+v", area, want)
	}
}
