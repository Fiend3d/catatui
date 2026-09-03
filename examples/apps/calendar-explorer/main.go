// Command calendar-explorer draws a whole year of Monthly calendars, with the
// holidays and the solstices marked, in each of the styles the widget offers.
//
//	go run ./examples/apps/calendar-explorer
//
// s changes the style, n and p (or Tab and Shift-Tab) move a month, h/j/k/l or
// the arrows move a day or a week, q quits. The year shown is the year of
// whichever date is selected, so moving off either end of December or January
// redraws the lot.
//
// Port of examples/apps/calendar-explorer @ ratatui-v0.30.2
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{selected: today()}

	for !a.quit {
		if err := terminal.Draw(func(f *catatui.Frame) { a.render(f) }); err != nil {
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

// today is the current date with the time of day thrown away, which is what
// the calendar and the event store both compare on.
func today() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// app is the date the cursor is on and which style the calendars are drawn in.
type app struct {
	selected time.Time
	style    calendarStyle
	quit     bool
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('s'):
		a.style = a.style.next()
	case ev.IsRune('n'), ev.IsKey(term.KeyTab):
		a.selected = addMonths(a.selected, 1)
	case ev.IsRune('p'), ev.IsKey(term.KeyBackTab):
		a.selected = addMonths(a.selected, -1)
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.selected = a.selected.AddDate(0, 0, -1)
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.selected = a.selected.AddDate(0, 0, 7)
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.selected = a.selected.AddDate(0, 0, -7)
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.selected = a.selected.AddDate(0, 0, 1)
	}
}

// addMonths moves n months, keeping the day where the new month is long
// enough for it and clamping to the last day where it is not.
//
// Go's AddDate would roll the 31st of January forward into March rather than
// stopping at the end of February, and ratatui's replace_month refuses the
// date outright; clamping is the behaviour a date picker wants.
func addMonths(date time.Time, n int) time.Time {
	first := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	moved := first.AddDate(0, n, 0)
	return time.Date(moved.Year(), moved.Month(),
		min(date.Day(), daysInMonth(moved)), 0, 0, 0, 0, date.Location())
}

// daysInMonth is how many days the month containing date has. Day zero of the
// next month is the last day of this one.
func daysInMonth(date time.Time) int {
	return time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()
}

// render draws the header and the year under it.
func (a *app) render(f *catatui.Frame) {
	header := catatui.NewText(
		catatui.LineFromStyledString("Calendar Example",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.LineFromString(
			"<q> Quit | <s> Change Style | <n> Next Month | <p> Previous Month, <hjkl> Move"),
		catatui.LineFromString(fmt.Sprintf("Current date: %s | Current style: %v",
			a.selected.Format("2006-01-02"), a.style)),
	).Centered()

	rows := catatui.VerticalLayout(
		catatui.Length(uint16(header.Height())),
		catatui.Fill(1),
	).Split(f.Area())

	f.RenderWidget(widgets.NewParagraphFromText(header), rows[0])
	a.renderYear(f, rows[1])
}

// renderYear draws the twelve months of the selected date's year, three rows
// of four.
func (a *app) renderYear(f *catatui.Frame, area catatui.Rect) {
	events := a.events()

	rows := catatui.VerticalLayout(
		catatui.Ratio(1, 3), catatui.Ratio(1, 3), catatui.Ratio(1, 3),
	).Split(area.Inner(catatui.Margin{Horizontal: 1, Vertical: 1}))

	var month int
	for _, row := range rows {
		columns := catatui.HorizontalLayout(
			catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4),
		).Split(row)
		for _, cell := range columns {
			first := time.Date(a.selected.Year(), time.Month(month+1), 1, 0, 0, 0, 0, a.selected.Location())
			f.RenderWidget(a.style.monthly(first, events), cell)
			month++
		}
	}
}

// events marks the holidays, the solstices and equinoxes, today, and wherever
// the cursor is.
func (a *app) events() widgets.CalendarEventStore {
	var (
		selected = catatui.NewStyle().
				Fg(catatui.ColorWhite).
				Bg(catatui.ColorRed).
				AddModifier(catatui.ModifierBold)
		holiday = catatui.NewStyle().
			Fg(catatui.ColorRed).
			AddModifier(catatui.ModifierUnderlined)
		season = catatui.NewStyle().
			Fg(catatui.ColorGreen).
			Bg(catatui.ColorBlack).
			AddModifier(catatui.ModifierUnderlined)
	)

	store := widgets.CalendarEventStoreToday(
		catatui.NewStyle().AddModifier(catatui.ModifierBold).Bg(catatui.ColorBlue))

	year := a.selected.Year()
	date := func(y int, month time.Month, day int) time.Time {
		return time.Date(y, month, day, 0, 0, 0, 0, a.selected.Location())
	}

	for _, holidayDate := range []time.Time{
		date(year, time.January, 1),
		// Next new year's day too, so December has something to show when it
		// is drawn with the surrounding days on.
		date(year+1, time.January, 1),
		date(year, time.February, 2),  // groundhog day
		date(year, time.April, 1),     // april fool's
		date(year, time.April, 22),    // earth day
		date(year, time.May, 4),       // star wars day
		date(year, time.December, 23), // festivus
		date(year, time.December, 31), // new year's eve
	} {
		store.Add(holidayDate, holiday)
	}

	for _, seasonDate := range []time.Time{
		date(year, time.March, 22),     // spring equinox
		date(year, time.June, 21),      // summer solstice
		date(year, time.September, 22), // autumn equinox
		date(year, time.December, 21),  // winter solstice
	} {
		store.Add(seasonDate, season)
	}

	store.Add(a.selected, selected)
	return store
}

// calendarStyle is which of the Monthly headers and surrounding days are
// turned on, cycled through with s.
type calendarStyle int

const (
	styleDefault calendarStyle = iota
	styleSurrounding
	styleWeekdaysHeader
	styleSurroundingAndWeekdaysHeader
	styleMonthHeader
	styleMonthAndWeekdaysHeader
)

// next cycles to the following style, and round to the first from the last.
func (s calendarStyle) next() calendarStyle {
	return (s + 1) % calendarStyle(len(calendarStyleNames))
}

func (s calendarStyle) String() string {
	if int(s) < len(calendarStyleNames) {
		return calendarStyleNames[s]
	}
	return "Unknown"
}

var calendarStyleNames = []string{
	styleDefault:                      "Default",
	styleSurrounding:                  "Show Surrounding",
	styleWeekdaysHeader:               "Show Weekdays Header",
	styleSurroundingAndWeekdaysHeader: "Show Surrounding and Weekdays Header",
	styleMonthHeader:                  "Show Month Header",
	styleMonthAndWeekdaysHeader:       "Show Month Header and Weekdays Header",
}

// monthly builds the calendar for one month in the current style. Every style
// shows the month header; what changes is whether the surrounding days and the
// weekday names come with it, and how they are drawn.
func (s calendarStyle) monthly(month time.Time, events widgets.CalendarEventStore) widgets.Monthly {
	calendar := widgets.NewMonthly(month, events).
		DefaultStyle(catatui.NewStyle().
			AddModifier(catatui.ModifierBold).
			Bg(catatui.Rgb(50, 50, 50))).
		ShowMonthHeader(catatui.NewStyle())

	dim := catatui.NewStyle().AddModifier(catatui.ModifierDim)
	boldGreen := catatui.NewStyle().
		AddModifier(catatui.ModifierBold).
		Fg(catatui.ColorGreen)

	switch s {
	case styleSurrounding:
		calendar = calendar.ShowSurrounding(dim)
	case styleWeekdaysHeader:
		calendar = calendar.ShowWeekdaysHeader(boldGreen)
	case styleSurroundingAndWeekdaysHeader:
		calendar = calendar.ShowSurrounding(dim).ShowWeekdaysHeader(boldGreen)
	case styleMonthHeader:
		calendar = calendar.ShowMonthHeader(boldGreen)
	case styleMonthAndWeekdaysHeader:
		calendar = calendar.ShowWeekdaysHeader(catatui.NewStyle().
			AddModifier(catatui.ModifierBold).
			AddModifier(catatui.ModifierDim).
			Fg(catatui.ColorLightYellow))
	}
	return calendar
}
