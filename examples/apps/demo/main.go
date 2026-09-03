// Command demo is the original tui-rs demo, the one the ratatui README shows:
// three tabs of gauges, lists, charts, a table and a world map, all animating
// on a tick.
//
//	go run ./examples/apps/demo
//	go run ./examples/apps/demo -tick-rate 100 -unicode=false
//
// Arrow keys or h/j/k/l move between tabs and down the list, t toggles the
// chart, q quits.
//
// Port of examples/apps/demo @ ratatui-v0.30.2
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	tickRate := flag.Int("tick-rate", 250, "milliseconds between ticks")
	unicode := flag.Bool("unicode", true, "use the finer unicode symbols")
	flag.Parse()

	if err := run(time.Duration(*tickRate)*time.Millisecond, *unicode); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run draws the demo until it is quit, ticking the animation along in between
// key presses.
func run(tickRate time.Duration, enhancedGraphics bool) error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	a := newApp("Catatui Demo", enhancedGraphics)

	for !a.shouldQuit {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, a) }); err != nil {
			return err
		}

		// Whichever comes first: something from the keyboard, or the next
		// tick. Nothing else moves the demo along.
		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			handle(a, ev)
		case <-ticker.C:
			a.onTick()
		}
	}
	return nil
}

// handle applies one event to the app.
func handle(a *app, ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.onLeft()
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.onDown()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.onUp()
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.onRight()
	case ev.IsCtrl('c'):
		a.shouldQuit = true
	case ev.Key == term.KeyRune:
		a.onKey(ev.Rune)
	}
}
