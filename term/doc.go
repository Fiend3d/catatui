// Package term is catatui's terminal driver: raw mode, the alternate screen,
// input events, and the Backend that a catatui.Terminal writes through. It is
// what crossterm is to ratatui, in one package with no dependencies beyond
// golang.org/x/sys and golang.org/x/term.
//
// # Starting and stopping
//
// Init does everything a program needs to start drawing and returns the
// function that undoes it. The restore function is safe to call twice, and is
// also registered to run if the program panics — a panic that left the terminal
// in raw mode on the alternate screen would leave the user with an unusable
// shell and an invisible stack trace.
//
//	defer term.RecoverAndRestore()
//
//	terminal, restore, err := term.Init(term.WithMouse())
//	if err != nil {
//		return err
//	}
//	defer restore()
//
// Options cover what the terminal reports and how much of it catatui owns:
// WithMouse, WithBracketedPaste, WithFocusReporting, WithCursorShape,
// WithoutAlternateScreen, WithViewport and WithIO.
//
// # Input
//
// NewEventReader starts reading in a goroutine and delivers Events on a
// channel: keys, mouse, resize, focus and paste. Resizes come from SIGWINCH on
// Unix and from polling on Windows.
//
//	events := term.NewEventReader(os.Stdin, os.Stdout)
//	defer events.Close()
//
//	for {
//		if err := terminal.Draw(render); err != nil {
//			return err
//		}
//		ev, ok := <-events.Events()
//		if !ok {
//			return events.Err()
//		}
//		if ev.IsRune('q') || ev.IsCtrl('c') {
//			return nil
//		}
//	}
//
// Blocking on the channel is the point: an idle UI costs nothing. A program
// that also has work of its own selects over the channel and a ticker, and one
// that redraws on every event should drain whatever else is already queued
// first, so that a fast mouse drag becomes one frame rather than thirty. The
// events guide under docs/concepts shows both.
//
// # Output
//
// Backend implements catatui.Backend over VT escape sequences, and tracks the
// state it has already put the terminal into so that a redraw emits only what
// changed. That is what keeps a full-screen frame from becoming tens of
// kilobytes of colour sequences, which is visible as tearing over ssh.
//
// Cursor position is the one thing Backend does not ask the terminal for:
// GetCursorPosition returns the position it has been tracking, because querying
// would race with the application's own input loop. QueryCursorPosition does
// ask, and Init uses it once, before any EventReader exists, to place an inline
// viewport where the shell's cursor actually is.
package term
