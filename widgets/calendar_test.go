// Tests ported from ratatui-widgets/src/calendar.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
)

// date builds a calendar date in the way ratatui's tests call
// Date::from_calendar_date.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestCalendarEventStore(t *testing.T) {
	aDate, aStyle := date(2023, time.January, 1), catatui.NewStyle()
	bDate, bStyle := date(2023, time.January, 2), catatui.NewStyle().Bg(catatui.ColorRed).Fg(catatui.ColorBlue)
	s := NewCalendarEventStore()
	s.Add(bDate, bStyle)

	if got := s.GetStyle(aDate); got != aStyle {
		t.Errorf("date not added to the styler should look up as the empty style, got %+v", got)
	}
	if got := s.GetStyle(bDate); got != bStyle {
		t.Errorf("date added to styler should return the provided style, got %+v", got)
	}
}

// TestCalendarEventStoreIgnoresClockAndZone pins the Go-specific rule that a
// date is matched by year, month and day only.
func TestCalendarEventStoreIgnoresClockAndZone(t *testing.T) {
	style := catatui.NewStyle().Fg(catatui.ColorRed)
	s := NewCalendarEventStore()
	s.Add(time.Date(2023, time.January, 2, 13, 45, 0, 0, time.FixedZone("x", 3600)), style)
	if got := s.GetStyle(date(2023, time.January, 2)); got != style {
		t.Errorf("lookup by date only should find the style, got %+v", got)
	}
}

func TestCalendarToday(t *testing.T) {
	CalendarEventStoreToday(catatui.NewStyle())
}

func TestCalendarRenderInMinimalBuffer(t *testing.T) {
	calendar := NewMonthly(date(1984, time.January, 1), NewCalendarEventStore())
	// This should not panic, even if the buffer is too small to render the
	// calendar.
	buf := renderToBuffer(calendar, 1, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestCalendarRenderInZeroSizeBuffer(t *testing.T) {
	calendar := NewMonthly(date(1984, time.January, 1), NewCalendarEventStore())
	// This should not panic, even if the buffer has zero size.
	renderToBuffer(calendar, 0, 0)
}

func TestCalendarWidthReflectsGridLayout(t *testing.T) {
	calendar := NewMonthly(date(2023, time.January, 1), NewCalendarEventStore())
	if got := calendar.Width(); got != 21 {
		t.Errorf("Width() = %d, want 21", got)
	}
}

func TestCalendarHeightCountsWeeksAndHeaders(t *testing.T) {
	d := date(2015, time.February, 1)
	base := NewMonthly(d, NewCalendarEventStore())
	if got := base.Height(); got != 4 {
		t.Errorf("Height() = %d, want 4", got)
	}

	decorated := NewMonthly(d, NewCalendarEventStore()).
		ShowMonthHeader(catatui.NewStyle()).
		ShowWeekdaysHeader(catatui.NewStyle())
	if got := decorated.Height(); got != 6 {
		t.Errorf("Height() = %d, want 6", got)
	}
}

func TestCalendarDimensionsExamples(t *testing.T) {
	check := func(name string, got, want uint16) {
		t.Helper()
		if got != want {
			t.Errorf("%s: got %d, want %d", name, got, want)
		}
	}

	// Feb 2015 starts Sunday and spans 4 rows.
	feb2015 := date(2015, time.February, 1)
	cal := NewMonthly(feb2015, NewCalendarEventStore())
	check("4w base width", cal.Width(), 21)
	check("Feb 2015 rows", cal.Height(), 4)

	cal = NewMonthly(feb2015, NewCalendarEventStore()).
		ShowMonthHeader(catatui.NewStyle()).
		ShowWeekdaysHeader(catatui.NewStyle())
	check("Headers add 2 rows", cal.Height(), 6)

	block := Bordered().Padding(NewPadding(2, 3, 1, 2))
	cal = NewMonthly(feb2015, NewCalendarEventStore()).Block(block)
	check("Padding widens width", cal.Width(), 28)
	check("Padding grows height", cal.Height(), 9)

	// Feb 2024 starts Thursday and spans 5 rows.
	feb2024 := date(2024, time.February, 1)
	cal = NewMonthly(feb2024, NewCalendarEventStore())
	check("5w base width", cal.Width(), 21)
	check("Feb 2024 rows", cal.Height(), 5)

	cal = NewMonthly(feb2024, NewCalendarEventStore()).
		ShowMonthHeader(catatui.NewStyle()).
		ShowWeekdaysHeader(catatui.NewStyle())
	check("Headers add 2 rows (5w)", cal.Height(), 7)

	cal = NewMonthly(feb2024, NewCalendarEventStore()).Block(Bordered())
	check("Border adds 2 cols", cal.Width(), 23)
	check("Border adds 2 rows", cal.Height(), 7)

	// Apr 2023 starts Saturday and spans 6 rows.
	apr2023 := date(2023, time.April, 1)
	cal = NewMonthly(apr2023, NewCalendarEventStore())
	check("6w base width", cal.Width(), 21)
	check("Apr 2023 rows", cal.Height(), 6)

	cal = NewMonthly(apr2023, NewCalendarEventStore()).
		ShowMonthHeader(catatui.NewStyle()).
		ShowWeekdaysHeader(catatui.NewStyle())
	check("Headers add 2 rows (6w)", cal.Height(), 8)

	block = Bordered().Padding(NewPadding(1, 1, 1, 1))
	cal = NewMonthly(apr2023, NewCalendarEventStore()).Block(block)
	check("Symmetric padding width", cal.Width(), 25)
	check("Symmetric padding height", cal.Height(), 10)
}

func TestCalendarSundayBasedWeeksShapes(t *testing.T) {
	sundayStart := date(2015, time.February, 11)
	saturdayStart := date(2023, time.April, 9)
	leapYear := date(2024, time.February, 29)

	if got := sundayBasedWeeks(sundayStart); got != 4 {
		t.Errorf("sundayBasedWeeks(Feb 2015) = %d, want 4", got)
	}
	if got := sundayBasedWeeks(saturdayStart); got != 6 {
		t.Errorf("sundayBasedWeeks(Apr 2023) = %d, want 6", got)
	}
	if got := sundayBasedWeeks(leapYear); got != 5 {
		t.Errorf("sundayBasedWeeks(Feb 2024) = %d, want 5", got)
	}
}

// TestCalendarRenderMonth pins the full layout: headers, the Sunday-first
// grid, blank slots outside the month, and an event style on one day.
func TestCalendarRenderMonth(t *testing.T) {
	store := NewCalendarEventStore()
	event := catatui.NewStyle().Fg(catatui.ColorRed)
	store.Add(date(2024, time.February, 14), event)
	cal := NewMonthly(date(2024, time.February, 1), store).
		ShowMonthHeader(catatui.NewStyle()).
		ShowWeekdaysHeader(catatui.NewStyle())

	buf := renderToBuffer(cal, 21, 7)
	want := catatui.NewBufferWithStrings(
		"    February 2024    ",
		" Su Mo Tu We Th Fr Sa",
		"              1  2  3",
		"  4  5  6  7  8  9 10",
		" 11 12 13 14 15 16 17",
		" 18 19 20 21 22 23 24",
		" 25 26 27 28 29      ",
	)
	want.SetStyle(catatui.NewRect(10, 4, 2, 1), event)
	catatui.AssertBuffer(t, buf, want)
}

// TestCalendarShowSurrounding checks that the days of the neighbouring months
// fill the empty slots in the given style.
func TestCalendarShowSurrounding(t *testing.T) {
	surrounding := catatui.NewStyle().Fg(catatui.ColorDarkGray)
	cal := NewMonthly(date(2024, time.February, 1), NewCalendarEventStore()).
		ShowSurrounding(surrounding)

	buf := renderToBuffer(cal, 21, 5)
	want := catatui.NewBufferWithStrings(
		" 28 29 30 31  1  2  3",
		"  4  5  6  7  8  9 10",
		" 11 12 13 14 15 16 17",
		" 18 19 20 21 22 23 24",
		" 25 26 27 28 29  1  2",
	)
	for _, r := range []catatui.Rect{
		catatui.NewRect(1, 0, 2, 1), catatui.NewRect(4, 0, 2, 1),
		catatui.NewRect(7, 0, 2, 1), catatui.NewRect(10, 0, 2, 1),
		catatui.NewRect(16, 4, 2, 1), catatui.NewRect(19, 4, 2, 1),
	} {
		want.SetStyle(r, surrounding)
	}
	catatui.AssertBuffer(t, buf, want)
}
