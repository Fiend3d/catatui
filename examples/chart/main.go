// Command chart shows catatui's Chart widget: a line plot and a filled area
// plot sharing a pair of axes.
//
//	go run ./examples/chart
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/chart.rs @ ratatui-v0.30.2
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

// render draws the title and the chart below it.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(title("Chart Widget"), rows[0])
	renderChart(f, rows[1])
}

// renderChart draws two datasets: one line going up, one area going down.
func renderChart(f *catatui.Frame, area catatui.Rect) {
	upward := widgets.NewDataset().
		Name("Stonks").
		Marker(symbols.Braille).
		GraphType(widgets.GraphTypeLine).
		Style(catatui.NewStyle().Fg(catatui.ColorBlue)).
		Data([][2]float64{
			{0, 10}, {1, 14}, {2, 12}, {3, 15}, {4, 12.5}, {5, 16},
			{6, 13}, {7, 18}, {8, 17}, {9, 19}, {10, 20},
		})

	downward := widgets.NewDataset().
		Name("Not stonks").
		Marker(symbols.Braille).
		GraphType(widgets.GraphTypeArea).
		FillToY(0).
		Style(catatui.NewStyle().Fg(catatui.ColorRed)).
		Data([][2]float64{
			{0, 10}, {1, 8}, {2, 8.5}, {3, 6}, {4, 7}, {5, 5},
			{6, 5.5}, {7, 4}, {8, 3.5}, {9, 1.5}, {10, 2.5},
		})

	blue := catatui.NewStyle().Fg(catatui.ColorBlue)
	xAxis := widgets.NewAxis().
		TitleLine(catatui.LineFromStyledString("Hustle", blue)).
		Bounds([2]float64{0, 10}).
		Labels("0%", "50%", "100%")
	yAxis := widgets.NewAxis().
		TitleLine(catatui.LineFromStyledString("Profit", blue)).
		Bounds([2]float64{0, 20}).
		Labels("0", "10", "20")

	chart := widgets.NewChart(downward, upward).XAxis(xAxis).YAxis(yAxis)
	f.RenderWidget(chart, area)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
