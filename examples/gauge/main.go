// Command gauge shows catatui's Gauge and LineGauge widgets.
//
//	go run ./examples/gauge
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/gauge.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

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

// render draws a block gauge and a line gauge.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Max(2), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(title("Gauge Widget"), rows[0])
	renderGauge(f, rows[1])
	renderLineGauge(f, rows[2])
}

// renderGauge draws a gauge with a label of its own.
func renderGauge(f *catatui.Frame, area catatui.Rect) {
	gauge := widgets.NewGauge().
		Style(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		GaugeStyle(catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorBlack)).
		Label("Year Progress").
		Percent(80)
	f.RenderWidget(gauge, area)
}

// renderLineGauge draws the one-row form, which needs no block.
func renderLineGauge(f *catatui.Frame, area catatui.Rect) {
	lineGauge := widgets.NewLineGauge().
		FilledStyle(catatui.NewStyle().
			Fg(catatui.ColorWhite).
			Bg(catatui.ColorRed).
			AddModifier(catatui.ModifierBold)).
		UnfilledStyle(catatui.NewStyle().Fg(catatui.ColorGray).Bg(catatui.ColorBlack)).
		Label("❤️ HP").
		Ratio(0.42).
		FilledSymbol(symbols.ThickHorizontal).
		UnfilledSymbol(symbols.ThickHorizontal)
	f.RenderWidget(lineGauge, area)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
