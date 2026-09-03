package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// keyRune is a press of a character key, which is all these tests send.
func keyRune(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

// TestRender draws every tab at sizes from nothing to bigger than a screen, at
// both ends of its scroll. Rendering outside the area given panics in catatui,
// so this is what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for selected := range tabs {
		a := &app{selectedTab: tab(selected)}
		for _, offset := range []uint16{0, 3, a.maxScrollOffset()} {
			a.scrollOffset = offset
			for _, size := range sizes {
				terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
				if err != nil {
					t.Fatalf("tab %d, offset %d, %dx%d: %v", selected, offset, size[0], size[1], err)
				}
				err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
				if err != nil {
					t.Fatalf("tab %d, offset %d, %dx%d: %v", selected, offset, size[0], size[1], err)
				}
			}
		}
	}
}

// TestEveryTabHasExamples checks no tab is empty, since maxScrollOffset counts
// back from the number of examples and would underflow on an empty one.
func TestEveryTabHasExamples(t *testing.T) {
	for i, tb := range tabs {
		if len(tb.examples) == 0 {
			t.Errorf("tab %d (%s) has no examples", i, tb.name)
		}
		if tb.name == "" {
			t.Errorf("tab %d has no name", i)
		}
	}
}

// TestScrollStaysInRange checks the scroll stops at the last example and at the
// first, and that changing tab starts the new one from the top.
func TestScrollStaysInRange(t *testing.T) {
	a := &app{}
	for range int(a.maxScrollOffset()) + 10 {
		a.handle(keyRune('j'))
	}
	if got, want := a.scrollOffset, a.maxScrollOffset(); got != want {
		t.Errorf("scrolled to %d, want it to stop at %d", got, want)
	}
	for range int(a.maxScrollOffset()) + 10 {
		a.handle(keyRune('k'))
	}
	if a.scrollOffset != 0 {
		t.Errorf("scrolled back to %d, want 0", a.scrollOffset)
	}

	a.handle(keyRune('G'))
	a.handle(keyRune('l'))
	if a.scrollOffset != 0 {
		t.Errorf("changing tab left the scroll at %d, want it back at the top", a.scrollOffset)
	}
}

// TestTabsStayInRange checks the tabs stop at both ends rather than wrapping,
// which is what ratatui's example does.
func TestTabsStayInRange(t *testing.T) {
	a := &app{}
	for range len(tabs) + 5 {
		a.handle(keyRune('l'))
	}
	if got, want := a.selectedTab, tab(len(tabs)-1); got != want {
		t.Errorf("selected tab %d, want %d", got, want)
	}
	for range len(tabs) + 5 {
		a.handle(keyRune('h'))
	}
	if a.selectedTab != 0 {
		t.Errorf("selected tab %d, want 0", a.selectedTab)
	}
}
