// Port of ratatui-core/src/buffer/buffer.rs and buffer/diff.rs @ ratatui-v0.30.2

package catatui

import (
	"fmt"
	"strings"
)

// Buffer is a rectangular grid of Cells: the surface widgets draw into.
//
// Cells are stored row-major in Content, so the cell at (x, y) lives at index
// (y-Area.Y)*Area.Width + (x-Area.X). Coordinates are absolute, not relative to
// the buffer's origin: a Buffer whose Area starts at (10, 5) is indexed with
// x >= 10, y >= 5.
//
// Invariant: len(Content) == Area.Area(). Resize and Merge maintain it; if you
// assign to the fields directly, you must too.
type Buffer struct {
	// Area is the region of the terminal this buffer maps to.
	Area Rect
	// Content holds Area.Area() cells in row-major order.
	Content []Cell
}

// NewBuffer returns a buffer covering area with every cell empty.
func NewBuffer(area Rect) *Buffer {
	return NewBufferFilled(area, EmptyCell())
}

// NewBufferFilled returns a buffer covering area with every cell a copy of cell.
func NewBufferFilled(area Rect, cell Cell) *Buffer {
	content := make([]Cell, area.Area())
	for i := range content {
		content[i] = cell
	}
	return &Buffer{Area: area, Content: content}
}

// NewBufferWithLines returns a buffer just large enough to hold the given
// lines, with them drawn into it starting at the origin. Its width is that of
// the widest line.
//
// This is the workhorse of widget tests: build the buffer a widget should have
// produced, then compare. See AssertBuffer.
func NewBufferWithLines(lines ...Line) *Buffer {
	width := 0
	for _, l := range lines {
		width = max(width, l.Width())
	}
	buf := NewBuffer(NewRect(0, 0, uint16(min(width, maxU16)), uint16(min(len(lines), maxU16))))
	for y, l := range lines {
		buf.SetLine(0, uint16(y), l, uint16(min(width, maxU16)))
	}
	return buf
}

// NewBufferWithStrings is NewBufferWithLines for unstyled text, which is how
// most buffer assertions are written.
func NewBufferWithStrings(lines ...string) *Buffer {
	ls := make([]Line, len(lines))
	for i, s := range lines {
		ls[i] = LineFromString(s)
	}
	return NewBufferWithLines(ls...)
}

// IndexOf returns the index in Content of the cell at the absolute position
// (x, y). It panics if the position is outside the buffer, since that always
// means a widget drew outside the area it was given.
func (b *Buffer) IndexOf(x, y uint16) int {
	i, ok := b.indexOfOpt(Position{X: x, Y: y})
	if !ok {
		panic(fmt.Sprintf("catatui: index outside of buffer: the area is %+v but index is (%d, %d)", b.Area, x, y))
	}
	return i
}

func (b *Buffer) indexOfOpt(p Position) (int, bool) {
	if !b.Area.Contains(p) {
		return 0, false
	}
	return int(p.Y-b.Area.Y)*int(b.Area.Width) + int(p.X-b.Area.X), true
}

// PosOf returns the absolute position of the cell at the given index in Content.
// It panics if the index is out of range.
func (b *Buffer) PosOf(index int) (x, y uint16) {
	if index < 0 || index >= len(b.Content) {
		panic(fmt.Sprintf("catatui: trying to get the position of a cell outside the buffer: i=%d len=%d", index, len(b.Content)))
	}
	w := int(b.Area.Width)
	return uint16(index%w + int(b.Area.X)), uint16(index/w + int(b.Area.Y))
}

// Cell returns a pointer to the cell at the given position, or nil if the
// position lies outside the buffer. Writing through the pointer edits the
// buffer.
func (b *Buffer) Cell(p Position) *Cell {
	i, ok := b.indexOfOpt(p)
	if !ok {
		return nil
	}
	return &b.Content[i]
}

// CellAt is Cell taking loose coordinates.
func (b *Buffer) CellAt(x, y uint16) *Cell { return b.Cell(Position{X: x, Y: y}) }

// Get returns a pointer to the cell at (x, y), panicking if it is out of
// bounds. Use Cell when a position outside the buffer is expected.
func (b *Buffer) Get(x, y uint16) *Cell { return &b.Content[b.IndexOf(x, y)] }

// SetString draws s at (x, y) in the given style, stopping at the right edge of
// the buffer.
func (b *Buffer) SetString(x, y uint16, s string, style Style) (nextX, nextY uint16) {
	return b.SetStringn(x, y, s, maxU16, style)
}

// SetStringn draws at most maxWidth columns of s at (x, y) in the given style,
// and returns the position just past what it drew.
//
// This is the single place text becomes cells, and its edge behaviour is load
// bearing:
//
//   - Grapheme clusters are the unit, so a base character and its combining
//     marks land in one cell together.
//   - Control characters and zero-width clusters are dropped.
//   - A cluster wider than the space left is NOT clipped. It is dropped, the
//     cell keeps whatever it held, and drawing stops there — the rest of the
//     string is not drawn either. Callers that must fill every column have to
//     pad the remainder themselves.
//   - The continuation columns of a wide cluster are reset to blank cells, not
//     to the style being drawn.
func (b *Buffer) SetStringn(x, y uint16, s string, maxWidth uint16, style Style) (nextX, nextY uint16) {
	remaining := MinU16(SatSub(b.Area.Right(), x), maxWidth)
	for g := range AllGraphemes(s) {
		// A cluster that does not fit ends the draw entirely, rather than
		// being truncated into a partial glyph.
		if g.Width > remaining {
			break
		}
		remaining -= g.Width

		cell := b.Cell(Position{X: x, Y: y})
		if cell == nil {
			break
		}
		cell.SetSymbol(g.Symbol).SetStyle(style)

		// Blank the columns the cluster covers beyond its first, so nothing
		// shows through from the previous frame.
		end := x + g.Width
		x++
		for x < end {
			if c := b.Cell(Position{X: x, Y: y}); c != nil {
				c.Reset()
			}
			x++
		}
	}
	return x, y
}

// SetSpan draws a span at (x, y), in the span's own style, clipped to maxWidth.
func (b *Buffer) SetSpan(x, y uint16, span Span, maxWidth uint16) (nextX, nextY uint16) {
	return b.SetStringn(x, y, span.content, maxWidth, span.style)
}

// SetLine draws a line at (x, y), clipped to maxWidth. Each span is drawn in
// the line's style patched with the span's own.
func (b *Buffer) SetLine(x, y uint16, line Line, maxWidth uint16) (nextX, nextY uint16) {
	remaining := maxWidth
	for _, span := range line.spans {
		if remaining == 0 {
			break
		}
		nx, _ := b.SetStringn(x, y, span.content, remaining, line.style.Patch(span.style))
		remaining = SatSub(remaining, nx-x)
		x = nx
	}
	return x, y
}

// SetStyle applies a style to every cell in the given area, clipped to the
// buffer. It changes styling only, leaving the symbols alone.
func (b *Buffer) SetStyle(area Rect, style Style) {
	area = b.Area.Intersection(area)
	for y := area.Top(); y < area.Bottom(); y++ {
		for x := area.Left(); x < area.Right(); x++ {
			b.Get(x, y).SetStyle(style)
		}
	}
}

// Resize remaps the buffer onto a new area, growing or shrinking Content to
// match. Cells are not moved, so after a resize the contents are only
// meaningful once redrawn.
func (b *Buffer) Resize(area Rect) {
	length := int(area.Area())
	switch {
	case len(b.Content) > length:
		b.Content = b.Content[:length]
	case len(b.Content) < length:
		for len(b.Content) < length {
			b.Content = append(b.Content, EmptyCell())
		}
	}
	b.Area = area
}

// Reset blanks every cell, keeping the area.
func (b *Buffer) Reset() {
	for i := range b.Content {
		b.Content[i].Reset()
	}
}

// Merge draws another buffer into this one, growing this buffer's area to the
// union of the two. Where they overlap, the other buffer wins.
func (b *Buffer) Merge(other *Buffer) {
	area := b.Area.Union(other.Area)
	length := int(area.Area())
	for len(b.Content) < length {
		b.Content = append(b.Content, EmptyCell())
	}

	// Move this buffer's existing cells to where they belong in the larger
	// area. Walking backwards keeps a cell from being overwritten before it
	// has been moved, since indices only ever grow.
	size := int(b.Area.Area())
	for i := size - 1; i >= 0; i-- {
		x, y := b.PosOf(i)
		k := int(y-area.Y)*int(area.Width) + int(x-area.X)
		if i != k {
			b.Content[k] = b.Content[i]
			b.Content[i].Reset()
		}
	}

	for i := 0; i < int(other.Area.Area()); i++ {
		x, y := other.PosOf(i)
		k := int(y-area.Y)*int(area.Width) + int(x-area.X)
		b.Content[k] = other.Content[i]
	}
	b.Area = area
}

// PositionedCell is one cell of a diff: the cell together with where it goes.
type PositionedCell struct {
	X    uint16
	Y    uint16
	Cell Cell
}

// String renders the buffer as one string per row, joined by newlines, with all
// styling dropped. It is meant for test failure messages.
func (b *Buffer) String() string {
	var sb strings.Builder
	for y := b.Area.Top(); y < b.Area.Bottom(); y++ {
		if y > b.Area.Top() {
			sb.WriteByte('\n')
		}
		for x := b.Area.Left(); x < b.Area.Right(); x++ {
			sb.WriteString(b.Get(x, y).GetSymbol())
		}
	}
	return sb.String()
}
