package term

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

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
//
// Skipping the cursor move depends on knowing how far printing a symbol moved
// the cursor, and that is only knowable for ASCII. Every terminal advances one
// column for a printable ASCII byte; for anything else the terminal's shaping
// engine decides, and it does not have to agree with Unicode's tables. Windows
// Terminal in particular shapes Devanagari, Bengali, Tamil, Telugu, Thai and
// Arabic into cluster widths that unicode-width does not predict. Guessing
// wrong desynchronises the writer from the real cursor, and every later cell in
// the row lands in the wrong column, which leaves stale glyphs behind wherever
// the row was supposed to be overwritten.
//
// So the cursor is only tracked through ASCII. After any other symbol the
// position is forgotten and the next cell re-anchors with an absolute move.
// That costs one cursor move per non-ASCII cell and nothing at all on ASCII
// text, which is the case that has to stay fast.
func (b *Backend) Draw(cells []catatui.PositionedCell) error {
	for _, pc := range cells {
		b.w.moveTo(pc.X, pc.Y)
		c := pc.Cell
		b.w.setStyle(orReset(c.Fg), orReset(c.Bg), orReset(c.UnderlineColor), c.Modifier)
		symbol := c.GetSymbol()
		b.w.w.WriteString(symbol)
		if len(symbol) == 1 {
			b.w.advance(1)
		} else {
			b.w.invalidateCursor()
		}
	}
	// A frame ends on default attributes. Anything that happens between frames
	// and paints cells we did not ask for — the terminal reflowing on a window
	// resize, a scroll, an erase — uses whatever colour is left set, so leaving
	// the last cell's style in effect is what turns a resize into a screenful
	// of that colour.
	b.w.resetStyle()
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

// QueryCursorPosition asks the terminal where the cursor is and reads the
// reply from in, which must be the terminal's input.
//
// GetCursorPosition returns the position this backend has been tracking, which
// is right once drawing has started but not before: at startup the cursor is
// wherever the shell left it, and only the terminal knows where that is. An
// inline viewport has to know, because it is placed at the cursor, so Init
// queries the position once while nothing else is reading input.
//
// This must not be called while an EventReader is running: both read the same
// input, and the reply would go to whichever got there first. Input that
// arrives before the reply — keys typed before the program started — is
// discarded, and a terminal that does not answer within timeout gives
// ErrCursorPositionUnknown rather than blocking for ever.
func (b *Backend) QueryCursorPosition(in *os.File, timeout time.Duration) (catatui.Position, error) {
	// DSR 6: report the cursor position as ESC [ row ; col R.
	b.w.csi("6n")
	if err := b.Flush(); err != nil {
		return catatui.Position{}, err
	}

	// One read at a time, so that once the reply has been parsed no read is
	// left outstanding to swallow the user's first keystroke.
	type result struct {
		data []byte
		err  error
	}
	wanted := make(chan struct{})
	results := make(chan result, 1)
	go func() {
		buf := make([]byte, 64)
		for range wanted {
			n, err := in.Read(buf)
			results <- result{append([]byte(nil), buf[:n]...), err}
			if err != nil {
				return
			}
		}
	}()
	defer close(wanted)

	deadline := time.After(timeout)
	var pending []byte
	for {
		select {
		case wanted <- struct{}{}:
		case <-deadline:
			return catatui.Position{}, ErrCursorPositionUnknown
		}

		select {
		case res := <-results:
			pending = append(pending, res.data...)
			if pos, ok := parseCursorPositionReport(pending); ok {
				return pos, nil
			}
			if res.err != nil {
				return catatui.Position{}, ErrCursorPositionUnknown
			}
		case <-deadline:
			return catatui.Position{}, ErrCursorPositionUnknown
		}
	}
}

// ErrCursorPositionUnknown is returned by QueryCursorPosition when the terminal
// does not report its cursor position in time.
var ErrCursorPositionUnknown = errors.New("catatui/term: terminal did not report the cursor position")

// parseCursorPositionReport looks for a CPR reply (ESC [ row ; col R, or the
// DECXCPR form with a leading ?) anywhere in b, and converts its 1-based
// coordinates to catatui's 0-based ones.
func parseCursorPositionReport(b []byte) (catatui.Position, bool) {
	for i := 0; i+1 < len(b); i++ {
		if b[i] != 0x1b || b[i+1] != '[' {
			continue
		}
		j := i + 2
		if j < len(b) && b[j] == '?' {
			j++
		}
		row, j, ok := parseUint16(b, j)
		if !ok || j >= len(b) || b[j] != ';' {
			continue
		}
		col, j, ok := parseUint16(b, j+1)
		if !ok || j >= len(b) || b[j] != 'R' {
			continue
		}
		return catatui.Position{X: catatui.SatSub(col, 1), Y: catatui.SatSub(row, 1)}, true
	}
	return catatui.Position{}, false
}

// parseUint16 reads decimal digits from b at i, saturating rather than
// overflowing, and returns the position just past them.
func parseUint16(b []byte, i int) (value uint16, next int, ok bool) {
	start := i
	for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
		value = catatui.SatAdd(catatui.SatMul(value, 10), uint16(b[i]-'0'))
	}
	return value, i, i > start
}

// setTrackedCursor records where the cursor is without moving it, which is how
// Init hands the position it queried to the terminal that is about to place an
// inline viewport there.
func (b *Backend) setTrackedCursor(p catatui.Position) {
	b.w.curX, b.w.curY = p.X, p.Y
	// The position is right, but the next move is still written in full: the
	// cursor may since have moved for reasons this writer cannot see.
	b.w.invalidateCursor()
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
	// Erasing fills with the background colour currently in effect, so the last
	// cell drawn would become the colour of everything erased. See
	// writer.resetStyle.
	b.w.resetStyle()
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
	// Scrolling fills the lines it exposes with the current background colour,
	// the same trap as erasing.
	b.w.resetStyle()
	for range n {
		b.w.w.WriteString("\n")
	}
	b.w.invalidateCursor()
	return nil
}

var _ catatui.Backend = (*Backend)(nil)
