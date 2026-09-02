// Command shadow shows catatui's Shadow widget: four ways to cast a shadow
// behind a popup.
//
//	go run ./examples/shadow
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/shadow.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

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

// run draws the UI and waits for a key.
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

// render draws one popup per shadow kind, in a two-by-two grid.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	f.RenderWidget(title("Shadow Widget"), rows[0])

	halves := catatui.VerticalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Spacing(catatui.Space(1)).Split(rows[1])
	top := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Spacing(catatui.Space(1)).Split(halves[0])
	bottom := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Spacing(catatui.Space(1)).Split(halves[1])

	renderOverlayShadow(f, top[0])
	renderBlockShadow(f, top[1])
	renderSymbolShadow(f, bottom[0])
	renderDimmedShadow(f, bottom[1])
}

// renderOverlayShadow restyles the cells behind the popup without drawing
// anything of its own, so the text under it stays readable.
func renderOverlayShadow(f *catatui.Frame, area catatui.Rect) {
	renderBackground(f, area, catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorBlue))
	shadow := widgets.ShadowOverlay().
		Style(catatui.NewStyle().Bg(catatui.ColorDarkGray))
	block := widgets.Bordered().
		Title("Overlay shadow").
		Style(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorYellow)).
		Shadow(shadow)
	renderPopup(f, area, block)
}

// renderBlockShadow fills the shadow with solid blocks.
func renderBlockShadow(f *catatui.Frame, area catatui.Rect) {
	renderBackground(f, area, catatui.NewStyle().Fg(catatui.ColorWhite))
	shadow := widgets.ShadowBlock().
		Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)).
		Offset(catatui.Offset{X: 2, Y: 1})
	block := widgets.Bordered().
		Title("Block shadow").
		Style(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorYellow)).
		Shadow(shadow)
	renderPopup(f, area, block)
}

// renderSymbolShadow fills the shadow with a character of your choosing.
func renderSymbolShadow(f *catatui.Frame, area catatui.Rect) {
	renderBackground(f, area, catatui.NewStyle().Fg(catatui.ColorWhite))
	shadow := widgets.ShadowSymbol("$").
		Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)).
		Offset(catatui.Offset{X: 2, Y: 1})
	block := widgets.Bordered().
		Title("Symbol shadow").
		Style(catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorRed)).
		Shadow(shadow)
	renderPopup(f, area, block)
}

// renderDimmedShadow dims whatever the shadow falls on.
func renderDimmedShadow(f *catatui.Frame, area catatui.Rect) {
	renderBackground(f, area, catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorBlue))
	shadow := widgets.ShadowCustom(widgets.NewDimmed()).
		Style(catatui.NewStyle().Bg(catatui.ColorDarkGray)).
		Offset(catatui.Offset{X: 2, Y: 1})
	block := widgets.Bordered().
		Title("Dimmed shadow").
		Style(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorGreen)).
		Shadow(shadow)
	renderPopup(f, area, block)
}

// renderBackground fills the area with text for the shadow to fall on.
func renderBackground(f *catatui.Frame, area catatui.Rect, style catatui.Style) {
	sentence := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
	background := widgets.NewParagraph(strings.Repeat(sentence, int(area.Height))).
		Block(widgets.Bordered()).
		Wrap(widgets.Wrap{Trim: true}).
		Style(style)
	f.RenderWidget(background, area)
}

// renderPopup clears a centred area and draws the block into it. Clear is what
// stops the background text showing through the popup.
func renderPopup(f *catatui.Frame, area catatui.Rect, block widgets.Block) {
	popup := centered(area,
		catatui.SatSub(area.Width, 18),
		catatui.SatSub(area.Height, 8))
	f.RenderWidget(widgets.Clear{}, popup)
	f.RenderWidget(block, popup)
}

// centered returns a rect of the given size in the middle of area, which is
// what ratatui's Rect::centered does with a pair of centred layouts.
func centered(area catatui.Rect, width, height uint16) catatui.Rect {
	cols := catatui.HorizontalLayout(catatui.Length(width)).
		Flex(catatui.FlexCenter).Split(area)
	rows := catatui.VerticalLayout(catatui.Length(height)).
		Flex(catatui.FlexCenter).Split(cols[0])
	return rows[0]
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
