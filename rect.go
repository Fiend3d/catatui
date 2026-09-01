// Port of ratatui-core/src/layout/rect.rs, position.rs, size.rs, margin.rs,
// offset.rs and rect/ops.rs @ ratatui-v0.30.2
//
// Coordinates are uint16 throughout, as in ratatui. The saturating arithmetic
// below is not incidental: widget code all over the library relies on
// subtraction clamping at zero rather than wrapping, and on rects never
// extending past uint16 max. Switching these to int would silently change
// results at the edges.

package catatui

// Position is a single point in the terminal, in cells, measured from the top
// left corner.
type Position struct {
	X uint16
	Y uint16
}

// PositionOrigin is the top left corner.
var PositionOrigin = Position{0, 0}

// Size is a width and height in cells, with no position.
type Size struct {
	Width  uint16
	Height uint16
}

// Margin is a horizontal and vertical inset, applied to both sides of each axis.
type Margin struct {
	Horizontal uint16
	Vertical   uint16
}

// NewMargin returns a Margin with the given horizontal and vertical insets.
func NewMargin(horizontal, vertical uint16) Margin {
	return Margin{Horizontal: horizontal, Vertical: vertical}
}

// Offset is a signed displacement applied to a Rect. Positive X moves right,
// positive Y moves down.
type Offset struct {
	X int32
	Y int32
}

// Rect is a rectangular area of the terminal, in cells.
//
// The zero value is an empty rect at the origin. A Rect built with NewRect is
// guaranteed never to extend past uint16 max: the constructor clamps Width and
// Height so that Right and Bottom stay in range.
type Rect struct {
	X      uint16
	Y      uint16
	Width  uint16
	Height uint16
}

// ZeroRect is an empty rect at the origin.
var ZeroRect = Rect{}

// NewRect returns a Rect, clamping Width and Height so the right and bottom
// edges stay within uint16.
func NewRect(x, y, width, height uint16) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  SatAdd(x, width) - x,
		Height: SatAdd(y, height) - y,
	}
}

// Area is the number of cells in the rect. It is a uint32 because a full
// uint16-by-uint16 rect overflows uint16.
func (r Rect) Area() uint32 { return uint32(r.Width) * uint32(r.Height) }

// IsEmpty reports whether the rect has no area.
func (r Rect) IsEmpty() bool { return r.Width == 0 || r.Height == 0 }

// Left is the x coordinate of the leftmost column in the rect.
func (r Rect) Left() uint16 { return r.X }

// Right is the first x coordinate to the right of the rect, clamped to uint16 max.
func (r Rect) Right() uint16 { return SatAdd(r.X, r.Width) }

// Top is the y coordinate of the topmost row in the rect.
func (r Rect) Top() uint16 { return r.Y }

// Bottom is the first y coordinate below the rect, clamped to uint16 max.
func (r Rect) Bottom() uint16 { return SatAdd(r.Y, r.Height) }

// Inner returns the rect inset by the margin on each side. If the margin is
// larger than the rect, the result is empty.
func (r Rect) Inner(m Margin) Rect {
	dh := SatMul(m.Horizontal, 2)
	dv := SatMul(m.Vertical, 2)
	if r.Width < dh || r.Height < dv {
		return ZeroRect
	}
	return Rect{
		X:      SatAdd(r.X, m.Horizontal),
		Y:      SatAdd(r.Y, m.Vertical),
		Width:  r.Width - dh,
		Height: r.Height - dv,
	}
}

// Outer returns the rect grown by the margin on each side, truncated to keep
// the bounds within uint16. The result may fall outside the containing area, so
// consider Clamp before using it.
func (r Rect) Outer(m Margin) Rect {
	x := SatSub(r.X, m.Horizontal)
	y := SatSub(r.Y, m.Vertical)
	return Rect{
		X:      x,
		Y:      y,
		Width:  SatSub(SatAdd(r.Right(), m.Horizontal), x),
		Height: SatSub(SatAdd(r.Bottom(), m.Vertical), y),
	}
}

// Offset moves the rect without changing its size. If the offset would push an
// edge outside uint16, the position is clamped to the nearest edge.
func (r Rect) Offset(o Offset) Rect {
	maxX := int32(maxU16 - r.Width)
	maxY := int32(maxU16 - r.Height)
	r.X = uint16(clampI32(int32(r.X)+o.X, 0, maxX))
	r.Y = uint16(clampI32(int32(r.Y)+o.Y, 0, maxY))
	return r
}

// Resize changes the size, keeping the position and clamping so that Right and
// Bottom stay within uint16.
func (r Rect) Resize(s Size) Rect {
	r.Width = SatSub(SatAdd(r.X, s.Width), r.X)
	r.Height = SatSub(SatAdd(r.Y, s.Height), r.Y)
	return r
}

// Union returns the smallest rect containing both r and other.
func (r Rect) Union(other Rect) Rect {
	x1 := MinU16(r.X, other.X)
	y1 := MinU16(r.Y, other.Y)
	x2 := MaxU16(r.Right(), other.Right())
	y2 := MaxU16(r.Bottom(), other.Bottom())
	return Rect{X: x1, Y: y1, Width: SatSub(x2, x1), Height: SatSub(y2, y1)}
}

// Intersection returns the overlap of r and other, which is empty if they do
// not overlap.
func (r Rect) Intersection(other Rect) Rect {
	x1 := MaxU16(r.X, other.X)
	y1 := MaxU16(r.Y, other.Y)
	x2 := MinU16(r.Right(), other.Right())
	y2 := MinU16(r.Bottom(), other.Bottom())
	return Rect{X: x1, Y: y1, Width: SatSub(x2, x1), Height: SatSub(y2, y1)}
}

// Intersects reports whether r and other overlap.
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.Right() && r.Right() > other.X &&
		r.Y < other.Bottom() && r.Bottom() > other.Y
}

// Contains reports whether the position lies inside the rect. Positions on the
// left and top borders are inside; the right and bottom edges are not.
func (r Rect) Contains(p Position) bool {
	return p.X >= r.X && p.X < r.Right() && p.Y >= r.Y && p.Y < r.Bottom()
}

// Clamp moves and shrinks r so it fits inside other.
//
// This differs from Intersection: Clamp slides r to fit, preserving as much of
// its size as it can, while Intersection keeps r's position and truncates.
func (r Rect) Clamp(other Rect) Rect {
	width := MinU16(r.Width, other.Width)
	height := MinU16(r.Height, other.Height)
	x := ClampU16(r.X, other.X, SatSub(other.Right(), width))
	y := ClampU16(r.Y, other.Y, SatSub(other.Bottom(), height))
	return NewRect(x, y, width, height)
}

// AsPosition returns the rect's top left corner.
func (r Rect) AsPosition() Position { return Position{X: r.X, Y: r.Y} }

// AsSize returns the rect's size, discarding its position.
func (r Rect) AsSize() Size { return Size{Width: r.Width, Height: r.Height} }

// Rows returns one height-1 rect per row of r, top to bottom.
func (r Rect) Rows() []Rect {
	rows := make([]Rect, 0, r.Height)
	for y := r.Y; y < r.Bottom(); y++ {
		rows = append(rows, Rect{X: r.X, Y: y, Width: r.Width, Height: 1})
	}
	return rows
}

// Columns returns one width-1 rect per column of r, left to right.
func (r Rect) Columns() []Rect {
	cols := make([]Rect, 0, r.Width)
	for x := r.X; x < r.Right(); x++ {
		cols = append(cols, Rect{X: x, Y: r.Y, Width: 1, Height: r.Height})
	}
	return cols
}

// Positions returns every position in r in row-major order.
func (r Rect) Positions() []Position {
	ps := make([]Position, 0, r.Area())
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			ps = append(ps, Position{X: x, Y: y})
		}
	}
	return ps
}

// --- saturating uint16 helpers -------------------------------------------
//
// Go has no saturating arithmetic, and ratatui's layout and widget code depends
// on it pervasively, so these are the only form of uint16 arithmetic used above.
// They are exported because anyone writing a widget needs exactly this: a
// widget that subtracts a border width from an area must clamp at zero rather
// than wrap to 65535.

const maxU16 = 1<<16 - 1

// SatAdd returns a+b, clamped at uint16 max instead of wrapping.
func SatAdd(a, b uint16) uint16 {
	if s := uint32(a) + uint32(b); s <= maxU16 {
		return uint16(s)
	}
	return maxU16
}

// SatSub returns a-b, clamped at zero instead of wrapping.
func SatSub(a, b uint16) uint16 {
	if a < b {
		return 0
	}
	return a - b
}

// SatMul returns a*b, clamped at uint16 max instead of wrapping.
func SatMul(a, b uint16) uint16 {
	if p := uint32(a) * uint32(b); p <= maxU16 {
		return uint16(p)
	}
	return maxU16
}

// MinU16 returns the smaller of a and b.
func MinU16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

// MaxU16 returns the larger of a and b.
func MaxU16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

// ClampU16 returns v confined to the range [lo, hi]. An empty range pins to lo.
func ClampU16(v, lo, hi uint16) uint16 {
	if hi < lo {
		// The caller passed an empty range, which Rust's clamp treats as a
		// precondition violation. Pin to lo, as the saturating math intends.
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampI32(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
