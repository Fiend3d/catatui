// Command popup shows how to draw a popup over the top of the rest of the UI.
//
//	go run ./examples/popup
//
// Press p to toggle the popup, q to quit.
//
// Port of examples/apps/popup @ ratatui-v0.30.2
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

// run keeps one flag across frames. An application with more state than this
// would put it in a struct; a single bool does not need one.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	showPopup := false

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, showPopup) }); err != nil {
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
		case ev.IsRune('p'):
			showPopup = !showPopup
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return nil
		}
	}
}

// render draws the content, and the popup over it when it is showing.
func render(f *catatui.Frame, showPopup bool) {
	area := f.Area()

	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).Split(area)
	instructions, content := rows[0], rows[1]

	f.RenderWidget(
		catatui.LineFromString("Press 'p' to toggle popup, 'q' to quit").Centered(),
		instructions)

	f.RenderWidget(
		widgets.Bordered().
			Title("Content").
			Style(catatui.NewStyle().Bg(catatui.ColorBlue)),
		content)

	if !showPopup {
		return
	}

	popupArea := centered(area, catatui.Percentage(60), catatui.Percentage(20))

	// Clear wipes whatever was drawn underneath, so the blue content block
	// does not show through the popup.
	f.RenderWidget(widgets.Clear{}, popupArea)
	f.RenderWidget(
		widgets.NewParagraph("Lorem ipsum").Block(widgets.Bordered().Title("Popup")),
		popupArea)

	// Rendering into the block's inner area instead of over the whole popup
	// is the other way round to do this:
	//
	//	inner := popupBlock.Inner(popupArea)
	//	f.RenderWidget(yourWidget, inner)
}

// centered cuts a rect of the given size out of the middle of area. This is
// what ratatui's Rect::centered does: one centred layout per axis.
func centered(area catatui.Rect, horizontal, vertical catatui.Constraint) catatui.Rect {
	row := catatui.VerticalLayout(vertical).Flex(catatui.FlexCenter).Split(area)[0]
	return catatui.HorizontalLayout(horizontal).Flex(catatui.FlexCenter).Split(row)[0]
}
