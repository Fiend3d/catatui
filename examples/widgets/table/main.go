// Command table shows catatui's Table widget with row, column and cell
// selection.
//
//	go run ./examples/widgets/table
//
// Press h/j/k/l or the arrow keys to move, g/G for first/last row, q to quit.
//
// Port of ratatui-widgets/examples/table.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run keeps the table state across frames.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	var state widgets.TableState
	state.SelectFirst()
	state.SelectFirstColumn()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, &state) }); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind != term.EventKey {
			continue
		}
		switch {
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return nil
		case ev.IsRune('j'), ev.IsKey(term.KeyDown):
			state.SelectNext()
		case ev.IsRune('k'), ev.IsKey(term.KeyUp):
			state.SelectPrevious()
		case ev.IsRune('l'), ev.IsKey(term.KeyRight):
			state.SelectNextColumn()
		case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
			state.SelectPreviousColumn()
		case ev.IsRune('g'):
			state.SelectFirst()
		case ev.IsRune('G'):
			state.SelectLast()
		}
	}
}

// render draws a title and the table below it.
func render(f *catatui.Frame, state *widgets.TableState) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("Table Widget",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press 'q' to quit and arrow keys to navigate)"),
	).Centered(), rows[0])

	renderTable(f, rows[1], state)
}

// renderTable draws a recipe as a table with a header, a footer and separate
// highlight styles for the selected row, column and cell.
func renderTable(f *catatui.Frame, area catatui.Rect, state *widgets.TableState) {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)

	header := widgets.NewRowFromStrings("Ingredient", "Quantity", "Macros").
		Style(bold).
		BottomMargin(1)

	rows := []widgets.Row{
		widgets.NewRowFromStrings("Eggplant", "1 medium", "25 kcal, 6g carbs, 1g protein"),
		widgets.NewRowFromStrings("Tomato", "2 large", "44 kcal, 10g carbs, 2g protein"),
		widgets.NewRowFromStrings("Zucchini", "1 medium", "33 kcal, 7g carbs, 2g protein"),
		widgets.NewRowFromStrings("Bell Pepper", "1 medium", "24 kcal, 6g carbs, 1g protein"),
		widgets.NewRowFromStrings("Garlic", "2 cloves", "9 kcal, 2g carbs, 0.4g protein"),
	}

	footer := widgets.NewRowFromStrings(
		"Ratatouille Recipe", "", "135 kcal, 31g carbs, 6.4g protein",
	).Style(catatui.NewStyle().AddModifier(catatui.ModifierItalic))

	table := widgets.NewTable(rows,
		catatui.Percentage(30), catatui.Percentage(20), catatui.Percentage(50)).
		Header(header).
		Footer(footer).
		ColumnSpacing(1).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		RowHighlightStyle(catatui.NewStyle().
			Bg(catatui.ColorBlack).
			AddModifier(catatui.ModifierBold)).
		ColumnHighlightStyle(catatui.NewStyle().Fg(catatui.ColorGray)).
		CellHighlightStyle(catatui.NewStyle().
			Fg(catatui.ColorYellow).
			AddModifier(catatui.ModifierReversed)).
		HighlightSymbol("🍴 ")

	catatui.RenderStatefulWidget(table, area, f.Buffer(), state)
}
