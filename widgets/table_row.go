// Port of ratatui-widgets/src/table/row.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// Row is one line of a Table: a list of Cells, a height, margins above and
// below, and a style beneath every cell's own.
//
//	widgets.NewRowFromStrings("id", "name", "size").
//		Style(catatui.NewStyle().AddModifier(catatui.ModifierBold))
//
// A row built with NewRow or NewRowFromStrings is one cell high. The zero
// value has height zero, matching ratatui's Row::default, and draws nothing
// until Height is set.
type Row struct {
	cells        []Cell
	height       uint16
	topMargin    uint16
	bottomMargin uint16
	style        catatui.Style
}

// NewRow returns a row of the given cells, one line high.
func NewRow(cells ...Cell) Row {
	return Row{cells: cells, height: 1}
}

// NewRowFromStrings returns a row with one unstyled cell per string, one line
// high.
func NewRowFromStrings(cells ...string) Row {
	cs := make([]Cell, len(cells))
	for i, s := range cells {
		cs[i] = NewCell(s)
	}
	return Row{cells: cs, height: 1}
}

// Cells returns a copy of r with the cells replaced.
func (r Row) Cells(cells ...Cell) Row { r.cells = cells; return r }

// Height returns a copy of r with a fixed height. A cell with more lines than
// this is cut off.
func (r Row) Height(h uint16) Row { r.height = h; return r }

// TopMargin returns a copy of r with the given number of blank lines drawn
// before it.
func (r Row) TopMargin(n uint16) Row { r.topMargin = n; return r }

// BottomMargin returns a copy of r with the given number of blank lines drawn
// after it.
func (r Row) BottomMargin(n uint16) Row { r.bottomMargin = n; return r }

// Style returns a copy of r with a style applied beneath every cell's own.
func (r Row) Style(s catatui.Style) Row { r.style = s; return r }

// GetStyle returns the row's style.
func (r Row) GetStyle() catatui.Style { return r.style }

// heightWithMargin is the number of lines the row takes including its margins.
func (r Row) heightWithMargin() uint16 {
	return catatui.SatAdd(catatui.SatAdd(r.height, r.topMargin), r.bottomMargin)
}
