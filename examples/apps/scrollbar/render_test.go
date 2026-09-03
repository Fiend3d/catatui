package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the panes unscrolled and scrolled a long way past the end of
// the text, at sizes from nothing to bigger than a screen. Rendering outside the
// area given panics in catatui, so this is what keeps the example honest when
// the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, scroll := range []int{0, 5, 500} {
		a := &app{verticalScroll: scroll, horizontalScroll: scroll}
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("scroll %d, %dx%d: %v", scroll, size[0], size[1], err)
			}
			if err := terminal.Draw(a.render); err != nil {
				t.Fatalf("scroll %d, %dx%d: %v", scroll, size[0], size[1], err)
			}
		}
	}
}

// TestScrollStopsAtTheStart checks neither scroll runs below zero, which would
// wrap the uint16 the paragraph is scrolled by.
func TestScrollStopsAtTheStart(t *testing.T) {
	a := &app{}
	for range 5 {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyUp})
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	}
	if a.verticalScroll != 0 || a.horizontalScroll != 0 {
		t.Errorf("scrolled to %d,%d, want both to stop at 0", a.horizontalScroll, a.verticalScroll)
	}
}

// TestScrollFollowsTheState checks the scrollbar state is moved along with the
// scroll, since it is what puts the thumb in the right place.
func TestScrollFollowsTheState(t *testing.T) {
	a := &app{}
	for range 3 {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyDown})
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRight})
	}
	if got := a.verticalState.GetPosition(); got != a.verticalScroll {
		t.Errorf("the vertical thumb is at %d and the text at %d", got, a.verticalScroll)
	}
	if got := a.horizontalState.GetPosition(); got != a.horizontalScroll {
		t.Errorf("the horizontal thumb is at %d and the text at %d", got, a.horizontalScroll)
	}
}

// TestMaskedHidesEveryCharacter checks the password is covered one mask
// character per grapheme, as ratatui's Masked does.
func TestMaskedHidesEveryCharacter(t *testing.T) {
	if got, want := masked("password", '*'), "********"; got != want {
		t.Errorf("masked = %q, want %q", got, want)
	}
	// A flag is one grapheme cluster made of two runes, and gets one star.
	if got, want := masked("🇬🇧", '*'), "*"; got != want {
		t.Errorf("masked flag = %q, want %q", got, want)
	}
}
