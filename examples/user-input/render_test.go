package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the example in both modes, empty and full, at sizes from
// nothing to bigger than a screen. Rendering outside the area given panics in
// catatui, so this is what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	apps := []*app{
		{},
		{mode: modeEditing, input: []rune("hello"), cursor: 5},
		{mode: modeEditing, input: []rune("日本語テキスト"), cursor: 3},
		{mode: modeNormal, messages: []string{"first", "second", "third"}},
	}
	for i, a := range apps {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
			if err := terminal.Draw(func(f *catatui.Frame) { a.render(f) }); err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
		}
	}
}

// TestEditing types a message, edits it in the middle, and files it away.
func TestEditing(t *testing.T) {
	a := &app{}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'e'})
	if a.mode != modeEditing {
		t.Fatalf("e did not start editing")
	}
	for _, r := range "helo" {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r})
	}
	// Back up over the "lo" and put the missing l in.
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'l'})
	if got := string(a.input); got != "hello" {
		t.Errorf("input is %q, want %q", got, "hello")
	}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
	if len(a.messages) != 1 || a.messages[0] != "hello" {
		t.Errorf("messages are %q, want [hello]", a.messages)
	}
	if len(a.input) != 0 || a.cursor != 0 {
		t.Errorf("input is %q with the cursor at %d, want it emptied", string(a.input), a.cursor)
	}
}

// TestCursorStaysInRange checks the cursor stops at both ends of the input
// rather than running off it, which is what would index out of bounds.
func TestCursorStaysInRange(t *testing.T) {
	a := &app{mode: modeEditing, input: []rune("hi")}
	for range 10 {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRight})
	}
	if a.cursor != len(a.input) {
		t.Errorf("cursor at %d, want it to stop at %d", a.cursor, len(a.input))
	}
	for range 10 {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	}
	if a.cursor != 0 {
		t.Errorf("cursor at %d, want it to stop at 0", a.cursor)
	}
	// Backspace at the start of the input has nothing to delete.
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyBackspace})
	if got := string(a.input); got != "hi" {
		t.Errorf("input is %q, want it left alone", got)
	}
}
