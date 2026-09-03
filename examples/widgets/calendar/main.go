// Command calendar shows catatui's Monthly widget with two styled months.
//
//	go run ./examples/widgets/calendar
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/calendar.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"time"

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

// render draws two monthly calendars side by side.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	cols := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Spacing(catatui.Space(1)).Split(rows[1])

	f.RenderWidget(title("Calendar Widget"), rows[0])
	renderCurrentMonth(f, cols[0])
	renderStyledMonth(f, cols[1])
}

// renderCurrentMonth draws this month with today picked out.
func renderCurrentMonth(f *catatui.Frame, area catatui.Rect) {
	red := catatui.NewStyle().Fg(catatui.ColorRed).AddModifier(catatui.ModifierBold)

	monthly := widgets.NewMonthly(time.Now(), widgets.CalendarEventStoreToday(red)).
		Block(widgets.NewBlock().Padding(widgets.NewPadding(0, 0, 2, 0))).
		ShowMonthHeader(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		ShowWeekdaysHeader(catatui.NewStyle().AddModifier(catatui.ModifierItalic))
	f.RenderWidget(monthly, area)
}

// renderStyledMonth draws a month in the past with every style turned on.
func renderStyledMonth(f *catatui.Frame, area catatui.Rect) {
	// Release date of the movie Ratatouille.
	date := time.Date(2007, time.June, 29, 0, 0, 0, 0, time.Local)

	red := catatui.NewStyle().Fg(catatui.ColorRed).AddModifier(catatui.ModifierBold)
	events := widgets.CalendarEventStoreToday(red)
	events.Add(date, catatui.NewStyle().
		Fg(catatui.ColorBlue).
		AddModifier(catatui.ModifierItalic))

	monthly := widgets.NewMonthly(date, events).
		ShowSurrounding(catatui.NewStyle().AddModifier(catatui.ModifierDim)).
		ShowMonthHeader(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		ShowWeekdaysHeader(catatui.NewStyle().
			Fg(catatui.ColorGreen).
			AddModifier(catatui.ModifierBold)).
		DefaultStyle(catatui.NewStyle().
			Bg(catatui.Rgb(50, 50, 50)).
			AddModifier(catatui.ModifierBold))
	f.RenderWidget(monthly, area)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
