// Command panic shows what a panic does to the terminal, with catatui's
// recovery in place and without it.
//
//	go run ./examples/apps/panic
//
// Press p to panic, e to fail with an error, h to start again with the
// recovery off, q to quit.
//
// A panic while the terminal is in raw mode on the alternate screen leaves the
// shell unusable and the stack trace scrawled across a screen that is still in
// graphics mode. Deferring term.RecoverAndRestore puts every terminal Init
// opened back the way it was and then re-panics, so the trace still reaches the
// user and the shell still works afterwards.
//
// Go's recovery is a deferred call rather than a process-wide hook, so it is
// chosen when the function starts and cannot be taken away while it runs. That
// is why h starts the program over instead of flipping a flag, which is what
// ratatui's example does with std::panic::take_hook.
//
// Port of examples/apps/panic @ ratatui-v0.30.2
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	// Start with the recovery in place. Pressing h comes back here and runs
	// the app again without it, so the two can be compared.
	recovery := true
	for {
		again, err := run(recovery)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if !again {
			return
		}
		recovery = false
	}
}

// run draws the app until it is quit, and reports whether to start again with
// the recovery off.
func run(recovery bool) (again bool, err error) {
	if recovery {
		// This one line is the whole mechanism, and deferring it before Init
		// is deliberate: a panic inside Init itself is then covered too.
		defer term.RecoverAndRestore()
	}

	terminal, restore, err := term.Init()
	if err != nil {
		return false, err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, recovery) }); err != nil {
			return false, err
		}
		ev, ok := <-events.Events()
		if !ok {
			return false, events.Err()
		}
		if ev.Kind != term.EventKey {
			continue
		}
		switch {
		case ev.IsRune('p'):
			panic("intentional demo panic")
		case ev.IsRune('e'):
			// Returning an error is the ordinary path: the deferred restore
			// has already run by the time main prints it, so the message
			// lands on a terminal that works.
			return false, errors.New("intentional demo error")
		case ev.IsRune('h'):
			return true, nil
		case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return false, nil
		}
	}
}

// render draws the instructions and which way the recovery is set.
func render(f *catatui.Frame, recovery bool) {
	state := "disabled"
	if recovery {
		state = "enabled"
	}

	text := catatui.NewText(
		catatui.LineFromString("Panic recovery is currently: "+state),
		catatui.LineFromString(""),
		catatui.LineFromString("Press `p` to cause a panic"),
		catatui.LineFromString("Press `e` to cause an error"),
		catatui.LineFromString("Press `h` to start again without the recovery"),
		catatui.LineFromString("Press `q` to quit"),
		catatui.LineFromString(""),
		catatui.LineFromString("Panicking without the recovery leaves the terminal in raw mode on"),
		catatui.LineFromString("the alternate screen, and you will likely have to run `reset`"),
		catatui.LineFromString(""),
		catatui.LineFromString("Try it first with the recovery on, and then with it off, to see"),
		catatui.LineFromString("the difference"),
	)

	f.RenderWidget(
		widgets.NewParagraphFromText(text).
			Block(widgets.Bordered().Title("Panic Handler Demo")).
			Centered(),
		f.Area())
}
