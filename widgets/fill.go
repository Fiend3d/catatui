// Port of ratatui-widgets/src/fill.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// Fill paints every cell of its area with one symbol in one style.
//
// It is the building block for solid regions: a background, a separator, a
// scrollbar track, a hand-drawn border. It saves writing the nested loop and,
// like Clear, silently clips to the buffer.
//
//	fill := widgets.NewFill("X").Style(catatui.NewStyle().Fg(catatui.ColorBlue))
//	f.RenderWidget(fill, area)
//
// renders as
//
//	XXXXXXXXXX
//	XXXXXXXXXX
//	XXXXXXXXXX
type Fill struct {
	symbol string
	style  catatui.Style
}

// NewFill returns a fill that paints symbol into every cell, in the empty
// style. Use Style to color it.
func NewFill(symbol string) Fill {
	return Fill{symbol: symbol}
}

// Style returns a copy of f with the style each cell is painted in.
func (f Fill) Style(s catatui.Style) Fill { f.style = s; return f }

// Symbol returns a copy of f painting a different symbol.
func (f Fill) Symbol(symbol string) Fill { f.symbol = symbol; return f }

// Render paints every cell in area that lies within the buffer.
func (f Fill) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	for _, p := range area.Positions() {
		buf.Get(p.X, p.Y).SetSymbol(f.symbol).SetStyle(f.style)
	}
}

var _ catatui.Widget = Fill{}
