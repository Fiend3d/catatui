package main

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// sampleEvents is one of each kind, which is what describe has to cope with.
var sampleEvents = []term.Event{
	{Kind: term.EventKey, Key: term.KeyRune, Rune: 'a'},
	{Kind: term.EventKey, Key: term.KeyEnter, Mods: term.ModCtrl},
	{Kind: term.EventKey, Key: term.KeyF7},
	{Kind: term.EventMouse, MouseKind: term.MouseDown, Button: term.MouseButtonLeft, X: 3, Y: 4},
	{Kind: term.EventMouse, MouseKind: term.MouseScrollUp, X: 1, Y: 2, Mods: term.ModShift},
	{Kind: term.EventResize, Size: catatui.Size{Width: 80, Height: 24}},
	{Kind: term.EventPaste, Text: "pasted"},
	{Kind: term.EventFocus, Focused: true},
	{Kind: term.EventFocus},
}

// TestRender draws the example empty and full, at sizes from nothing to bigger
// than a screen. Rendering outside the area given panics in catatui, so this is
// what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, recent := range [][]term.Event{nil, sampleEvents} {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("%d events, %dx%d: %v", len(recent), size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { render(f, recent) })
			if err != nil {
				t.Fatalf("%d events, %dx%d: %v", len(recent), size[0], size[1], err)
			}
		}
	}
}

// TestDescribeEveryKind checks each kind of event gets a line of its own, and
// that none of them falls through to the catch-all.
func TestDescribeEveryKind(t *testing.T) {
	for _, ev := range sampleEvents {
		got := describe(ev)
		if got == "" || strings.Contains(got, "unknown event") {
			t.Errorf("describe(%+v) = %q", ev, got)
		}
		if strings.Contains(got, "\n") {
			t.Errorf("describe(%+v) = %q, want one line", ev, got)
		}
	}
	if got, want := describe(sampleEvents[1]), "key Enter [ctrl]"; got != want {
		t.Errorf("describe of ctrl-enter = %q, want %q", got, want)
	}
}

// TestShouldExitOnlyOnQ checks the loop ends on q and nothing else, since the
// key is read back out of the recorded events rather than handled directly.
func TestShouldExitOnlyOnQ(t *testing.T) {
	if shouldExit(sampleEvents) {
		t.Errorf("exited on events with no q in them")
	}
	withQ := append([]term.Event{{Kind: term.EventKey, Key: term.KeyRune, Rune: 'q'}}, sampleEvents...)
	if !shouldExit(withQ) {
		t.Errorf("did not exit on q")
	}
}

// TestLogLevelDefaultsToDebug checks the env var is read, and that an unset or
// unreadable one still leaves logging on.
func TestLogLevelDefaultsToDebug(t *testing.T) {
	for _, tc := range []struct {
		env  string
		want string
	}{
		{"", "DEBUG"},
		{"nonsense", "DEBUG"},
		{"INFO", "INFO"},
		{"DEBUG-4", "DEBUG-4"},
	} {
		t.Setenv("CATATUI_LOG", tc.env)
		if got := logLevel().String(); got != tc.want {
			t.Errorf("CATATUI_LOG=%q gives level %s, want %s", tc.env, got, tc.want)
		}
	}
}
