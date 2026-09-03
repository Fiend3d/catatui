package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the example with nothing selected, with the first item
// selected and with the last, at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, selected := range []int{-1, 0, 5} {
		a := newApp()
		if selected >= 0 {
			a.todoList.state.Select(selected)
		}
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("selected %d, %dx%d: %v", selected, size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
			if err != nil {
				t.Fatalf("selected %d, %dx%d: %v", selected, size[0], size[1], err)
			}
		}
	}
}

// TestToggleStatus checks that toggling turns an item over and back, and that
// it does nothing at all while nothing is selected.
func TestToggleStatus(t *testing.T) {
	a := newApp()
	before := a.todoList.items[0].status

	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'l'})
	if a.todoList.items[0].status != before {
		t.Errorf("toggled with nothing selected")
	}

	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'j'})
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'l'})
	if a.todoList.items[0].status != statusCompleted {
		t.Errorf("item 0 is %v, want it completed", a.todoList.items[0].status)
	}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
	if a.todoList.items[0].status != statusTodo {
		t.Errorf("item 0 is %v, want it back to todo", a.todoList.items[0].status)
	}
}

// TestSelectionStaysInRange checks the selection stops at both ends of the
// list, and that unselecting leaves nothing selected.
//
// The state is clamped when the list is rendered, not when a key is pressed:
// until then the number of items is not known. So this drives the app the way
// the event loop does, drawing a frame between key presses.
func TestSelectionStaysInRange(t *testing.T) {
	a := newApp()
	terminal, err := catatui.NewTerminal(catatui.NewTestBackend(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	press := func(ev term.Event) {
		t.Helper()
		a.handle(ev)
		if err := terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) }); err != nil {
			t.Fatal(err)
		}
	}

	for range len(a.todoList.items) + 5 {
		press(term.Event{Kind: term.EventKey, Key: term.KeyDown})
	}
	if i, ok := a.todoList.state.Selected(); !ok || i != len(a.todoList.items)-1 {
		t.Errorf("selected %d (%v), want the last item", i, ok)
	}
	for range len(a.todoList.items) + 5 {
		press(term.Event{Kind: term.EventKey, Key: term.KeyUp})
	}
	if i, ok := a.todoList.state.Selected(); !ok || i != 0 {
		t.Errorf("selected %d (%v), want the first item", i, ok)
	}
	press(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	if _, ok := a.todoList.state.Selected(); ok {
		t.Errorf("still selected after unselecting")
	}
	// Toggling with nothing selected must not index off the end of the list.
	press(term.Event{Kind: term.EventKey, Key: term.KeyRight})
}
