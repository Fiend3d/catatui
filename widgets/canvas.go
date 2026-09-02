// Port of ratatui-widgets/src/canvas.rs @ ratatui-v0.30.2

package widgets

import (
	"math"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// Shape is something that can be drawn on a Canvas.
//
// The library ships CanvasLine, FilledLine, Rectangle, Circle, Points and Map.
// Implement Draw to add your own: convert canvas coordinates to grid points
// with Painter.GetPoint and light them with Painter.Paint.
//
//	type Cross struct{ X, Y float64 }
//
//	func (c Cross) Draw(p *widgets.Painter) {
//		if x, y, ok := p.GetPoint(c.X, c.Y); ok {
//			p.Paint(x, y, catatui.ColorRed)
//		}
//	}
type Shape interface {
	Draw(p *Painter)
}

// label is a line of text placed at a canvas coordinate by Context.Print.
type label struct {
	x, y float64
	line catatui.Line
}

// layer is one saved state of the grid. A canvas is drawn as a stack of
// layers, later ones over earlier ones.
type layer struct {
	contents []layerCell
}

// layerCell is what one layer contributes to a terminal cell: possibly a
// symbol, possibly a foreground, possibly a background. Whatever is not set is
// left to the layers beneath. An unset Color (the zero value) means "not
// provided", exactly as ratatui's Option<Color>::None does.
type layerCell struct {
	symbol    rune
	hasSymbol bool
	fg        catatui.Color
	bg        catatui.Color
}

// grid is a region of the screen measured in cells, painted at a resolution
// that may be finer than the cells: a braille grid has 2x4 dots per cell, so
// a 10x10 grid has a resolution of 20x40 dots.
type grid interface {
	// resolution is the size of the grid in dots.
	resolution() (float64, float64)
	// paint lights one dot, counted from the top left of the grid. This is
	// not the canvas coordinate system; Painter.GetPoint converts.
	paint(x, y int, color catatui.Color)
	// save copies the grid into a layer to be rendered.
	save() layer
	// reset clears the grid.
	reset()
}

// paintColor turns an unset Color into ColorReset, so that a shape built with
// the zero Color paints exactly as ratatui's Color::default() (which is
// Color::Reset) does, rather than silently painting nothing.
func paintColor(c catatui.Color) catatui.Color {
	if c.IsSet() {
		return c
	}
	return catatui.ColorReset
}

// patternCell is the state of one cell of a patternGrid: which of its
// pseudo-pixels are lit, held in row-major order in the low bits (for a 2x4
// pattern, bit 0 is the top left pixel and bit 7 the bottom right), and the
// single foreground color the whole pattern is drawn in.
type patternCell struct {
	pattern uint8
	color   catatui.Color
}

// patternGrid is a grid whose cells each hold one character from a table of
// W x H pixel patterns: braille (2x4), quadrants (2x2), sextants (2x3) or
// octants (2x4). Only one foreground color is possible per cell, because a
// character has only one foreground.
//
// It is ratatui's PatternGrid<W, H>, with the pattern size as fields rather
// than type parameters.
type patternGrid struct {
	width     uint16
	height    uint16
	w, h      int
	cells     []patternCell
	charTable []rune
}

// newPatternGrid returns a grid of width x height cells, each a w x h pattern
// looked up in charTable, which must hold 1<<(w*h) runes indexed by pattern.
func newPatternGrid(width, height uint16, w, h int, charTable []rune) *patternGrid {
	if w*h > 8 {
		panic("catatui: pattern grid cells must fit in 8 bits")
	}
	return &patternGrid{
		width:     width,
		height:    height,
		w:         w,
		h:         h,
		cells:     make([]patternCell, int(width)*int(height)),
		charTable: charTable,
	}
}

func (g *patternGrid) resolution() (float64, float64) {
	return float64(g.width) * float64(g.w), float64(g.height) * float64(g.h)
}

func (g *patternGrid) save() layer {
	contents := make([]layerCell, len(g.cells))
	for i, cell := range g.cells {
		// A blank pattern is skipped so that layers underneath show through.
		if cell.pattern != 0 {
			contents[i].symbol = g.charTable[cell.pattern]
			contents[i].hasSymbol = true
		}
		// Patterns only affect the foreground.
		contents[i].fg = cell.color
	}
	return layer{contents: contents}
}

func (g *patternGrid) reset() {
	for i := range g.cells {
		g.cells[i] = patternCell{}
	}
}

func (g *patternGrid) paint(x, y int, color catatui.Color) {
	// Out-of-range dots are ignored rather than panicking, as in ratatui,
	// whose saturating index arithmetic lands them outside the cell vector.
	if x < 0 || y < 0 {
		return
	}
	row, col := y/g.h, x/g.w
	if row >= int(g.height) {
		return
	}
	index := row*int(g.width) + col
	if index >= len(g.cells) {
		return
	}
	cell := &g.cells[index]
	cell.pattern |= 1 << ((x % g.w) + g.w*(y%g.h))
	cell.color = paintColor(color)
}

// charGrid is a grid with one character per cell, so a resolution of one dot
// per cell. It is what the Dot, Block, Bar and Custom markers use.
type charGrid struct {
	width  uint16
	height uint16
	// cells holds the color of each cell; an unset Color is an unpainted cell.
	cells []catatui.Color
	// cellChar is drawn in every painted cell.
	cellChar rune
	// applyColorToBg also paints the background, which the Block marker uses
	// so that the block overwrites any earlier symbol yet leaves a background
	// a later layer can draw a symbol over.
	applyColorToBg bool
}

func newCharGrid(width, height uint16, cellChar rune) *charGrid {
	return &charGrid{
		width:    width,
		height:   height,
		cells:    make([]catatui.Color, int(width)*int(height)),
		cellChar: cellChar,
	}
}

func (g *charGrid) withColorOnBg() *charGrid {
	g.applyColorToBg = true
	return g
}

func (g *charGrid) resolution() (float64, float64) {
	return float64(g.width), float64(g.height)
}

func (g *charGrid) save() layer {
	contents := make([]layerCell, len(g.cells))
	for i, color := range g.cells {
		if !color.IsSet() {
			continue
		}
		contents[i] = layerCell{symbol: g.cellChar, hasSymbol: true, fg: color}
		if g.applyColorToBg {
			contents[i].bg = color
		}
	}
	return layer{contents: contents}
}

func (g *charGrid) reset() {
	for i := range g.cells {
		g.cells[i] = catatui.Color{}
	}
}

func (g *charGrid) paint(x, y int, color catatui.Color) {
	if x < 0 || y < 0 || y >= int(g.height) {
		return
	}
	index := y*int(g.width) + x
	if index >= len(g.cells) {
		return
	}
	g.cells[index] = paintColor(color)
}

// halfBlockGrid is a grid with two vertically stacked pixels per cell, drawn
// with the upper half block, lower half block and full block characters. A
// cell has a foreground and a background, so each half can have its own
// color, which a patternGrid cannot do.
type halfBlockGrid struct {
	width  uint16
	height uint16
	// pixels holds one color per pixel, in rows of width, height*2 rows deep.
	pixels [][]catatui.Color
}

func newHalfBlockGrid(width, height uint16) *halfBlockGrid {
	pixels := make([][]catatui.Color, int(height)*2)
	for i := range pixels {
		pixels[i] = make([]catatui.Color, width)
	}
	return &halfBlockGrid{width: width, height: height, pixels: pixels}
}

func (g *halfBlockGrid) resolution() (float64, float64) {
	return float64(g.width), float64(g.height) * 2
}

// save pairs each two rows of pixels into one row of cells. A cell can be in
// four states:
//
//  1. upper unset, lower unset: ' ' with no colors
//  2. upper unset, lower color: '▄' fg lower
//  3. upper color, lower unset: '▀' fg upper
//  4. upper color, lower color: '▀' fg upper, bg lower
//
// Case 2 uses the lower half block rather than an upper half block with a
// background, because the default foreground and background colors usually
// differ, so the reset half has to stay a real reset. When both halves are the
// same color a full block is used instead of case 4, so that tests can treat
// the cell as one character.
func (g *halfBlockGrid) save() layer {
	contents := make([]layerCell, 0, int(g.width)*int(g.height))
	for i := 0; i+1 < len(g.pixels); i += 2 {
		upperRow, lowerRow := g.pixels[i], g.pixels[i+1]
		for x := range upperRow {
			upper, lower := upperRow[x], lowerRow[x]
			var cell layerCell
			switch {
			case !upper.IsSet() && !lower.IsSet():
			case !upper.IsSet():
				cell = layerCell{symbol: halfBlockLowerRune, hasSymbol: true, fg: lower}
			case !lower.IsSet():
				cell = layerCell{symbol: halfBlockUpperRune, hasSymbol: true, fg: upper}
			case upper == lower:
				cell = layerCell{symbol: halfBlockFullRune, hasSymbol: true, fg: upper, bg: lower}
			default:
				cell = layerCell{symbol: halfBlockUpperRune, hasSymbol: true, fg: upper, bg: lower}
			}
			contents = append(contents, cell)
		}
	}
	return layer{contents: contents}
}

func (g *halfBlockGrid) reset() {
	for _, row := range g.pixels {
		for x := range row {
			row[x] = catatui.Color{}
		}
	}
}

func (g *halfBlockGrid) paint(x, y int, color catatui.Color) {
	// ratatui indexes the pixel vector directly and would panic here; a
	// silent ignore is friendlier and matches the other grids.
	if y < 0 || y >= len(g.pixels) || x < 0 || x >= len(g.pixels[y]) {
		return
	}
	g.pixels[y][x] = paintColor(color)
}

var (
	dotRune            = firstRune(symbols.DotFull)
	blockRune          = firstRune(symbols.BlockFull)
	barRune            = firstRune(symbols.BarHalf)
	halfBlockUpperRune = firstRune(symbols.HalfBlockUpper)
	halfBlockLowerRune = firstRune(symbols.HalfBlockLower)
	halfBlockFullRune  = firstRune(symbols.HalfBlockFull)
)

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return ' '
}

// Painter is what a Shape draws with: it converts canvas coordinates to grid
// dots and lights them. Think of it as the Buffer of the canvas world.
type Painter struct {
	context    *Context
	resolution [2]float64
}

// NewPainter returns a painter for the context's current grid. Context.Draw
// creates one for each shape; you only need this to drive a Context by hand.
func NewPainter(ctx *Context) *Painter {
	rx, ry := ctx.grid.resolution()
	return &Painter{context: ctx, resolution: [2]float64{rx, ry}}
}

// GetPoint converts canvas coordinates to the grid dot they land on, or
// reports false if the point is outside the canvas bounds.
//
// Canvas coordinates have their origin at the bottom left and run over the
// canvas's X and Y bounds; grid coordinates have their origin at the top left
// and run over [0, width-1] and [0, height-1] in dots. Points are rounded to
// the nearest dot, with a point exactly between two dots rounding up.
//
//	ctx := widgets.NewContext(2, 2, [2]float64{1, 2}, [2]float64{0, 2}, symbols.Braille)
//	p := widgets.NewPainter(ctx)
//	p.GetPoint(1, 0)   // 0, 7, true
//	p.GetPoint(1.5, 1) // 2, 4, true
//	p.GetPoint(0, 0)   // 0, 0, false
//	p.GetPoint(2, 2)   // 3, 0, true
func (p *Painter) GetPoint(x, y float64) (gridX, gridY int, ok bool) {
	left, right := p.context.xBounds[0], p.context.xBounds[1]
	bottom, top := p.context.yBounds[0], p.context.yBounds[1]
	if x < left || x > right || y < bottom || y > top {
		return 0, 0, false
	}
	width := right - left
	height := top - bottom
	if width <= 0 || height <= 0 {
		return 0, 0, false
	}
	gridX = f64ToIndex(math.Round((x - left) * (p.resolution[0] - 1) / width))
	gridY = f64ToIndex(math.Round((top - y) * (p.resolution[1] - 1) / height))
	return gridX, gridY, true
}

// Paint lights one dot of the grid in the given color. Dots outside the grid
// are ignored. An unset color paints ColorReset.
func (p *Painter) Paint(x, y int, color catatui.Color) {
	p.context.grid.paint(x, y, color)
}

// Bounds returns the canvas's X and Y bounds, each as [low, high].
func (p *Painter) Bounds() (xBounds, yBounds [2]float64) {
	return p.context.xBounds, p.context.yBounds
}

// f64ToIndex converts as Rust's `as usize` does: NaN and negatives become 0,
// and values too large clamp rather than wrap.
func f64ToIndex(f float64) int {
	if f != f || f <= 0 {
		return 0
	}
	if f >= float64(math.MaxInt) {
		return math.MaxInt
	}
	return int(f)
}

// f64ToU16 converts as Rust's `as u16` does: NaN and negatives become 0, and
// values too large clamp at uint16 max.
func f64ToU16(f float64) uint16 {
	if f != f || f <= 0 {
		return 0
	}
	if f >= 0xFFFF {
		return 0xFFFF
	}
	return uint16(f)
}

// Context is the state of a Canvas while it is being painted: the grid being
// drawn into, the layers already saved, and the labels to print on top. It is
// the canvas's counterpart of a Frame.
//
// A Canvas creates one and hands it to the Paint callback; NewContext is for
// driving the machinery directly.
type Context struct {
	// width and height are in cells, not dots.
	width   uint16
	height  uint16
	xBounds [2]float64
	yBounds [2]float64
	grid    grid
	dirty   bool
	layers  []layer
	labels  []label
}

// NewContext returns an empty context of width x height cells, mapping the
// given canvas bounds onto it and drawing with the given marker. xBounds is
// [left, right] and yBounds is [bottom, top].
//
//	ctx := widgets.NewContext(100, 100,
//		[2]float64{-180, 180}, [2]float64{-90, 90}, symbols.Braille)
func NewContext(width, height uint16, xBounds, yBounds [2]float64, marker symbols.Marker) *Context {
	return &Context{
		width:   width,
		height:  height,
		xBounds: xBounds,
		yBounds: yBounds,
		grid:    markerToGrid(width, height, marker),
	}
}

func markerToGrid(width, height uint16, marker symbols.Marker) grid {
	switch marker.Kind {
	case symbols.MarkerBlock:
		return newCharGrid(width, height, blockRune).withColorOnBg()
	case symbols.MarkerBar:
		return newCharGrid(width, height, barRune)
	case symbols.MarkerBraille:
		return newPatternGrid(width, height, 2, 4, symbols.BrailleTable[:])
	case symbols.MarkerHalfBlock:
		return newHalfBlockGrid(width, height)
	case symbols.MarkerQuadrant:
		return newPatternGrid(width, height, 2, 2, symbols.Quadrants[:])
	case symbols.MarkerSextant:
		return newPatternGrid(width, height, 2, 3, symbols.Sextants[:])
	case symbols.MarkerOctant:
		return newPatternGrid(width, height, 2, 4, symbols.Octants[:])
	case symbols.MarkerCustom:
		return newCharGrid(width, height, marker.Rune)
	default:
		return newCharGrid(width, height, dotRune)
	}
}

// Marker switches the marker used from here on. Anything drawn so far is
// saved as a layer first, and a fresh grid of the new kind is started.
func (c *Context) Marker(marker symbols.Marker) {
	c.finish()
	c.grid = markerToGrid(c.width, c.height, marker)
}

// Draw draws a shape into the current layer.
func (c *Context) Draw(shape Shape) {
	c.dirty = true
	shape.Draw(NewPainter(c))
}

// Layer saves what has been drawn so far as a layer and starts a new, empty
// one on top of it. Later layers draw over earlier ones cell by cell, so this
// is how shapes are ordered: draw the background, call Layer, draw the rest.
func (c *Context) Layer() {
	c.layers = append(c.layers, c.grid.save())
	c.grid.reset()
	c.dirty = false
}

// Print places a line of text at a canvas coordinate. Labels are always drawn
// on top of every layer, and are not affected by Layer.
func (c *Context) Print(x, y float64, line catatui.Line) {
	c.labels = append(c.labels, label{x: x, y: y, line: line})
}

// finish saves the current grid as a layer if anything was drawn since the
// last Layer call.
func (c *Context) finish() {
	if c.dirty {
		c.Layer()
	}
}

// Canvas draws shapes (lines, rectangles, circles, a world map, or your own)
// on a grid of pseudo-pixels.
//
// By default the grid is made of braille patterns, which give 2x4 dots per
// cell. Marker selects a different set: Octant is as fine as braille but
// densely packed, Quadrant and Sextant are coarser, HalfBlock gives 1x2 pixels
// each with its own color, and Dot, Block, Bar and Custom draw one character
// per cell for terminals without the fancier glyphs.
//
// Drawing happens in the Paint callback, which receives a Context. Context.Draw
// draws a shape, Context.Layer starts a new layer over what has been drawn,
// and Context.Print puts text on top of everything.
//
//	widgets.NewCanvas().
//		Block(widgets.Bordered().Title("Canvas")).
//		XBounds([2]float64{-180, 180}).
//		YBounds([2]float64{-90, 90}).
//		Paint(func(ctx *widgets.Context) {
//			ctx.Draw(widgets.Map{Resolution: widgets.MapResolutionHigh, Color: catatui.ColorWhite})
//			ctx.Layer()
//			ctx.Draw(widgets.CanvasLine{X1: 0, Y1: 10, X2: 10, Y2: 10, Color: catatui.ColorWhite})
//			ctx.Draw(widgets.Rectangle{X: 10, Y: 20, Width: 10, Height: 10, Color: catatui.ColorRed})
//		})
//
// The zero Canvas draws with the Dot marker; NewCanvas gives ratatui's
// default, which is Braille.
type Canvas struct {
	block           Block
	hasBlock        bool
	xBounds         [2]float64
	yBounds         [2]float64
	paintFunc       func(*Context)
	backgroundColor catatui.Color
	marker          symbols.Marker
}

// NewCanvas returns a canvas with zero bounds, no paint function, a reset
// background and the braille marker, matching ratatui's default.
func NewCanvas() Canvas {
	return Canvas{backgroundColor: catatui.ColorReset, marker: symbols.Braille}
}

// Block returns a copy of c drawn inside the given block.
func (c Canvas) Block(b Block) Canvas { c.block, c.hasBlock = b, true; return c }

// XBounds returns a copy of c showing the given [left, right] range of the
// coordinate system. Change the bounds to zoom or pan.
func (c Canvas) XBounds(bounds [2]float64) Canvas { c.xBounds = bounds; return c }

// YBounds returns a copy of c showing the given [bottom, top] range of the
// coordinate system.
func (c Canvas) YBounds(bounds [2]float64) Canvas { c.yBounds = bounds; return c }

// Paint returns a copy of c that draws by calling f with a fresh Context each
// time the canvas is rendered.
func (c Canvas) Paint(f func(*Context)) Canvas { c.paintFunc = f; return c }

// BackgroundColor returns a copy of c with the whole canvas area painted in
// the given background color before anything is drawn.
func (c Canvas) BackgroundColor(color catatui.Color) Canvas {
	c.backgroundColor = color
	return c
}

// Marker returns a copy of c drawing with the given marker.
//
// Braille is the default and the finest. Use Dot or Block if the target
// terminal lacks braille glyphs, and HalfBlock for a middle ground that keeps
// a separate color per half cell.
func (c Canvas) Marker(m symbols.Marker) Canvas { c.marker = m; return c }

// Render draws the block, paints the background, runs the paint function and
// copies its layers and labels into the buffer.
func (c Canvas) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	canvasArea := area
	if c.hasBlock {
		c.block.Render(area, buf)
		canvasArea = c.block.Inner(area)
	}
	if canvasArea.IsEmpty() {
		return
	}

	buf.SetStyle(canvasArea, catatui.NewStyle().Bg(paintColor(c.backgroundColor)))

	width := int(canvasArea.Width)

	if c.paintFunc == nil {
		return
	}

	// Paint into a blank context the size of the canvas.
	ctx := NewContext(canvasArea.Width, canvasArea.Height, c.xBounds, c.yBounds, c.marker)
	c.paintFunc(ctx)
	ctx.finish()

	// Copy each layer's cells over, bottom layer first.
	for _, l := range ctx.layers {
		for index, lc := range l.contents {
			x := uint16(index%width) + canvasArea.Left()
			y := uint16(index/width) + canvasArea.Top()
			cell := buf.Get(x, y)
			if lc.hasSymbol {
				cell.SetChar(lc.symbol)
			}
			if lc.fg.IsSet() {
				cell.SetFg(lc.fg)
			}
			if lc.bg.IsSet() {
				cell.SetBg(lc.bg)
			}
		}
	}

	// Labels go on last. Unlike shapes they are placed by truncation, not
	// rounding, and measured against the cell grid rather than the dot grid.
	left, right := c.xBounds[0], c.xBounds[1]
	top, bottom := c.yBounds[1], c.yBounds[0]
	boundsWidth := math.Abs(right - left)
	boundsHeight := math.Abs(top - bottom)
	resX := float64(canvasArea.Width - 1)
	resY := float64(canvasArea.Height - 1)
	for _, lb := range ctx.labels {
		if lb.x < left || lb.x > right || lb.y > top || lb.y < bottom {
			continue
		}
		x := catatui.SatAdd(f64ToU16((lb.x-left)*resX/boundsWidth), canvasArea.Left())
		y := catatui.SatAdd(f64ToU16((top-lb.y)*resY/boundsHeight), canvasArea.Top())
		buf.SetLine(x, y, lb.line, catatui.SatSub(canvasArea.Right(), x))
	}
}

var _ catatui.Widget = Canvas{}
