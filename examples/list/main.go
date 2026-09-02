// Command list shows catatui's List widget, including a bottom-to-top list.
//
//	go run ./examples/list
//
// Press j/k or the arrow keys to move the selection, q to quit.
//
// Port of ratatui-widgets/examples/list.rs @ ratatui-v0.30.2
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

// run keeps the list state across frames, which is what a stateful widget is
// for: the widget is rebuilt every frame, the selection is not.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	state := widgets.NewListState().WithSelected(0)

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
		case ev.IsRune('j'), ev.IsKey(term.KeyDown):
			state.SelectNext()
		case ev.IsRune('k'), ev.IsKey(term.KeyUp):
			state.SelectPrevious()
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return nil
		}
	}
}

// render draws a title and two lists.
func render(f *catatui.Frame, state *widgets.ListState) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("List Widget",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press 'q' to quit and arrow keys to navigate)"),
	).Centered(), rows[0])

	renderList(f, rows[1], state)
	renderBottomList(f, rows[2])
}

// renderList draws the list whose selection the key handler moves.
func renderList(f *catatui.Frame, area catatui.Rect, state *widgets.ListState) {
	list := widgets.NewListFromStrings("Item 1", "Item 2", "Item 3", "Item 4").
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
		HighlightSymbol("> ")
	catatui.RenderStatefulWidget(list, area, f.Buffer(), state)
}

// renderBottomList draws a list that fills from the bottom up, the way a chat
// log does, with multi-line items.
func renderBottomList(f *catatui.Frame, area catatui.Rect) {
	list := widgets.NewListFromStrings(
		"[Remy]: I'm building one now.\nIt even supports multiline text!",
		"[Gusteau]: With enough passion, yes.",
		"[Remy]: But can anyone build a TUI in Go?",
		"[Gusteau]: Anyone can cook!",
	).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		HighlightStyle(catatui.NewStyle().
			Fg(catatui.ColorYellow).
			AddModifier(catatui.ModifierItalic)).
		HighlightSymbolLine(catatui.LineFromStyledString("> ",
			catatui.NewStyle().Fg(catatui.ColorRed))).
		ScrollPadding(1).
		Direction(widgets.ListDirectionBottomToTop).
		RepeatHighlightSymbol(true)

	state := widgets.NewListState()
	state.SelectFirst()
	catatui.RenderStatefulWidget(list, area, f.Buffer(), &state)
}
