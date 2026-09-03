// Command constraints shows what each kind of Constraint does, a tab per kind.
//
//	go run ./examples/apps/constraints
//
// h/l or the left and right arrows change tab, j/k or up and down scroll, g/G
// jump to the ends, q quits.
//
// Where the flex example varies the Flex mode over one set of constraints, this
// one varies the constraints: each tab stacks up the cases that show what its
// kind of constraint gives way to.
//
// Port of examples/apps/constraints @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run draws the app and waits for a key between frames. Nothing animates, so a
// blocking read is all it takes.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{}

	for !a.quit {
		// The app is a widget in its own right, as in ratatui, so drawing a
		// frame is one RenderWidget call.
		if err := terminal.Draw(func(f *catatui.Frame) {
			f.RenderWidget(a, f.Area())
		}); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		a.handle(ev)
	}
	return nil
}

// app is which tab is shown and how far it is scrolled.
type app struct {
	selectedTab  tab
	scrollOffset uint16
	quit         bool
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.nextTab()
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.previousTab()
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.scrollOffset = min(catatui.SatAdd(a.scrollOffset, 1), a.maxScrollOffset())
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.scrollOffset = catatui.SatSub(a.scrollOffset, 1)
	case ev.IsRune('g'), ev.IsKey(term.KeyHome):
		a.scrollOffset = 0
	case ev.IsRune('G'), ev.IsKey(term.KeyEnd):
		a.scrollOffset = a.maxScrollOffset()
	}
}

// nextTab moves to the next kind of constraint, stopping at the last, and
// starts it from the top.
func (a *app) nextTab() {
	if int(a.selectedTab) < len(tabs)-1 {
		a.selectedTab++
		a.scrollOffset = 0
	}
}

// previousTab moves to the previous kind, stopping at the first.
func (a *app) previousTab() {
	if a.selectedTab > 0 {
		a.selectedTab--
		a.scrollOffset = 0
	}
}

// maxScrollOffset stops the scroll with the last example of the current tab in
// view.
//
// ratatui hard-codes a count per tab, and the one for Length has drifted from
// the number of examples it actually draws; counting them is the same answer
// everywhere else and the right one there.
func (a *app) maxScrollOffset() uint16 {
	return uint16(len(tabs[a.selectedTab].examples)-1) * exampleHeight
}
