# Terminal

This page covers the `term` package, which is catatui's terminal driver and the
equivalent of the crossterm crate that ratatui uses: `term.Init` and its
options, raw mode and the alternate screen, the three viewport kinds,
`InsertBefore`, panic recovery, the headless `TestBackend`, and how the diff
writer keeps output small. It is for anyone starting a catatui program, and the
last two sections are for anyone debugging flicker or writing tests.

## term.Init

`term.Init` does everything needed to start drawing: it puts the terminal into
raw mode, switches to the alternate screen, hides the cursor, turns on whatever
reporting the options ask for, and returns a ready `*catatui.Terminal` together
with a restore function that undoes all of it. The restore function is safe to
call more than once, and it is also registered to run if the program panics.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func run() error {
	terminal, restore, err := term.Init(term.WithMouse(), term.WithBracketedPaste())
	if err != nil {
		return err
	}
	defer restore()

	return terminal.Draw(func(f *catatui.Frame) {
		f.Buffer().SetString(0, 0, "ready", catatui.NewStyle())
	})
}
```

The options are functions of type `term.Option`.

| Option | Effect |
|---|---|
| `WithMouse()` | Enables mouse reporting, so `EventMouse` events arrive. |
| `WithBracketedPaste()` | Pasted text arrives as one `EventPaste` instead of a flood of key presses. |
| `WithFocusReporting()` | Enables `EventFocus` events. |
| `WithoutAlternateScreen()` | Stays on the main screen, so output remains in the scrollback after exit. |
| `WithCursorShape(s)` | Sets the cursor shape for the life of the program; restore puts it back. |
| `WithViewport(v)` | Draws into the given `catatui.Viewport` instead of the whole terminal. |
| `WithIO(in, out)` | Uses the given files instead of `os.Stdin` and `os.Stdout`. |

Without options, `Init` uses the alternate screen and standard input and output,
and enables nothing else. On Unix, `Init` fails if standard input is not a
terminal; on Windows, it fails if neither handle is a console.

## Raw mode and the alternate screen

Raw mode turns off line buffering, echo and signal generation from control
characters, so that every key press reaches the program immediately and
`ctrl+c` arrives as a key event rather than killing the process. On Windows the
console is additionally switched to virtual terminal mode on both handles, so
that escape sequences are interpreted on output and keys arrive as escape
sequences on input.

The alternate screen is a second screen buffer with no scrollback. Drawing on it
leaves the user's shell history intact, and leaving it restores whatever was
there before. Use `WithoutAlternateScreen()` when the program's output should
stay visible after it exits, which is usual for inline viewports.

The restore function reverses the sequence: it resets the cursor shape, turns
off focus, paste and mouse reporting, shows the cursor, resets text attributes,
leaves the alternate screen, and finally leaves raw mode.

## Viewports

A `catatui.Viewport` is the region of the terminal a `Terminal` draws into.
There are three kinds.

| Constructor | Behaviour |
|---|---|
| `catatui.FullscreenViewport()` | The whole terminal. Resizes with it. The default. |
| `catatui.InlineViewport(height)` | `height` lines at the bottom of the terminal, below the existing output. Resizes with the terminal's width. |
| `catatui.FixedViewport(area)` | An explicit `Rect`. Never resizes. |

An inline viewport is positioned at the cursor when `Init` runs; if there is not
enough room below it, the terminal is scrolled up to make room. Combine it with
`WithoutAlternateScreen()`, since the point of an inline viewport is to coexist
with normal output.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func inline() (*catatui.Terminal, func(), error) {
	return term.Init(
		term.WithoutAlternateScreen(),
		term.WithViewport(catatui.InlineViewport(5)),
	)
}
```

`Terminal.ViewportArea()` returns the current region, which is what
`Frame.Area()` reports inside `Draw`. A fixed viewport is the one case where
`Terminal.Autoresize` does nothing.

## InsertBefore

`Terminal.InsertBefore(height, draw)` draws `height` lines above the viewport
and scrolls the terminal so they become part of the scrollback, while the
viewport keeps redrawing below. This is how a program with an inline viewport
emits log lines that stay on screen.

```go
package main

import "github.com/Fiend3d/catatui"

func logLine(terminal *catatui.Terminal, msg string) error {
	return terminal.InsertBefore(1, func(buf *catatui.Buffer) {
		buf.SetString(buf.Area.X, buf.Area.Y, msg, catatui.NewStyle())
	})
}
```

The callback receives a buffer whose `Area` is the strip being inserted, at
absolute coordinates, so draw at `buf.Area.X` and `buf.Area.Y` rather than at
zero.

## RecoverAndRestore

A panic while the terminal is in raw mode on the alternate screen leaves the
user with an unusable shell and no visible stack trace. `Init` registers its
restore function against this, but Go has no global panic hook, so the recovery
has to be deferred in the goroutine that panics. `term.RecoverAndRestore`
restores every terminal opened by `Init` and then re-panics, so the trace is
printed on a sane screen.

```go
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

	return terminal.Draw(func(f *catatui.Frame) {})
}
```

Defer it first in `main` and in any goroutine that draws. `term.RestoreAll`
does the same restoration without the re-panic, for use from a signal handler or
similar.

## Cursor shape

`catatui.Backend` mirrors ratatui's backend trait and has no notion of cursor
shape, so changing the shape while running goes through the concrete
`term.Backend`. `term.BackendOf(terminal)` returns it, or `nil` when the
terminal is driven by something else such as a `TestBackend`. The change takes
effect on the next flush, which every `Draw` performs.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func useBarCursor(terminal *catatui.Terminal) {
	if b := term.BackendOf(terminal); b != nil {
		_ = b.SetCursorShape(term.CursorSteadyBar)
	}
}
```

The shapes are `CursorDefault`, `CursorBlinkingBlock`, `CursorSteadyBlock`,
`CursorBlinkingUnderline`, `CursorSteadyUnderline`, `CursorBlinkingBar` and
`CursorSteadyBar`. There is no way to query the current shape, so the restore
function always resets it.

## TestBackend for headless use

`catatui.TestBackend` implements `catatui.Backend` by drawing into an in-memory
`Buffer`. It needs no tty, so it is how widgets and whole applications are
tested, and it also works for rendering a UI to a string in any headless
context.

| Method | Purpose |
|---|---|
| `NewTestBackend(width, height)` | A blank backend of the given size. |
| `Buffer()` | Everything drawn so far. |
| `Resize(width, height)` | Changes the size, discarding contents; the next `Draw` sees a resize. |
| `Scrollback()` | Lines pushed off the top by `AppendLines`, for asserting `InsertBefore`. |
| `CursorHidden()` | Whether the last frame hid the cursor. |

```go
package main

import (
	"fmt"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func renderToString() (string, error) {
	backend := catatui.NewTestBackend(12, 3)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		return "", err
	}
	err = terminal.Draw(func(f *catatui.Frame) {
		f.RenderWidget(widgets.NewParagraph("hi").Block(widgets.Bordered()), f.Area())
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprint(backend.Buffer()), nil
}
```

`Buffer.String()` renders the rows joined by newlines with styling dropped. The
testing page describes the assertion helpers built on top of this.

## How the diff writer works, and why it matters for flicker

Flicker in a terminal UI comes from two things: clearing the screen and
redrawing everything, and sending far more bytes than the terminal can paint
between refreshes. catatui avoids both.

The `Terminal` keeps two buffers. After your callback fills one, it is diffed
against the buffer from the previous frame, and only the cells that changed are
passed to `Backend.Draw`. Nothing is cleared between frames, so an unchanged
region is never repainted and cannot flicker. A full clear happens only on
resize, when the terminal has already reflowed the screen anyway.

The `term.Backend` then turns those cells into the smallest escape sequence
stream it can. Its writer tracks the state it has already put the terminal
into:

- A cursor move is emitted only when the next cell is not directly after the
  previous one. Position is tracked through ASCII only, because the terminal's
  shaping engine decides how far a non-ASCII glyph advances and it does not
  have to agree with Unicode's tables; after any such cell the writer forgets
  the position and the next cell re-anchors with an absolute move.
- A style change (SGR sequence) is emitted only when the color or modifiers
  actually differ from the previous cell. Removing a modifier resets the whole
  style and rebuilds it, since per-attribute off codes are unreliable.
- Every frame ends with attributes reset, and every erase or scroll resets them
  first. Terminals fill erased cells with the background color currently in
  effect, so forgetting this turns a resize into a screen painted in the color
  of the last cell drawn.

All of this goes through a 16 KiB buffered writer and is flushed once per
frame, so the terminal receives one contiguous burst rather than a trickle. On a
full 200x50 redraw the naive approach would emit a color sequence per cell,
tens of times more output than necessary, and that is visible as tearing over
ssh. If you implement your own `Backend`, the contract is the same: `Draw`
receives changed cells in row-major order at absolute positions, they are not
necessarily contiguous, and `Flush` is called once at the end of the frame.

**Differences from ratatui.** ratatui delegates to crossterm, termion or termwiz.
catatui has one driver, written for the library, that covers Windows and Unix.
Its `Backend.GetCursorPosition` returns the position the writer tracks rather
than querying the terminal, because reading the reply would race with the
application's own input loop.
