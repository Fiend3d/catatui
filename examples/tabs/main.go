// Command tabs shows catatui's Tabs widget over a block of content.
//
//	go run ./examples/tabs
//
// Press h/l or the arrow keys to change tab, q to quit.
//
// Port of ratatui-widgets/examples/tabs.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// tabCount is how many tabs there are to cycle through.
const tabCount = 3

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run keeps the selected tab across frames.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	selected := 0

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
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return nil
		case ev.IsRune('l'), ev.IsKey(term.KeyRight):
			selected = (selected + 1) % tabCount
		case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
			selected = (selected + tabCount - 1) % tabCount
		}
	}
}

// render draws the content first and the tabs over its top border, which is
// what makes the tabs look attached to the block.
func render(f *catatui.Frame, selected int) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("Tabs Widget",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press 'q' to quit, arrow keys to navigate tabs)"),
	).Centered(), rows[0])

	renderContent(f, rows[1], selected)
	renderTabs(f, rows[1].Offset(catatui.Offset{X: 1}), selected)
}

// renderTabs draws the row of tab titles.
func renderTabs(f *catatui.Frame, area catatui.Rect, selected int) {
	tabs := widgets.NewTabs("Tab1", "Tab2", "Tab3").
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		HighlightStyle(catatui.NewStyle().
			Fg(catatui.ColorMagenta).
			Bg(catatui.ColorBlack).
			AddModifier(catatui.ModifierBold)).
		Select(selected).
		Divider(symbols.DotFull).
		Padding(" ", " ")
	f.RenderWidget(tabs, area)
}

// renderContent draws whatever the selected tab shows.
func renderContent(f *catatui.Frame, area catatui.Rect, selected int) {
	var text catatui.Text
	switch selected {
	case 0:
		text = catatui.TextFromString("Great terminal interfaces start with a single widget.")
	case 1:
		text = catatui.TextFromString("In the terminal, we don't just render widgets; we create dreams.")
	default:
		text = catatui.TextFromStyledString("Render boldly, style with purpose.",
			catatui.NewStyle().AddModifier(catatui.ModifierBold))
	}

	paragraph := widgets.NewParagraphFromText(text).
		Alignment(catatui.AlignmentCenter).
		Block(widgets.Bordered())
	f.RenderWidget(paragraph, area)
}
