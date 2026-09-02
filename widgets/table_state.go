// Port of ratatui-widgets/src/table/state.rs @ ratatui-v0.30.2

package widgets

import "math"

// TableState is the caller-owned state of a Table: how far it is scrolled and
// which row, column or cell is selected.
//
// Keep it across frames and hand it to RenderStateful each time; the table
// updates the offset so that the selected row stays in view, and clamps the
// selection to the rows and columns it actually has.
//
//	var state widgets.TableState
//	state.SelectNext()
//	catatui.RenderStatefulWidget(table, area, buf, &state)
//
// The selection methods do not know how many rows there are, so SelectLast
// and SelectNext past the end pick a very large index that the next render
// clamps down. That is how ratatui does it too.
type TableState struct {
	offset            int
	selected          int
	hasSelected       bool
	selectedColumn    int
	hasSelectedColumn bool
}

// NewTableState returns a state with no selection and no scroll offset, which
// is also the zero value.
func NewTableState() TableState { return TableState{} }

// WithOffset returns a copy of s scrolled so that the given row is the first
// one shown.
func (s TableState) WithOffset(offset int) TableState { s.offset = offset; return s }

// WithSelected returns a copy of s with the given row selected.
func (s TableState) WithSelected(index int) TableState {
	s.selected, s.hasSelected = index, true
	return s
}

// WithSelectedNone returns a copy of s with no row selected.
func (s TableState) WithSelectedNone() TableState {
	s.selected, s.hasSelected = 0, false
	return s
}

// WithSelectedColumn returns a copy of s with the given column selected.
func (s TableState) WithSelectedColumn(index int) TableState {
	s.selectedColumn, s.hasSelectedColumn = index, true
	return s
}

// WithSelectedColumnNone returns a copy of s with no column selected.
func (s TableState) WithSelectedColumnNone() TableState {
	s.selectedColumn, s.hasSelectedColumn = 0, false
	return s
}

// WithSelectedCell returns a copy of s with the given row and column selected.
func (s TableState) WithSelectedCell(row, column int) TableState {
	s.selected, s.hasSelected = row, true
	s.selectedColumn, s.hasSelectedColumn = column, true
	return s
}

// WithSelectedCellNone returns a copy of s with neither row nor column
// selected.
func (s TableState) WithSelectedCellNone() TableState {
	s.selected, s.hasSelected = 0, false
	s.selectedColumn, s.hasSelectedColumn = 0, false
	return s
}

// Offset returns the index of the first row shown.
func (s TableState) Offset() int { return s.offset }

// SetOffset scrolls so that the given row is the first one shown. The next
// render adjusts it if the selection would otherwise be out of view.
func (s *TableState) SetOffset(offset int) { s.offset = offset }

// Selected returns the selected row, and whether there is one.
func (s TableState) Selected() (int, bool) { return s.selected, s.hasSelected }

// SelectedColumn returns the selected column, and whether there is one.
func (s TableState) SelectedColumn() (int, bool) { return s.selectedColumn, s.hasSelectedColumn }

// SelectedCell returns the selected row and column, and whether both are
// selected.
func (s TableState) SelectedCell() (row, column int, ok bool) {
	if s.hasSelected && s.hasSelectedColumn {
		return s.selected, s.selectedColumn, true
	}
	return 0, 0, false
}

// Select selects a row. An index past the end is clamped on the next render.
func (s *TableState) Select(index int) {
	s.selected, s.hasSelected = index, true
}

// SelectNone clears the row selection and scrolls back to the top.
func (s *TableState) SelectNone() {
	s.selected, s.hasSelected = 0, false
	s.offset = 0
}

// SelectColumn selects a column. An index past the end is clamped on the
// next render.
func (s *TableState) SelectColumn(index int) {
	s.selectedColumn, s.hasSelectedColumn = index, true
}

// SelectColumnNone clears the column selection.
func (s *TableState) SelectColumnNone() {
	s.selectedColumn, s.hasSelectedColumn = 0, false
}

// SelectCell selects a row and a column together.
func (s *TableState) SelectCell(row, column int) {
	s.selected, s.hasSelected = row, true
	s.selectedColumn, s.hasSelectedColumn = column, true
}

// SelectCellNone clears both selections and scrolls back to the top.
func (s *TableState) SelectCellNone() {
	s.offset = 0
	s.selected, s.hasSelected = 0, false
	s.selectedColumn, s.hasSelectedColumn = 0, false
}

// SelectNext selects the row after the current one, or the first row if none
// is selected.
func (s *TableState) SelectNext() {
	next := 0
	if s.hasSelected {
		next = satAddInt(s.selected, 1)
	}
	s.Select(next)
}

// SelectNextColumn selects the column after the current one, or the first
// column if none is selected.
func (s *TableState) SelectNextColumn() {
	next := 0
	if s.hasSelectedColumn {
		next = satAddInt(s.selectedColumn, 1)
	}
	s.SelectColumn(next)
}

// SelectPrevious selects the row before the current one, or the last row if
// none is selected.
func (s *TableState) SelectPrevious() {
	previous := math.MaxInt
	if s.hasSelected {
		previous = satSubInt(s.selected, 1)
	}
	s.Select(previous)
}

// SelectPreviousColumn selects the column before the current one, or the last
// column if none is selected.
func (s *TableState) SelectPreviousColumn() {
	previous := math.MaxInt
	if s.hasSelectedColumn {
		previous = satSubInt(s.selectedColumn, 1)
	}
	s.SelectColumn(previous)
}

// SelectFirst selects the first row.
func (s *TableState) SelectFirst() { s.Select(0) }

// SelectFirstColumn selects the first column.
func (s *TableState) SelectFirstColumn() { s.SelectColumn(0) }

// SelectLast selects the last row.
func (s *TableState) SelectLast() { s.Select(math.MaxInt) }

// SelectLastColumn selects the last column.
func (s *TableState) SelectLastColumn() { s.SelectColumn(math.MaxInt) }

// ScrollDownBy moves the selection down by the given number of rows, starting
// from the first row if none is selected.
func (s *TableState) ScrollDownBy(amount uint16) {
	s.Select(satAddInt(s.selectedOrZero(), int(amount)))
}

// ScrollUpBy moves the selection up by the given number of rows, stopping at
// the first.
func (s *TableState) ScrollUpBy(amount uint16) {
	s.Select(satSubInt(s.selectedOrZero(), int(amount)))
}

// ScrollRightBy moves the column selection right by the given number of
// columns, starting from the first column if none is selected.
func (s *TableState) ScrollRightBy(amount uint16) {
	s.SelectColumn(satAddInt(s.selectedColumnOrZero(), int(amount)))
}

// ScrollLeftBy moves the column selection left by the given number of
// columns, stopping at the first.
func (s *TableState) ScrollLeftBy(amount uint16) {
	s.SelectColumn(satSubInt(s.selectedColumnOrZero(), int(amount)))
}

func (s TableState) selectedOrZero() int {
	if s.hasSelected {
		return s.selected
	}
	return 0
}

func (s TableState) selectedColumnOrZero() int {
	if s.hasSelectedColumn {
		return s.selectedColumn
	}
	return 0
}

// satSubInt subtracts non-negative ints, stopping at zero. It stands in for
// usize::saturating_sub; satAddInt lives in list_state.go.
func satSubInt(a, b int) int {
	if a < b {
		return 0
	}
	return a - b
}
