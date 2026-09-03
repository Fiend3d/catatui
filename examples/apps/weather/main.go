// Command weather draws a day of made-up temperatures as a vertical bar chart,
// one bar per hour, coloured from yellow through to red.
//
//	go run ./examples/apps/weather
//
// Any key quits.
//
// Port of examples/apps/weather @ ratatui-v0.30.2
package main

import (
	"fmt"
	"math/rand/v2"
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	// The readings are made up once and then kept, so the chart does not
	// change shape underneath a redraw.
	temperatures := newTemperatures()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, temperatures) }); err != nil {
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

// newTemperatures makes up one reading per hour, between 50 and 89 degrees,
// which is the range temperatureStyle knows how to colour.
func newTemperatures() []uint8 {
	temperatures := make([]uint8, 24)
	for i := range temperatures {
		temperatures[i] = uint8(50 + rand.IntN(40))
	}
	return temperatures
}

// render draws the title and the chart under it.
func render(f *catatui.Frame, temperatures []uint8) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).
		Split(f.Area())

	f.RenderWidget(
		catatui.LineFromStyledString("Weather demo",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)).Centered(),
		rows[0])
	f.RenderWidget(verticalBarChart(temperatures), rows[1])
}

// verticalBarChart is one bar per hour, wide enough for the "00:00" labels
// underneath them.
func verticalBarChart(temperatures []uint8) widgets.BarChart {
	bars := make([]widgets.Bar, len(temperatures))
	for hour, temperature := range temperatures {
		bars[hour] = verticalBar(hour, temperature)
	}
	return widgets.NewBarChart().
		Data(widgets.NewBarGroup(bars...)).
		BarWidth(5)
}

// verticalBar is one hour: the reading, the hour under it, and the temperature
// written into the bar itself.
func verticalBar(hour int, temperature uint8) widgets.Bar {
	style := temperatureStyle(temperature)
	return widgets.NewBar(uint64(temperature)).
		Label(catatui.LineFromString(fmt.Sprintf("%02d:00", hour))).
		TextValue(fmt.Sprintf("%3d°", temperature)).
		Style(style).
		ValueStyle(style.AddModifier(catatui.ModifierReversed))
}

// temperatureStyle runs from yellow at 50 degrees to red at 90, by taking the
// green out of it.
func temperatureStyle(value uint8) catatui.Style {
	green := uint8(255.0 * (1.0 - float64(value-50)/40.0))
	return catatui.NewStyle().Fg(catatui.Rgb(255, green, 0))
}
