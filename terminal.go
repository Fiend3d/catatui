// Port of ratatui-core/src/terminal.rs, terminal/frame.rs, buffers.rs,
// render.rs, resize.rs, cursor.rs and viewport.rs @ ratatui-v0.30.2

package catatui

import "fmt"

// ViewportKind selects how much of the terminal a Terminal owns.
type ViewportKind uint8

const (
	// ViewportFullscreen uses the whole terminal, which is the usual choice for
	// an application drawing on the alternate screen.
	ViewportFullscreen ViewportKind = iota
	// ViewportInline uses a fixed number of lines at the bottom of the terminal,
	// leaving the scrollback above it intact.
	ViewportInline
	// ViewportFixed uses an explicit region and never resizes with the terminal.
	ViewportFixed
)

// Viewport is the region of the terminal a Terminal draws into.
type Viewport struct {
	kind   ViewportKind
	height uint16
	area   Rect
}

// FullscreenViewport uses the whole terminal.
func FullscreenViewport() Viewport { return Viewport{kind: ViewportFullscreen} }

// InlineViewport uses height lines at the bottom of the terminal.
func InlineViewport(height uint16) Viewport {
	return Viewport{kind: ViewportInline, height: height}
}

// FixedViewport uses an explicit region that never autoresizes.
func FixedViewport(area Rect) Viewport { return Viewport{kind: ViewportFixed, area: area} }

// Kind returns which sort of viewport this is.
func (v Viewport) Kind() ViewportKind { return v.kind }

// Frame is the drawing surface for a single frame.
//
// A Frame is handed to the callback passed to Terminal.Draw and is only valid
// for the duration of that call; do not keep it.
type Frame struct {
	area           Rect
	buffer         *Buffer
	cursorPosition *Position
	count          uint64
}

// Area returns the region the frame covers.
func (f *Frame) Area() Rect { return f.area }

// Size returns the frame's area. It is a synonym for Area, kept because
// ratatui has both.
func (f *Frame) Size() Rect { return f.area }

// Buffer returns the frame's buffer, for widgets that draw cells directly
// rather than through a Widget. This is how a program can bypass the widget
// layer entirely and treat catatui as a cell grid.
func (f *Frame) Buffer() *Buffer { return f.buffer }

// Count returns how many frames have been drawn before this one, which is handy
// for driving animations.
func (f *Frame) Count() uint64 { return f.count }

// RenderWidget draws a widget into the given area of the frame.
func (f *Frame) RenderWidget(w Widget, area Rect) {
	RenderWidget(w, area, f.buffer)
}

// SetCursorPosition puts the terminal cursor at the given position after the
// frame is drawn. A frame that never calls this hides the cursor.
func (f *Frame) SetCursorPosition(p Position) {
	f.cursorPosition = &p
}

// SetCursor is SetCursorPosition taking loose coordinates.
func (f *Frame) SetCursor(x, y uint16) { f.SetCursorPosition(Position{X: x, Y: y}) }

// RenderStatefulWidgetOn draws a stateful widget into a frame.
//
// It is a free function rather than a method on Frame because Go methods cannot
// have their own type parameters. See StatefulWidget.
func RenderStatefulWidgetOn[S any](f *Frame, w StatefulWidget[S], area Rect, state *S) {
	RenderStatefulWidget(w, area, f.buffer, state)
}

// Terminal drives a Backend, keeping two buffers so that each frame only writes
// the cells that actually changed.
//
// Draw is the whole interface in normal use: it resizes if the terminal
// changed, hands a Frame to the callback, diffs the result against the previous
// frame, and writes only the difference.
type Terminal struct {
	backend      Backend
	buffers      [2]*Buffer
	current      int
	hiddenCursor bool
	viewport     Viewport
	viewportArea Rect
	lastKnownAre Rect
	lastCursor   Position
	frameCount   uint64

	// updates is the diff scratch, reused across frames so a redraw does not
	// allocate a screen-sized slice every time.
	updates []PositionedCell
}

// NewTerminal returns a Terminal drawing to the whole of the given backend.
func NewTerminal(backend Backend) (*Terminal, error) {
	return NewTerminalWithViewport(backend, FullscreenViewport())
}

// NewTerminalWithViewport returns a Terminal drawing into the given viewport.
func NewTerminalWithViewport(backend Backend, viewport Viewport) (*Terminal, error) {
	size, err := backend.Size()
	if err != nil {
		return nil, fmt.Errorf("catatui: could not read the terminal size: %w", err)
	}
	area := Rect{Width: size.Width, Height: size.Height}

	viewportArea := area
	switch viewport.kind {
	case ViewportFixed:
		viewportArea = viewport.area
	case ViewportInline:
		pos, err := backend.GetCursorPosition()
		if err != nil {
			return nil, fmt.Errorf("catatui: could not read the cursor position: %w", err)
		}
		viewportArea = inlineArea(backend, viewport.height, size, pos)
	}

	return &Terminal{
		backend:      backend,
		buffers:      [2]*Buffer{NewBuffer(viewportArea), NewBuffer(viewportArea)},
		viewport:     viewport,
		viewportArea: viewportArea,
		lastKnownAre: area,
	}, nil
}

// inlineArea works out where an inline viewport of the given height sits,
// scrolling the terminal up to make room if the cursor is too near the bottom.
func inlineArea(backend Backend, height uint16, size Size, cursor Position) Rect {
	height = MinU16(height, size.Height)
	top := cursor.Y
	// If there is not enough room below the cursor, push the existing lines up.
	if overflow := SatSub(SatAdd(top, height), size.Height); overflow > 0 {
		_ = backend.AppendLines(overflow)
		top = SatSub(top, overflow)
	}
	return Rect{X: 0, Y: top, Width: size.Width, Height: height}
}

// Backend returns the terminal's backend.
func (t *Terminal) Backend() Backend { return t.backend }

// Size reports the terminal's size in cells.
func (t *Terminal) Size() (Size, error) { return t.backend.Size() }

// ViewportArea returns the region the terminal draws into.
func (t *Terminal) ViewportArea() Rect { return t.viewportArea }

// FrameCount returns how many frames have been drawn.
func (t *Terminal) FrameCount() uint64 { return t.frameCount }

// CurrentBuffer returns the buffer being drawn into. It is exposed mainly for
// tests and for code that draws cells outside the normal Draw cycle.
func (t *Terminal) CurrentBuffer() *Buffer { return t.buffers[t.current] }

// GetFrame returns a Frame for the current buffer without drawing anything.
func (t *Terminal) GetFrame() *Frame {
	return &Frame{
		area:   t.viewportArea,
		buffer: t.buffers[t.current],
		count:  t.frameCount,
	}
}

// Draw renders one frame.
//
// It resizes to match the terminal if needed, calls render into a fresh Frame,
// writes only the cells that changed since the previous frame, and positions
// the cursor where the frame asked (hiding it if the frame did not ask).
func (t *Terminal) Draw(render func(*Frame)) error {
	return t.TryDraw(func(f *Frame) error {
		render(f)
		return nil
	})
}

// TryDraw is Draw for a render callback that can fail. If the callback returns
// an error, nothing is written to the backend.
func (t *Terminal) TryDraw(render func(*Frame) error) error {
	// Resize first: growing after widgets have laid out would let them draw
	// out of bounds, and shrinking afterwards would leave stale cells.
	if err := t.Autoresize(); err != nil {
		return err
	}

	frame := t.GetFrame()
	if err := render(frame); err != nil {
		return err
	}
	return t.applyBuffer(frame.cursorPosition)
}

func (t *Terminal) applyBuffer(cursor *Position) error {
	if err := t.flushDiff(); err != nil {
		return err
	}

	// The cursor can only be placed once the frame has been written, or the
	// writes would move it again.
	if cursor == nil {
		if err := t.HideCursor(); err != nil {
			return err
		}
	} else {
		if err := t.ShowCursor(); err != nil {
			return err
		}
		if err := t.SetCursorPosition(*cursor); err != nil {
			return err
		}
	}

	t.swapBuffers()
	if err := t.backend.Flush(); err != nil {
		return err
	}
	t.frameCount++
	return nil
}

// flushDiff writes the cells that differ between the previous frame and this one.
func (t *Terminal) flushDiff() error {
	previous := t.buffers[1-t.current]
	current := t.buffers[t.current]
	// Reuse the update slice across frames; the backend does not retain it.
	t.updates = previous.DiffInto(t.updates[:0], current)
	updates := t.updates
	if len(updates) > 0 {
		last := updates[len(updates)-1]
		t.lastCursor = Position{X: last.X, Y: last.Y}
	}
	return t.backend.Draw(updates)
}

// swapBuffers makes the frame just drawn the "previous" one and blanks the
// buffer that will be drawn into next.
func (t *Terminal) swapBuffers() {
	t.buffers[1-t.current].Reset()
	t.current = 1 - t.current
}

// Resize remaps the terminal onto a new area and clears it.
func (t *Terminal) Resize(area Rect) error {
	next := area
	var restoreCursor *Position

	if t.viewport.kind == ViewportInline {
		size, err := t.backend.Size()
		if err != nil {
			return err
		}
		offset := SatSub(t.lastCursor.Y, t.viewportArea.Top())
		next = inlineArea(t.backend, t.viewport.height, size, Position{Y: t.lastCursor.Y})
		pos := Position{X: 0, Y: SatAdd(next.Y, offset)}
		restoreCursor = &pos
	}

	// A narrower terminal rewraps whatever is on screen, so start from a clean
	// slate rather than trying to reconcile it.
	if next.Width < t.viewportArea.Width {
		next.Y = 0
		if err := t.backend.ClearRegion(ClearAll); err != nil {
			return err
		}
	}

	t.setViewportArea(next)
	if err := t.clearViewport(); err != nil {
		return err
	}
	if restoreCursor != nil {
		if err := t.backend.SetCursorPosition(*restoreCursor); err != nil {
			return err
		}
	}
	t.lastKnownAre = area
	return nil
}

// Autoresize resizes the terminal if it has changed size since the last frame.
// Fixed viewports are left alone.
func (t *Terminal) Autoresize() error {
	if t.viewport.kind == ViewportFixed {
		return nil
	}
	size, err := t.backend.Size()
	if err != nil {
		return err
	}
	area := Rect{Width: size.Width, Height: size.Height}
	if area != t.lastKnownAre {
		return t.Resize(area)
	}
	return nil
}

func (t *Terminal) setViewportArea(area Rect) {
	t.buffers[0].Resize(area)
	t.buffers[1].Resize(area)
	t.viewportArea = area
}

// clearViewport blanks the previous buffer so that the next diff rewrites every
// cell, and erases the corresponding region of the screen. It leaves the cursor
// wherever the erasing put it; callers that care put it back.
func (t *Terminal) clearViewport() error {
	t.buffers[1-t.current].Reset()
	switch t.viewport.kind {
	case ViewportFullscreen:
		return t.backend.ClearRegion(ClearAll)
	case ViewportInline:
		// Everything from the viewport's first row down belongs to the
		// program, so it goes in one call. What is above is scrollback, and is
		// left alone.
		if err := t.backend.SetCursorPosition(t.viewportArea.AsPosition()); err != nil {
			return err
		}
		return t.backend.ClearRegion(ClearAfterCursor)
	}
	// A fixed viewport shares the screen with whatever else is drawn on it, so
	// only its own rows are blanked.
	blank := make([]PositionedCell, 0, t.viewportArea.Area())
	for y := t.viewportArea.Top(); y < t.viewportArea.Bottom(); y++ {
		for x := t.viewportArea.Left(); x < t.viewportArea.Right(); x++ {
			blank = append(blank, PositionedCell{X: x, Y: y, Cell: EmptyCell()})
		}
	}
	return t.backend.Draw(blank)
}

// Clear erases the viewport and forces the next frame to redraw every cell.
func (t *Terminal) Clear() error {
	original, err := t.backend.GetCursorPosition()
	if err != nil {
		return err
	}
	if err := t.clearViewport(); err != nil {
		return err
	}
	if err := t.backend.SetCursorPosition(original); err != nil {
		return err
	}
	return t.backend.Flush()
}

// HideCursor hides the terminal cursor.
func (t *Terminal) HideCursor() error {
	if t.hiddenCursor {
		return nil
	}
	if err := t.backend.HideCursor(); err != nil {
		return err
	}
	t.hiddenCursor = true
	return nil
}

// ShowCursor shows the terminal cursor.
func (t *Terminal) ShowCursor() error {
	if !t.hiddenCursor {
		return nil
	}
	if err := t.backend.ShowCursor(); err != nil {
		return err
	}
	t.hiddenCursor = false
	return nil
}

// GetCursorPosition reports where the cursor is.
func (t *Terminal) GetCursorPosition() (Position, error) { return t.backend.GetCursorPosition() }

// SetCursorPosition moves the cursor and records the position, so that an
// inline viewport can restore it across a resize.
func (t *Terminal) SetCursorPosition(p Position) error {
	if err := t.backend.SetCursorPosition(p); err != nil {
		return err
	}
	t.lastCursor = p
	return nil
}

// InsertBefore draws height lines above the viewport, pushing it down the
// screen or scrolling the screen up to make room, so that the lines become part
// of the scrollback.
//
// This is how a program with an inline viewport emits log lines that stay on
// screen while the viewport keeps redrawing beneath them. It does nothing at
// all for a fullscreen or fixed viewport, which have no scrollback to write to.
//
// The buffer handed to draw starts at the origin and is height rows of the
// viewport's width, whatever row the lines end up on.
//
// The viewport is cleared afterwards, and the whole of it is rewritten by the
// next Draw: the screen has moved underneath the previous frame, so a diff
// against it would leave the parts that did not change where they used to be.
func (t *Terminal) InsertBefore(height uint16, draw func(*Buffer)) error {
	if t.viewport.kind != ViewportInline || t.lastKnownAre.Height == 0 {
		return nil
	}

	width := t.viewportArea.Width
	buf := NewBuffer(Rect{Width: width, Height: height})
	draw(buf)
	cells := buf.Content

	// int rather than uint16 throughout, so that neither the additions below
	// overflow nor the subtractions run past zero.
	drawnHeight := int(t.viewportArea.Top())
	bufferHeight := int(height)
	viewportHeight := int(t.viewportArea.Height)
	screenHeight := int(t.lastKnownAre.Height)

	// Draw as much as will fit, a screenful at a time, until what is left of
	// the buffer and the viewport together fit on the screen. Scrolling by the
	// smallest amount that makes room keeps the viewport off the middle of the
	// screen once this is done.
	for bufferHeight+viewportHeight > screenHeight {
		toDraw := min(bufferHeight, screenHeight)
		scrollUp := max(0, drawnHeight+toDraw-screenHeight)
		if err := t.scrollUp(uint16(scrollUp)); err != nil {
			return err
		}
		var err error
		cells, err = t.drawLines(uint16(drawnHeight-scrollUp), uint16(toDraw), width, cells)
		if err != nil {
			return err
		}
		drawnHeight += toDraw - scrollUp
		bufferHeight -= toDraw
	}

	// What is left now fits, though the existing text may still have to scroll
	// up to leave the screen exactly full. The viewport may not have started at
	// the bottom, so the amount can come out negative; scrolling by that would
	// push the lines the wrong way.
	scrollUp := max(0, drawnHeight+bufferHeight+viewportHeight-screenHeight)
	if err := t.scrollUp(uint16(scrollUp)); err != nil {
		return err
	}
	if _, err := t.drawLines(uint16(drawnHeight-scrollUp), uint16(bufferHeight), width, cells); err != nil {
		return err
	}
	drawnHeight += bufferHeight - scrollUp

	area := t.viewportArea
	area.Y = uint16(drawnHeight)
	t.setViewportArea(area)

	// Clearing last rather than first is deliberate: the lines drawn above are
	// a full rectangle and overwrite whatever they land on, and on tmux a
	// full clear followed immediately by a scroll puts rubbish in the
	// scrollback.
	return t.Clear()
}

// scrollUp pushes the screen up by n lines, by putting the cursor on the last
// row and appending there.
func (t *Terminal) scrollUp(n uint16) error {
	if n == 0 {
		return nil
	}
	if err := t.SetCursorPosition(Position{Y: SatSub(t.lastKnownAre.Height, 1)}); err != nil {
		return err
	}
	return t.backend.AppendLines(n)
}

// drawLines writes rows of cells straight to the backend at an absolute screen
// row, bypassing the frame diff, and returns the cells it did not use.
func (t *Terminal) drawLines(y, rows, width uint16, cells []Cell) ([]Cell, error) {
	n := int(width) * int(rows)
	toDraw, remainder := cells[:n], cells[n:]
	if rows == 0 {
		return remainder, nil
	}
	positioned := make([]PositionedCell, 0, n)
	for i, c := range toDraw {
		positioned = append(positioned, PositionedCell{
			X:    uint16(i % int(width)),
			Y:    y + uint16(i/int(width)),
			Cell: c,
		})
	}
	if err := t.backend.Draw(positioned); err != nil {
		return nil, err
	}
	return remainder, t.backend.Flush()
}
