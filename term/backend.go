package term

import (
	"fmt"
	"io"
	"os"

	"github.com/Fiend3d/catatui"
)

// Backend drives a real terminal over VT escape sequences. It implements
// catatui.Backend.
type Backend struct {
	out    io.Writer
	w      *writer
	sizer  func() (catatui.Size, error)
	winSzr func() (catatui.WindowSize, error)
}

// NewBackend returns a Backend writing to out.
//
// Size is read from out if it is a terminal file; otherwise it falls back to a
// fixed 80x24, which keeps the backend usable when output is redirected.
func NewBackend(out io.Writer) *Backend {
	b := &Backend{out: out, w: newWriter(out)}
	if f, ok := out.(*os.File); ok {
		b.sizer = func() (catatui.Size, error) { return terminalSize(f) }
		b.winSzr = func() (catatui.WindowSize, error) { return terminalWindowSize(f) }
	} else {
		b.sizer = func() (catatui.Size, error) { return catatui.Size{Width: 80, Height: 24}, nil }
		b.winSzr = func() (catatui.WindowSize, error) {
			return catatui.WindowSize{Columns: catatui.Size{Width: 80, Height: 24}}, nil
		}
	}
	return b
}

// Draw writes the given cells to the terminal.
//
// The cells arrive in row-major order but are not necessarily adjacent, so the
// writer emits a cursor move only when there is a gap, and a colour change only
// when the style actually differs from the previous cell.
func (b *Backend) Draw(cells []catatui.PositionedCell) error {
	for _, pc := range cells {
		b.w.moveTo(pc.X, pc.Y)
		c := pc.Cell
		b.w.setStyle(orReset(c.Fg), orReset(c.Bg), orReset(c.UnderlineColor), c.Modifier)
		symbol := c.GetSymbol()
		b.w.w.WriteString(symbol)
		b.w.advance(max(cellColumns(symbol), 1))
	}
	return nil
}

// orReset normalizes an unset color, which a Cell should not carry, to a reset.
func orReset(c catatui.Color) catatui.Color {
	if c.IsSet() {
		return c
	}
	return catatui.ColorReset
}

// HideCursor hides the terminal cursor.
func (b *Backend) HideCursor() error {
	b.w.esc(seqCursorHide)
	return nil
}

// ShowCursor shows the terminal cursor.
func (b *Backend) ShowCursor() error {
	b.w.esc(seqCursorShow)
	return nil
}

// GetCursorPosition reports where the cursor is.
//
// Querying the terminal requires reading its reply from stdin, which would race
// with the application's own input loop, so the position tracked by the writer
// is returned instead. That is accurate because every cursor move goes through
// this backend.
func (b *Backend) GetCursorPosition() (catatui.Position, error) {
	return catatui.Position{X: b.w.curX, Y: b.w.curY}, nil
}

// SetCursorPosition moves the cursor.
func (b *Backend) SetCursorPosition(p catatui.Position) error {
	b.w.moveTo(p.X, p.Y)
	return nil
}

// SetCursorShape changes how the terminal draws the cursor.
//
// This is beyond catatui.Backend, which mirrors ratatui's backend trait and has
// no notion of cursor shape, so reach it through the concrete backend:
//
//	if b := term.BackendOf(terminal); b != nil {
//		b.SetCursorShape(term.CursorSteadyBar)
//	}
//
// The change takes effect on the next Flush, which Terminal.Draw does at the
// end of every frame.
func (b *Backend) SetCursorShape(s CursorShape) error {
	b.w.setCursorShape(s)
	return nil
}

// BackendOf returns the term.Backend driving a Terminal, or nil if the terminal
// is driven by some other backend such as catatui.TestBackend.
//
// It exists so that a program can reach the terminal-specific parts — cursor
// shape today — without type-asserting by hand at every call site.
func BackendOf(t *catatui.Terminal) *Backend {
	b, _ := t.Backend().(*Backend)
	return b
}

// Clear erases the whole screen.
func (b *Backend) Clear() error { return b.ClearRegion(catatui.ClearAll) }

// ClearRegion erases part of the screen.
func (b *Backend) ClearRegion(t catatui.ClearType) error {
	switch t {
	case catatui.ClearAll:
		b.w.csi("2J")
	case catatui.ClearAfterCursor:
		b.w.csi("0J")
	case catatui.ClearBeforeCursor:
		b.w.csi("1J")
	case catatui.ClearCurrentLine:
		b.w.csi("2K")
	case catatui.ClearUntilNewLine:
		b.w.csi("0K")
	default:
		return fmt.Errorf("catatui/term: unknown clear type %d", t)
	}
	// Erasing does not move the cursor, but a full clear resets enough terminal
	// state that assuming otherwise is not worth the risk.
	if t == catatui.ClearAll {
		b.w.invalidateCursor()
	}
	return nil
}

// Size reports the terminal's size in cells.
func (b *Backend) Size() (catatui.Size, error) { return b.sizer() }

// WindowSize reports the terminal's size in cells and, where available, pixels.
func (b *Backend) WindowSize() (catatui.WindowSize, error) { return b.winSzr() }

// Flush writes buffered output to the terminal.
func (b *Backend) Flush() error { return b.w.flush() }

// AppendLines scrolls the terminal up by n lines.
func (b *Backend) AppendLines(n uint16) error {
	for range n {
		b.w.w.WriteString("\n")
	}
	b.w.invalidateCursor()
	return nil
}

var _ catatui.Backend = (*Backend)(nil)
