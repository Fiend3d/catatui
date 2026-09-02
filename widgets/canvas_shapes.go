// Port of ratatui-widgets/src/canvas/{line,rectangle,circle,points,map}.rs
// @ ratatui-v0.30.2
//
// The line clipping is a port of the cohen_sutherland module of the
// line-clipping crate (0.3.7), which ratatui's Line shape uses.

package widgets

import (
	"fmt"
	"math"

	"github.com/Fiend3d/catatui"
)

// --- Line ----------------------------------------------------------------

// CanvasLine is a straight line between two canvas points.
//
// It is ratatui's canvas Line, renamed so that it cannot be confused with
// catatui.Line, the line of text.
//
//	ctx.Draw(widgets.NewCanvasLine(0, 0, 1, 0, catatui.ColorRed))
//	ctx.Draw(widgets.NewCanvasLine(1, 0, 0.5, 1, catatui.ColorRed))
//	ctx.Draw(widgets.NewCanvasLine(0.5, 1, 0, 0, catatui.ColorRed))
type CanvasLine struct {
	// X1 and Y1 are the starting point.
	X1, Y1 float64
	// X2 and Y2 are the ending point.
	X2, Y2 float64
	// Color of the line. Unset paints ColorReset.
	Color catatui.Color
}

// NewCanvasLine returns a line from (x1, y1) to (x2, y2).
func NewCanvasLine(x1, y1, x2, y2 float64, color catatui.Color) CanvasLine {
	return CanvasLine{X1: x1, Y1: y1, X2: x2, Y2: y2, Color: color}
}

// Draw clips the line to the canvas bounds and draws it dot by dot.
func (l CanvasLine) Draw(p *Painter) {
	xBounds, yBounds := p.Bounds()
	wx1, wy1, wx2, wy2, ok := clipLine(xBounds, yBounds, l.X1, l.Y1, l.X2, l.Y2)
	if !ok {
		return
	}
	x1, y1, ok := p.GetPoint(wx1, wy1)
	if !ok {
		return
	}
	x2, y2, ok := p.GetPoint(wx2, wy2)
	if !ok {
		return
	}
	drawLine(p, x1, y1, x2, y2, l.Color)
}

// clipRegion is the Cohen-Sutherland outcode: which sides of the window a
// point lies beyond.
type clipRegion uint8

const (
	regionLeft clipRegion = 1 << iota
	regionRight
	regionBottom
	regionTop
)

func (r clipRegion) contains(other clipRegion) bool { return r&other == other }

func (r clipRegion) intersects(other clipRegion) bool { return r&other != 0 }

func (r clipRegion) isOutside() bool { return r != 0 }

// regionOf classifies a point against the window [xmin, xmax] x [ymin, ymax].
func regionOf(x, y, xmin, xmax, ymin, ymax float64) clipRegion {
	var r clipRegion
	if x < xmin {
		r |= regionLeft
	} else if x > xmax {
		r |= regionRight
	}
	if y < ymin {
		r |= regionBottom
	} else if y > ymax {
		r |= regionTop
	}
	return r
}

// clipIntersection moves p1, which lies in region, onto the window edge along
// the line to p2. The edges are tried in the crate's order: left, right,
// bottom, top.
func clipIntersection(p1x, p1y, p2x, p2y float64, region clipRegion, xmin, xmax, ymin, ymax float64) (float64, float64) {
	dx := p2x - p1x
	dy := p2y - p1y
	if region.contains(regionLeft) {
		return xmin, p1y + (xmin-p1x)*dy/dx
	}
	if region.contains(regionRight) {
		return xmax, p1y + (xmax-p1x)*dy/dx
	}
	if region.contains(regionBottom) {
		return p1x + (ymin-p1y)*dx/dy, ymin
	}
	return p1x + (ymax-p1y)*dx/dy, ymax
}

// clipLine is the Cohen-Sutherland algorithm: it returns the part of the
// segment inside the bounds, or false if the segment misses them entirely.
func clipLine(xBounds, yBounds [2]float64, x1, y1, x2, y2 float64) (cx1, cy1, cx2, cy2 float64, ok bool) {
	xmin, xmax := xBounds[0], xBounds[1]
	ymin, ymax := yBounds[0], yBounds[1]
	r1 := regionOf(x1, y1, xmin, xmax, ymin, ymax)
	r2 := regionOf(x2, y2, xmin, xmax, ymin, ymax)
	for {
		switch {
		case r1.intersects(r2):
			return 0, 0, 0, 0, false
		case r1.isOutside():
			x1, y1 = clipIntersection(x1, y1, x2, y2, r1, xmin, xmax, ymin, ymax)
			r1 = regionOf(x1, y1, xmin, xmax, ymin, ymax)
		case r2.isOutside():
			x2, y2 = clipIntersection(x2, y2, x1, y1, r2, xmin, xmax, ymin, ymax)
			r2 = regionOf(x2, y2, xmin, xmax, ymin, ymax)
		default:
			return x1, y1, x2, y2, true
		}
	}
}

// drawLine paints every dot on the Bresenham line between two grid points.
func drawLine(p *Painter, x1, y1, x2, y2 int, color catatui.Color) {
	forEachLinePoint(x1, y1, x2, y2, func(x, y int) { p.Paint(x, y, color) })
}

// forEachLinePoint calls f for each dot on the Bresenham line from (x1, y1)
// to (x2, y2), inclusive.
//
// The rounding here decides exactly which dots light up, and the rendering
// tests depend on it, so it follows ratatui step for step: horizontal and
// vertical lines are walked directly, shallow lines are walked along x from
// the left end, and steep lines along y from the top end.
func forEachLinePoint(x1, y1, x2, y2 int, f func(x, y int)) {
	dx := absDiff(x2, x1)
	dy := absDiff(y2, y1)

	switch {
	case dx == 0:
		for y := min(y1, y2); y <= max(y1, y2); y++ {
			f(x1, y)
		}
	case dy == 0:
		for x := min(x1, x2); x <= max(x1, x2); x++ {
			f(x, y1)
		}
	case dy < dx:
		if x1 > x2 {
			forEachLinePointLow(x2, y2, x1, y1, f)
		} else {
			forEachLinePointLow(x1, y1, x2, y2, f)
		}
	case y1 > y2:
		forEachLinePointHigh(x2, y2, x1, y1, f)
	default:
		forEachLinePointHigh(x1, y1, x2, y2, f)
	}
}

// forEachLinePointLow walks a shallow line (|dy| < dx) with x1 <= x2.
func forEachLinePointLow(x1, y1, x2, y2 int, f func(x, y int)) {
	dx := x2 - x1
	dy := absDiff(y2, y1)
	d := 2*dy - dx
	y := y1
	for x := x1; x <= x2; x++ {
		f(x, y)
		if d > 0 {
			if y1 > y2 {
				y = max(y-1, 0)
			} else {
				y++
			}
			d -= 2 * dx
		}
		d += 2 * dy
	}
}

// forEachLinePointHigh walks a steep line (dy >= dx) with y1 <= y2.
func forEachLinePointHigh(x1, y1, x2, y2 int, f func(x, y int)) {
	dx := absDiff(x2, x1)
	dy := y2 - y1
	d := 2*dx - dy
	x := x1
	for y := y1; y <= y2; y++ {
		f(x, y)
		if d > 0 {
			if x1 > x2 {
				x = max(x-1, 0)
			} else {
				x++
			}
			d -= 2 * dy
		}
		d += 2 * dx
	}
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

// FilledLine is a line with the area between it and a horizontal level filled
// in, which is how an area chart is drawn.
type FilledLine struct {
	// X1 and Y1 are the starting point.
	X1, Y1 float64
	// X2 and Y2 are the ending point.
	X2, Y2 float64
	// FillToY is the level to fill to; the area between the line and this Y
	// is painted.
	FillToY float64
	// Color of the line and the fill.
	Color catatui.Color
}

// NewFilledLine returns a line from (x1, y1) to (x2, y2) filled to fillToY.
func NewFilledLine(x1, y1, x2, y2, fillToY float64, color catatui.Color) FilledLine {
	return FilledLine{X1: x1, Y1: y1, X2: x2, Y2: y2, FillToY: fillToY, Color: color}
}

// Draw clips the line to the canvas, then paints a vertical run from each dot
// of the line to the fill level.
func (l FilledLine) Draw(p *Painter) {
	xBounds, yBounds := p.Bounds()
	wx1, wy1, wx2, wy2, ok := clipLine(xBounds, yBounds, l.X1, l.Y1, l.X2, l.Y2)
	if !ok {
		return
	}
	x1, y1, ok := p.GetPoint(wx1, wy1)
	if !ok {
		return
	}
	x2, y2, ok := p.GetPoint(wx2, wy2)
	if !ok {
		return
	}

	yFill := clampF64(l.FillToY, yBounds[0], yBounds[1])
	_, yFillGrid, ok := p.GetPoint(wx1, yFill)
	if !ok {
		return
	}

	forEachLinePoint(x1, y1, x2, y2, func(x, y int) {
		for fy := min(y, yFillGrid); fy <= max(y, yFillGrid); fy++ {
			p.Paint(x, fy, l.Color)
		}
	})
}

// clampF64 is Rust's f64::clamp.
func clampF64(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- Rectangle -----------------------------------------------------------

// Rectangle is an axis-aligned rectangle outline. Its size is in canvas
// units, not cells, and it is positioned by its bottom left corner.
type Rectangle struct {
	// X and Y are the bottom left corner.
	X, Y float64
	// Width and Height are the size in canvas units.
	Width, Height float64
	// Color of the outline.
	Color catatui.Color
}

// NewRectangle returns a rectangle with its bottom left corner at (x, y).
func NewRectangle(x, y, width, height float64, color catatui.Color) Rectangle {
	return Rectangle{X: x, Y: y, Width: width, Height: height, Color: color}
}

// Draw draws the four edges as lines.
func (r Rectangle) Draw(p *Painter) {
	lines := [4]CanvasLine{
		{X1: r.X, Y1: r.Y, X2: r.X, Y2: r.Y + r.Height, Color: r.Color},
		{X1: r.X, Y1: r.Y + r.Height, X2: r.X + r.Width, Y2: r.Y + r.Height, Color: r.Color},
		{X1: r.X + r.Width, Y1: r.Y, X2: r.X + r.Width, Y2: r.Y + r.Height, Color: r.Color},
		{X1: r.X, Y1: r.Y, X2: r.X + r.Width, Y2: r.Y, Color: r.Color},
	}
	for _, l := range lines {
		l.Draw(p)
	}
}

// --- Circle --------------------------------------------------------------

// Circle is a circle outline with a given center and radius.
type Circle struct {
	// X and Y are the center.
	X, Y float64
	// Radius in canvas units.
	Radius float64
	// Color of the outline.
	Color catatui.Color
}

// NewCircle returns a circle centered at (x, y).
func NewCircle(x, y, radius float64, color catatui.Color) Circle {
	return Circle{X: x, Y: y, Radius: radius, Color: color}
}

// Draw paints one dot per degree around the circumference.
func (c Circle) Draw(p *Painter) {
	for angle := 0; angle < 360; angle++ {
		radians := float64(angle) * (math.Pi / 180)
		circleX := math.FMA(c.Radius, math.Cos(radians), c.X)
		circleY := math.FMA(c.Radius, math.Sin(radians), c.Y)
		if x, y, ok := p.GetPoint(circleX, circleY); ok {
			p.Paint(x, y, c.Color)
		}
	}
}

// --- Points --------------------------------------------------------------

// Points is a scatter of points in one color.
type Points struct {
	// Coords are the (x, y) canvas coordinates to draw.
	Coords [][2]float64
	// Color of the points.
	Color catatui.Color
}

// NewPoints returns a Points shape. The slice is used as given, not copied.
func NewPoints(coords [][2]float64, color catatui.Color) Points {
	return Points{Coords: coords, Color: color}
}

// Draw paints each point that falls within the canvas bounds.
func (pts Points) Draw(p *Painter) {
	for _, c := range pts.Coords {
		if x, y, ok := p.GetPoint(c[0], c[1]); ok {
			p.Paint(x, y, pts.Color)
		}
	}
}

// --- Map -----------------------------------------------------------------

// MapResolution is how many points a Map is drawn with.
type MapResolution uint8

const (
	// MapResolutionLow uses about 1000 points, and is the default.
	MapResolutionLow MapResolution = iota
	// MapResolutionHigh uses about 5000 points; use it with the braille
	// marker.
	MapResolutionHigh
)

// String returns "Low" or "High".
func (r MapResolution) String() string {
	switch r {
	case MapResolutionHigh:
		return "High"
	default:
		return "Low"
	}
}

// ParseMapResolution parses "Low" or "High".
func ParseMapResolution(s string) (MapResolution, error) {
	switch s {
	case "Low":
		return MapResolutionLow, nil
	case "High":
		return MapResolutionHigh, nil
	}
	return MapResolutionLow, fmt.Errorf("catatui: unknown map resolution %q", s)
}

func (r MapResolution) data() [][2]float64 {
	if r == MapResolutionHigh {
		return worldHighResolution
	}
	return worldLowResolution
}

// Map is a world map in the EPSG:4326 coordinate system: longitude on X from
// -180 to 180, latitude on Y from -90 to 90. Set the canvas bounds to those
// ranges to see the whole world.
type Map struct {
	// Resolution is how many points to draw.
	Resolution MapResolution
	// Color of the points. Unset paints ColorReset.
	Color catatui.Color
}

// Draw paints each coastline point within the canvas bounds.
func (m Map) Draw(p *Painter) {
	for _, c := range m.Resolution.data() {
		if x, y, ok := p.GetPoint(c[0], c[1]); ok {
			p.Paint(x, y, m.Color)
		}
	}
}

// Compile-time checks that every shape implements Shape.
var (
	_ Shape = CanvasLine{}
	_ Shape = FilledLine{}
	_ Shape = Rectangle{}
	_ Shape = Circle{}
	_ Shape = Points{}
	_ Shape = Map{}
)
