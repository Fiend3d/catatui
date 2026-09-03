// Command table is a table of people with a scrollbar, a moving selection and
// a colour scheme that can be changed while it runs.
//
//	go run ./examples/apps/table
//
// j/k or up and down move between rows, h/l or left and right between columns,
// Shift with either (or H/L) changes the colour scheme, q quits. The selected
// row, column and cell are each styled separately, which is what makes the one
// cell where they cross stand out.
//
// Port of examples/apps/table @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// itemHeight is how many rows of the terminal one person takes: a blank line,
// the text, a blank line, and a gap.
const itemHeight = 4

// palettes are the schemes Shift and an arrow move between.
var palettes = []tailwind.Palette{
	tailwind.Blue,
	tailwind.Emerald,
	tailwind.Indigo,
	tailwind.Red,
}

// infoText is the footer, which says what the keys do.
var infoText = []string{
	"(Esc) quit | (↑) move up | (↓) move down | (←) move left | (→) move right",
	"(Shift + →) next color | (Shift + ←) previous color",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := newApp()

	for !a.quit {
		if err := terminal.Draw(a.render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		a.handle(ev)
	}
	return nil
}

// person is one row of the table. The address runs to two lines, which is why
// every row is four high.
type person struct {
	name    string
	address string
	email   string
}

// fields is the person's columns in the order the table shows them.
func (p person) fields() [3]string { return [3]string{p.name, p.address, p.email} }

// app is the table's state, the data, and which colour scheme is in use.
type app struct {
	state           widgets.TableState
	scrollState     widgets.ScrollbarState
	items           []person
	longestItemLens [3]uint16
	colorIndex      int
	quit            bool
}

func newApp() *app {
	return &app{
		state:           widgets.NewTableState().WithSelected(0),
		scrollState:     widgets.NewScrollbarState((len(people) - 1) * itemHeight),
		items:           people,
		longestItemLens: constraintLens(people),
	}
}

// colors is the scheme currently in use.
func (a *app) colors() tableColors { return newTableColors(palettes[a.colorIndex]) }

// handle applies one event.
//
// Shift with an arrow changes the colour scheme rather than moving the
// selection. A shifted letter arrives as its capital, so H and L do the same,
// which is what a terminal actually sends where ratatui's example asks for
// shift-plus-lowercase.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	shift := ev.Mods.Contains(term.ModShift)

	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.nextRow()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.previousRow()
	case ev.Rune == 'L', ev.Key == term.KeyRight && shift:
		a.colorIndex = (a.colorIndex + 1) % len(palettes)
	case ev.Rune == 'H', ev.Key == term.KeyLeft && shift:
		a.colorIndex = (a.colorIndex + len(palettes) - 1) % len(palettes)
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.state.SelectNextColumn()
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.state.SelectPreviousColumn()
	}
}

// nextRow moves down a row, round to the first from the last, and takes the
// scrollbar with it.
func (a *app) nextRow() {
	i, ok := a.state.Selected()
	switch {
	case !ok:
		i = 0
	case i >= len(a.items)-1:
		i = 0
	default:
		i++
	}
	a.selectRow(i)
}

// previousRow moves up a row, round to the last from the first.
func (a *app) previousRow() {
	i, ok := a.state.Selected()
	switch {
	case !ok:
		i = 0
	case i == 0:
		i = len(a.items) - 1
	default:
		i--
	}
	a.selectRow(i)
}

// selectRow moves the selection and the scrollbar together: the scrollbar is
// measured in terminal rows, so the row index is scaled by the row height.
func (a *app) selectRow(i int) {
	a.state.Select(i)
	a.scrollState = a.scrollState.Position(i * itemHeight)
}

// render draws the table with its scrollbar, and the footer under it.
func (a *app) render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Min(5), catatui.Length(4)).Split(f.Area())

	a.renderTable(f, rows[0])
	a.renderScrollbar(f, rows[0])
	a.renderFooter(f, rows[1])
}

func (a *app) renderTable(f *catatui.Frame, area catatui.Rect) {
	colors := a.colors()

	header := widgets.NewRowFromStrings("Name", "Address", "Email").
		Style(catatui.NewStyle().Fg(colors.headerFg).Bg(colors.headerBg)).
		Height(1)

	rows := make([]widgets.Row, len(a.items))
	for i, item := range a.items {
		// The rows alternate backgrounds, which is what lets the eye follow one
		// across a wide table.
		background := colors.normalRowColor
		if i%2 == 1 {
			background = colors.altRowColor
		}
		rows[i] = a.row(i, item).
			Style(catatui.NewStyle().Fg(colors.rowFg).Bg(background)).
			Height(4)
	}

	// The highlight symbol is a column of its own, blank on the row's first
	// and last lines so the bar sits beside the text rather than the padding.
	const bar = " █ "
	symbol := catatui.NewText(
		catatui.LineFromString(""),
		catatui.LineFromString(bar),
		catatui.LineFromString(bar),
		catatui.LineFromString(""),
	)

	table := widgets.NewTable(rows,
		// One column wider than the longest entry, for padding.
		catatui.Length(a.longestItemLens[0]+1),
		catatui.Min(a.longestItemLens[1]+1),
		catatui.Min(a.longestItemLens[2]),
	).
		Header(header).
		RowHighlightStyle(catatui.NewStyle().
			AddModifier(catatui.ModifierReversed).
			Fg(colors.selectedRowStyleFg)).
		ColumnHighlightStyle(catatui.NewStyle().Fg(colors.selectedColumnStyleFg)).
		CellHighlightStyle(catatui.NewStyle().
			AddModifier(catatui.ModifierReversed).
			Fg(colors.selectedCellStyleFg)).
		HighlightSymbolText(symbol).
		Style(catatui.NewStyle().Bg(colors.bufferBg)).
		HighlightSpacing(widgets.HighlightSpacingAlways)

	catatui.RenderStatefulWidgetOn(f, table, area, &a.state)
}

// row is one person's cells, blank-padded top and bottom so the text sits in
// the middle of the four rows it is given.
//
// One row is missing its details, and says so across both of the columns it
// would have filled: that is what a cell spanning two columns is for.
func (a *app) row(i int, item person) widgets.Row {
	if i == 3 {
		return widgets.NewRow(
			widgets.NewCell("\n"+item.name+"\n"),
			widgets.NewCell("\n[no address or email address is available for this person]\n").
				ColumnSpan(2),
		)
	}
	fields := item.fields()
	cells := make([]widgets.Cell, len(fields))
	for j, content := range fields {
		cells[j] = widgets.NewCell("\n" + content + "\n")
	}
	return widgets.NewRow(cells...)
}

func (a *app) renderScrollbar(f *catatui.Frame, area catatui.Rect) {
	catatui.RenderStatefulWidgetOn(f,
		widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
			BeginSymbolNone().
			EndSymbolNone(),
		area.Inner(catatui.Margin{Horizontal: 1, Vertical: 1}),
		&a.scrollState)
}

func (a *app) renderFooter(f *catatui.Frame, area catatui.Rect) {
	colors := a.colors()

	lines := make([]catatui.Line, len(infoText))
	for i, s := range infoText {
		lines[i] = catatui.LineFromString(s)
	}

	f.RenderWidget(
		widgets.NewParagraphFromText(catatui.NewText(lines...)).
			Style(catatui.NewStyle().Fg(colors.rowFg).Bg(colors.bufferBg)).
			Centered().
			Block(widgets.Bordered().
				BorderType(widgets.BorderDouble).
				BorderStyle(catatui.NewStyle().Fg(colors.footerBorderColor))),
		area)
}

// tableColors is one scheme, mixed from a tailwind ramp and the slate one.
type tableColors struct {
	bufferBg              catatui.Color
	headerBg              catatui.Color
	headerFg              catatui.Color
	rowFg                 catatui.Color
	selectedRowStyleFg    catatui.Color
	selectedColumnStyleFg catatui.Color
	selectedCellStyleFg   catatui.Color
	normalRowColor        catatui.Color
	altRowColor           catatui.Color
	footerBorderColor     catatui.Color
}

func newTableColors(palette tailwind.Palette) tableColors {
	return tableColors{
		bufferBg:              tailwind.Slate.C950,
		headerBg:              palette.C900,
		headerFg:              tailwind.Slate.C200,
		rowFg:                 tailwind.Slate.C200,
		selectedRowStyleFg:    palette.C400,
		selectedColumnStyleFg: palette.C400,
		selectedCellStyleFg:   palette.C600,
		normalRowColor:        tailwind.Slate.C950,
		altRowColor:           tailwind.Slate.C900,
		footerBorderColor:     palette.C400,
	}
}

// constraintLens is the widest entry in each column, which is what the column
// widths are built from. The address is two lines, so each is measured on its
// own.
func constraintLens(items []person) [3]uint16 {
	var lens [3]uint16
	for _, item := range items {
		lens[0] = max(lens[0], uint16(catatui.StringWidth(item.name)))
		for _, line := range strings.Split(item.address, "\n") {
			lens[1] = max(lens[1], uint16(catatui.StringWidth(line)))
		}
		lens[2] = max(lens[2], uint16(catatui.StringWidth(item.email)))
	}
	return lens
}

// people is the table's contents, made up and sorted by name.
//
// ratatui generates them with the fakeit crate; hard-coding them keeps the
// example's only dependency the library itself, and means it draws the same
// thing twice running.
var people = []person{
	{"Ada Lovelace", "12 Analytical Way\nLondon, ENG SW1A 1AA", "ada.lovelace@example.com"},
	{"Alan Turing", "5 Bletchley Road\nMilton Keynes, ENG MK3 6EB", "alan.turing@example.com"},
	{"Barbara Liskov", "77 Substitution Street\nCambridge, MA 02139", "barbara.liskov@example.com"},
	{"Claude Shannon", "1 Information Lane\nMurray Hill, NJ 07974", "claude.shannon@example.com"},
	{"Donald Knuth", "3 Literate Avenue\nStanford, CA 94305", "donald.knuth@example.com"},
	{"Edsger Dijkstra", "9 Shortest Path\nNuenen, NB 5671", "edsger.dijkstra@example.com"},
	{"Frances Allen", "44 Optimising Drive", "frances.allen@example.com"},
	{"Grace Hopper", "8 Compiler Court\nArlington, VA 22204", "grace.hopper@example.com"},
	{"Hedy Lamarr", "16 Frequency Hop\nVienna, W 1010", "hedy.lamarr@example.com"},
	{"Jean Bartik", "2 ENIAC Place\nPhiladelphia, PA 19104", "jean.bartik@example.com"},
	{"John McCarthy", "6 Recursion Road\nStanford, CA 94305", "john.mccarthy@example.com"},
	{"Katherine Johnson", "23 Orbit Circle\nHampton, VA 23666", "katherine.johnson@example.com"},
	{"Ken Thompson", "11 Pipe Street\nMurray Hill, NJ 07974", "ken.thompson@example.com"},
	{"Margaret Hamilton", "1 Apollo Terrace\nCambridge, MA 02142", "margaret.hamilton@example.com"},
	{"Radia Perlman", "4 Spanning Tree\nRedmond, WA 98052", "radia.perlman@example.com"},
	{"Rob Pike", "7 Goroutine Grove\nSydney, NSW 2000", "rob.pike@example.com"},
	{"Robert Kahn", "10 Datagram Drive\nReston, VA 20190", "robert.kahn@example.com"},
	{"Rosalind Franklin", "51 Diffraction Row\nLondon, ENG WC2R 2LS", "rosalind.franklin@example.com"},
	{"Sophie Wilson", "32 Reduced Instruction\nCambridge, ENG CB1 2JD", "sophie.wilson@example.com"},
	{"Vint Cerf", "10 Protocol Parade\nReston, VA 20190", "vint.cerf@example.com"},
}
