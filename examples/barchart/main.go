// Command barchart shows catatui's BarChart widget in both directions.
//
//	go run ./examples/barchart
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/barchart.rs @ ratatui-v0.30.2
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
	// RecoverAndRestore puts the terminal back if anything below panics.
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
		// A resize event redraws; any key quits.
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

// render draws a title and two bar charts.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	cols := catatui.HorizontalLayout(catatui.Length(28), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(rows[1])

	f.RenderWidget(title("BarChart Widget"), rows[0])
	renderVerticalBarChart(f, cols[0])
	renderHorizontalBarChart(f, cols[1])
}

// renderVerticalBarChart draws bars growing upwards.
func renderVerticalBarChart(f *catatui.Frame, area catatui.Rect) {
	f.RenderWidget(widgets.VerticalBarChart(sampleBars()...).BarWidth(6), area)
}

// renderHorizontalBarChart draws bars growing to the right.
func renderHorizontalBarChart(f *catatui.Frame, area catatui.Rect) {
	f.RenderWidget(widgets.HorizontalBarChart(sampleBars()...).BarWidth(3), area)
}

// sampleBars returns one labelled bar per colour.
func sampleBars() []widgets.Bar {
	return []widgets.Bar{
		bar("Red", 30, catatui.ColorRed),
		bar("Blue", 20, catatui.ColorBlue),
		bar("Green", 15, catatui.ColorGreen),
		bar("Yellow", 10, catatui.ColorYellow),
	}
}

// bar returns a labelled bar in the given colour.
func bar(label string, value uint64, color catatui.Color) widgets.Bar {
	return widgets.BarWithLabel(catatui.LineFromString(label), value).
		Style(catatui.NewStyle().Fg(color))
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
