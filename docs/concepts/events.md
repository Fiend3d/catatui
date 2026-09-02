# Events

This page explains how input reaches a catatui program: the `EventReader`, the
`Event` struct and its kinds, the drain-the-queue loop that the hello example
uses, what to do with each kind of event, and a minimal application structure.
It is for anyone writing an interactive program rather than a one-shot render.

## EventReader

`term.NewEventReader(in, sizeSrc)` starts a goroutine that reads bytes from
`in`, parses them into events, and delivers them on a buffered channel. If
`sizeSrc` is not nil, a second goroutine watches that file for size changes
and delivers `EventResize` events; pass the same file for both when reading
from a terminal. On Unix the resize watcher is driven by `SIGWINCH`, and on
Windows it polls.

```go
package main

import (
	"os"

	"github.com/Fiend3d/catatui/term"
)

func readOne() (term.Event, bool) {
	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()
	ev, ok := <-events.Events()
	return ev, ok
}
```

| Method | Purpose |
|---|---|
| `Events()` | The receive-only channel. It is closed when input ends or `Close` is called. |
| `Err()` | The error that ended reading, or nil. End of file is reported as nil. |
| `Close()` | Stops reading. Safe to call more than once. |

The reader must run in raw mode, which `term.Init` sets up, or key presses
arrive line-buffered and echoed.

A lone `ESC` byte is ambiguous: it is the escape key, and it is also the first
byte of every arrow and function key. The reader waits 50 ms for the rest of a
sequence before reporting a bare escape. That is the usual compromise and it
is why pressing escape feels a hair slower than other keys.

## The Event struct

Every event is one `term.Event` value. `Kind` says what happened, and which of
the other fields carry meaning depends on it.

| Kind | Meaningful fields |
|---|---|
| `EventKey` | `Key`, `Rune`, `Mods` |
| `EventMouse` | `MouseKind`, `Button`, `X`, `Y`, `Mods` |
| `EventResize` | `Size` |
| `EventPaste` | `Text` |
| `EventFocus` | `Focused` |

`Key` is a `KeyCode`. Printable characters use `KeyRune` with the character in
`Rune`; everything else has its own constant: `KeyEnter`, `KeyEscape`,
`KeyBackspace`, `KeyTab`, `KeyBackTab`, `KeyDelete`, `KeyInsert`, `KeyLeft`,
`KeyRight`, `KeyUp`, `KeyDown`, `KeyHome`, `KeyEnd`, `KeyPageUp`,
`KeyPageDown` and `KeyF1` through `KeyF12`.

`Mods` is a bit set of `ModShift`, `ModAlt` and `ModCtrl`, with a `Contains`
method and a `String` that prints as `ctrl+shift`. A control character such as
`ctrl+c` arrives as `KeyRune` with `Rune` set to `'c'` and `ModCtrl` set.

Three helpers cover the common tests and all of them require no modifiers
except where stated:

| Helper | True when |
|---|---|
| `ev.IsKey(k)` | A press of key `k` with no modifiers. |
| `ev.IsRune(r)` | A press of character `r` with no modifiers. |
| `ev.IsCtrl(r)` | Character `r` with control held. |

## The drain-the-queue loop

The hello example's loop is the pattern to copy. It blocks on the channel when
idle, so an idle UI costs no CPU, and after handling one event it drains every
other event already queued before drawing again. A fast wheel spin or a mouse
drag produces dozens of events, and collapsing them into one frame is what
keeps the UI feeling immediate.

```go
package main

import (
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

type app struct{ quit bool }

func (a *app) handle(ev term.Event) {
	if ev.IsRune('q') || ev.IsCtrl('c') {
		a.quit = true
	}
}

func (a *app) draw(f *catatui.Frame) {}

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{}
	for !a.quit {
		if err := terminal.Draw(a.draw); err != nil {
			return err
		}

		ev, ok := <-events.Events()
		if !ok {
			break
		}
		a.handle(ev)

		for drained := true; drained; {
			select {
			case ev, ok := <-events.Events():
				if !ok {
					a.quit = true
					drained = false
					break
				}
				a.handle(ev)
			default:
				drained = false
			}
		}
	}
	return events.Err()
}
```

Two details are easy to get wrong. The `break` inside the `select` only leaves
the `select`, which is why the loop variable is cleared explicitly. And
`events.Err()` is the value to return at the end, because a closed channel can
mean either a clean end of input or a read error.

A program that also needs a tick, for an animation or a clock, adds a
`time.Ticker` channel as a second case in the blocking receive. Keep the drain
loop as it is; a tick is just another reason to draw.

## Handling each kind

The right response to each kind is straightforward once the fields are known.

### Keys

Compare with the helpers for the common cases and fall back to the fields for
combinations.

```go
package main

import "github.com/Fiend3d/catatui/term"

type editor struct {
	cursor int
	quit   bool
}

func (e *editor) handleKey(ev term.Event) {
	switch {
	case ev.IsRune('q'), ev.IsCtrl('c'), ev.IsKey(term.KeyEscape):
		e.quit = true
	case ev.IsKey(term.KeyLeft):
		if e.cursor > 0 {
			e.cursor--
		}
	case ev.IsKey(term.KeyRight):
		e.cursor++
	case ev.Key == term.KeyLeft && ev.Mods.Contains(term.ModShift):
		e.cursor = 0
	}
}
```

### Mouse

`MouseKind` is one of `MouseDown`, `MouseUp`, `MouseDrag`, `MouseMove`,
`MouseScrollUp` and `MouseScrollDown`. `Button` is `MouseButtonLeft`,
`MouseButtonMiddle`, `MouseButtonRight`, or `MouseButtonNone` for a bare move or
a wheel event. `X` and `Y` are absolute cell coordinates, which is why the usual
test is `Rect.Contains`.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func clickedInside(ev term.Event, area catatui.Rect) bool {
	return ev.Kind == term.EventMouse &&
		ev.MouseKind == term.MouseDown &&
		ev.Button == term.MouseButtonLeft &&
		area.Contains(catatui.Position{X: ev.X, Y: ev.Y})
}
```

Mouse events arrive only when `term.Init` was given `term.WithMouse()`.

### Resize

`Terminal.Draw` already resizes its buffers on every call, so a resize event
needs no special handling to keep the screen correct. Its `Size` is useful
when the application caches something derived from the size, and the event is
in any case a reason to draw.

### Paste

With `term.WithBracketedPaste()`, a paste arrives as one `EventPaste` with the
whole text in `Text`. Without it, each pasted character is a separate key event
and a pasted `q` will quit your program.

### Focus

With `term.WithFocusReporting()`, the terminal reports the window gaining or
losing focus in `Focused`. A program can dim itself or pause a tick while it is
in the background.

## A minimal app: struct with handle and draw

The structure that scales from the hello example to a real program is a single
struct holding all application state, with a `handle` method that mutates it in
response to an event and a `draw` method that renders it into a `Frame`. Neither
knows about the other, and nothing draws from inside `handle`.

```go
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

type counter struct {
	n    int
	quit bool
}

func (c *counter) handle(ev term.Event) {
	switch {
	case ev.IsRune('q'), ev.IsCtrl('c'):
		c.quit = true
	case ev.IsKey(term.KeyUp):
		c.n++
	case ev.IsKey(term.KeyDown):
		c.n--
	}
}

func (c *counter) draw(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Fill(1),
	).Split(f.Area())

	f.RenderWidget(
		widgets.NewParagraph(fmt.Sprintf("count: %d", c.n)).
			Block(widgets.Bordered().Title("counter")),
		rows[0])
	f.RenderWidget(
		widgets.NewParagraph("up/down to change, q to quit").
			Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)),
		rows[1])
}

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

	c := &counter{}
	for !c.quit {
		if err := terminal.Draw(c.draw); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			break
		}
		c.handle(ev)
		for drained := true; drained; {
			select {
			case ev, ok := <-events.Events():
				if !ok {
					c.quit = true
					drained = false
					break
				}
				c.handle(ev)
			default:
				drained = false
			}
		}
	}
	return events.Err()
}
```

Because `handle` is a plain method on a plain struct, it is testable by
constructing `term.Event` values directly, and `draw` is testable through a
`TestBackend` as the testing page describes.

**Differences from ratatui.** ratatui has no event type of its own; programs
use crossterm's. catatui's `term.Event` is a single struct with a `Kind`
discriminator rather than an enum with per-variant payloads, and the key event
carries `KeyRune` plus a `Rune` rather than a `KeyCode::Char` variant. There
is no key-release reporting.
