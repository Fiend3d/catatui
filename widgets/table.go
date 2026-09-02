// Port of ratatui-widgets/src/table.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// Table draws rows of cells in columns, with an optional header and footer,
// and lets the caller select a row, a column or a single cell through a
// TableState.
//
//	table := widgets.NewTable(
//		[]widgets.Row{
//			widgets.NewRowFromStrings("1", "main.go", "2 KiB"),
//			widgets.NewRowFromStrings("2", "go.mod", "80 B"),
//		},
//		catatui.Length(3), catatui.Fill(1), catatui.Length(8),
//	).
//		Header(widgets.NewRowFromStrings("#", "name", "size")).
//		Block(widgets.Bordered().Title("Files")).
//		RowHighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
//		HighlightSymbol(">> ")
//
//	catatui.RenderStatefulWidget(table, area, buf, &state)
//
// Column widths come from the Constraints given to NewTable or Widths, solved
// by the core Layout; with no constraints, the width is shared equally among
// the columns. Rendering it as a plain Widget uses a fresh, unselected state.
//
// The zero value of Table has a column spacing of zero; NewTable sets the
// ratatui default of one.
type Table struct {
	rows                 []Row
	header               Row
	hasHeader            bool
	footer               Row
	hasFooter            bool
	widths               []catatui.Constraint
	columnSpacing        uint16
	block                Block
	hasBlock             bool
	style                catatui.Style
	rowHighlightStyle    catatui.Style
	columnHighlightStyle catatui.Style
	cellHighlightStyle   catatui.Style
	highlightSymbol      catatui.Text
	highlightSpacing     HighlightSpacing
	flex                 catatui.Flex
}

// NewTable returns a table of the given rows with one constraint per column.
//
// It panics if any Percentage constraint is above 100, as ratatui does. Pass
// no widths to share the space equally among the columns.
func NewTable(rows []Row, widths ...catatui.Constraint) Table {
	ensurePercentagesLessThan100(widths)
	return Table{rows: rows, widths: widths, columnSpacing: 1}
}

// Rows returns a copy of t with the rows replaced.
func (t Table) Rows(rows ...Row) Table { t.rows = rows; return t }

// Header returns a copy of t with a header row drawn above the rows. It does
// not scroll with them.
func (t Table) Header(header Row) Table { t.header, t.hasHeader = header, true; return t }

// Footer returns a copy of t with a footer row drawn below the rows. It does
// not scroll with them.
func (t Table) Footer(footer Row) Table { t.footer, t.hasFooter = footer, true; return t }

// Widths returns a copy of t with the column constraints replaced. It panics
// if any Percentage constraint is above 100.
func (t Table) Widths(widths ...catatui.Constraint) Table {
	ensurePercentagesLessThan100(widths)
	t.widths = widths
	return t
}

// ColumnSpacing returns a copy of t with the given number of blank columns
// between adjacent columns. The default is one.
func (t Table) ColumnSpacing(spacing uint16) Table { t.columnSpacing = spacing; return t }

// Block returns a copy of t drawn inside the given block.
func (t Table) Block(b Block) Table { t.block, t.hasBlock = b, true; return t }

// Style returns a copy of t with a style applied to the whole area, beneath
// every row's and cell's own.
func (t Table) Style(s catatui.Style) Table { t.style = s; return t }

// RowHighlightStyle returns a copy of t with the style laid over the selected
// row, including the highlight symbol column.
func (t Table) RowHighlightStyle(s catatui.Style) Table { t.rowHighlightStyle = s; return t }

// ColumnHighlightStyle returns a copy of t with the style laid over the
// selected column, for the full height of the rows area.
func (t Table) ColumnHighlightStyle(s catatui.Style) Table { t.columnHighlightStyle = s; return t }

// CellHighlightStyle returns a copy of t with the style laid over the cell
// where the selected row and column meet, on top of both of their styles.
func (t Table) CellHighlightStyle(s catatui.Style) Table { t.cellHighlightStyle = s; return t }

// HighlightSymbol returns a copy of t drawing the given text at the left of
// the selected row. Its width is the width of the column reserved for it.
func (t Table) HighlightSymbol(symbol string) Table {
	t.highlightSymbol = catatui.TextFromString(symbol)
	return t
}

// HighlightSymbolText returns a copy of t drawing styled text at the left of
// the selected row.
func (t Table) HighlightSymbolText(symbol catatui.Text) Table {
	t.highlightSymbol = symbol
	return t
}

// HighlightSpacing returns a copy of t with the rule for when the highlight
// symbol column is reserved.
func (t Table) HighlightSpacing(h HighlightSpacing) Table { t.highlightSpacing = h; return t }

// FlexMode returns a copy of t with the given handling of leftover width when
// the columns do not fill the area. This is ratatui's Table::flex.
func (t Table) FlexMode(f catatui.Flex) Table { t.flex = f; return t }

// GetStyle returns the table's style.
func (t Table) GetStyle() catatui.Style { return t.style }

// Render draws the table with a fresh state: nothing selected, scrolled to
// the top.
func (t Table) Render(area catatui.Rect, buf *catatui.Buffer) {
	var state TableState
	t.RenderStateful(area, buf, &state)
}

// RenderStateful draws the table, scrolling so that the selected row is in
// view and clamping the selection to the rows and columns that exist.
func (t Table) RenderStateful(area catatui.Rect, buf *catatui.Buffer, state *TableState) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, t.style)
	tableArea := area
	if t.hasBlock {
		t.block.Render(area, buf)
		tableArea = t.block.Inner(area)
	}
	if tableArea.IsEmpty() {
		return
	}

	if state.hasSelected && state.selected >= len(t.rows) {
		state.Select(satSubInt(len(t.rows), 1))
	}
	if len(t.rows) == 0 {
		state.SelectNone()
	}

	columnCount := t.columnCount()
	if state.hasSelectedColumn && state.selectedColumn >= columnCount {
		state.SelectColumn(satSubInt(columnCount, 1))
	}
	if columnCount == 0 {
		state.SelectColumnNone()
	}

	selectionWidth := t.selectionWidth(*state)
	columnWidths := t.getColumnWidths(tableArea.Width, selectionWidth, columnCount)
	headerArea, rowsArea, footerArea := t.layout(tableArea)

	t.renderHeader(headerArea, buf, columnWidths)
	t.renderRows(rowsArea, buf, selectionWidth, state, columnWidths)
	t.renderFooter(footerArea, buf, columnWidths)
}

// layout splits the area into the header, rows and footer areas, with the
// header and footer margins between them.
func (t Table) layout(area catatui.Rect) (headerArea, rowsArea, footerArea catatui.Rect) {
	var header, footer Row
	if t.hasHeader {
		header = t.header
	}
	if t.hasFooter {
		footer = t.footer
	}
	parts := catatui.VerticalLayout(
		catatui.Length(header.topMargin),
		catatui.Length(header.height),
		catatui.Length(header.bottomMargin),
		catatui.Min(0),
		catatui.Length(footer.topMargin),
		catatui.Length(footer.height),
		catatui.Length(footer.bottomMargin),
	).Split(area)
	return parts[1], parts[3], parts[5]
}

func (t Table) renderHeader(area catatui.Rect, buf *catatui.Buffer, columnWidths []catatui.Rect) {
	if !t.hasHeader {
		return
	}
	t.renderFixedRow(t.header, area, buf, columnWidths)
}

func (t Table) renderFooter(area catatui.Rect, buf *catatui.Buffer, columnWidths []catatui.Rect) {
	if !t.hasFooter {
		return
	}
	t.renderFixedRow(t.footer, area, buf, columnWidths)
}

// renderFixedRow draws a header or footer: one cell per column, ignoring
// column spans, as ratatui does for these rows.
func (t Table) renderFixedRow(row Row, area catatui.Rect, buf *catatui.Buffer, columnWidths []catatui.Rect) {
	buf.SetStyle(area, row.style)
	for i, cell := range row.cells {
		if i >= len(columnWidths) {
			break
		}
		cellArea := columnWidths[i]
		x := catatui.SatAdd(area.X, cellArea.X)
		cell.render(catatui.NewRect(x, area.Y, cellArea.Width, area.Height), buf)
	}
}

func (t Table) renderRows(area catatui.Rect, buf *catatui.Buffer, selectionWidth uint16, state *TableState, columnWidths []catatui.Rect) {
	if len(t.rows) == 0 {
		return
	}

	start, end := t.visibleRows(*state, area)
	state.offset = start

	var yOffset uint16
	var selectedRowArea catatui.Rect
	hasSelectedRow := false
	for i := start; i < end; i++ {
		row := t.rows[i]
		y := catatui.SatAdd(catatui.SatAdd(area.Y, yOffset), row.topMargin)
		height := catatui.SatSub(catatui.MinU16(catatui.SatAdd(y, row.height), area.Bottom()), y)
		rowArea := catatui.Rect{X: area.X, Y: y, Width: area.Width, Height: height}
		buf.SetStyle(rowArea, row.style)

		isSelected := state.hasSelected && state.selected == i
		if selectionWidth > 0 && isSelected {
			t.setSelectionStyle(buf, selectionWidth, rowArea, row)
		}
		t.renderRowCells(buf, columnWidths, row.cells, rowArea)
		if isSelected {
			selectedRowArea, hasSelectedRow = rowArea, true
		}
		yOffset = catatui.SatAdd(yOffset, row.heightWithMargin())
	}

	// The selection is clamped by the column count, but the caller may have
	// given fewer widths than there are columns, so this must not index past
	// the widths.
	var selectedColumnArea catatui.Rect
	hasSelectedColumn := false
	if state.hasSelectedColumn && state.selectedColumn < len(columnWidths) {
		cellArea := columnWidths[state.selectedColumn]
		selectedColumnArea = catatui.Rect{
			X:      catatui.SatAdd(cellArea.X, area.X),
			Y:      area.Y,
			Width:  cellArea.Width,
			Height: area.Height,
		}
		hasSelectedColumn = true
	}

	switch {
	case hasSelectedRow && hasSelectedColumn:
		buf.SetStyle(selectedRowArea, t.rowHighlightStyle)
		buf.SetStyle(selectedColumnArea, t.columnHighlightStyle)
		buf.SetStyle(selectedRowArea.Intersection(selectedColumnArea), t.cellHighlightStyle)
	case hasSelectedRow:
		buf.SetStyle(selectedRowArea, t.rowHighlightStyle)
	case hasSelectedColumn:
		buf.SetStyle(selectedColumnArea, t.columnHighlightStyle)
	}
}

// renderRowCells draws a row's cells left to right, each taking as many
// columns as its span says.
func (t Table) renderRowCells(buf *catatui.Buffer, columnWidths []catatui.Rect, cells []Cell, rowArea catatui.Rect) {
	remaining := columnWidths
	for _, cell := range cells {
		cellArea, rest, ok := getCellArea(remaining, cell.columnSpan, t.columnSpacing)
		remaining = rest
		if !ok {
			continue
		}
		x := catatui.SatAdd(rowArea.X, cellArea.X)
		cell.render(catatui.NewRect(x, rowArea.Y, cellArea.Width, rowArea.Height), buf)
	}
}

// setSelectionStyle draws the highlight symbol in the column reserved for it
// at the left of a selected row.
func (t Table) setSelectionStyle(buf *catatui.Buffer, selectionWidth uint16, rowArea catatui.Rect, row Row) {
	selectionArea := rowArea
	selectionArea.Width = selectionWidth
	buf.SetStyle(selectionArea, row.style)
	t.highlightSymbol.Render(selectionArea, buf)
}

// getCellArea takes the columns a cell covers off the front of columns and
// returns the area they make together: the columns' widths plus the spacing
// between them, starting at the first column's position.
//
// A span of zero takes no column and yields no area. A span longer than the
// columns left takes all of them. Nothing is returned once the columns are
// used up.
func getCellArea(columns []catatui.Rect, columnSpan, columnSpacing uint16) (area catatui.Rect, rest []catatui.Rect, ok bool) {
	if columnSpan == 0 {
		return catatui.Rect{}, columns, false
	}
	if len(columns) == 0 {
		return catatui.Rect{}, columns, false
	}
	first := columns[0]
	taken := uint16(1)
	width := first.Width
	rest = columns[1:]
	for taken < columnSpan && len(rest) > 0 {
		width = catatui.SatAdd(width, rest[0].Width)
		taken++
		rest = rest[1:]
	}
	width = catatui.SatAdd(width, catatui.SatMul(taken-1, columnSpacing))
	return catatui.NewRect(first.X, first.Y, width, 1), rest, true
}

// visibleRows returns the range [start, end) of rows to draw: as many whole
// rows as fit from the offset, moved so that the selected row is among them,
// plus one partial row if there is space left over.
func (t Table) visibleRows(state TableState, area catatui.Rect) (start, end int) {
	lastRow := satSubInt(len(t.rows), 1)
	start = min(state.offset, lastRow)

	if state.hasSelected {
		start = min(start, state.selected)
	}

	end = start
	var height uint16

	for _, row := range t.rows[start:] {
		if catatui.SatAdd(height, row.height) > area.Height {
			break
		}
		height = catatui.SatAdd(height, row.heightWithMargin())
		end++
	}

	if state.hasSelected {
		selected := min(state.selected, lastRow)

		// Scroll down until the selected row is visible.
		for selected >= end {
			height = catatui.SatAdd(height, t.rows[end].heightWithMargin())
			end++
			for height > area.Height {
				height = catatui.SatSub(height, t.rows[start].heightWithMargin())
				start++
			}
		}
	}

	// Include a partial row if there is space.
	if height < area.Height && end < len(t.rows) {
		end++
	}

	return start, end
}

// getColumnWidths solves the column layout for the given width, after first
// reserving the highlight symbol column at the left.
//
// The returned rects are relative to the table area, one row high, and there
// is one per constraint. Without constraints the width is shared equally.
func (t Table) getColumnWidths(maxWidth, selectionWidth uint16, colCount int) []catatui.Rect {
	widths := t.widths
	if len(widths) == 0 {
		// Divide the space between each column equally.
		each := catatui.Length(maxWidth / uint16(max(colCount, 1)))
		widths = make([]catatui.Constraint, colCount)
		for i := range widths {
			widths[i] = each
		}
	}
	// This always allocates a selection area, even when it is zero wide.
	areas := catatui.HorizontalLayout(catatui.Length(selectionWidth), catatui.Fill(0)).
		Split(catatui.NewRect(0, 0, maxWidth, 1))
	columnsArea := areas[1]
	rects := catatui.HorizontalLayout(widths...).
		Flex(t.flex).
		Spacing(catatui.Space(t.columnSpacing)).
		Split(columnsArea)
	out := make([]catatui.Rect, len(rects))
	for i, r := range rects {
		out[i] = catatui.NewRect(r.X, 0, r.Width, 1)
	}
	return out
}

// columnCount is the widest row, counting the header and footer too.
func (t Table) columnCount() int {
	n := 0
	for _, r := range t.rows {
		n = max(n, len(r.cells))
	}
	if t.hasFooter {
		n = max(n, len(t.footer.cells))
	}
	if t.hasHeader {
		n = max(n, len(t.header.cells))
	}
	return n
}

// selectionWidth is the width of the highlight symbol column, or zero when
// the spacing rule says not to reserve it.
func (t Table) selectionWidth(state TableState) uint16 {
	if t.highlightSpacing.ShouldAdd(state.hasSelected) {
		return uint16(min(t.highlightSymbol.Width(), 0xFFFF))
	}
	return 0
}

func ensurePercentagesLessThan100(widths []catatui.Constraint) {
	for _, w := range widths {
		if p, ok := w.IsPercentage(); ok && p > 100 {
			panic("Percentages should be between 0 and 100 inclusively.")
		}
	}
}

var (
	_ catatui.Widget                     = Table{}
	_ catatui.StatefulWidget[TableState] = Table{}
)
