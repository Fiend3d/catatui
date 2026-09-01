// Port of ratatui-core/src/backend.rs @ ratatui-v0.30.2

package catatui

import "fmt"

// ClearType selects how much of the screen a ClearRegion call erases.
type ClearType uint8

const (
	// ClearAll erases the whole screen.
	ClearAll ClearType = iota
	// ClearAfterCursor erases from the cursor to the end of the screen.
	ClearAfterCursor
	// ClearBeforeCursor erases from the start of the screen to the cursor.
	ClearBeforeCursor
	// ClearCurrentLine erases the line the cursor is on.
	ClearCurrentLine
	// ClearUntilNewLine erases from the cursor to the end of its line.
	ClearUntilNewLine
)

// String returns the clear type's name.
func (c ClearType) String() string {
	switch c {
	case ClearAfterCursor:
		return "AfterCursor"
	case ClearBeforeCursor:
		return "BeforeCursor"
	case ClearCurrentLine:
		return "CurrentLine"
	case ClearUntilNewLine:
		return "UntilNewLine"
	default:
		return "All"
	}
}

// WindowSize is the terminal's size in cells and, where the terminal reports it,
// in pixels. A zero pixel size means the terminal did not say.
type WindowSize struct {
	Columns Size
	Pixels  Size
}

// Backend is the boundary between catatui and an actual terminal.
//
// Terminal drives a Backend: it works out which cells changed and hands them to
// Draw, then flushes. Implementations exist for a real terminal (package
// catatui/term) and for tests (TestBackend), and anything else that can paint
// cells can implement this.
//
// Draw receives only the cells that changed, in row-major order, and the
// positions are absolute. An implementation must not assume the cells are
// contiguous.
type Backend interface {
	// Draw writes the given cells to the terminal.
	Draw(cells []PositionedCell) error

	// HideCursor hides the terminal cursor.
	HideCursor() error

	// ShowCursor shows the terminal cursor.
	ShowCursor() error

	// GetCursorPosition reports where the cursor is.
	GetCursorPosition() (Position, error)

	// SetCursorPosition moves the cursor.
	SetCursorPosition(p Position) error

	// Clear erases the whole screen.
	Clear() error

	// ClearRegion erases part of the screen.
	ClearRegion(t ClearType) error

	// Size reports the terminal's size in cells.
	Size() (Size, error)

	// WindowSize reports the terminal's size in cells and pixels.
	WindowSize() (WindowSize, error)

	// Flush writes any buffered output to the terminal.
	Flush() error

	// AppendLines scrolls the terminal up by n lines, for drawing above an
	// inline viewport.
	AppendLines(n uint16) error
}

// --- TestBackend ----------------------------------------------------------

// TestBackend is a Backend that draws into an in-memory Buffer instead of a
// terminal.
//
// It is how widget and terminal behaviour is tested without a tty: render a
// frame, then compare the backend's buffer against an expected one.
//
//	backend := catatui.NewTestBackend(10, 3)
//	term, _ := catatui.NewTerminal(backend)
//	term.Draw(func(f *catatui.Frame) {
//		f.RenderWidget(myWidget, f.Area())
//	})
//	catatui.AssertBuffer(t, backend.Buffer(), expected)
type TestBackend struct {
	buffer       *Buffer
	cursorHidden bool
	cursor       Position
	// scrollback holds lines pushed off the top by AppendLines, so that inline
	// viewport behaviour can be asserted.
	scrollback []string
}

// NewTestBackend returns a TestBackend with a blank buffer of the given size.
func NewTestBackend(width, height uint16) *TestBackend {
	return &TestBackend{
		buffer:       NewBuffer(NewRect(0, 0, width, height)),
		cursorHidden: true,
	}
}

// Buffer returns the backend's buffer, holding everything drawn so far.
func (b *TestBackend) Buffer() *Buffer { return b.buffer }

// Scrollback returns the lines that AppendLines has pushed off the top.
func (b *TestBackend) Scrollback() []string { return b.scrollback }

// CursorHidden reports whether the cursor is currently hidden.
func (b *TestBackend) CursorHidden() bool { return b.cursorHidden }

// Resize changes the backend's size, discarding its contents.
func (b *TestBackend) Resize(width, height uint16) {
	b.buffer = NewBuffer(NewRect(0, 0, width, height))
}

// Draw writes cells into the backend's buffer.
func (b *TestBackend) Draw(cells []PositionedCell) error {
	for _, pc := range cells {
		c := b.buffer.Cell(Position{X: pc.X, Y: pc.Y})
		if c == nil {
			return fmt.Errorf("catatui: TestBackend draw outside the buffer at (%d, %d), area is %+v",
				pc.X, pc.Y, b.buffer.Area)
		}
		*c = pc.Cell
	}
	return nil
}

// HideCursor hides the cursor.
func (b *TestBackend) HideCursor() error { b.cursorHidden = true; return nil }

// ShowCursor shows the cursor.
func (b *TestBackend) ShowCursor() error { b.cursorHidden = false; return nil }

// GetCursorPosition reports the cursor position.
func (b *TestBackend) GetCursorPosition() (Position, error) { return b.cursor, nil }

// SetCursorPosition moves the cursor.
func (b *TestBackend) SetCursorPosition(p Position) error { b.cursor = p; return nil }

// Clear blanks the whole buffer.
func (b *TestBackend) Clear() error { return b.ClearRegion(ClearAll) }

// ClearRegion blanks part of the buffer.
func (b *TestBackend) ClearRegion(t ClearType) error {
	area := b.buffer.Area
	var region Rect
	switch t {
	case ClearAll:
		region = area
	case ClearAfterCursor:
		i := b.buffer.IndexOf(b.cursor.X, b.cursor.Y) + 1
		b.clearFrom(i, len(b.buffer.Content))
		return nil
	case ClearBeforeCursor:
		i := b.buffer.IndexOf(b.cursor.X, b.cursor.Y)
		b.clearFrom(0, i)
		return nil
	case ClearCurrentLine:
		region = NewRect(area.X, b.cursor.Y, area.Width, 1)
	case ClearUntilNewLine:
		region = NewRect(b.cursor.X, b.cursor.Y, SatSub(area.Right(), b.cursor.X), 1)
	}
	region = area.Intersection(region)
	for y := region.Top(); y < region.Bottom(); y++ {
		for x := region.Left(); x < region.Right(); x++ {
			b.buffer.Get(x, y).Reset()
		}
	}
	return nil
}

func (b *TestBackend) clearFrom(start, end int) {
	for i := start; i < end && i < len(b.buffer.Content); i++ {
		b.buffer.Content[i].Reset()
	}
}

// Size reports the buffer's size.
func (b *TestBackend) Size() (Size, error) { return b.buffer.Area.AsSize(), nil }

// WindowSize reports the buffer's size, with no pixel dimensions.
func (b *TestBackend) WindowSize() (WindowSize, error) {
	return WindowSize{Columns: b.buffer.Area.AsSize()}, nil
}

// Flush does nothing; a TestBackend has no output to flush.
func (b *TestBackend) Flush() error { return nil }

// AppendLines scrolls the buffer up by n lines, recording what fell off the top
// in Scrollback.
func (b *TestBackend) AppendLines(n uint16) error {
	area := b.buffer.Area
	if area.Height == 0 {
		return nil
	}
	if n >= area.Height {
		n = area.Height
	}
	for y := area.Top(); y < area.Top()+n; y++ {
		var line []rune
		for x := area.Left(); x < area.Right(); x++ {
			line = append(line, []rune(b.buffer.Get(x, y).GetSymbol())...)
		}
		b.scrollback = append(b.scrollback, string(line))
	}
	// Shift the remaining rows up and blank the vacated ones at the bottom.
	for y := area.Top(); y < area.Bottom(); y++ {
		for x := area.Left(); x < area.Right(); x++ {
			if y+n < area.Bottom() {
				*b.buffer.Get(x, y) = *b.buffer.Get(x, y+n)
			} else {
				b.buffer.Get(x, y).Reset()
			}
		}
	}
	return nil
}
