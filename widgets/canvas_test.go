// Tests ported from ratatui-widgets/src/canvas.rs and canvas/{line,rectangle,
// circle,map}.rs @ ratatui-v0.30.2

package widgets

import (
	"math"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// filledBuffer returns a buffer of the given size with every cell showing s,
// the counterpart of ratatui's Buffer::filled(area, Cell::new(s)).
func filledBuffer(width, height uint16, s string) *catatui.Buffer {
	return catatui.NewBufferFilled(catatui.NewRect(0, 0, width, height), catatui.NewCell(s))
}

func TestCanvasHorizontalWithVertical(t *testing.T) {
	cases := []struct {
		name   string
		marker symbols.Marker
		want   []string
	}{
		{"block", symbols.Block, []string{
			"█xxxx",
			"█xxxx",
			"█xxxx",
			"█xxxx",
			"█████",
		}},
		{"half block", symbols.HalfBlock, []string{
			"█xxxx",
			"█xxxx",
			"█xxxx",
			"█xxxx",
			"█▄▄▄▄",
		}},
		{"bar", symbols.Bar, []string{
			"▄xxxx",
			"▄xxxx",
			"▄xxxx",
			"▄xxxx",
			"▄▄▄▄▄",
		}},
		{"braille", symbols.Braille, []string{
			"⡇xxxx",
			"⡇xxxx",
			"⡇xxxx",
			"⡇xxxx",
			"⣇⣀⣀⣀⣀",
		}},
		{"quadrant", symbols.Quadrant, []string{
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"▙▄▄▄▄",
		}},
		{"sextant", symbols.Sextant, []string{
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"🬲🬭🬭🬭🬭",
		}},
		{"octant", symbols.Octant, []string{
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"▌xxxx",
			"𜷀▂▂▂▂",
		}},
		{"x sign", symbols.Custom('×'), []string{
			"×xxxx",
			"×xxxx",
			"×xxxx",
			"×xxxx",
			"×××××",
		}},
		{"plus sign", symbols.Custom('+'), []string{
			"+xxxx",
			"+xxxx",
			"+xxxx",
			"+xxxx",
			"+++++",
		}},
		{"dot", symbols.Dot, []string{
			"•xxxx",
			"•xxxx",
			"•xxxx",
			"•xxxx",
			"•••••",
		}},
	}
	horizontal := CanvasLine{X1: 0, Y1: 0, X2: 10, Y2: 0, Color: catatui.ColorReset}
	vertical := CanvasLine{X1: 0, Y1: 0, X2: 0, Y2: 10, Color: catatui.ColorReset}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := filledBuffer(5, 5, "x")
			NewCanvas().
				Marker(c.marker).
				Paint(func(ctx *Context) {
					ctx.Draw(vertical)
					ctx.Draw(horizontal)
				}).
				XBounds([2]float64{0, 10}).
				YBounds([2]float64{0, 10}).
				Render(buf.Area, buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestCanvasDiagonalLines(t *testing.T) {
	cases := []struct {
		name   string
		marker symbols.Marker
		want   []string
	}{
		{"block", symbols.Block, []string{
			"█xxx█",
			"x█x█x",
			"xx█xx",
			"x█x█x",
			"█xxx█",
		}},
		{"half block", symbols.HalfBlock, []string{
			"█xxx█",
			"x█x█x",
			"xx█xx",
			"x█x█x",
			"█xxx█",
		}},
		{"bar", symbols.Bar, []string{
			"▄xxx▄",
			"x▄x▄x",
			"xx▄xx",
			"x▄x▄x",
			"▄xxx▄",
		}},
		{"braille", symbols.Braille, []string{
			"⢣xxx⡜",
			"x⢣x⡜x",
			"xx⣿xx",
			"x⡜x⢣x",
			"⡜xxx⢣",
		}},
		{"quadrant", symbols.Quadrant, []string{
			"▚xxx▞",
			"x▚x▞x",
			"xx█xx",
			"x▞x▚x",
			"▞xxx▚",
		}},
		{"sextant", symbols.Sextant, []string{
			"🬧xxx🬔",
			"x🬧x🬔x",
			"xx█xx",
			"x🬘x🬣x",
			"🬘xxx🬣",
		}},
		{"octant", symbols.Octant, []string{
			"▚xxx▞",
			"x▚x▞x",
			"xx█xx",
			"x▞x▚x",
			"▞xxx▚",
		}},
		{"x sign", symbols.Custom('×'), []string{
			"×xxx×",
			"x×x×x",
			"xx×xx",
			"x×x×x",
			"×xxx×",
		}},
		{"plus sign", symbols.Custom('+'), []string{
			"+xxx+",
			"x+x+x",
			"xx+xx",
			"x+x+x",
			"+xxx+",
		}},
		{"dot", symbols.Dot, []string{
			"•xxx•",
			"x•x•x",
			"xx•xx",
			"x•x•x",
			"•xxx•",
		}},
	}
	diagonalUp := CanvasLine{X1: 0, Y1: 0, X2: 10, Y2: 10, Color: catatui.ColorReset}
	diagonalDown := CanvasLine{X1: 0, Y1: 10, X2: 10, Y2: 0, Color: catatui.ColorReset}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := filledBuffer(5, 5, "x")
			NewCanvas().
				Marker(c.marker).
				Paint(func(ctx *Context) {
					ctx.Draw(diagonalDown)
					ctx.Draw(diagonalUp)
				}).
				XBounds([2]float64{0, 10}).
				YBounds([2]float64{0, 10}).
				Render(buf.Area, buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

// TestCanvasCheckPaintMax is ratatui's check_canvas_paint_max: the grids do a
// lot of index arithmetic, so paint the extreme corners of huge grids and make
// sure nothing overflows or panics.
func TestCanvasCheckPaintMax(t *testing.T) {
	bGrid := newPatternGrid(math.MaxUint16, 2, 2, 4, symbols.Octants[:])
	cGrid := newCharGrid(math.MaxUint16, 2, 'd')

	const maxU16 = math.MaxUint16

	bGrid.paint(0, 0, catatui.ColorRed)
	bGrid.paint(0, maxU16, catatui.ColorRed)
	bGrid.paint(maxU16, 0, catatui.ColorRed)
	bGrid.paint(maxU16, maxU16, catatui.ColorRed)

	cGrid.paint(0, 0, catatui.ColorRed)
	cGrid.paint(0, maxU16, catatui.ColorRed)
	cGrid.paint(maxU16, 0, catatui.ColorRed)
	cGrid.paint(maxU16, maxU16, catatui.ColorRed)
}

// TestCanvasCheckPaintOverflow deliberately paints outside the grid, including
// at the largest possible index, and expects it to be ignored.
func TestCanvasCheckPaintOverflow(t *testing.T) {
	bGrid := newPatternGrid(math.MaxUint16, 3, 2, 4, symbols.BrailleTable[:])
	cGrid := newCharGrid(math.MaxUint16, 3, 'd')
	hGrid := newHalfBlockGrid(math.MaxUint16, 3)

	const over = math.MaxUint16 + 10

	bGrid.paint(over, over, catatui.ColorRed)
	cGrid.paint(over, over, catatui.ColorRed)
	hGrid.paint(over, over, catatui.ColorRed)

	bGrid.paint(math.MaxInt, math.MaxInt, catatui.ColorRed)
	cGrid.paint(math.MaxInt, math.MaxInt, catatui.ColorRed)
	hGrid.paint(math.MaxInt, math.MaxInt, catatui.ColorRed)

	bGrid.paint(-1, -1, catatui.ColorRed)
	cGrid.paint(-1, -1, catatui.ColorRed)
	hGrid.paint(-1, -1, catatui.ColorRed)
}

func TestCanvasRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	canvas := NewCanvas().
		XBounds([2]float64{0, 10}).
		YBounds([2]float64{0, 10}).
		Paint(func(*Context) {})
	// This should not panic, even if the buffer is too small for the canvas.
	canvas.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestCanvasRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	canvas := NewCanvas().
		XBounds([2]float64{0, 10}).
		YBounds([2]float64{0, 10}).
		Paint(func(*Context) {})
	// This should not panic, even if the buffer has zero size.
	canvas.Render(buf.Area, buf)
}

// TestCanvasPainterGetPoint is the doc example on Painter::get_point.
func TestCanvasPainterGetPoint(t *testing.T) {
	ctx := NewContext(2, 2, [2]float64{1, 2}, [2]float64{0, 2}, symbols.Braille)
	p := NewPainter(ctx)
	cases := []struct {
		x, y   float64
		gx, gy int
		ok     bool
	}{
		{1, 0, 0, 7, true},
		{1.5, 1, 2, 4, true},
		{0, 0, 0, 0, false},
		{2, 2, 3, 0, true},
		{1, 2, 0, 0, true},
	}
	for _, c := range cases {
		gx, gy, ok := p.GetPoint(c.x, c.y)
		if ok != c.ok || (ok && (gx != c.gx || gy != c.gy)) {
			t.Errorf("GetPoint(%v, %v) = (%d, %d, %v), want (%d, %d, %v)",
				c.x, c.y, gx, gy, ok, c.gx, c.gy, c.ok)
		}
	}
}

// TestCanvasLayersAndLabels checks that a later layer draws over an earlier
// one and that labels go on top of everything.
func TestCanvasLayersAndLabels(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 3))
	NewCanvas().
		Marker(symbols.Block).
		XBounds([2]float64{0, 4}).
		YBounds([2]float64{0, 2}).
		Paint(func(ctx *Context) {
			ctx.Draw(CanvasLine{X1: 0, Y1: 2, X2: 4, Y2: 2, Color: catatui.ColorBlue})
			ctx.Layer()
			ctx.Marker(symbols.Custom('o'))
			ctx.Draw(Points{Coords: [][2]float64{{0, 2}}, Color: catatui.ColorRed})
			ctx.Print(0, 0, catatui.LineFromString("label"))
		}).
		Render(buf.Area, buf)

	want := catatui.NewBufferWithStrings(
		"o████",
		"     ",
		"label",
	)
	want.SetStyle(catatui.NewRect(0, 0, 5, 1), catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorBlue))
	want.SetStyle(catatui.NewRect(0, 0, 1, 1), catatui.NewStyle().Fg(catatui.ColorRed))
	catatui.AssertBuffer(t, buf, want)
}

// --- line.rs -------------------------------------------------------------

// redDots is the expectation builder shared by the line tests: every dot in
// the expected buffer is drawn in red, as ratatui's tests set_style it.
func redDots(lines ...string) *catatui.Buffer {
	buf := catatui.NewBufferWithStrings(lines...)
	for i := range buf.Content {
		if buf.Content[i].GetSymbol() == "•" {
			buf.Content[i].SetStyle(catatui.NewStyle().Fg(catatui.ColorRed))
		}
	}
	return buf
}

func repeatCanvasLines(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func renderDotShape(shape Shape) *catatui.Buffer {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10))
	NewCanvas().
		Marker(symbols.Dot).
		XBounds([2]float64{0, 10}).
		YBounds([2]float64{0, 10}).
		Paint(func(ctx *Context) { ctx.Draw(shape) }).
		Render(buf.Area, buf)
	return buf
}

func TestCanvasLine(t *testing.T) {
	blank := repeatCanvasLines("          ", 10)
	cases := []struct {
		name string
		line CanvasLine
		want []string
	}{
		{"off grid 1", NewCanvasLine(-1, 0, -1, 10, catatui.ColorRed), blank},
		{"off grid 2", NewCanvasLine(0, -1, 10, -1, catatui.ColorRed), blank},
		{"off grid 3", NewCanvasLine(-10, 5, -1, 5, catatui.ColorRed), blank},
		{"off grid 4", NewCanvasLine(5, 11, 5, 20, catatui.ColorRed), blank},
		{"off grid 5", NewCanvasLine(-10, 0, 5, 0, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••••••    ",
		}},
		{"off grid 6", NewCanvasLine(-1, -1, 10, 10, catatui.ColorRed), []string{
			"         •",
			"        • ",
			"       •  ",
			"      •   ",
			"     •    ",
			"    •     ",
			"   •      ",
			"  •       ",
			" •        ",
			"•         ",
		}},
		{"off grid 7", NewCanvasLine(0, 0, 11, 11, catatui.ColorRed), []string{
			"         •",
			"        • ",
			"       •  ",
			"      •   ",
			"     •    ",
			"    •     ",
			"   •      ",
			"  •       ",
			" •        ",
			"•         ",
		}},
		{"off grid 8", NewCanvasLine(-1, -1, 11, 11, catatui.ColorRed), []string{
			"         •",
			"        • ",
			"       •  ",
			"      •   ",
			"     •    ",
			"    •     ",
			"   •      ",
			"  •       ",
			" •        ",
			"•         ",
		}},
		{"horizontal 1", NewCanvasLine(0, 0, 10, 0, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••••••••••",
		}},
		{"horizontal 2", NewCanvasLine(10, 10, 0, 10, catatui.ColorRed), []string{
			"••••••••••",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"vertical 1", NewCanvasLine(0, 0, 0, 10, catatui.ColorRed), repeatCanvasLines("•         ", 10)},
		{"vertical 2", NewCanvasLine(10, 10, 10, 0, catatui.ColorRed), repeatCanvasLines("         •", 10)},
		// dy < dx, x1 < x2
		{"diagonal 1", NewCanvasLine(0, 0, 10, 5, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"        ••",
			"      ••  ",
			"    ••    ",
			"  ••      ",
			"••        ",
		}},
		// dy < dx, x1 > x2
		{"diagonal 2", NewCanvasLine(10, 0, 0, 5, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••        ",
			"  ••      ",
			"    ••    ",
			"      ••  ",
			"        ••",
		}},
		// dy > dx, y1 < y2
		{"diagonal 3", NewCanvasLine(0, 0, 5, 10, catatui.ColorRed), []string{
			"     •    ",
			"    •     ",
			"    •     ",
			"   •      ",
			"   •      ",
			"  •       ",
			"  •       ",
			" •        ",
			" •        ",
			"•         ",
		}},
		// dy > dx, y1 > y2
		{"diagonal 4", NewCanvasLine(0, 10, 5, 0, catatui.ColorRed), []string{
			"•         ",
			" •        ",
			" •        ",
			"  •       ",
			"  •       ",
			"   •      ",
			"   •      ",
			"    •     ",
			"    •     ",
			"     •    ",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, renderDotShape(c.line), redDots(c.want...))
		})
	}
}

func TestCanvasFilledLine(t *testing.T) {
	blank := repeatCanvasLines("          ", 10)
	full := repeatCanvasLines("••••••••••", 10)
	cases := []struct {
		name string
		line FilledLine
		want []string
	}{
		{"off grid 1", NewFilledLine(-1, 0, -1, 10, 0, catatui.ColorRed), blank},
		{"off grid 2", NewFilledLine(0, -1, 10, -1, 0, catatui.ColorRed), blank},
		{"off grid 3", NewFilledLine(-10, 5, -1, 5, 0, catatui.ColorRed), blank},
		{"off grid 4", NewFilledLine(5, 11, 5, 20, 0, catatui.ColorRed), blank},
		{"off grid 5", NewFilledLine(-10, 0, 5, 0, -10, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••••••    ",
		}},
		{"off grid 6", NewFilledLine(0, 0, 10, 10, 0, catatui.ColorRed), []string{
			"         •",
			"        ••",
			"       •••",
			"      ••••",
			"     •••••",
			"    ••••••",
			"   •••••••",
			"  ••••••••",
			" •••••••••",
			"••••••••••",
		}},
		{"off grid 7", NewFilledLine(0, 0, 11, 11, 0, catatui.ColorRed), []string{
			"         •",
			"        ••",
			"       •••",
			"      ••••",
			"     •••••",
			"    ••••••",
			"   •••••••",
			"  ••••••••",
			" •••••••••",
			"••••••••••",
		}},
		{"off grid 8", NewFilledLine(-1, -1, 11, 11, 0, catatui.ColorRed), []string{
			"         •",
			"        ••",
			"       •••",
			"      ••••",
			"     •••••",
			"    ••••••",
			"   •••••••",
			"  ••••••••",
			" •••••••••",
			"••••••••••",
		}},
		{"horizontal 1", NewFilledLine(0, 0, 10, 0, 0, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••••••••••",
		}},
		{"horizontal 2", NewFilledLine(0, 0, 10, 0, 10, catatui.ColorRed), full},
		{"horizontal 3", NewFilledLine(10, 10, 0, 10, 10, catatui.ColorRed), []string{
			"••••••••••",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"horizontal 4", NewFilledLine(10, 10, 0, 10, 0, catatui.ColorRed), full},
		{"vertical 1", NewFilledLine(0, 0, 0, 10, 0, catatui.ColorRed), repeatCanvasLines("•         ", 10)},
		{"vertical 2", NewFilledLine(10, 10, 10, 0, 0, catatui.ColorRed), repeatCanvasLines("         •", 10)},
		// dy < dx, x1 < x2
		{"diagonal 1", NewFilledLine(0, 0, 10, 5, 0, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"        ••",
			"      ••••",
			"    ••••••",
			"  ••••••••",
			"••••••••••",
		}},
		// dy < dx, x1 > x2
		{"diagonal 2", NewFilledLine(10, 0, 0, 5, 0, catatui.ColorRed), []string{
			"          ",
			"          ",
			"          ",
			"          ",
			"          ",
			"••        ",
			"••••      ",
			"••••••    ",
			"••••••••  ",
			"••••••••••",
		}},
		// dy > dx, y1 < y2
		{"diagonal 3", NewFilledLine(0, 0, 5, 10, 0, catatui.ColorRed), []string{
			"     •    ",
			"    ••    ",
			"    ••    ",
			"   •••    ",
			"   •••    ",
			"  ••••    ",
			"  ••••    ",
			" •••••    ",
			" •••••    ",
			"••••••    ",
		}},
		// dy > dx, y1 > y2
		{"diagonal 4", NewFilledLine(0, 10, 5, 0, 0, catatui.ColorRed), []string{
			"•         ",
			"••        ",
			"••        ",
			"•••       ",
			"•••       ",
			"••••      ",
			"••••      ",
			"•••••     ",
			"•••••     ",
			"••••••    ",
		}},
		{"split 1", NewFilledLine(0, 0, 10, 10, 5, catatui.ColorRed), []string{
			"         •",
			"        ••",
			"       •••",
			"      ••••",
			"     •••••",
			"••••••••••",
			"••••      ",
			"•••       ",
			"••        ",
			"•         ",
		}},
		{"split 2", NewFilledLine(0, 0, 10, 10, 7, catatui.ColorRed), []string{
			"         •",
			"        ••",
			"       •••",
			"••••••••••",
			"••••••    ",
			"•••••     ",
			"••••      ",
			"•••       ",
			"••        ",
			"•         ",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, renderDotShape(c.line), redDots(c.want...))
		})
	}
}

// --- rectangle.rs --------------------------------------------------------

func renderRectangle(marker symbols.Marker, rect catatui.Rect) *catatui.Buffer {
	buf := catatui.NewBuffer(rect)
	NewCanvas().
		Marker(marker).
		XBounds([2]float64{0, 10}).
		YBounds([2]float64{0, 10}).
		Paint(func(ctx *Context) {
			ctx.Draw(Rectangle{X: 0, Y: 0, Width: 10, Height: 10, Color: catatui.ColorRed})
		}).
		Render(buf.Area, buf)
	return buf
}

func TestCanvasRectangleDrawBlockLines(t *testing.T) {
	want := catatui.NewBufferWithStrings(
		"██████████",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"██████████",
	)
	rect := catatui.NewRect(0, 0, 10, 10)
	want.SetStyle(rect, catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 1)), catatui.ResetStyle())
	catatui.AssertBuffer(t, renderRectangle(symbols.Block, rect), want)
}

func TestCanvasRectangleDrawHalfBlockLines(t *testing.T) {
	want := catatui.NewBufferWithStrings(
		"█▀▀▀▀▀▀▀▀█",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█        █",
		"█▄▄▄▄▄▄▄▄█",
	)
	rect := catatui.NewRect(0, 0, 10, 10)
	want.SetStyle(rect, catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 0)), catatui.ResetStyle().Fg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 1)), catatui.ResetStyle())
	catatui.AssertBuffer(t, renderRectangle(symbols.HalfBlock, rect), want)
}

func TestCanvasRectangleDrawBrailleLines(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10))
	NewCanvas().
		Marker(symbols.Braille).
		XBounds([2]float64{0, 20}).
		YBounds([2]float64{0, 20}).
		Paint(func(ctx *Context) {
			// A rectangle that draws the outside part of the braille cells.
			ctx.Draw(Rectangle{X: 0, Y: 0, Width: 20, Height: 20, Color: catatui.ColorRed})
			// A rectangle that draws the inside part of the braille cells.
			ctx.Draw(Rectangle{X: 4, Y: 4, Width: 12, Height: 12, Color: catatui.ColorGreen})
		}).
		Render(buf.Area, buf)
	want := catatui.NewBufferWithStrings(
		"⡏⠉⠉⠉⠉⠉⠉⠉⠉⢹",
		"⡇        ⢸",
		"⡇ ⡏⠉⠉⠉⠉⢹ ⢸",
		"⡇ ⡇    ⢸ ⢸",
		"⡇ ⡇    ⢸ ⢸",
		"⡇ ⡇    ⢸ ⢸",
		"⡇ ⡇    ⢸ ⢸",
		"⡇ ⣇⣀⣀⣀⣀⣸ ⢸",
		"⡇        ⢸",
		"⣇⣀⣀⣀⣀⣀⣀⣀⣀⣸",
	)
	want.SetStyle(buf.Area, catatui.NewStyle().Fg(catatui.ColorRed))
	want.SetStyle(buf.Area.Inner(catatui.NewMargin(1, 1)), catatui.ResetStyle())
	want.SetStyle(buf.Area.Inner(catatui.NewMargin(2, 2)), catatui.NewStyle().Fg(catatui.ColorGreen))
	want.SetStyle(buf.Area.Inner(catatui.NewMargin(3, 3)), catatui.ResetStyle())
	catatui.AssertBuffer(t, buf, want)
}

func TestCanvasRectangleDrawXLines(t *testing.T) {
	want := catatui.NewBufferWithStrings(
		"××××××××××",
		"×        ×",
		"×        ×",
		"×        ×",
		"×        ×",
		"×        ×",
		"×        ×",
		"×        ×",
		"×        ×",
		"××××××××××",
	)
	rect := catatui.NewRect(0, 0, 10, 10)
	want.SetStyle(rect, catatui.NewStyle().Fg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 0)), catatui.ResetStyle().Fg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 1)), catatui.ResetStyle())
	catatui.AssertBuffer(t, renderRectangle(symbols.Custom('×'), rect), want)
}

func TestCanvasRectangleDrawPlusLines(t *testing.T) {
	want := catatui.NewBufferWithStrings(
		"++++++++++",
		"+        +",
		"+        +",
		"+        +",
		"+        +",
		"+        +",
		"+        +",
		"+        +",
		"+        +",
		"++++++++++",
	)
	rect := catatui.NewRect(0, 0, 10, 10)
	want.SetStyle(rect, catatui.NewStyle().Fg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 0)), catatui.ResetStyle().Fg(catatui.ColorRed))
	want.SetStyle(rect.Inner(catatui.NewMargin(1, 1)), catatui.ResetStyle())
	catatui.AssertBuffer(t, renderRectangle(symbols.Custom('+'), rect), want)
}

// --- circle.rs -----------------------------------------------------------

func TestCanvasCircleItDrawsACircle(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 5))
	NewCanvas().
		Paint(func(ctx *Context) {
			ctx.Draw(Circle{X: 5, Y: 2, Radius: 5, Color: catatui.ColorReset})
		}).
		Marker(symbols.Braille).
		XBounds([2]float64{-10, 10}).
		YBounds([2]float64{-10, 10}).
		Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"      ⣀⣀⣀ ",
		"     ⡞⠁ ⠈⢣",
		"     ⢇⡀ ⢀⡼",
		"      ⠉⠉⠉ ",
		"          ",
	))
}

// --- map.rs --------------------------------------------------------------

func TestCanvasMapResolutionToString(t *testing.T) {
	if got := MapResolutionLow.String(); got != "Low" {
		t.Errorf("Low.String() = %q", got)
	}
	if got := MapResolutionHigh.String(); got != "High" {
		t.Errorf("High.String() = %q", got)
	}
}

func TestCanvasMapResolutionFromStr(t *testing.T) {
	if got, err := ParseMapResolution("Low"); err != nil || got != MapResolutionLow {
		t.Errorf("ParseMapResolution(Low) = %v, %v", got, err)
	}
	if got, err := ParseMapResolution("High"); err != nil || got != MapResolutionHigh {
		t.Errorf("ParseMapResolution(High) = %v, %v", got, err)
	}
	if _, err := ParseMapResolution(""); err == nil {
		t.Errorf("ParseMapResolution(\"\") should fail")
	}
}

// TestCanvasMapDefault is ratatui's `default` test. The zero Map draws at low
// resolution; its zero Color is unset, which the canvas paints as ColorReset,
// matching ratatui's Color::default().
func TestCanvasMapDefault(t *testing.T) {
	var m Map
	if m.Resolution != MapResolutionLow {
		t.Errorf("zero Map resolution = %v, want Low", m.Resolution)
	}
	if m.Color.IsSet() {
		t.Errorf("zero Map color should be unset, got %v", m.Color)
	}
	if got := paintColor(m.Color); got != catatui.ColorReset {
		t.Errorf("unset color should paint as Reset, got %v", got)
	}
}

func TestCanvasMapDrawLow(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 80, 40))
	NewCanvas().
		Marker(symbols.Dot).
		XBounds([2]float64{-180, 180}).
		YBounds([2]float64{-90, 90}).
		Paint(func(ctx *Context) { ctx.Draw(Map{}) }).
		Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                                                                                ",
		"                               •                                                ",
		"               • •• •••••••• ••   ••••    •••••  ••• ••     •••                 ",
		"             •••••••••••••••       •      ••••      • •   •••••••     •••       ",
		"    • •••• ••••••••••••••• ••     ••  •     •••    ••  ••••    ••  ••••••• •••  ",
		"•••••     •••••••••••• •••• •  ••••••     •••• • ••• •••••                     •",
		"   ••  • •   •••• ••••••••  ••••   ••  • •• •  •••                        •• •••",
		"    •••• •••   •••••• •••••   •       •• ••••••                       • •••••   ",
		"•••••     •••     •  ••   ••         •••••••                          ••  •• •• ",
		"            ••    ••••  •••••          ••       •  • •                ••        ",
		"            •  •    •••••••           •• •••• ••• •• •  ••          • ••        ",
		"            •          ••             ••••••••• • ••             •••• •         ",
		"             ••       ••              • • • •• •                  •••••         ",
		"              ••   •••               •      ••••  •               • •           ",
		"               •  •   ••             •         ••  •• •           •             ",
		"    ••          • •••••••           •           •   •  •   •   •• •             ",
		"                 •••••••••          •           •• •   •  • •• •  ••            ",
		"                    ••  ••          •            •••     •   •••  ••            ",
		"                     •••  • •        •  •         •     ••  •••  •••            ",
		"                      •               •  ••                   • ••              ",
		"                   •  •     •••                • •            •••   •••         ",
		"                                •         •     •              • •    •••       ",
		"  •                                        •    • •                  • • •      ",
		"                       •       •                • •               ••• ••       •",
		"                        •      •          •    • ••              •      •   •   ",
		"                        •    •                   •               •       •      ",
		"                        •   •              •   •                    •           ",
		"                           ••               ••                   ••  ••  •   •  ",
		"                       •  •                                           •••    •• ",
		"                       •  •                                            ••   ••  ",
		"                       • •                                                      ",
		"                       •••••                                                    ",
		"                                                                                ",
		"                          ••                                                    ",
		"                         •••           •       • ••••• • •••• • • •• •• ••      ",
		"            •    • • ••••••        ••••••••• • ••      ••                  •••  ",
		"•    ••• •••• ••••   • •  • ••• • •                                        ••• •",
		"   •• •                •  ••  • ••                                         ••   ",
		"•      •                                                                      • ",
		"                                                                                ",
	))
}

func TestCanvasMapDrawHigh(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 80, 40))
	NewCanvas().
		Marker(symbols.Braille).
		XBounds([2]float64{-180, 180}).
		YBounds([2]float64{-90, 90}).
		Paint(func(ctx *Context) { ctx.Draw(Map{Resolution: MapResolutionHigh}) }).
		Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                                                                                ",
		"                   ⢀⣀⣤⠄⠤⠤⣤⣀⡀⣀⣀⡄⠄⢄⣀⣄⡄⢀⡀                                          ",
		"             ⢀⣀⣤⠰⢤⣼⡯⢽⡟⣀⢶⣺⡛⠁       ⠈⢰⠃⠁    ⢖⣒⣾⡟⠂  ⠈⠛⠁        ⠺⢩⢖⡄                ",
		"            ⡬⢍⣿⣟⣿⣻⣿⣿⣿⡾⣯⡀⠈⠁⠁⢦      ⢀⡿       ⠈       ⢠⢶⠘⠋⡁⣀⢠⠤⠖⠘⠉⠁⠈⠼⡧⡄⣄⡀ ⢫⣗⠒⠆      ",
		"⣓  ⣠⠖⠓⠒⠢⠤⢄⠤⠶⠽⠽⣶⣃⣽⡮⣿⡷⣗⣤⡭⣍⢓⡄ ⠸⣷   ⢀⣀⠿⠇       ⢀⠔⠒⠲⠄⢄⢀⡀⢙⣑⡄⠴⡍⣟⠉          ⠑⠉⠉  ⠑⠐⠦⠤⣤⠤⢞",
		"⠶⢧⣗⢾⡆         ⠈⠈⠁⠈⠉⢀⣹⣶⣩⣽⣐⢮⠃ ⣇ ⢀⡔⠊ ⢰⣖⣲    ⢀⡐⠁⣰⠦ ⢲⣶⠛⠋    ⠐⠋                      ⡤",
		"  ⠉⣮⣀⣀⣴⡤⣠⡀         ⡎ ⠛⢫⠙⢫⢫  ⠈⠦⠼          ⡃⡀⢸⠼⣤⡄                        ⡀⣀⣀⡐⡶⣣⢤⠖⠉",
		"   ⢀⡽⠟⠃  ⠈⠱⡀       ⠙⠢⣀⣨⠆⠈⠁⢧⡀          ⣸⣷ ⢹⣷⣼⣸⠃                       ⢀⡐⢀ ⠁⡚⣨⠆   ",
		"          ⠘⢳⡀        ⠈⠾  ⣀⣀⣽         ⠸⢼⣇⡧⠋⠉⠁                          ⠉⣿  ⠢⠂    ",
		"           ⠈⢻           ⠜⢹⣵⠻⠇         ⠈⢻  ⢀⡀  ⢠⣠⡤ ⢀⢤                  ⢰⣯        ",
		"            ⢼          ⢀⣾⠛⠉          ⠐⡖⠒⡰⢺⣞⣵⡄⢀⣏⡭⣙⡄⢕⢫⡀             ⢀ ⢠⠖⢱⡿⠃       ",
		"            ⠸         ⠠⡎             ⠰⣅⣰⣃⣘⡣⡿⢻⡿⣁⣀  ⠸⣽             ⠐⣿⣽⣫ ⡸⡇        ",
		"             ⠳⣄       ⡰⠃             ⢀⠎⠉  ⢧⡀⣠⣛⠈⢻                  ⢻⠘⢺⡿⠚⠁        ",
		"              ⢻⣇  ⣠⠲⠖⢲⡇              ⡸     ⠉⠃⠈⠉⣿  ⢰⣆              ⢸ ⠈⠁          ",
		"              ⠈⢿⣆ ⡟  ⣘⣻             ⡸          ⢸⢇ ⠈⠯⢿⡒⠲⡀   ⢀⡀    ⣀⢾             ",
		"    ⠈⢳          ⠸⡀⢳⣠⢾⠉⢹⣦⣤⣀          ⡇           ⡿⡄  ⢰⠃ ⠑⡂ ⢠⠏⢣  ⣼⡮⠁⢈⡀            ",
		"                 ⠙⠲⢆⡿⢦⠈⠉⠁⠁          ⡇           ⠱⣇⣀⠼⠃   ⡃⢰⠃ ⠸⢶ ⠘⠄ ⢾⡁            ",
		"                    ⠙⣾⣀⡴⡶⢤⣤         ⢳            ⠻⠵⡆    ⠸⣸   ⢸⡳⡤⠃⢀⡾⣿            ",
		"                     ⠘⢻⠁  ⠈⠦⣄        ⢧⣀⣀⠤⣀        ⢐⠁    ⠈⠩⠆  ⣘⣧⠁ ⡸⡔⢿            ",
		"                      ⡸     ⢨         ⠁  ⠉⡇      ⢀⠎          ⢻⢿⠄⡴⢑⣧⡠⡄           ",
		"                      ⡇     ⠈⠋⠦⡄         ⠈⡆     ⢠⠃            ⢏⡇⢧⣼⣾⣧⣽⣿⠶⢤⡀⣤      ",
		"                      ⣇        ⠈⡇         ⢸     ⢸             ⠈⠶⣦⣄⣋⣁⡀⠸⣵⢠⣻⠋⠷⣄    ",
		"                      ⠰⡀       ⣰⠁         ⢘⠆    ⢸ ⢠⡀              ⠙⠋⢠⠦⡄⣷⠙⠃ ⠙    ",
		"⠄                      ⠣⡀      ⡃          ⢸     ⣸⢡⢾⠆               ⡞⠛⠘⢧⡏⡆   ⠸⠄ ⡤",
		"                        ⠱     ⢠⠃          ⠸⡀   ⢸⠁⢸⢨              ⡤⠚     ⠱⡀  ⢦  ⠁",
		"                        ⠅    ⡖⠉            ⡇   ⡜ ⠸⠔              ⡇       ⢳      ",
		"                        ⡇   ⢀⠃             ⢱⡀ ⢰⠃                 ⣇  ⢀⡀   ⢸      ",
		"                       ⢀⠃  ⡦⠏              ⠈⠷⠖⠃                  ⠾⠴⠊⠁⠹⣦  ⡞    ⣄ ",
		"                       ⢸  ⡤⠃                                          ⠘⢲⠖⠃    ⣽⡆",
		"                       ⢸ ⣸⠁                                            ⠈⠿   ⢀⢼⠏ ",
		"                       ⠞ ⡗                             ⣄                    ⠈⠋  ",
		"                       ⢧⡼⡁⠲⠂                                                    ",
		"                        ⠙⠉                                                      ",
		"                           ⡀                                                    ",
		"                         ⣴⠏⠁                      ⣀⡤⢤⣀⣀  ⢀⣀⣤⣀⣀⡴⣄⡤⢤⣀⠤⠤⠴⣄⣀⡀       ",
		"                 ⣀⣀    ⣠⣿⡍⣆          ⣠⣤⣤⠤⠴⠶⠖⠲⠤⠔⠛⠒⠉   ⠈⠨⣇⠖⠋              ⠈⠉⠓⠢⠤⢄  ",
		"     ⡀ ⣠⠤⠴⠒⠚⠛⠛⠒⠢⠤⠿⠙⠉⠉⠑⢋⣚⣉⠥⠚      ⢀⣀⡠⠟⠁                                      ⡴⠋  ",
		"   ⠐⠶⣛⣫⡤              ⠐⢏⣀⣤⣤ ⣴⣋⢇⢀⣮⡥                                         ⣴⠓   ",
		"⠤⠤⠤⠤⡀⣈⢣⣠⡄                 ⠉⠊⠉⠉⠉                                            ⠈⠓⠆⠤⠤",
		"                                                                                ",
	))
}
