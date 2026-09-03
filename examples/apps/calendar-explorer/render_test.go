package main

import (
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws every style at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	// A leap year, so February fills its rows exactly, and the 31st of a long
	// month, which is the date the month keys have to be careful with.
	dates := []time.Time{
		date(2024, time.February, 29),
		date(2024, time.January, 31),
		date(2025, time.December, 31),
	}
	for _, selected := range dates {
		for style := styleDefault; int(style) < len(calendarStyleNames); style++ {
			a := &app{selected: selected, style: style}
			for _, size := range sizes {
				terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
				if err != nil {
					t.Fatalf("%v, style %v, %dx%d: %v", selected, style, size[0], size[1], err)
				}
				if err := terminal.Draw(a.render); err != nil {
					t.Fatalf("%v, style %v, %dx%d: %v", selected, style, size[0], size[1], err)
				}
			}
		}
	}
}

// TestAddMonthsClampsTheDay checks moving by months keeps the day where it can
// and stops at the end of the month where it cannot, rather than rolling over
// into the month after.
func TestAddMonthsClampsTheDay(t *testing.T) {
	for _, tc := range []struct {
		from time.Time
		n    int
		want time.Time
	}{
		{date(2024, time.January, 31), 1, date(2024, time.February, 29)}, // leap year
		{date(2025, time.January, 31), 1, date(2025, time.February, 28)},
		{date(2025, time.March, 31), -1, date(2025, time.February, 28)},
		{date(2025, time.January, 15), 1, date(2025, time.February, 15)},
		{date(2025, time.December, 15), 1, date(2026, time.January, 15)},
		{date(2025, time.January, 15), -1, date(2024, time.December, 15)},
	} {
		if got := addMonths(tc.from, tc.n); !got.Equal(tc.want) {
			t.Errorf("%s plus %d months = %s, want %s",
				tc.from.Format("2006-01-02"), tc.n,
				got.Format("2006-01-02"), tc.want.Format("2006-01-02"))
		}
	}
}

// TestStyleCyclesRoundAndBack checks s reaches every style and comes back to
// the first, so none of them is unreachable.
func TestStyleCyclesRoundAndBack(t *testing.T) {
	a := &app{selected: date(2025, time.June, 1)}
	seen := map[calendarStyle]bool{a.style: true}
	for range len(calendarStyleNames) {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 's'})
		seen[a.style] = true
	}
	if len(seen) != len(calendarStyleNames) {
		t.Errorf("cycling reached %d of %d styles", len(seen), len(calendarStyleNames))
	}
	if a.style != styleDefault {
		t.Errorf("a full cycle ended on %v, want it back at %v", a.style, styleDefault)
	}
	for style := styleDefault; int(style) < len(calendarStyleNames); style++ {
		if style.String() == "Unknown" {
			t.Errorf("style %d has no name", style)
		}
	}
}

// TestKeysMoveTheDate checks each movement key moves by what it says.
func TestKeysMoveTheDate(t *testing.T) {
	start := date(2025, time.June, 10)
	for _, tc := range []struct {
		key  rune
		want time.Time
	}{
		{'h', date(2025, time.June, 9)},
		{'l', date(2025, time.June, 11)},
		{'k', date(2025, time.June, 3)},
		{'j', date(2025, time.June, 17)},
		{'n', date(2025, time.July, 10)},
		{'p', date(2025, time.May, 10)},
	} {
		a := &app{selected: start}
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: tc.key})
		if !a.selected.Equal(tc.want) {
			t.Errorf("%q moved to %s, want %s", tc.key,
				a.selected.Format("2006-01-02"), tc.want.Format("2006-01-02"))
		}
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}
