// Port of ratatui-widgets/src/table/cell.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// Cell is one entry in a table Row: a block of styled Text with a style of
// its own beneath it.
//
// A cell can hold several lines, but only as many as the row's height will
// show; the rest are cut off.
//
//	widgets.NewCell("Total").Style(catatui.NewStyle().AddModifier(catatui.ModifierBold))
type Cell struct {
	content    catatui.Text
	style      catatui.Style
	columnSpan uint16
}

// NewCell returns a cell holding the given text, split on newlines, spanning
// one column.
func NewCell(content string) Cell {
	return Cell{content: catatui.TextFromString(content), columnSpan: 1}
}

// NewCellFromText returns a cell holding already-styled text, spanning one
// column.
func NewCellFromText(text catatui.Text) Cell {
	return Cell{content: text, columnSpan: 1}
}

// NewCellFromLine returns a cell holding a single styled line, spanning one
// column. A line that sets its own alignment keeps it inside the cell.
func NewCellFromLine(line catatui.Line) Cell {
	return Cell{content: catatui.NewText(line), columnSpan: 1}
}

// Content returns a copy of c holding the given text, split on newlines.
func (c Cell) Content(content string) Cell {
	c.content = catatui.TextFromString(content)
	return c
}

// ContentText returns a copy of c holding already-styled text.
func (c Cell) ContentText(text catatui.Text) Cell { c.content = text; return c }

// ColumnSpan returns a copy of c spread across the given number of columns,
// including the spacing between them. A span of zero makes the cell take no
// column at all, so the next cell lands where this one would have.
//
// Note that the zero value of Cell has a span of zero, as ratatui's
// Cell::default does; the constructors set it to one.
func (c Cell) ColumnSpan(n uint16) Cell { c.columnSpan = n; return c }

// Style returns a copy of c with a style applied beneath the text's own.
func (c Cell) Style(s catatui.Style) Cell { c.style = s; return c }

// GetStyle returns the cell's style.
func (c Cell) GetStyle() catatui.Style { return c.style }

// GetColumnSpan returns how many columns the cell spreads across.
func (c Cell) GetColumnSpan() uint16 { return c.columnSpan }

// render draws the cell into its area: the cell style first, then the text
// over it.
func (c Cell) render(area catatui.Rect, buf *catatui.Buffer) {
	buf.SetStyle(area, c.style)
	c.content.Render(area, buf)
}
