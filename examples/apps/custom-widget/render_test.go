package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func keyRune(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

// TestRender draws the buttons in every state at sizes from nothing to bigger
// than a screen. Rendering outside the area given panics in catatui, so this is
// what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	states := []buttonState{stateNormal, stateSelected, stateActive}
	for _, state := range states {
		a := &app{states: [3]buttonState{state, stateSelected, stateActive}}
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("state %d, %dx%d: %v", state, size[0], size[1], err)
			}
			if err := terminal.Draw(a.render); err != nil {
				t.Fatalf("state %d, %dx%d: %v", state, size[0], size[1], err)
			}
		}
	}
}

// TestButtonFitsItsArea draws one button at every size up to a comfortable one.
// The label is centred by hand rather than by a layout, so this is where an
// off-by-one in that arithmetic would show up as a panic.
func TestButtonFitsItsArea(t *testing.T) {
	for width := uint16(0); width <= 20; width++ {
		for height := uint16(0); height <= 5; height++ {
			area := catatui.NewRect(0, 0, width, height)
			buf := catatui.NewBuffer(area)
			for _, label := range []string{"", "Red", "A rather long label"} {
				button{label: label, theme: redTheme, state: stateActive}.Render(area, buf)
			}
		}
	}
}

// TestSelectionStaysInRange checks the keys stop at both ends rather than
// running off the row of buttons.
func TestSelectionStaysInRange(t *testing.T) {
	a := &app{states: [3]buttonState{stateSelected, stateNormal, stateNormal}}
	for range len(a.states) + 3 {
		a.handle(keyRune('l'))
	}
	if a.selected != len(a.states)-1 {
		t.Errorf("selected %d, want it to stop at %d", a.selected, len(a.states)-1)
	}
	for range len(a.states) + 3 {
		a.handle(keyRune('h'))
	}
	if a.selected != 0 {
		t.Errorf("selected %d, want it to stop at 0", a.selected)
	}
}

// TestActiveButtonStaysActive checks moving away from a pressed button leaves
// it pressed, and that only the selected button toggles.
func TestActiveButtonStaysActive(t *testing.T) {
	a := &app{states: [3]buttonState{stateSelected, stateNormal, stateNormal}}
	a.handle(keyRune(' '))
	if a.states[0] != stateActive {
		t.Fatalf("space left button 0 in state %d", a.states[0])
	}

	a.handle(keyRune('l'))
	if a.states[0] != stateActive {
		t.Errorf("moving away un-pressed button 0 (state %d)", a.states[0])
	}
	if a.states[1] != stateSelected {
		t.Errorf("button 1 is in state %d, want it selected", a.states[1])
	}

	a.handle(keyRune(' '))
	a.handle(keyRune(' '))
	if a.states[1] != stateNormal {
		t.Errorf("toggling twice left button 1 in state %d", a.states[1])
	}
	if a.states[0] != stateActive {
		t.Errorf("toggling button 1 changed button 0 to state %d", a.states[0])
	}
}

// TestMouseSelectsTheButtonUnderIt checks the pointer picks the button its
// column falls in, and that anything past the last one still lands on it.
func TestMouseSelectsTheButtonUnderIt(t *testing.T) {
	for _, tc := range []struct {
		x    uint16
		want int
	}{
		{0, 0}, {14, 0}, {15, 1}, {29, 1}, {30, 2}, {44, 2}, {200, 2},
	} {
		a := &app{states: [3]buttonState{stateSelected, stateNormal, stateNormal}}
		a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseMove, X: tc.x})
		if a.selected != tc.want {
			t.Errorf("column %d selected button %d, want %d", tc.x, a.selected, tc.want)
		}
	}

	// A click presses whatever the pointer selected.
	a := &app{states: [3]buttonState{stateSelected, stateNormal, stateNormal}}
	a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseMove, X: 20})
	a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseDown, Button: term.MouseButtonLeft, X: 20})
	if a.states[1] != stateActive {
		t.Errorf("clicking button 1 left it in state %d", a.states[1])
	}
}
