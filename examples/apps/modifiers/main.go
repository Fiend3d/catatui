// Command modifiers draws every modifier against every pairing of five
// foreground and background colours, so you can see which of them your
// terminal actually renders.
//
//	go run ./examples/apps/modifiers
//
// Any key quits.
//
// Port of examples/apps/modifiers @ ratatui-v0.30.2
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

// colors are the five the grid pairs off against each other, foreground
// against background.
var colors = []catatui.Color{
	catatui.ColorBlack,
	catatui.ColorDarkGray,
	catatui.ColorGray,
	catatui.ColorWhite,
	catatui.ColorRed,
}

// allModifiers is every modifier there is, with the empty set first. Modifier
// is a bit set, so this is its nine flags one at a time rather than the 512
// combinations of them.
var allModifiers = []catatui.Modifier{
	catatui.ModifierNone,
	catatui.ModifierBold,
	catatui.ModifierDim,
	catatui.ModifierItalic,
	catatui.ModifierUnderlined,
	catatui.ModifierSlowBlink,
	catatui.ModifierRapidBlink,
	catatui.ModifierReversed,
	catatui.ModifierHidden,
	catatui.ModifierCrossedOut,
}

// render draws the warning and then the grid: five backgrounds by five
// foregrounds by ten modifiers is 250 cells, laid out fifty rows of five.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Min(0)).Split(f.Area())
	textArea, mainArea := rows[0], rows[1]

	f.RenderWidget(
		widgets.NewParagraph("Note: not all terminals support all modifiers").
			Style(catatui.NewStyle().
				Fg(catatui.ColorRed).
				AddModifier(catatui.ModifierBold)),
		textArea)

	for i, area := range gridCells(mainArea) {
		// The grid runs background slowest, then foreground, then modifier,
		// so each block of ten rows shares a background.
		bg := colors[i/(len(colors)*len(allModifiers))]
		fg := colors[i/len(allModifiers)%len(colors)]
		modifier := allModifiers[i%len(allModifiers)]

		// The name is padded out to the width of the longest of them, so that
		// the colours make solid blocks rather than ragged ones.
		style := catatui.NewStyle().Fg(fg).Bg(bg).AddModifier(modifier)
		label := fmt.Sprintf("%-12s", modifier)
		f.RenderWidget(
			widgets.NewParagraphFromText(
				catatui.NewText(catatui.LineFromStyledString(label, style))),
			area)
	}
}

// gridCells cuts the area into the 250 single-row cells the grid needs, in
// reading order. A row that does not fit comes back empty rather than missing,
// which is what keeps the indexing above simple on a small terminal.
func gridCells(area catatui.Rect) []catatui.Rect {
	const (
		gridRows    = 50
		gridColumns = 5
	)
	rowConstraints := make([]catatui.Constraint, gridRows)
	for i := range rowConstraints {
		rowConstraints[i] = catatui.Length(1)
	}
	columnConstraints := make([]catatui.Constraint, gridColumns)
	for i := range columnConstraints {
		columnConstraints[i] = catatui.Percentage(20)
	}

	cells := make([]catatui.Rect, 0, gridRows*gridColumns)
	for _, row := range catatui.VerticalLayout(rowConstraints...).Split(area) {
		cells = append(cells, catatui.HorizontalLayout(columnConstraints...).Split(row)...)
	}
	return cells
}
