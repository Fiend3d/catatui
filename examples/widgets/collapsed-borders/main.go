// Command collapsed-borders shows four blocks sharing their borders, with the
// selected one drawn on top in a thick border.
//
//	go run ./examples/widgets/collapsed-borders
//
// Press the arrow keys to select a pane, q to quit.
//
// Port of ratatui-widgets/examples/collapsed-borders.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// pane names one of the four blocks.
type pane int

const (
	paneTop pane = iota
	paneLeft
	paneRight
	paneBottom
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run keeps the selected pane across frames.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	selected := paneTop

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, selected) }); err != nil {
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
		case ev.IsKey(term.KeyUp):
			selected = paneTop
		case ev.IsKey(term.KeyLeft):
			selected = paneLeft
		case ev.IsKey(term.KeyRight):
			selected = paneRight
		case ev.IsKey(term.KeyDown):
			selected = paneBottom
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return nil
		}
	}
}

// render draws the title and the panes below it.
func render(f *catatui.Frame, selected pane) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("Block With Collapsed Borders",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press 'q' to quit)"),
	).Centered(), rows[0])

	renderPanes(f, rows[1], selected)
}

// renderPanes draws the four blocks. The recipe for collapsed borders is:
//
//  1. merge borders with symbols.MergeExact (or MergeFuzzy);
//  2. overlap the areas by one cell with catatui.Overlap(1), so neighbours
//     share a row or column;
//  3. give the selected pane a thick border so it stands out;
//  4. render the selected pane last, so it ends up on top.
func renderPanes(f *catatui.Frame, area catatui.Rect, selected pane) {
	rows := catatui.VerticalLayout(catatui.Fill(1), catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Overlap(1)).Split(area)
	middle := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Overlap(1)).Split(rows[1])

	panes := []struct {
		pane  pane
		area  catatui.Rect
		title string
	}{
		{paneTop, rows[0], "Top Block"},
		{paneLeft, middle[0], "Left Block"},
		{paneRight, middle[1], "Right Block"},
		{paneBottom, rows[2], "Bottom Block"},
	}

	// Everything but the selected pane first.
	for _, p := range panes {
		if p.pane == selected {
			continue
		}
		block := widgets.Bordered().
			MergeBorders(symbols.MergeExact).
			Title(p.title)
		f.RenderWidget(block, p.area)
	}

	// Then the selected one, over the top of its neighbours.
	for _, p := range panes {
		if p.pane != selected {
			continue
		}
		block := widgets.Bordered().
			MergeBorders(symbols.MergeExact).
			BorderType(widgets.BorderThick).
			BorderStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
			Title(p.title)
		f.RenderWidget(block, p.area)
	}
}
