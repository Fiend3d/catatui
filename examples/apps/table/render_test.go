package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func keyRune(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

// TestRender draws the table in each colour scheme, at the top and bottom of
// the list, at sizes from nothing to bigger than a screen. Rendering outside
// the area given panics in catatui, so this is what keeps the example honest
// when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for color := range palettes {
		for _, selected := range []int{0, 3, len(people) - 1} {
			a := newApp()
			a.colorIndex = color
			a.selectRow(selected)
			a.state.SelectColumn(1)
			for _, size := range sizes {
				terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
				if err != nil {
					t.Fatalf("colour %d, row %d, %dx%d: %v", color, selected, size[0], size[1], err)
				}
				if err := terminal.Draw(a.render); err != nil {
					t.Fatalf("colour %d, row %d, %dx%d: %v", color, selected, size[0], size[1], err)
				}
			}
		}
	}
}

// TestConstraintLens is ratatui's own test for this function, with its data:
// the address is two lines and each is measured on its own, so the longest is
// the longer line rather than the whole string.
func TestConstraintLens(t *testing.T) {
	items := []person{
		{
			name:    "Emirhan Tala",
			address: "Cambridgelaan 6XX\n3584 XX Utrecht",
			email:   "tala.emirhan@gmail.com",
		},
		{
			name:    "thistextis26characterslong",
			address: "this line is 31 characters long\nbottom line is 33 characters long",
			email:   "thisemailis40caharacterslong@ratatui.com",
		},
	}
	if got, want := constraintLens(items), [3]uint16{26, 33, 40}; got != want {
		t.Errorf("constraintLens = %v, want %v", got, want)
	}
}

// TestRowsWrapAround checks the selection goes round both ways rather than
// stopping, which is what ratatui's example does.
func TestRowsWrapAround(t *testing.T) {
	a := newApp()
	for range len(a.items) {
		a.handle(keyRune('j'))
	}
	if i, _ := a.state.Selected(); i != 0 {
		t.Errorf("moving down past the last row selected %d, want it to wrap to 0", i)
	}

	a.handle(keyRune('k'))
	if i, _ := a.state.Selected(); i != len(a.items)-1 {
		t.Errorf("moving up from the first row selected %d, want the last", i)
	}
}

// TestScrollbarFollowsTheSelection checks the scrollbar is moved with the
// selection, scaled by the height of a row.
func TestScrollbarFollowsTheSelection(t *testing.T) {
	a := newApp()
	for range 3 {
		a.handle(keyRune('j'))
	}
	i, _ := a.state.Selected()
	if got, want := a.scrollState.GetPosition(), i*itemHeight; got != want {
		t.Errorf("the scrollbar is at %d for row %d, want %d", got, i, want)
	}
}

// TestShiftChangesTheColorScheme checks a shifted arrow cycles the colours
// rather than moving the selection, and that the plain arrow still moves it.
func TestShiftChangesTheColorScheme(t *testing.T) {
	a := newApp()

	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRight, Mods: term.ModShift})
	if a.colorIndex != 1 {
		t.Errorf("shift-right gave scheme %d, want 1", a.colorIndex)
	}
	if _, ok := a.state.SelectedColumn(); ok {
		t.Errorf("shift-right also moved the column selection")
	}

	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft, Mods: term.ModShift})
	if a.colorIndex != 0 {
		t.Errorf("shift-left gave scheme %d, want 0", a.colorIndex)
	}
	// And round the other way, off the front of the list.
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft, Mods: term.ModShift})
	if a.colorIndex != len(palettes)-1 {
		t.Errorf("shift-left from the first scheme gave %d, want the last", a.colorIndex)
	}

	// A capital L is what a terminal sends for shift-l, and does the same.
	a.colorIndex = 0
	a.handle(keyRune('L'))
	if a.colorIndex != 1 {
		t.Errorf("L gave scheme %d, want 1", a.colorIndex)
	}

	a.handle(keyRune('l'))
	if _, ok := a.state.SelectedColumn(); !ok {
		t.Errorf("a plain l did not move the column selection")
	}
}

// TestEveryPersonHasDetails checks the data is filled in, since the widths of
// the columns are measured from it.
func TestEveryPersonHasDetails(t *testing.T) {
	for i, p := range people {
		if p.name == "" || p.address == "" || p.email == "" {
			t.Errorf("person %d is missing something: %+v", i, p)
		}
		if i > 0 && people[i-1].name > p.name {
			t.Errorf("person %d (%s) is out of order after %s", i, p.name, people[i-1].name)
		}
	}
}
