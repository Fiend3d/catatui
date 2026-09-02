// Command logo draws the catatui logo in an inline viewport: three rows at the
// shell's cursor, leaving the scrollback above them untouched.
//
//	go run ./examples/logo [small]
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/logo.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run draws the logo in three rows at the cursor.
func run() error {
	defer term.RecoverAndRestore()

	// Tiny is the default size, as in ratatui.
	size := widgets.CatatuiLogoTiny
	if len(os.Args) > 1 && os.Args[1] == "small" {
		size = widgets.CatatuiLogoSmall
	}

	// An inline viewport owns three rows at the cursor and nothing else, so
	// whatever was on screen before stays visible above the logo. It implies
	// no alternate screen: that is the whole point of drawing inline.
	terminal, restore, err := term.Init(term.WithViewport(catatui.InlineViewport(3)))
	if err != nil {
		return err
	}
	defer func() {
		// Leave the cursor on the viewport's last row and print a newline
		// once the terminal is back to normal, so the shell prompt comes up
		// below the logo instead of over it. restore flushes the move.
		bottom := catatui.SatSub(terminal.ViewportArea().Bottom(), 1)
		_ = terminal.SetCursorPosition(catatui.Position{Y: bottom})
		restore()
		fmt.Println()
	}()

	// The reader has to start after Init: placing an inline viewport means
	// asking the terminal where its cursor is, and that reply comes back on
	// the same input this reader would be draining.
	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, size) }); err != nil {
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

// render draws the logo under a caption.
func render(f *catatui.Frame, size widgets.CatatuiLogoSize) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).Split(f.Area())
	f.RenderWidget(catatui.LineFromString("Powered by"), rows[0])
	f.RenderWidget(widgets.NewCatatuiLogo(size), rows[1])
}
