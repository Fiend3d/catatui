// Command barchart-grouped shows catatui's BarChart widget with grouped bars.
//
//	go run ./examples/widgets/barchart-grouped
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/barchart-grouped.rs @ ratatui-v0.30.2
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

// render draws the same grouped data both ways up.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	cols := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(rows[1])

	f.RenderWidget(title("BarChart Widget (Grouped)"), rows[0])
	renderBarChart(f, cols[0], catatui.Vertical, 6)
	renderBarChart(f, cols[1], catatui.Horizontal, 1)
}

// company is one series of the chart: a label and the colour it is drawn in.
type company struct {
	label string
	color catatui.Color
}

// renderBarChart draws quarterly revenue, one group per month and one bar per
// company within each group.
func renderBarChart(f *catatui.Frame, area catatui.Rect, direction catatui.Direction, barWidth uint16) {
	companies := []company{
		{"BITE", catatui.ColorBlue},
		{"TART", catatui.ColorWhite},
		{"BAKE", catatui.ColorLightRed},
	}
	revenues := []struct {
		period string
		values [3]uint64
	}{
		{"Jan", [3]uint64{8500, 6500, 7000}},
		{"Feb", [3]uint64{9000, 7500, 8500}},
		{"Mar", [3]uint64{9500, 4500, 8200}},
		{"Apr", [3]uint64{6300, 4000, 5000}},
	}

	chart := widgets.NewBarChart().
		BarGap(0).
		BarWidth(barWidth).
		GroupGap(2).
		Direction(direction)

	for _, revenue := range revenues {
		bars := make([]widgets.Bar, len(companies))
		for i, c := range companies {
			bars[i] = bar(c.label, revenue.values[i], c.color)
		}
		group := widgets.NewBarGroup(bars...).
			Label(catatui.LineFromString(revenue.period).Centered())
		chart = chart.Data(group)
	}

	f.RenderWidget(chart, area)
}

// bar returns one bar, labelled and with its value shown in millions.
func bar(label string, value uint64, color catatui.Color) widgets.Bar {
	return widgets.NewBar(value).
		Label(catatui.LineFromString(label)).
		TextValue(fmt.Sprintf("%.1fM", float64(value)/1000)).
		Style(catatui.NewStyle().Fg(color)).
		ValueStyle(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(color))
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
