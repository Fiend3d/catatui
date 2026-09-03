// Command block shows catatui's Block widget: borders, styles and border types.
//
//	go run ./examples/widgets/block
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/block.rs @ ratatui-v0.30.2
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

// render draws three blocks side by side.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	cols := catatui.HorizontalLayout(
		catatui.Percentage(33), catatui.Percentage(33), catatui.Percentage(33),
	).Spacing(catatui.Space(1)).Split(rows[1])

	f.RenderWidget(title("Block Widget"), rows[0])
	renderBorderedBlock(f, cols[0])
	renderStyledBlock(f, cols[1])
	renderCustomBorderedBlock(f, cols[2])
}

// renderBorderedBlock draws a plain block with a title.
func renderBorderedBlock(f *catatui.Frame, area catatui.Rect) {
	f.RenderWidget(widgets.Bordered().Title("Bordered block"), area)
}

// renderStyledBlock draws a block whose style covers everything inside it.
func renderStyledBlock(f *catatui.Frame, area catatui.Rect) {
	block := widgets.Bordered().
		Style(catatui.NewStyle().
			Fg(catatui.ColorBlue).
			Bg(catatui.ColorBlack).
			AddModifier(catatui.ModifierBold | catatui.ModifierItalic)).
		Title("Styled block")
	f.RenderWidget(block, area)
}

// renderCustomBorderedBlock draws a block with rounded, coloured borders.
func renderCustomBorderedBlock(f *catatui.Frame, area catatui.Rect) {
	block := widgets.Bordered().
		BorderType(widgets.BorderRounded).
		BorderStyle(catatui.NewStyle().Fg(catatui.ColorRed)).
		Title("Custom borders")
	f.RenderWidget(block, area)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
