// Command scrollbar shows catatui's Scrollbar widget, vertical and horizontal,
// scrolling a paragraph.
//
//	go run ./examples/scrollbar
//
// Press h/j/k/l or the arrow keys to scroll, q to quit.
//
// Port of ratatui-widgets/examples/scrollbar.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run keeps one scrollbar state per axis across frames.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	vertical := widgets.NewScrollbarState(100)
	horizontal := widgets.NewScrollbarState(100)

	for {
		err := terminal.Draw(func(f *catatui.Frame) {
			render(f, &vertical, &horizontal)
		})
		if err != nil {
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
			vertical.Next()
		case ev.IsRune('k'), ev.IsKey(term.KeyUp):
			vertical.Prev()
		case ev.IsRune('l'), ev.IsKey(term.KeyRight):
			horizontal.Next()
		case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
			horizontal.Prev()
		}
	}
}

// render draws the content with a scrollbar down its right edge and another
// along its bottom.
func render(f *catatui.Frame, vertical, horizontal *widgets.ScrollbarState) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("Scrollbar Widget",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press 'q' to quit, arrow keys to scroll)"),
	).Centered(), rows[0])

	main := rows[1]
	renderContent(f, main, vertical, horizontal)
	renderVerticalScrollbar(f, main, vertical)
	renderHorizontalScrollbar(f, main, horizontal)
}

// renderVerticalScrollbar draws the scrollbar on the right, inset by a row at
// each end so it does not sit in the corners.
func renderVerticalScrollbar(f *catatui.Frame, area catatui.Rect, state *widgets.ScrollbarState) {
	scrollbar := widgets.NewScrollbar(widgets.ScrollbarVerticalRight)
	inner := area.Inner(catatui.NewMargin(0, 1))
	catatui.RenderStatefulWidget(scrollbar, inner, f.Buffer(), state)
}

// renderHorizontalScrollbar draws the scrollbar along the bottom with symbols
// and styles of its own.
func renderHorizontalScrollbar(f *catatui.Frame, area catatui.Rect, state *widgets.ScrollbarState) {
	scrollbar := widgets.NewScrollbar(widgets.ScrollbarHorizontalBottom).
		Symbols(symbols.ScrollbarSet{Track: "-", Thumb: "▮", Begin: "<", End: ">"}).
		TrackStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
		BeginStyle(catatui.NewStyle().Fg(catatui.ColorGreen)).
		EndStyle(catatui.NewStyle().Fg(catatui.ColorRed))
	inner := area.Inner(catatui.NewMargin(1, 0))
	catatui.RenderStatefulWidget(scrollbar, inner, f.Buffer(), state)
}

// renderContent draws the text the scrollbars scroll, and reports both
// positions so it is clear which key moved what.
func renderContent(f *catatui.Frame, area catatui.Rect, vertical, horizontal *widgets.ScrollbarState) {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	yellow := catatui.NewStyle().Fg(catatui.ColorYellow)

	content := catatui.NewText(
		catatui.LineFromString("This is a paragraph with a vertical and horizontal scrollbar."),
		catatui.LineFromString(strings.Repeat(
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.", 10)),
		catatui.NewLine(
			catatui.NewStyledSpan("Horizontal: ", bold),
			catatui.NewStyledSpan(fmt.Sprint(horizontal.GetPosition()), yellow),
		),
		catatui.NewLine(
			catatui.NewStyledSpan("Vertical: ", bold),
			catatui.NewStyledSpan(fmt.Sprint(vertical.GetPosition()), yellow),
		),
	)

	paragraph := widgets.NewParagraphFromText(content).
		Scroll(catatui.Position{
			X: uint16(horizontal.GetPosition()),
			Y: uint16(vertical.GetPosition()),
		})
	f.RenderWidget(paragraph, area)
}
