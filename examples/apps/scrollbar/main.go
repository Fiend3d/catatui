// Command scrollbar draws four scrollbars over the same text, each built a
// different way.
//
//	go run ./examples/apps/scrollbar
//
// h/j/k/l or the arrow keys scroll, q quits. The two vertical bars move
// together and so do the two horizontal ones: what differs is the arrows, the
// track and the thumb.
//
// Port of examples/apps/scrollbar @ ratatui-v0.30.2
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{}

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

// app is how far the text is scrolled each way, and the two scrollbar states
// that follow it.
type app struct {
	verticalScroll   int
	horizontalScroll int
	verticalState    widgets.ScrollbarState
	horizontalState  widgets.ScrollbarState
	quit             bool
}

// handle applies one event. The scroll and the scrollbar state are moved
// together: the state is what puts the thumb in the right place.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.verticalScroll++
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.verticalScroll = max(a.verticalScroll-1, 0)
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.horizontalScroll++
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.horizontalScroll = max(a.horizontalScroll-1, 0)
	}
	a.verticalState = a.verticalState.Position(a.verticalScroll)
	a.horizontalState = a.horizontalState.Position(a.horizontalScroll)
}

// render draws the heading and the four scrolling panes.
func (a *app) render(f *catatui.Frame) {
	area := f.Area()

	// A line long enough to need scrolling sideways whatever the window is.
	longLine := longLine(area.Width)
	text := sampleText(longLine)

	// The content lengths are what the thumb's size and travel are worked out
	// from, and they depend on the text, so they are set here rather than kept
	// in the app.
	a.verticalState = a.verticalState.ContentLength(len(text.GetLines()))
	a.horizontalState = a.horizontalState.ContentLength(catatui.StringWidth(longLine))

	rows := catatui.VerticalLayout(
		catatui.Min(1),
		catatui.Percentage(25),
		catatui.Percentage(25),
		catatui.Percentage(25),
		catatui.Percentage(25),
	).Split(area)

	f.RenderWidget(
		widgets.NewBlock().
			TitleAlignment(catatui.AlignmentCenter).
			TitleLine(catatui.LineFromStyledString("Use h j k l or ◄ ▲ ▼ ► to scroll ",
				catatui.NewStyle().AddModifier(catatui.ModifierBold))),
		rows[0])

	a.renderVerticalScroll(f, text, rows[1], rows[2])
	a.renderHorizontalScroll(f, text, rows[3], rows[4])
}

// renderVerticalScroll draws the two vertical bars: one with arrows on the
// right, one without arrows or a track on the left.
func (a *app) renderVerticalScroll(f *catatui.Frame, text catatui.Text, withArrows, plain catatui.Rect) {
	f.RenderWidget(
		paragraph(text).
			Block(createBlock("Vertical scrollbar with arrows")).
			Scroll(catatui.Position{Y: uint16(a.verticalScroll)}),
		withArrows)
	catatui.RenderStatefulWidgetOn(f,
		widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
			BeginSymbol("↑").
			EndSymbol("↓"),
		withArrows, &a.verticalState)

	f.RenderWidget(
		paragraph(text).
			Block(createBlock("Vertical scrollbar without arrows, without track symbol and mirrored")).
			Scroll(catatui.Position{Y: uint16(a.verticalScroll)}),
		plain)
	catatui.RenderStatefulWidgetOn(f,
		widgets.NewScrollbar(widgets.ScrollbarVerticalLeft).
			Symbols(symbols.ScrollbarVertical).
			BeginSymbolNone().
			TrackSymbolNone().
			EndSymbolNone(),
		// Inside the block's top and bottom border, so the bar runs the height
		// of the text rather than of the box.
		plain.Inner(catatui.Margin{Vertical: 1}), &a.verticalState)
}

// renderHorizontalScroll draws the two horizontal bars: one with a begin arrow
// and a custom thumb, one with no arrows and a custom track.
func (a *app) renderHorizontalScroll(f *catatui.Frame, text catatui.Text, withArrow, plain catatui.Rect) {
	f.RenderWidget(
		paragraph(text).
			Block(createBlock("Horizontal scrollbar with only begin arrow & custom thumb symbol")).
			Scroll(catatui.Position{X: uint16(a.horizontalScroll)}),
		withArrow)
	catatui.RenderStatefulWidgetOn(f,
		widgets.NewScrollbar(widgets.ScrollbarHorizontalBottom).
			ThumbSymbol("🬋").
			EndSymbolNone(),
		withArrow.Inner(catatui.Margin{Horizontal: 1}), &a.horizontalState)

	f.RenderWidget(
		paragraph(text).
			Block(createBlock("Horizontal scrollbar without arrows & custom thumb and track symbol")).
			Scroll(catatui.Position{X: uint16(a.horizontalScroll)}),
		plain)
	catatui.RenderStatefulWidgetOn(f,
		widgets.NewScrollbar(widgets.ScrollbarHorizontalBottom).
			ThumbSymbol("░").
			TrackSymbol("─"),
		plain.Inner(catatui.Margin{Horizontal: 1}), &a.horizontalState)
}

// paragraph is the text in grey, which every pane draws.
func paragraph(text catatui.Text) widgets.Paragraph {
	return widgets.NewParagraphFromText(text).
		Style(catatui.NewStyle().Fg(catatui.ColorGray))
}

// createBlock is the bordered box each pane sits in.
func createBlock(title string) widgets.Block {
	return widgets.Bordered().
		Style(catatui.NewStyle().Fg(catatui.ColorGray)).
		TitleLine(catatui.LineFromStyledString(title,
			catatui.NewStyle().AddModifier(catatui.ModifierBold)))
}

// longLine is a line wide enough to scroll sideways in a window of the given
// width, with the words drawn out to show where they break.
func longLine(width uint16) string {
	const s = "Veeeeeeeeeeeeeeeery    loooooooooooooooooong   striiiiiiiiiiiiiiiiiiiiiiiiiing.   "
	return strings.Repeat(s, int(width)/len(s)+4)
}

// sampleText is what all four panes scroll: a few styled lines, the long one,
// and a masked password, twice over so there is something to scroll to.
func sampleText(long string) catatui.Text {
	var lines []catatui.Line
	for range 2 {
		lines = append(lines,
			catatui.LineFromString("This is a line "),
			catatui.LineFromStyledString("This is a line   ",
				catatui.NewStyle().Fg(catatui.ColorRed)),
			catatui.LineFromStyledString("This is a line",
				catatui.NewStyle().Bg(catatui.ColorDarkGray)),
			catatui.LineFromStyledString("This is a longer line",
				catatui.NewStyle().AddModifier(catatui.ModifierCrossedOut)),
			catatui.LineFromString(long),
			catatui.LineFromStyledString("This is a line", catatui.ResetStyle()),
			catatui.NewLine(
				catatui.NewSpan("Masked text: "),
				catatui.NewStyledSpan(masked("password", '*'),
					catatui.NewStyle().Fg(catatui.ColorRed)),
			),
		)
	}
	return catatui.NewText(lines...)
}

// masked hides a string behind a mask character, one per grapheme cluster, the
// way ratatui's Masked does.
func masked(s string, mask rune) string {
	width := 0
	for range catatui.AllGraphemes(s) {
		width++
	}
	return strings.Repeat(string(mask), width)
}
