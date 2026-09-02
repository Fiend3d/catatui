// Port of the App loop in examples/apps/flex/src/main.rs @ ratatui-v0.30.2

package main

import (
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// run draws the demo and waits for a key between frames. Nothing here animates,
// so a blocking read is all it takes.
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
		a.down()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.up()
	case ev.IsRune('g'), ev.IsKey(term.KeyHome):
		a.scrollOffset = 0
	case ev.IsRune('G'), ev.IsKey(term.KeyEnd):
		a.scrollOffset = maxScrollOffset()
	case ev.IsRune('+'):
		a.spacing = catatui.SatAdd(a.spacing, 1)
	case ev.IsRune('-'):
		a.spacing = catatui.SatSub(a.spacing, 1)
	}
}

// nextTab moves to the next flex mode, stopping at the last one.
func (a *app) nextTab() {
	if int(a.selectedTab) < len(tabs)-1 {
		a.selectedTab++
	}
}

// previousTab moves to the previous flex mode, stopping at the first.
func (a *app) previousTab() {
	if a.selectedTab > 0 {
		a.selectedTab--
	}
}

func (a *app) up() { a.scrollOffset = catatui.SatSub(a.scrollOffset, 1) }

func (a *app) down() {
	a.scrollOffset = catatui.MinU16(catatui.SatAdd(a.scrollOffset, 1), maxScrollOffset())
}
