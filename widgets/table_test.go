// Tests ported from ratatui-widgets/src/table.rs, table/row.rs, table/cell.rs
// and table/state.rs @ ratatui-v0.30.2

package widgets

import (
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

var (
	tableRed   = catatui.NewStyle().Fg(catatui.ColorRed)
	tableBlue  = catatui.NewStyle().Fg(catatui.ColorBlue)
	tableGreen = catatui.NewStyle().Fg(catatui.ColorGreen)
)

// tableStyle is the style ratatui's tests build with Style::default().red().italic().
var tableStyle = catatui.NewStyle().Fg(catatui.ColorRed).AddModifier(catatui.ModifierItalic)

// lengths returns n copies of Length(w), standing in for Rust's [Length(w); n].
func lengths(w uint16, n int) []catatui.Constraint {
	out := make([]catatui.Constraint, n)
	for i := range out {
		out[i] = catatui.Length(w)
	}
	return out
}

func assertRects(t *testing.T, got, want []catatui.Rect) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("column widths = %v, want %v", got, want)
	}
}

func assertTableSelected(t *testing.T, s TableState, want int, wantOk bool) {
	t.Helper()
	got, ok := s.Selected()
	if ok != wantOk || (ok && got != want) {
		t.Errorf("Selected() = (%d, %v), want (%d, %v)", got, ok, want, wantOk)
	}
}

func assertTableSelectedColumn(t *testing.T, s TableState, want int, wantOk bool) {
	t.Helper()
	got, ok := s.SelectedColumn()
	if ok != wantOk || (ok && got != want) {
		t.Errorf("SelectedColumn() = (%d, %v), want (%d, %v)", got, ok, want, wantOk)
	}
}

// --- Row -------------------------------------------------------------------

func TestRowNew(t *testing.T) {
	cells := []Cell{NewCell("")}
	row := NewRow(cells...)
	if !reflect.DeepEqual(row.cells, cells) {
		t.Errorf("cells = %v, want %v", row.cells, cells)
	}
	if row.height != 1 {
		t.Errorf("height = %d, want 1", row.height)
	}
}

func TestRowCells(t *testing.T) {
	cells := []Cell{NewCell("")}
	row := Row{}.Cells(cells...)
	if !reflect.DeepEqual(row.cells, cells) {
		t.Errorf("cells = %v, want %v", row.cells, cells)
	}
}

func TestRowHeight(t *testing.T) {
	if row := (Row{}).Height(2); row.height != 2 {
		t.Errorf("height = %d, want 2", row.height)
	}
}

func TestRowTopMargin(t *testing.T) {
	if row := (Row{}).TopMargin(1); row.topMargin != 1 {
		t.Errorf("topMargin = %d, want 1", row.topMargin)
	}
}

func TestRowBottomMargin(t *testing.T) {
	if row := (Row{}).BottomMargin(1); row.bottomMargin != 1 {
		t.Errorf("bottomMargin = %d, want 1", row.bottomMargin)
	}
}

func TestRowStyle(t *testing.T) {
	if row := (Row{}).Style(tableStyle); row.style != tableStyle {
		t.Errorf("style = %v, want %v", row.style, tableStyle)
	}
}

// --- Cell ------------------------------------------------------------------

func TestCellNew(t *testing.T) {
	cell := NewCell("")
	if !reflect.DeepEqual(cell.content, catatui.TextFromString("")) {
		t.Errorf("content = %v, want empty text", cell.content)
	}
}

func TestCellContent(t *testing.T) {
	cell := Cell{}.Content("")
	if !reflect.DeepEqual(cell.content, catatui.TextFromString("")) {
		t.Errorf("content = %v, want empty text", cell.content)
	}
}

func TestCellStyle(t *testing.T) {
	if cell := (Cell{}).Style(tableStyle); cell.style != tableStyle {
		t.Errorf("style = %v, want %v", cell.style, tableStyle)
	}
}

// --- TableState --------------------------------------------------------------

func TestTableStateNew(t *testing.T) {
	state := NewTableState()
	if state.offset != 0 {
		t.Errorf("offset = %d, want 0", state.offset)
	}
	assertTableSelected(t, state, 0, false)
	assertTableSelectedColumn(t, state, 0, false)
}

func TestTableStateWithOffset(t *testing.T) {
	if state := NewTableState().WithOffset(1); state.offset != 1 {
		t.Errorf("offset = %d, want 1", state.offset)
	}
}

func TestTableStateWithSelected(t *testing.T) {
	assertTableSelected(t, NewTableState().WithSelected(1), 1, true)
}

func TestTableStateWithSelectedColumn(t *testing.T) {
	assertTableSelectedColumn(t, NewTableState().WithSelectedColumn(1), 1, true)
}

func TestTableStateWithSelectedCellNone(t *testing.T) {
	state := NewTableState().WithSelectedCell(1, 5).WithSelectedCellNone()
	assertTableSelected(t, state, 0, false)
	assertTableSelectedColumn(t, state, 0, false)
}

func TestTableStateOffset(t *testing.T) {
	if got := NewTableState().Offset(); got != 0 {
		t.Errorf("Offset() = %d, want 0", got)
	}
}

func TestTableStateSetOffset(t *testing.T) {
	state := NewTableState()
	state.SetOffset(1)
	if state.offset != 1 {
		t.Errorf("offset = %d, want 1", state.offset)
	}
}

func TestTableStateSelected(t *testing.T) {
	assertTableSelected(t, NewTableState(), 0, false)
}

func TestTableStateSelectedColumn(t *testing.T) {
	assertTableSelectedColumn(t, NewTableState(), 0, false)
}

func TestTableStateSelectedCell(t *testing.T) {
	if _, _, ok := NewTableState().SelectedCell(); ok {
		t.Error("SelectedCell() reported a selection on a fresh state")
	}
}

func TestTableStateSelect(t *testing.T) {
	state := NewTableState()
	state.Select(1)
	assertTableSelected(t, state, 1, true)
}

func TestTableStateSelectNone(t *testing.T) {
	state := NewTableState().WithSelected(1)
	state.SelectNone()
	assertTableSelected(t, state, 0, false)
}

func TestTableStateSelectColumn(t *testing.T) {
	state := NewTableState()
	state.SelectColumn(1)
	assertTableSelectedColumn(t, state, 1, true)
}

func TestTableStateSelectColumnNone(t *testing.T) {
	state := NewTableState().WithSelectedColumn(1)
	state.SelectColumnNone()
	assertTableSelectedColumn(t, state, 0, false)
}

func TestTableStateSelectCell(t *testing.T) {
	state := NewTableState()
	state.SelectCell(1, 5)
	row, col, ok := state.SelectedCell()
	if !ok || row != 1 || col != 5 {
		t.Errorf("SelectedCell() = (%d, %d, %v), want (1, 5, true)", row, col, ok)
	}
}

func TestTableStateSelectCellNone(t *testing.T) {
	state := NewTableState().WithSelectedCell(1, 5)
	state.SelectCellNone()
	assertTableSelected(t, state, 0, false)
	assertTableSelectedColumn(t, state, 0, false)
	if _, _, ok := state.SelectedCell(); ok {
		t.Error("SelectedCell() still reports a selection")
	}
}

func TestTableStateNavigation(t *testing.T) {
	var state TableState
	state.SelectFirst()
	assertTableSelected(t, state, 0, true)

	state.SelectPrevious() // should not go below 0
	assertTableSelected(t, state, 0, true)

	state.SelectNext()
	assertTableSelected(t, state, 1, true)

	state.SelectPrevious()
	assertTableSelected(t, state, 0, true)

	state.SelectLast()
	assertTableSelected(t, state, math.MaxInt, true)

	state.SelectNext() // should not go above MaxInt
	assertTableSelected(t, state, math.MaxInt, true)

	state.SelectPrevious()
	assertTableSelected(t, state, math.MaxInt-1, true)

	state.SelectNext()
	assertTableSelected(t, state, math.MaxInt, true)

	state = TableState{}
	state.SelectNext()
	assertTableSelected(t, state, 0, true)

	state = TableState{}
	state.SelectPrevious()
	assertTableSelected(t, state, math.MaxInt, true)

	state = TableState{}
	state.Select(2)
	state.ScrollDownBy(4)
	assertTableSelected(t, state, 6, true)

	state = TableState{}
	state.ScrollUpBy(3)
	assertTableSelected(t, state, 0, true)

	state.Select(6)
	state.ScrollUpBy(4)
	assertTableSelected(t, state, 2, true)

	state.ScrollUpBy(4)
	assertTableSelected(t, state, 0, true)

	state = TableState{}
	state.SelectFirstColumn()
	assertTableSelectedColumn(t, state, 0, true)

	state.SelectPreviousColumn()
	assertTableSelectedColumn(t, state, 0, true)

	state.SelectNextColumn()
	assertTableSelectedColumn(t, state, 1, true)

	state.SelectPreviousColumn()
	assertTableSelectedColumn(t, state, 0, true)

	state.SelectLastColumn()
	assertTableSelectedColumn(t, state, math.MaxInt, true)

	state.SelectPreviousColumn()
	assertTableSelectedColumn(t, state, math.MaxInt-1, true)

	state = TableState{}.WithSelectedColumn(12)
	state.ScrollRightBy(4)
	assertTableSelectedColumn(t, state, 16, true)

	state.ScrollLeftBy(20)
	assertTableSelectedColumn(t, state, 0, true)

	state.ScrollRightBy(100)
	assertTableSelectedColumn(t, state, 100, true)

	state.ScrollLeftBy(20)
	assertTableSelectedColumn(t, state, 80, true)
}

// --- Table builders ----------------------------------------------------------

func TestTableNew(t *testing.T) {
	rows := []Row{NewRow(NewCell(""))}
	widths := []catatui.Constraint{catatui.Percentage(100)}
	table := NewTable(rows, widths...)
	if !reflect.DeepEqual(table.rows, rows) {
		t.Errorf("rows = %v, want %v", table.rows, rows)
	}
	if table.hasHeader || table.hasFooter || table.hasBlock {
		t.Error("a new table has a header, footer or block")
	}
	if !slices.Equal(table.widths, widths) {
		t.Errorf("widths = %v, want %v", table.widths, widths)
	}
	if table.columnSpacing != 1 {
		t.Errorf("columnSpacing = %d, want 1", table.columnSpacing)
	}
	if !table.style.IsEmpty() || !table.rowHighlightStyle.IsEmpty() {
		t.Error("a new table has a style")
	}
	if !reflect.DeepEqual(table.highlightSymbol, catatui.Text{}) {
		t.Errorf("highlightSymbol = %v, want empty", table.highlightSymbol)
	}
	if table.highlightSpacing != HighlightSpacingWhenSelected {
		t.Errorf("highlightSpacing = %v, want WhenSelected", table.highlightSpacing)
	}
	if table.flex != catatui.FlexStart {
		t.Errorf("flex = %v, want Start", table.flex)
	}
}

func TestTableDefault(t *testing.T) {
	table := NewTable(nil)
	if len(table.rows) != 0 || len(table.widths) != 0 {
		t.Error("a default table has rows or widths")
	}
	if table.hasHeader || table.hasFooter || table.hasBlock {
		t.Error("a default table has a header, footer or block")
	}
	if table.columnSpacing != 1 {
		t.Errorf("columnSpacing = %d, want 1", table.columnSpacing)
	}
	if !table.style.IsEmpty() || !table.rowHighlightStyle.IsEmpty() {
		t.Error("a default table has a style")
	}
	if table.highlightSpacing != HighlightSpacingWhenSelected {
		t.Errorf("highlightSpacing = %v, want WhenSelected", table.highlightSpacing)
	}
	if table.flex != catatui.FlexStart {
		t.Errorf("flex = %v, want Start", table.flex)
	}
}

func TestTableWidths(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Length(100))
	assertConstraints(t, table.widths, []catatui.Constraint{catatui.Length(100)})

	table = NewTable(nil).Widths([]catatui.Constraint{catatui.Length(100)}...)
	assertConstraints(t, table.widths, []catatui.Constraint{catatui.Length(100)})
}

func assertConstraints(t *testing.T, got, want []catatui.Constraint) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("widths = %v, want %v", got, want)
	}
}

func TestTableRows(t *testing.T) {
	rows := []Row{NewRow(NewCell(""))}
	table := NewTable(nil).Rows(rows...)
	if !reflect.DeepEqual(table.rows, rows) {
		t.Errorf("rows = %v, want %v", table.rows, rows)
	}
}

func TestTableColumnSpacing(t *testing.T) {
	if table := NewTable(nil).ColumnSpacing(2); table.columnSpacing != 2 {
		t.Errorf("columnSpacing = %d, want 2", table.columnSpacing)
	}
}

func TestTableBlock(t *testing.T) {
	block := Bordered().Title("Table")
	table := NewTable(nil).Block(block)
	if !table.hasBlock || !reflect.DeepEqual(table.block, block) {
		t.Errorf("block = %v, want %v", table.block, block)
	}
}

func TestTableHeader(t *testing.T) {
	header := NewRow(NewCell(""))
	table := NewTable(nil).Header(header)
	if !table.hasHeader || !reflect.DeepEqual(table.header, header) {
		t.Errorf("header = %v, want %v", table.header, header)
	}
}

func TestTableFooter(t *testing.T) {
	footer := NewRow(NewCell(""))
	table := NewTable(nil).Footer(footer)
	if !table.hasFooter || !reflect.DeepEqual(table.footer, footer) {
		t.Errorf("footer = %v, want %v", table.footer, footer)
	}
}

func TestTableRowHighlightStyle(t *testing.T) {
	if table := NewTable(nil).RowHighlightStyle(tableStyle); table.rowHighlightStyle != tableStyle {
		t.Errorf("rowHighlightStyle = %v, want %v", table.rowHighlightStyle, tableStyle)
	}
}

func TestTableColumnHighlightStyle(t *testing.T) {
	if table := NewTable(nil).ColumnHighlightStyle(tableStyle); table.columnHighlightStyle != tableStyle {
		t.Errorf("columnHighlightStyle = %v, want %v", table.columnHighlightStyle, tableStyle)
	}
}

func TestTableCellHighlightStyle(t *testing.T) {
	if table := NewTable(nil).CellHighlightStyle(tableStyle); table.cellHighlightStyle != tableStyle {
		t.Errorf("cellHighlightStyle = %v, want %v", table.cellHighlightStyle, tableStyle)
	}
}

func TestTableHighlightSymbol(t *testing.T) {
	table := NewTable(nil).HighlightSymbol(">>")
	if !reflect.DeepEqual(table.highlightSymbol, catatui.TextFromString(">>")) {
		t.Errorf("highlightSymbol = %v, want >>", table.highlightSymbol)
	}
}

func TestTableHighlightSpacing(t *testing.T) {
	table := NewTable(nil).HighlightSpacing(HighlightSpacingAlways)
	if table.highlightSpacing != HighlightSpacingAlways {
		t.Errorf("highlightSpacing = %v, want Always", table.highlightSpacing)
	}
}

func TestTableInvalidPercentages(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Widths(Percentage(110)) did not panic")
		}
		if !strings.Contains(r.(string), "Percentages should be between 0 and 100 inclusively") {
			t.Errorf("panic = %v, want the percentage message", r)
		}
	}()
	_ = NewTable(nil).Widths(catatui.Percentage(110))
}

// --- state clamping ------------------------------------------------------------

func TestTableStateEmptyList(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10))
	var state TableState
	table := NewTable(nil, catatui.Percentage(100))
	state.SelectFirst()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelected(t, state, 0, false)
	assertTableSelectedColumn(t, state, 0, false)
}

func TestTableStateSingleItem(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 10))
	var state TableState
	table := NewTable([]Row{NewRowFromStrings("Item 1")}, catatui.Percentage(100))

	state.SelectFirst()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelected(t, state, 0, true)
	assertTableSelectedColumn(t, state, 0, false)

	state.SelectLast()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelected(t, state, 0, true)
	assertTableSelectedColumn(t, state, 0, false)

	state.SelectPrevious()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelected(t, state, 0, true)
	assertTableSelectedColumn(t, state, 0, false)

	state.SelectNext()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelected(t, state, 0, true)
	assertTableSelectedColumn(t, state, 0, false)

	state = TableState{}

	state.SelectFirstColumn()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelectedColumn(t, state, 0, true)
	assertTableSelected(t, state, 0, false)

	state.SelectLastColumn()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelectedColumn(t, state, 0, true)
	assertTableSelected(t, state, 0, false)

	state.SelectPreviousColumn()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelectedColumn(t, state, 0, true)
	assertTableSelected(t, state, 0, false)

	state.SelectNextColumn()
	table.RenderStateful(buf.Area, buf, &state)
	assertTableSelectedColumn(t, state, 0, true)
	assertTableSelected(t, state, 0, false)
}

// --- rendering ------------------------------------------------------------------

func TestTableRenderEmptyArea(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	table := NewTable([]Row{NewRowFromStrings("Cell1", "Cell2")}, lengths(5, 2)...)
	table.Render(catatui.NewRect(0, 0, 0, 0), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3)))
}

func TestTableRenderDefault(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	NewTable(nil).Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3)))
}

func TestTableRenderWithBlock(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).Block(Bordered().Title("Block"))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌Block────────┐",
		"│Cell1 Cell2  │",
		"└─────────────┘",
	))
}

func TestTableColspans2Cols(t *testing.T) {
	cases := []struct {
		name string
		rows []Row
		want []string
	}{
		{
			"all span one",
			[]Row{
				NewRow(NewCell("Cell1").ColumnSpan(1), NewCell("Cell2").ColumnSpan(1)),
				NewRow(NewCell("Cell3").ColumnSpan(1), NewCell("Cell4").ColumnSpan(1)),
			},
			[]string{"Cell1 Cell2    ", "Cell3 Cell4    "},
		},
		{
			"span zero takes no column",
			[]Row{
				NewRow(NewCell("Cell1").ColumnSpan(0), NewCell("Cell2").ColumnSpan(1)),
				NewRow(NewCell("Cell3").ColumnSpan(1), NewCell("Cell4").ColumnSpan(1)),
			},
			[]string{"Cell2          ", "Cell3 Cell4    "},
		},
		{
			"span two takes both",
			[]Row{
				NewRow(NewCell("Cell1").ColumnSpan(2), NewCell("Cell2").ColumnSpan(1)),
				NewRow(NewCell("Cell3").ColumnSpan(1), NewCell("Cell4").ColumnSpan(1)),
			},
			[]string{"Cell1          ", "Cell3 Cell4    "},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 2))
			NewTable(c.rows, lengths(5, 2)...).Render(catatui.NewRect(0, 0, 15, 2), buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestTableColspans3Cols(t *testing.T) {
	cases := []struct {
		name        string
		width       uint16
		columnWidth uint16
		rows        []Row
		want        []string
	}{
		{
			"first spans two", 17, 5,
			[]Row{
				NewRow(NewCell("Cell1").ColumnSpan(2), NewCell("Cell2").ColumnSpan(1)),
				NewRow(NewCell("Cell3").ColumnSpan(1), NewCell("Cell4").ColumnSpan(1), NewCell("Cell5").ColumnSpan(1)),
			},
			[]string{"Cell1       Cell2", "Cell3 Cell4 Cell5"},
		},
		{
			"middle spans two", 17, 5,
			[]Row{
				NewRow(NewCell("Cell1").ColumnSpan(1), NewCell("Cell2").ColumnSpan(2), NewCell("Cell3").ColumnSpan(1)),
				NewRow(NewCell("Cell4").ColumnSpan(1), NewCell("Cell5").ColumnSpan(1), NewCell("Cell6").ColumnSpan(1)),
			},
			[]string{"Cell1 Cell2      ", "Cell4 Cell5 Cell6"},
		},
		{
			"long content is cut", 15, 5,
			[]Row{
				NewRow(NewCell("11111111111111111111").ColumnSpan(2), NewCell("22222222222222222222").ColumnSpan(1)),
				NewRow(NewCell("33333333333333333333").ColumnSpan(1), NewCell("44444444444444444444").ColumnSpan(2), NewCell("55555555555555555555").ColumnSpan(1)),
			},
			[]string{"1111111111 2222", "3333 4444444444"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, c.width, 2))
			NewTable(c.rows, lengths(c.columnWidth, 3)...).Render(catatui.NewRect(0, 0, c.width, 2), buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestTableRenderWithHeader(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).Header(NewRowFromStrings("Head1", "Head2"))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Head1 Head2    ",
		"Cell1 Cell2    ",
		"Cell3 Cell4    ",
	))
}

func TestTableRenderWithFooter(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).Footer(NewRowFromStrings("Foot1", "Foot2"))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Cell1 Cell2    ",
		"Cell3 Cell4    ",
		"Foot1 Foot2    ",
	))
}

func TestTableRenderWithHeaderAndFooter(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{NewRowFromStrings("Cell1", "Cell2")}
	table := NewTable(rows, lengths(5, 2)...).
		Header(NewRowFromStrings("Head1", "Head2")).
		Footer(NewRowFromStrings("Foot1", "Foot2"))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Head1 Head2    ",
		"Cell1 Cell2    ",
		"Foot1 Foot2    ",
	))
}

func TestTableRenderWithHeaderMargin(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).
		Header(NewRowFromStrings("Head1", "Head2").BottomMargin(1))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Head1 Head2    ",
		"               ",
		"Cell1 Cell2    ",
	))
}

func TestTableRenderWithFooterMargin(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{NewRowFromStrings("Cell1", "Cell2")}
	table := NewTable(rows, lengths(5, 2)...).
		Footer(NewRowFromStrings("Foot1", "Foot2").TopMargin(1))
	table.Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Cell1 Cell2    ",
		"               ",
		"Foot1 Foot2    ",
	))
}

func TestTableRenderWithRowMargin(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2").BottomMargin(1),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	NewTable(rows, lengths(5, 2)...).Render(catatui.NewRect(0, 0, 15, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Cell1 Cell2    ",
		"               ",
		"Cell3 Cell4    ",
	))
}

func TestTableRenderWithTallRow(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 23, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRow(
			NewCell("Cell3-Line1\nCell3-Line2\nCell3-Line3"),
			NewCell("Cell4-Line1\nCell4-Line2\nCell4-Line3"),
		).Height(3),
	}
	NewTable(rows, lengths(11, 2)...).Render(catatui.NewRect(0, 0, 23, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Cell1       Cell2      ",
		"Cell3-Line1 Cell4-Line1",
		"Cell3-Line2 Cell4-Line2",
	))
}

func TestTableRenderWithAlignment(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 3))
	rows := []Row{
		NewRow(NewCellFromLine(catatui.LineFromString("Left").Left())),
		NewRow(NewCellFromLine(catatui.LineFromString("Center").Centered())),
		NewRow(NewCellFromLine(catatui.LineFromString("Right").Right())),
	}
	NewTable(rows, catatui.Percentage(100)).Render(catatui.NewRect(0, 0, 10, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Left      ", "  Center  ", "     Right"))
}

func TestTableRenderWithOverflowDoesNotPanic(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 3))
	table := NewTable(nil, catatui.Min(20)).
		Header(NewRow(NewCellFromLine(catatui.LineFromString("").Right()))).
		Footer(NewRow(NewCellFromLine(catatui.LineFromString("").Right())))
	table.Render(catatui.NewRect(0, 0, 20, 3), buf)
}

func TestTableRenderWithSelectedColumnAndIncorrectWidthCountDoesNotPanic(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 3))
	table := NewTable([]Row{NewRowFromStrings("Row1", "Row2", "Row3")}, catatui.Length(10))
	state := NewTableState().WithSelectedColumn(2)
	table.RenderStateful(catatui.NewRect(0, 0, 20, 3), buf, &state)
}

func TestTableRenderWithSelected(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).
		RowHighlightStyle(tableRed).
		HighlightSymbol(">>")
	state := NewTableState().WithSelected(0)
	table.RenderStateful(catatui.NewRect(0, 0, 15, 3), buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(catatui.NewStyledSpan(">>Cell1 Cell2  ", tableRed)),
		catatui.LineFromString("  Cell3 Cell4  "),
		catatui.LineFromString("               "),
	))
}

func TestTableRenderWithSelectedColumn(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2"),
		NewRowFromStrings("Cell3", "Cell4"),
	}
	table := NewTable(rows, lengths(5, 2)...).
		ColumnHighlightStyle(tableBlue).
		HighlightSymbol(">>")
	state := NewTableState().WithSelectedColumn(1)
	table.RenderStateful(catatui.NewRect(0, 0, 15, 3), buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(
			catatui.NewSpan("Cell1"),
			catatui.NewSpan(" "),
			catatui.NewStyledSpan("Cell2", tableBlue),
			catatui.NewSpan("    "),
		),
		catatui.NewLine(
			catatui.NewSpan("Cell3"),
			catatui.NewSpan(" "),
			catatui.NewStyledSpan("Cell4", tableBlue),
			catatui.NewSpan("    "),
		),
		catatui.NewLine(
			catatui.NewSpan("      "),
			catatui.NewStyledSpan("     ", tableBlue),
			catatui.NewSpan("    "),
		),
	))
}

func TestTableRenderWithSelectedCell(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 4))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2", "Cell3"),
		NewRowFromStrings("Cell4", "Cell5", "Cell6"),
		NewRowFromStrings("Cell7", "Cell8", "Cell9"),
	}
	table := NewTable(rows, lengths(5, 3)...).
		HighlightSymbol(">>").
		CellHighlightStyle(tableGreen)
	state := NewTableState().WithSelectedCell(1, 2)
	table.RenderStateful(catatui.NewRect(0, 0, 20, 4), buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(catatui.NewSpan("  Cell1 "), catatui.NewSpan("Cell2 "), catatui.NewSpan("Cell3")),
		catatui.NewLine(catatui.NewSpan(">>Cell4 Cell5 "), catatui.NewStyledSpan("Cell6", tableGreen), catatui.NewSpan(" ")),
		catatui.NewLine(catatui.NewSpan("  Cell7 "), catatui.NewSpan("Cell8 "), catatui.NewSpan("Cell9")),
		catatui.NewLine(catatui.NewSpan("                    ")),
	))
}

func TestTableRenderWithSelectedRowAndColumn(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 4))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2", "Cell3"),
		NewRowFromStrings("Cell4", "Cell5", "Cell6"),
		NewRowFromStrings("Cell7", "Cell8", "Cell9"),
	}
	table := NewTable(rows, lengths(5, 3)...).
		HighlightSymbol(">>").
		RowHighlightStyle(tableRed).
		ColumnHighlightStyle(tableBlue)
	state := NewTableState().WithSelected(1).WithSelectedColumn(2)
	table.RenderStateful(catatui.NewRect(0, 0, 20, 4), buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(catatui.NewSpan("  Cell1 "), catatui.NewSpan("Cell2 "), catatui.NewStyledSpan("Cell3", tableBlue)),
		catatui.NewLine(catatui.NewStyledSpan(">>Cell4 Cell5 ", tableRed), catatui.NewStyledSpan("Cell6", tableBlue), catatui.NewStyledSpan(" ", tableRed)),
		catatui.NewLine(catatui.NewSpan("  Cell7 "), catatui.NewSpan("Cell8 "), catatui.NewStyledSpan("Cell9", tableBlue)),
		catatui.NewLine(catatui.NewSpan("              "), catatui.NewStyledSpan("     ", tableBlue), catatui.NewSpan(" ")),
	))
}

func TestTableRenderWithSelectedRowAndColumnAndCell(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 4))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2", "Cell3"),
		NewRowFromStrings("Cell4", "Cell5", "Cell6"),
		NewRowFromStrings("Cell7", "Cell8", "Cell9"),
	}
	table := NewTable(rows, lengths(5, 3)...).
		HighlightSymbol(">>").
		RowHighlightStyle(tableRed).
		ColumnHighlightStyle(tableBlue).
		CellHighlightStyle(tableGreen)
	state := NewTableState().WithSelected(1).WithSelectedColumn(2)
	table.RenderStateful(catatui.NewRect(0, 0, 20, 4), buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(catatui.NewSpan("  Cell1 "), catatui.NewSpan("Cell2 "), catatui.NewStyledSpan("Cell3", tableBlue)),
		catatui.NewLine(catatui.NewStyledSpan(">>Cell4 Cell5 ", tableRed), catatui.NewStyledSpan("Cell6", tableGreen), catatui.NewStyledSpan(" ", tableRed)),
		catatui.NewLine(catatui.NewSpan("  Cell7 "), catatui.NewSpan("Cell8 "), catatui.NewStyledSpan("Cell9", tableBlue)),
		catatui.NewLine(catatui.NewSpan("              "), catatui.NewStyledSpan("     ", tableBlue), catatui.NewSpan(" ")),
	))
}

// TestTableRenderWithSelectionAndOffset includes a regression test for a bug
// where the table would not render the correct rows when there is no
// selection: https://github.com/ratatui/ratatui/issues/1179
func TestTableRenderWithSelectionAndOffset(t *testing.T) {
	cases := []struct {
		name           string
		selected       int // -1 for no selection
		expectedOffset int
		expectedItems  []string
	}{
		{"no selection", -1, 50, []string{"50", "51", "52", "53", "54"}},
		{"selection before offset", 20, 20, []string{"20", "21", "22", "23", "24"}},
		{"selection immediately before offset", 49, 49, []string{"49", "50", "51", "52", "53"}},
		{"selection at start of offset", 50, 50, []string{"50", "51", "52", "53", "54"}},
		{"selection at end of offset", 54, 50, []string{"50", "51", "52", "53", "54"}},
		{"selection immediately after offset", 55, 51, []string{"51", "52", "53", "54", "55"}},
		{"selection after offset", 80, 76, []string{"76", "77", "78", "79", "80"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Render 100 rows offset at 50, with a selected row.
			rows := make([]Row, 100)
			for i := range rows {
				rows[i] = NewRowFromStrings(strconv.Itoa(i))
			}
			table := NewTable(rows, catatui.Length(2))
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 2, 5))
			state := NewTableState().WithOffset(50)
			if c.selected >= 0 {
				state = state.WithSelected(c.selected)
			}

			table.RenderStateful(catatui.NewRect(0, 0, 5, 5), buf, &state)

			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.expectedItems...))
			if state.offset != c.expectedOffset {
				t.Errorf("offset = %d, want %d", state.offset, c.expectedOffset)
			}
		})
	}
}

// --- column widths ----------------------------------------------------------------

func rect(x, width uint16) catatui.Rect { return catatui.NewRect(x, 0, width, 1) }

func TestTableLengthConstraint(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Length(4), catatui.Length(4))

	// without selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 0, 0), []catatui.Rect{rect(0, 4), rect(5, 4)})

	// with selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 3, 0), []catatui.Rect{rect(3, 4), rect(8, 4)})

	// without selection, less than needed width
	assertRects(t, table.getColumnWidths(7, 0, 0), []catatui.Rect{rect(0, 3), rect(4, 3)})

	// with selection, less than needed width
	// <--------7px-------->
	// ┌────────┐x┌────────┐
	// │ (3, 2) │x│ (6, 1) │
	// └────────┘x└────────┘
	// column spacing (i.e. `x`) is always prioritized
	assertRects(t, table.getColumnWidths(7, 3, 0), []catatui.Rect{rect(3, 2), rect(6, 1)})
}

func TestTableMaxConstraint(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Max(4), catatui.Max(4))
	assertRects(t, table.getColumnWidths(20, 0, 0), []catatui.Rect{rect(0, 4), rect(5, 4)})
	assertRects(t, table.getColumnWidths(20, 3, 0), []catatui.Rect{rect(3, 4), rect(8, 4)})
	assertRects(t, table.getColumnWidths(7, 0, 0), []catatui.Rect{rect(0, 3), rect(4, 3)})
	assertRects(t, table.getColumnWidths(7, 3, 0), []catatui.Rect{rect(3, 2), rect(6, 1)})
}

func TestTableMinConstraint(t *testing.T) {
	// In its current stage, the "Min" constraint does not grow to use the
	// possible available length and enabling "expand_to_fill" will just
	// stretch the last constraint and not split it with all available
	// constraints.
	table := NewTable(nil).Widths(catatui.Min(4), catatui.Min(4))

	// without selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 0, 0), []catatui.Rect{rect(0, 10), rect(11, 9)})

	// with selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 3, 0), []catatui.Rect{rect(3, 8), rect(12, 8)})

	// without selection, less than needed width: allocates spacer
	assertRects(t, table.getColumnWidths(7, 0, 0), []catatui.Rect{rect(0, 3), rect(4, 3)})

	// with selection, less than needed width: always allocates selection and spacer
	assertRects(t, table.getColumnWidths(7, 3, 0), []catatui.Rect{rect(3, 2), rect(6, 1)})
}

func TestTablePercentageConstraint(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Percentage(30), catatui.Percentage(30))

	// without selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 0, 0), []catatui.Rect{rect(0, 6), rect(7, 6)})

	// with selection, more than needed width
	assertRects(t, table.getColumnWidths(20, 3, 0), []catatui.Rect{rect(3, 5), rect(9, 5)})

	// without selection, less than needed width
	// rounds from positions: [0.0, 0.0, 2.1, 3.1, 5.2, 7.0]
	assertRects(t, table.getColumnWidths(7, 0, 0), []catatui.Rect{rect(0, 2), rect(3, 2)})

	// with selection, less than needed width
	// rounds from positions: [0.0, 3.0, 5.1, 6.1, 7.0, 7.0]
	assertRects(t, table.getColumnWidths(7, 3, 0), []catatui.Rect{rect(3, 1), rect(5, 1)})
}

func TestTableRatioConstraint(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Ratio(1, 3), catatui.Ratio(1, 3))

	// without selection, more than needed width
	// rounds from positions: [0.00, 0.00, 6.67, 7.67, 14.33]
	assertRects(t, table.getColumnWidths(20, 0, 0), []catatui.Rect{rect(0, 7), rect(8, 6)})

	// with selection, more than needed width
	// rounds from positions: [0.00, 3.00, 10.67, 17.33, 20.00]
	assertRects(t, table.getColumnWidths(20, 3, 0), []catatui.Rect{rect(3, 6), rect(10, 5)})

	// without selection, less than needed width
	// rounds from positions: [0.00, 2.33, 3.33, 5.66, 7.00]
	assertRects(t, table.getColumnWidths(7, 0, 0), []catatui.Rect{rect(0, 2), rect(3, 3)})

	// with selection, less than needed width
	// rounds from positions: [0.00, 3.00, 5.33, 6.33, 7.00, 7.00]
	assertRects(t, table.getColumnWidths(7, 3, 0), []catatui.Rect{rect(3, 1), rect(5, 2)})
}

// TestTableUnderconstrainedFlex: when more width is available than requested,
// the behavior is controlled by flex.
func TestTableUnderconstrainedFlex(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Min(10), catatui.Min(10), catatui.Min(1))
	assertRects(t, table.getColumnWidths(62, 0, 0), []catatui.Rect{rect(0, 20), rect(21, 20), rect(42, 20)})

	table = table.FlexMode(catatui.FlexLegacy)
	assertRects(t, table.getColumnWidths(62, 0, 0), []catatui.Rect{rect(0, 10), rect(11, 10), rect(22, 40)})

	table = table.FlexMode(catatui.FlexSpaceBetween)
	assertRects(t, table.getColumnWidths(62, 0, 0), []catatui.Rect{rect(0, 20), rect(21, 20), rect(42, 20)})
}

func TestTableUnderconstrainedSegmentSize(t *testing.T) {
	table := NewTable(nil).Widths(catatui.Min(10), catatui.Min(10), catatui.Min(1))
	assertRects(t, table.getColumnWidths(62, 0, 0), []catatui.Rect{rect(0, 20), rect(21, 20), rect(42, 20)})

	table = table.FlexMode(catatui.FlexLegacy)
	assertRects(t, table.getColumnWidths(62, 0, 0), []catatui.Rect{rect(0, 10), rect(11, 10), rect(22, 40)})
}

func TestTableNoConstraintWithRows(t *testing.T) {
	table := NewTable(nil).
		Rows(NewRowFromStrings("a", "b"), NewRowFromStrings("c", "d", "e")).
		// rows should get precedence over header
		Header(NewRowFromStrings("f", "g")).
		Footer(NewRowFromStrings("h", "i")).
		ColumnSpacing(0)
	assertRects(t, table.getColumnWidths(30, 0, 3), []catatui.Rect{rect(0, 10), rect(10, 10), rect(20, 10)})
}

func TestTableNoConstraintWithHeader(t *testing.T) {
	table := NewTable(nil).Rows().Header(NewRowFromStrings("f", "g")).ColumnSpacing(0)
	assertRects(t, table.getColumnWidths(10, 0, 2), []catatui.Rect{rect(0, 5), rect(5, 5)})
}

func TestTableNoConstraintWithFooter(t *testing.T) {
	table := NewTable(nil).Rows().Footer(NewRowFromStrings("h", "i")).ColumnSpacing(0)
	assertRects(t, table.getColumnWidths(10, 0, 2), []catatui.Rect{rect(0, 5), rect(5, 5)})
}

// tableWithSelection renders one row "ABCDE" "12345" with the given highlight
// spacing, width, column spacing and selection (-1 for none), and checks the
// result. It is ratatui's test_table_with_selection.
func tableWithSelection(t *testing.T, spacing HighlightSpacing, columns, columnSpacing uint16, selection int, expected ...string) {
	t.Helper()
	table := NewTable(nil).
		Rows(NewRowFromStrings("ABCDE", "12345")).
		HighlightSpacing(spacing).
		HighlightSymbol(">>>").
		ColumnSpacing(columnSpacing)
	area := catatui.NewRect(0, 0, columns, 3)
	buf := catatui.NewBuffer(area)
	var state TableState
	if selection >= 0 {
		state = state.WithSelected(selection)
	}
	table.RenderStateful(area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(expected...))
}

func TestTableExcessAreaHighlightSymbolAndColumnSpacingAllocation(t *testing.T) {
	// no highlight_symbol rendered ever
	tableWithSelection(t, HighlightSpacingNever, 15, 0, -1,
		// default layout is Flex::Start but columns length constraints are
		// calculated as `max_area / n_columns`, i.e. they are distributed
		// amongst available space
		"ABCDE  12345   ",
		"               ",
		"               ",
	)

	// As reference, this is what happens when you manually specify widths.
	table := NewTable(nil).
		Rows(NewRowFromStrings("ABCDE", "12345")).
		Widths(catatui.Length(5), catatui.Length(5)).
		ColumnSpacing(0)
	area := catatui.NewRect(0, 0, 15, 3)
	buf := catatui.NewBuffer(area)
	table.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"ABCDE12345     ",
		"               ",
		"               ",
	))

	// no highlight_symbol rendered ever
	tableWithSelection(t, HighlightSpacingNever, 15, 0, 0,
		"ABCDE  12345   ",
		"               ",
		"               ",
	)

	// no highlight_symbol rendered because no selection is made
	tableWithSelection(t, HighlightSpacingWhenSelected, 15, 0, -1,
		"ABCDE  12345   ",
		"               ",
		"               ",
	)
	// highlight_symbol rendered because selection is made
	tableWithSelection(t, HighlightSpacingWhenSelected, 15, 0, 0,
		">>>ABCDE 12345 ",
		"               ",
		"               ",
	)

	// highlight_symbol always rendered even no selection is made
	tableWithSelection(t, HighlightSpacingAlways, 15, 0, -1,
		"   ABCDE 12345 ",
		"               ",
		"               ",
	)

	tableWithSelection(t, HighlightSpacingAlways, 15, 0, 0,
		">>>ABCDE 12345 ",
		"               ",
		"               ",
	)
}

func TestTableInsufficientAreaHighlightSymbolAndColumnSpacingAllocation(t *testing.T) {
	// column spacing is prioritized over every other constraint
	tableWithSelection(t, HighlightSpacingNever, 10, 1, -1,
		"ABCDE 1234", // spacing is prioritized and column is cut
		"          ",
		"          ",
	)
	tableWithSelection(t, HighlightSpacingWhenSelected, 10, 1, -1,
		"ABCDE 1234", // spacing is prioritized and column is cut
		"          ",
		"          ",
	)

	// This checks that space for highlight_symbol is always allocated, and
	// that space for the column spacing is allocated too. The highlight
	// symbol space is split off first, then the column widths are computed
	// in the remainder, where the spacing takes priority and the last column
	// ends up just 1 wide.
	tableWithSelection(t, HighlightSpacingAlways, 10, 1, -1,
		"   ABC 123", // highlight_symbol and spacing are prioritized
		"          ",
		"          ",
	)

	// the following are specification tests
	tableWithSelection(t, HighlightSpacingAlways, 9, 1, -1,
		"   ABC 12",
		"         ",
		"         ",
	)
	tableWithSelection(t, HighlightSpacingAlways, 8, 1, -1,
		"   AB 12",
		"        ",
		"        ",
	)
	tableWithSelection(t, HighlightSpacingAlways, 7, 1, -1,
		"   AB 1",
		"       ",
		"       ",
	)

	// highlight_symbol and spacing are prioritized but columns are evenly
	// distributed
	table := NewTable(nil).
		Rows(NewRowFromStrings("ABCDE", "12345")).
		HighlightSpacing(HighlightSpacingAlways).
		FlexMode(catatui.FlexLegacy).
		HighlightSymbol(">>>").
		ColumnSpacing(1)
	area := catatui.NewRect(0, 0, 10, 3)
	buf := catatui.NewBuffer(area)
	table.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   ABCDE 1",
		"          ",
		"          ",
	))

	table = NewTable(nil).
		Rows(NewRowFromStrings("ABCDE", "12345")).
		HighlightSpacing(HighlightSpacingAlways).
		FlexMode(catatui.FlexStart).
		HighlightSymbol(">>>").
		ColumnSpacing(1)
	buf = catatui.NewBuffer(area)
	table.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   ABC 123",
		"          ",
		"          ",
	))

	tableWithSelection(t, HighlightSpacingNever, 10, 1, 0,
		"ABCDE 1234", // spacing is prioritized
		"          ",
		"          ",
	)

	tableWithSelection(t, HighlightSpacingWhenSelected, 10, 1, 0,
		">>>ABC 123",
		"          ",
		"          ",
	)

	tableWithSelection(t, HighlightSpacingAlways, 10, 1, 0,
		">>>ABC 123", // highlight column and spacing are prioritized
		"          ",
		"          ",
	)
}

func TestTableInsufficientAreaHighlightSymbolAllocationWithNoColumnSpacing(t *testing.T) {
	tableWithSelection(t, HighlightSpacingNever, 10, 0, -1,
		"ABCDE12345",
		"          ",
		"          ",
	)
	tableWithSelection(t, HighlightSpacingWhenSelected, 10, 0, -1,
		"ABCDE12345",
		"          ",
		"          ",
	)
	// highlight symbol spacing is prioritized over all constraints, even if
	// the constraints are fixed length, because the highlight_symbol column
	// is separated before any of the constraint widths are calculated
	tableWithSelection(t, HighlightSpacingAlways, 10, 0, -1,
		"   ABCD123", // highlight column and spacing are prioritized
		"          ",
		"          ",
	)
	tableWithSelection(t, HighlightSpacingNever, 10, 0, 0,
		"ABCDE12345",
		"          ",
		"          ",
	)
	tableWithSelection(t, HighlightSpacingWhenSelected, 10, 0, 0,
		">>>ABCD123", // highlight column and spacing are prioritized
		"          ",
		"          ",
	)
	tableWithSelection(t, HighlightSpacingAlways, 10, 0, 0,
		">>>ABCD123", // highlight column and spacing are prioritized
		"          ",
		"          ",
	)
}

// --- column count, minimal buffers, cell areas -----------------------------------

func TestTableColumnCount(t *testing.T) {
	cases := []struct {
		name     string
		header   []string
		rows     [][]string
		footer   []string
		expected int
	}{
		{"no columns", nil, nil, nil, 0},
		{"only header", []string{"H1", "H2"}, nil, nil, 2},
		{"only rows", nil, [][]string{{"C1", "C2"}, {"C1", "C2", "C3"}}, nil, 3},
		{"only footer", nil, nil, []string{"F1", "F2", "F3", "F4"}, 4},
		{"header longer", []string{"H1", "H2", "H3", "H4"}, [][]string{{"C1", "C2"}, {"C1", "C2", "C3"}}, []string{"F1", "F2"}, 4},
		{"rows longer", []string{"H1", "H2"}, [][]string{{"C1", "C2"}, {"C1", "C2", "C3", "C4"}}, []string{"F1", "F2"}, 4},
		{"footer longer", []string{"H1", "H2"}, [][]string{{"C1", "C2"}, {"C1", "C2", "C3"}}, []string{"F1", "F2", "F3", "F4"}, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rows := make([]Row, len(c.rows))
			for i, r := range c.rows {
				rows[i] = NewRowFromStrings(r...)
			}
			table := NewTable(rows).
				Header(NewRowFromStrings(c.header...)).
				Footer(NewRowFromStrings(c.footer...))
			if got := table.columnCount(); got != c.expected {
				t.Errorf("columnCount() = %d, want %d", got, c.expected)
			}
		})
	}
}

func TestTableRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2", "Cell3"),
		NewRowFromStrings("Cell4", "Cell5", "Cell6"),
	}
	table := NewTable(rows, lengths(10, 3)...).
		Header(NewRowFromStrings("Header1", "Header2", "Header3")).
		Footer(NewRowFromStrings("Footer1", "Footer2", "Footer3"))
	// This should not panic, even if the buffer is too small to render the table.
	table.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestTableRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	rows := []Row{
		NewRowFromStrings("Cell1", "Cell2", "Cell3"),
		NewRowFromStrings("Cell4", "Cell5", "Cell6"),
	}
	table := NewTable(rows, lengths(10, 3)...).
		Header(NewRowFromStrings("Header1", "Header2", "Header3")).
		Footer(NewRowFromStrings("Footer1", "Footer2", "Footer3"))
	// This should not panic, even if the buffer has zero size.
	table.Render(buf.Area, buf)
}

func TestTableGetAreaForColumnSpanOneNoMoreColumns(t *testing.T) {
	if _, _, ok := getCellArea(nil, 1, 1); ok {
		t.Error("getCellArea returned an area with no columns left")
	}
}

func TestTableGetAreaForColumnSpanTwoNoMoreColumns(t *testing.T) {
	if _, _, ok := getCellArea(nil, 2, 1); ok {
		t.Error("getCellArea returned an area with no columns left")
	}
}

func colspanColumns(n int) []catatui.Rect {
	out := make([]catatui.Rect, n)
	for i := range out {
		out[i] = catatui.Rect{X: 3, Y: 0, Width: 2, Height: 1}
	}
	return out
}

func TestTableColspanWidthSingleColumnSpacing(t *testing.T) {
	cases := []struct {
		columns    int
		columnSpan uint16
		width      uint16
	}{
		{3, 2, 5},
		{2, 2, 5},
		{2, 1, 2},
		{2, 3, 5},
		{1, 1, 2},
		{1, 2, 2},
		{4, 3, 8},
		{3, 3, 8},
	}
	for _, c := range cases {
		area, _, ok := getCellArea(colspanColumns(c.columns), c.columnSpan, 1)
		if !ok {
			t.Errorf("columns=%d span=%d: no area", c.columns, c.columnSpan)
			continue
		}
		if area.Width != c.width {
			t.Errorf("columns=%d span=%d: width = %d, want %d", c.columns, c.columnSpan, area.Width, c.width)
		}
	}
}

func TestTableColspanWidthTwoColumnSpacing(t *testing.T) {
	cases := []struct {
		columns    int
		columnSpan uint16
		width      uint16
	}{
		{3, 3, 10},
		{1, 3, 2},
	}
	for _, c := range cases {
		area, _, ok := getCellArea(colspanColumns(c.columns), c.columnSpan, 2)
		if !ok {
			t.Errorf("columns=%d span=%d: no area", c.columns, c.columnSpan)
			continue
		}
		if area.Width != c.width {
			t.Errorf("columns=%d span=%d: width = %d, want %d", c.columns, c.columnSpan, area.Width, c.width)
		}
	}
}

func TestTableWithSelectionAndColumnSpans(t *testing.T) {
	cells := func() []Cell {
		return []Cell{
			NewCell("ABCDEFGHIJK").ColumnSpan(2),
			NewCell("12345678901"),
			NewCell("XYZXYZXYZXY"),
		}
	}
	cases := []struct {
		name      string
		spacing   HighlightSpacing
		selection int // -1 for none
		expected  []string
	}{
		{"always, no selection", HighlightSpacingAlways, -1, []string{"   ABCDEFGH 123", "               ", "               "}},
		{"always, selected", HighlightSpacingAlways, 0, []string{">>>ABCDEFGH 123", "               ", "               "}},
		{"when selected, no selection", HighlightSpacingWhenSelected, -1, []string{"ABCDEFGHIJ 1234", "               ", "               "}},
		{"when selected, selected", HighlightSpacingWhenSelected, 0, []string{">>>ABCDEFGH 123", "               ", "               "}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			table := NewTable(nil).
				Rows(NewRow(cells()...)).
				HighlightSpacing(c.spacing).
				HighlightSymbol(">>>").
				ColumnSpacing(1)
			area := catatui.NewRect(0, 0, 15, 3)
			buf := catatui.NewBuffer(area)
			var state TableState
			if c.selection >= 0 {
				state = state.WithSelected(c.selection)
			}
			table.RenderStateful(area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.expected...))
		})
	}
}
