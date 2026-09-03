// Command minimal is the smallest catatui program that draws something: set up
// the terminal, draw one frame, quit on the first key.
//
//	go run ./examples/apps/minimal
//
// Any key quits.
//
// There are many ways to structure an application loop and this is not meant to
// be prescriptive. hello is the one to read next: it has a real event loop, a
// layout, and drawing straight into the buffer.
//
// Port of examples/apps/minimal @ ratatui-v0.30.2
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

// render draws the whole frame. A Span is a Widget in its own right, so a bare
// string needs no widget wrapped around it — ratatui renders a `&str` here for
// the same reason.
func render(f *catatui.Frame) {
	f.RenderWidget(catatui.NewSpan("Hello World!"), f.Area())
}
