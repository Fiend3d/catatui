// Port of ratatui-widgets/src/calendar.rs @ ratatui-v0.30.2

package widgets

import (
	"fmt"
	"time"

	"github.com/Fiend3d/catatui"
)

// DateStyler provides a style for a given date. Monthly asks it once per day
// it draws, so any type can decide how dates look: a store of events, a set of
// holidays, or a function of the weekday.
//
// Dates are passed with the clock and zone stripped; only year, month and day
// carry meaning.
type DateStyler interface {
	// GetStyle returns the style for a date, or the empty style for a date
	// that has nothing special about it.
	GetStyle(date time.Time) catatui.Style
}

// calendarDate strips a time to its date, so that dates compare by (year,
// month, day) only.
func calendarDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// CalendarEventStore is a simple DateStyler backed by a map from date to
// style.
//
//	store := widgets.NewCalendarEventStore()
//	store.Add(time.Date(2024, time.March, 5, 0, 0, 0, 0, time.UTC),
//		catatui.NewStyle().Fg(catatui.ColorRed))
//	f.RenderWidget(widgets.NewMonthly(today, store), area)
type CalendarEventStore struct {
	events map[time.Time]catatui.Style
}

// NewCalendarEventStore returns an empty store.
func NewCalendarEventStore() CalendarEventStore {
	return CalendarEventStore{events: make(map[time.Time]catatui.Style, 4)}
}

// CalendarEventStoreToday returns a store with the current local date styled.
func CalendarEventStoreToday(style catatui.Style) CalendarEventStore {
	s := NewCalendarEventStore()
	s.Add(time.Now(), style)
	return s
}

// Add records a style for a date. Adding a date twice keeps the last style.
func (s *CalendarEventStore) Add(date time.Time, style catatui.Style) {
	if s.events == nil {
		s.events = make(map[time.Time]catatui.Style, 4)
	}
	// to simplify style nonsense, last write wins
	s.events[calendarDate(date)] = style
}

// GetStyle implements DateStyler, returning the empty style for a date that
// was never added.
func (s CalendarEventStore) GetStyle(date time.Time) catatui.Style {
	return s.events[calendarDate(date)]
}

// Monthly draws a calendar for the month containing its date, one week per
// row with Sunday first.
//
// Days are drawn in the default style unless ShowSurrounding is set, in which
// case days outside the month use that style, or the DateStyler returns a
// style for the day.
//
//	cal := widgets.NewMonthly(time.Now(), widgets.CalendarEventStoreToday(
//		catatui.NewStyle().AddModifier(catatui.ModifierBold))).
//		ShowMonthHeader(catatui.NewStyle()).
//		ShowWeekdaysHeader(catatui.NewStyle()).
//		Block(widgets.Bordered())
//	f.RenderWidget(cal, area)
type Monthly struct {
	displayDate        time.Time
	events             DateStyler
	showSurrounding    catatui.Style
	hasShowSurrounding bool
	showWeekday        catatui.Style
	hasShowWeekday     bool
	showMonth          catatui.Style
	hasShowMonth       bool
	defaultStyle       catatui.Style
	block              Block
	hasBlock           bool
}

// NewMonthly returns a calendar for the month containing date, styling days
// with events. A nil events styles nothing.
func NewMonthly(date time.Time, events DateStyler) Monthly {
	return Monthly{displayDate: calendarDate(date), events: events}
}

// ShowSurrounding returns a copy of m that also fills the slots for days not
// in the month, so every row is complete. The days are drawn in the given
// style, patched with the event style for the date if there is one.
func (m Monthly) ShowSurrounding(style catatui.Style) Monthly {
	m.showSurrounding, m.hasShowSurrounding = style, true
	return m
}

// ShowWeekdaysHeader returns a copy of m with a header of weekday
// abbreviations in the given style.
func (m Monthly) ShowWeekdaysHeader(style catatui.Style) Monthly {
	m.showWeekday, m.hasShowWeekday = style, true
	return m
}

// ShowMonthHeader returns a copy of m with a header showing the month and
// year in the given style.
func (m Monthly) ShowMonthHeader(style catatui.Style) Monthly {
	m.showMonth, m.hasShowMonth = style, true
	return m
}

// DefaultStyle returns a copy of m drawing otherwise unstyled dates in the
// given style.
func (m Monthly) DefaultStyle(style catatui.Style) Monthly { m.defaultStyle = style; return m }

// Block returns a copy of m drawn inside the given block.
func (m Monthly) Block(b Block) Monthly { m.block, m.hasBlock = b, true; return m }

// Width returns the width required to render the calendar.
func (m Monthly) Width() uint16 {
	const (
		daysPerWeek uint16 = 7
		gutterWidth uint16 = 1
		dayWidth    uint16 = 2
	)

	width := daysPerWeek * (gutterWidth + dayWidth)
	if m.hasBlock {
		left, right := blockHorizontalSpace(m.block)
		width = catatui.SatAdd(catatui.SatAdd(width, left), right)
	}
	return width
}

// Height returns the height required to render the calendar.
func (m Monthly) Height() uint16 {
	height := uint16(sundayBasedWeeks(m.displayDate))
	if m.hasShowMonth {
		height = catatui.SatAdd(height, 1)
	}
	if m.hasShowWeekday {
		height = catatui.SatAdd(height, 1)
	}

	if m.hasBlock {
		top, bottom := blockVerticalSpace(m.block)
		height = catatui.SatAdd(catatui.SatAdd(height, top), bottom)
	}

	return height
}

// defaultBg returns a style with only the background from the default style.
func (m Monthly) defaultBg() catatui.Style {
	if bg := m.defaultStyle.GetBg(); bg.IsSet() {
		return catatui.NewStyle().Bg(bg)
	}
	return catatui.NewStyle()
}

// eventStyle asks the DateStyler for a date, tolerating a nil styler.
func (m Monthly) eventStyle(date time.Time) catatui.Style {
	if m.events == nil {
		return catatui.NewStyle()
	}
	return m.events.GetStyle(date)
}

// formatDate is where all the logic to style a date lives.
func (m Monthly) formatDate(date time.Time) catatui.Span {
	if date.Month() == m.displayDate.Month() {
		return catatui.NewStyledSpan(
			fmt.Sprintf("%2d", date.Day()),
			m.defaultStyle.Patch(m.eventStyle(date)),
		)
	}
	if !m.hasShowSurrounding {
		return catatui.NewStyledSpan("  ", m.defaultBg())
	}
	style := m.defaultStyle.Patch(m.showSurrounding).Patch(m.eventStyle(date))
	return catatui.NewStyledSpan(fmt.Sprintf("%2d", date.Day()), style)
}

// Render draws the calendar.
func (m Monthly) Render(area catatui.Rect, buf *catatui.Buffer) {
	inner := area
	if m.hasBlock {
		m.block.Render(area, buf)
		inner = m.block.Inner(area)
	}
	m.renderMonthly(inner, buf)
}

func (m Monthly) renderMonthly(area catatui.Rect, buf *catatui.Buffer) {
	var monthRows, weekdayRows uint16
	if m.hasShowMonth {
		monthRows = 1
	}
	if m.hasShowWeekday {
		weekdayRows = 1
	}
	rows := catatui.VerticalLayout(
		catatui.Length(monthRows),
		catatui.Length(weekdayRows),
		catatui.Fill(1),
	).Split(area)
	monthHeader, daysHeader, daysArea := rows[0], rows[1], rows[2]

	// Draw the month name and year
	if m.hasShowMonth {
		catatui.LineFromStyledString(
			fmt.Sprintf("%s %d", m.displayDate.Month(), m.displayDate.Year()),
			m.showMonth,
		).Centered().Render(monthHeader, buf)
	}

	// Draw days of week
	if m.hasShowWeekday {
		catatui.NewStyledSpan(" Su Mo Tu We Th Fr Sa", m.showWeekday).Render(daysHeader, buf)
	}

	// Set the start of the calendar to the Sunday before the 1st (or the
	// Sunday of the first)
	firstOfMonth := time.Date(m.displayDate.Year(), m.displayDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	currDay := firstOfMonth.AddDate(0, 0, -int(firstOfMonth.Weekday()))
	nextMonth := firstOfMonth.AddDate(0, 1, 0).Month()

	y := daysArea.Y
	// go through all the weeks containing a day in the target month.
	for currDay.Month() != nextMonth {
		spans := make([]catatui.Span, 0, 14)
		for i := range 7 {
			// Draw the gutter. Do it here so we can avoid worrying about
			// styling the ' ' in the formatDate method
			if i == 0 {
				spans = append(spans, catatui.NewStyledSpan(" ", catatui.NewStyle()))
			} else {
				spans = append(spans, catatui.NewStyledSpan(" ", m.defaultBg()))
			}
			spans = append(spans, m.formatDate(currDay))
			currDay = currDay.AddDate(0, 0, 1)
		}
		if buf.Area.Height > y {
			buf.SetLine(daysArea.X, y, catatui.NewLine(spans...), area.Width)
		}
		y = catatui.SatAdd(y, 1)
	}
}

// sundayBasedWeek is the 0-based week of the year, counting weeks from
// Sunday; week 0 is the (possibly partial) week before the first Sunday. It
// matches the time crate's Date::sunday_based_week.
func sundayBasedWeek(date time.Time) int {
	return (date.YearDay() - int(date.Weekday()) + 6) / 7
}

// sundayBasedWeeks computes how many Sunday-based week rows are needed to
// render the month of date.
//
// It mirrors the rendering logic by taking the difference between the first
// and last day's Sunday-based week numbers (inclusive).
func sundayBasedWeeks(date time.Time) uint8 {
	firstOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	firstWeek := sundayBasedWeek(firstOfMonth)
	lastWeek := sundayBasedWeek(lastOfMonth)
	return uint8(max(lastWeek-firstWeek, 0) + 1)
}

// blockHorizontalSpace is the left and right space a block takes up: its
// padding plus one column for each vertical border. It is ratatui's
// Block::horizontal_space.
func blockHorizontalSpace(b Block) (left, right uint16) {
	left = b.padding.Left
	if b.borders.Contains(BordersLeft) {
		left = catatui.SatAdd(left, 1)
	}
	right = b.padding.Right
	if b.borders.Contains(BordersRight) {
		right = catatui.SatAdd(right, 1)
	}
	return left, right
}

// blockVerticalSpace is the top and bottom space a block takes up: its padding
// plus one row for each horizontal border or title edge. It is ratatui's
// Block::vertical_space.
func blockVerticalSpace(b Block) (top, bottom uint16) {
	top = b.padding.Top
	if b.borders.Contains(BordersTop) || b.hasTitleAt(TitleTop) {
		top = catatui.SatAdd(top, 1)
	}
	bottom = b.padding.Bottom
	if b.borders.Contains(BordersBottom) || b.hasTitleAt(TitleBottom) {
		bottom = catatui.SatAdd(bottom, 1)
	}
	return top, bottom
}

var (
	_ catatui.Widget = Monthly{}
	_ DateStyler     = CalendarEventStore{}
)
