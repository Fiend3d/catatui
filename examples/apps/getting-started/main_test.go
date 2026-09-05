package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func runeEvent(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

func TestCareForCat(t *testing.T) {
	a := newApp()
	selected := func(want int) {
		t.Helper()
		if got, ok := a.list.Selected(); !ok || got != want {
			t.Fatalf("selection = %d, %v; want %d", got, ok, want)
		}
	}
	selected(0)
	for _, ev := range []term.Event{runeEvent('k'), {Kind: term.EventKey, Key: term.KeyUp}} {
		a.handle(ev)
		selected(0)
	}
	for i := range 3 {
		a.handle(runeEvent(' '))
		if got := a.completed(); got != i+1 {
			t.Fatalf("completed = %d, want %d", got, i+1)
		}
		a.handle(runeEvent('j'))
		selected(min(i+1, 2))
	}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyDown})
	selected(2)
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
	if a.completed() != 2 || a.tasks[2].done {
		t.Fatal("Enter should undo completion of the selected task")
	}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyUp})
	selected(1)
	a.handle(runeEvent('r'))
	selected(0)
	if a.completed() != 0 || a.quit || a.list.Offset() != 0 {
		t.Fatal("reset should restore the initial state")
	}
}

func TestQuitAndUnrelatedEvents(t *testing.T) {
	for _, ev := range []term.Event{
		runeEvent('q'), {Kind: term.EventKey, Key: term.KeyEscape},
		{Kind: term.EventKey, Key: term.KeyRune, Rune: 'c', Mods: term.ModCtrl},
	} {
		a := newApp()
		a.handle(ev)
		if !a.quit {
			t.Errorf("%+v did not quit", ev)
		}
	}
	a := newApp()
	for _, ev := range []term.Event{
		{Kind: term.EventResize}, {Kind: term.EventPaste, Text: "q j "},
		{Kind: term.EventMouse}, runeEvent('z'),
		{Kind: term.EventKey, Key: term.KeyRune, Rune: 'r', Mods: term.ModCtrl},
	} {
		a.handle(ev)
	}
	i, _ := a.list.Selected()
	if a.quit || a.completed() != 0 || i != 0 {
		t.Fatal("unrelated events changed app state")
	}
}

func TestRenderAndResize(t *testing.T) {
	backend := catatui.NewTestBackend(80, 24)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}
	a := newApp()
	for completed := 0; completed <= 3; completed++ {
		for i := range a.tasks {
			a.tasks[i].done = i < completed
		}
		for _, size := range [][2]uint16{{80, 24}, {72, 18}, {71, 18}, {40, 18}, {39, 18}, {80, 17}, {10, 4}, {1, 1}, {0, 0}, {120, 30}} {
			t.Run(fmt.Sprintf("%d-done/%dx%d", completed, size[0], size[1]), func(t *testing.T) {
				backend.Resize(size[0], size[1])
				if err := terminal.Draw(a.draw); err != nil {
					t.Fatal(err)
				}
				text := backend.Buffer().String()
				if size[0] < 40 || size[1] < 18 {
					if size[0] >= 26 && size[1] >= 3 && !strings.Contains(text, "Resize to at least 40 x 18.") {
						t.Fatalf("missing resize hint:\n%s", text)
					}
					return
				}
				for _, want := range []string{"PURRFECT DAY", "Serve breakfast", "Prepare a cozy nap", "Space/Enter: tick", fmt.Sprintf("%d / 3 little wins", completed)} {
					if !strings.Contains(text, want) {
						t.Errorf("missing %q:\n%s", want, text)
					}
				}
				if strings.Count(text, "[x]") != completed {
					t.Errorf("wrong number of checked tasks:\n%s", text)
				}
				message := "A little care goes a long way."
				if completed == 3 {
					message = "One happy cat. Nicely done!"
				}
				if !strings.Contains(text, message) {
					t.Errorf("missing cat message:\n%s", text)
				}
			})
		}
	}
	// Redrawing after undoing a win must remove the previous celebration.
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
	if err := terminal.Draw(a.draw); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(backend.Buffer().String(), "One happy cat") {
		t.Fatal("celebration remained after undo")
	}
}

func TestTutorialFinalProgramMatchesExample(t *testing.T) {
	guide, err := os.ReadFile("../../../docs/getting-started.md")
	if err != nil {
		t.Fatal(err)
	}
	_, final, ok := strings.Cut(strings.ReplaceAll(string(guide), "\r\n", "\n"), "<!-- final-program -->\n```go\n")
	if !ok {
		t.Fatal("missing final program in tutorial")
	}
	final, _, ok = strings.Cut(final, "\n```")
	if !ok {
		t.Fatal("missing end of final program")
	}
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(final) != strings.TrimSpace(strings.ReplaceAll(string(source), "\r\n", "\n")) {
		t.Fatal("tutorial final program differs from runnable example")
	}
}
