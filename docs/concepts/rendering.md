# Rendering

This page explains how a catatui program gets pixels onto the screen: the
immediate-mode model, the path from a `Frame` through a `Buffer` and a diff to
the backend, what a `Cell` is, how wide characters occupy several cells, and the
two widget interfaces. It is for anyone writing widgets or drawing into the
buffer by hand. If you only compose the built-in widgets, the first two sections
are enough.

## The immediate-mode model

catatui, like ratatui, is immediate mode. There is no widget tree that persists
between frames. Every frame, your code builds widget values from scratch, renders
them into a buffer, and throws them away. Whatever state the UI needs lives in
your own structs, not inside the widgets.

The whole cycle runs inside `Terminal.Draw`:

1. `Terminal` checks the backend size and resizes its buffers if the terminal
   changed.
2. It hands your callback a `Frame` wrapping the current (blank) buffer.
3. Your callback renders widgets into the frame.
4. `Terminal` diffs the new buffer against the previous frame's buffer.
5. Only the cells that changed are sent to `Backend.Draw`, then the cursor is
   placed (or hidden) and the backend is flushed.
6. The two buffers swap roles and the now-stale one is blanked for the next
   frame.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	terminal, restore, err := term.Init()
	if err != nil {
		panic(err)
	}
	defer restore()

	err = terminal.Draw(func(f *catatui.Frame) {
		f.RenderWidget(widgets.NewParagraph("hello"), f.Area())
	})
	if err != nil {
		panic(err)
	}
}
```

`Terminal.TryDraw` is the same thing for a callback that can fail; if it returns
an error, nothing reaches the backend for that frame.

## Frame

A `Frame` is the drawing surface for exactly one call of `Draw`. Do not keep it
after the callback returns; the buffer underneath it is swapped and blanked.

| Method | What it gives you |
|---|---|
| `Area()` | The `Rect` the frame covers, which is the viewport area. |
| `Size()` | Same as `Area()`, kept because ratatui has both. |
| `Buffer()` | The underlying `*Buffer`, for direct cell drawing. |
| `Count()` | How many frames were drawn before this one; useful for animation. |
| `RenderWidget(w, area)` | Renders a `Widget` into `area`, clipped to the buffer. |
| `SetCursorPosition(p)` / `SetCursor(x, y)` | Where the terminal cursor goes after the frame is written. |

## Buffer and Cell

A `Buffer` is a rectangular grid of `Cell` values stored row-major in the
exported `Content` slice, together with the `Area` it maps to. Coordinates are
absolute terminal coordinates, not offsets from the buffer's origin. The
invariant `len(Content) == Area.Area()` must hold; the library maintains it, and
you must too if you assign the fields directly.

A `Cell` is one column of the terminal: a grapheme cluster and its style. Its
fields are exported.

| Field | Meaning |
|---|---|
| `Symbol` | The grapheme cluster shown. An empty string means a space. |
| `Fg`, `Bg`, `UnderlineColor` | Concrete colors. A blank cell holds `ColorReset` in each. |
| `Modifier` | The text attribute bits. |
| `DiffOption` | How the cell takes part in the frame diff (see below). |

Two cells built differently can still draw identically, so compare them with
`Cell.Equal` rather than `==`: an empty `Symbol` equals a single space, and an
unset color equals `ColorReset`.

The buffer's drawing methods are the only place text turns into cells:

| Method | What it does |
|---|---|
| `SetString(x, y, s, style)` | Draws `s`, stopping at the buffer's right edge. |
| `SetStringn(x, y, s, maxWidth, style)` | Same, but also stops after `maxWidth` columns. Returns the position just past the drawn text. |
| `SetSpan(x, y, span, maxWidth)` | Draws a `Span` in its own style. |
| `SetLine(x, y, line, maxWidth)` | Draws a `Line`, patching each span's style over the line's. |
| `SetStyle(area, style)` | Applies a style to every cell in `area` without touching symbols. |
| `Get(x, y)` | Pointer to a cell; panics outside the buffer. |
| `Cell(pos)` / `CellAt(x, y)` | Pointer to a cell, or `nil` outside the buffer. |

`Get` panics on purpose. A widget that draws outside the area it was given is a
bug, and a panic at the offending coordinate is far easier to find than a
corrupted row somewhere else on screen.

## Wide graphemes and continuation cells

A grapheme cluster can be wider than one column: CJK ideographs and most emoji
take two, and some Indic clusters take two even though they are drawn as one
glyph. `SetStringn` handles this with three rules, all pinned down by tests
ported from ratatui.

1. The cluster goes into the cell at its first column, with the requested style.
2. The remaining columns it covers are **reset to blank cells**, not styled.
   These are the continuation cells. The diff knows to skip them, because the
   terminal draws the wide glyph over them itself.
3. A cluster that does not fit in the columns left is **dropped, not clipped**,
   and drawing stops there. Nothing after it in the string is drawn either.
   Callers that need every column filled must pad the remainder themselves.

Control characters and zero-width clusters are dropped before they reach a cell.

```go
package main

import "github.com/Fiend3d/catatui"

func main() {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 1))
	// "日本" is four columns wide; the "語" would need two more and is dropped.
	nextX, _ := buf.SetString(0, 0, "日本語", catatui.NewStyle())
	_ = nextX // 4
	_ = buf.Get(1, 0).GetSymbol() // " ": the continuation cell of 日
}
```

## The one-width-function rule

Everything in catatui that needs to know how many columns a string takes calls
`catatui.StringWidth`. It counts grapheme clusters the same way `SetStringn`
draws them, skipping the clusters a buffer would not draw, so the number it
returns is exactly the number of columns `SetString` fills. A test asserts this.

Do not measure text with `len`, `utf8.RuneCountInString`, or a width function
from another library when laying out a widget. Two width functions that disagree
on a single character are enough to make rows drift out of alignment, and that
is precisely the failure this library was written to avoid.

For code that segments text itself, `catatui.Graphemes` and the allocation-free
`catatui.AllGraphemes` return each cluster with its width, and
`catatui.TrimLeftColumns` cuts columns off the left on cluster boundaries.

```go
package main

import "github.com/Fiend3d/catatui"

func fits(label string, width uint16) bool {
	return catatui.StringWidth(label) <= int(width)
}
```

**Differences from ratatui.** ratatui has two width functions: `Span::width`
uses the `unicode-width` crate while `Buffer::set_stringn` uses `cell_width`,
and they disagree on halfwidth katakana sound marks and some emoji sequences.
catatui measures everything with one function.

## Drawing directly with Frame.Buffer

Nothing forces you to go through a `Widget`. `Frame.Buffer()` returns the
frame's buffer, and you can draw cells into it however you like. This is how the
hello example places its marker, and it is the intended way to do anything the
widget library does not cover yet.

```go
package main

import "github.com/Fiend3d/catatui"

func drawMarker(f *catatui.Frame, x, y uint16) {
	buf := f.Buffer()
	if f.Area().Contains(catatui.Position{X: x, Y: y}) {
		buf.SetString(x, y, "@",
			catatui.NewStyle().Fg(catatui.ColorYellow).AddModifier(catatui.ModifierBold))
	}
}
```

The check against `f.Area()` matters: `SetString` clips at the buffer's right
edge but does not check the row, and `Get` panics out of bounds.

## Cursor handling

The terminal cursor is hidden by default. A frame that wants a visible cursor,
such as one containing a text field, calls `Frame.SetCursorPosition` (or
`SetCursor`) during rendering. The terminal shows the cursor and moves it there
after the frame's cells are written, since writing cells would move it again. A
frame that never asks hides the cursor.

```go
package main

import "github.com/Fiend3d/catatui"

func drawInput(f *catatui.Frame, area catatui.Rect, text string) {
	f.Buffer().SetString(area.X, area.Y, text, catatui.NewStyle())
	col := catatui.SatAdd(area.X, uint16(min(catatui.StringWidth(text), 0xFFFF)))
	f.SetCursor(col, area.Y)
}
```

Measure the cursor offset with `StringWidth`, not `len(text)`; otherwise the
cursor lands in the wrong column as soon as a wide character is typed.

## The diff and CellDiffOption

`Buffer.Diff`, `DiffInto` and `DiffSeq` compute the cells that must be written
to turn one buffer into another. The `Terminal` uses `DiffInto` with a reused
slice so that a redraw does not allocate. The diff is not a plain cell-by-cell
comparison: it steps over the continuation columns of wide characters, and it
force-rewrites the trailing columns of a wide character that is being replaced
by a narrow one when residue could remain (a colored background, a visible
modifier, or an emoji presentation sequence).

A cell can opt out of the normal comparison through its `DiffOption`:

| Option | Effect |
|---|---|
| `CellDiffNone` | Default. Written when it differs from the previous frame. |
| `CellDiffSkip` | Never written. For cells covered by something outside catatui's control, such as a terminal image. |
| `CellDiffAlwaysUpdate` | Written every frame even if unchanged. |
| `CellForcedWidth(n)` | The cell counts as `n` columns regardless of its symbol. |

Set them with `Cell.SetDiffOption` or the shorthand `Cell.SetSkip`.

**Differences from ratatui.** ratatui 0.30.2 blanks the trailing column of an
emoji presentation sequence explicitly when drawing it. catatui does not, because
that blank lands on the column the terminal has just shaped the glyph into and
breaks ZWJ sequences. The clearing happens only when such a sequence is
replaced.

## Widget and StatefulWidget

A `Widget` is anything with a `Render(area Rect, buf *Buffer)` method.
Implement it on a value receiver and treat the widget as consumed by the call.
`Render` must not draw outside `area`.

```go
package main

import "github.com/Fiend3d/catatui"

type Banner struct{ Text string }

func (b Banner) Render(area catatui.Rect, buf *catatui.Buffer) {
	if area.IsEmpty() {
		return
	}
	buf.SetStringn(area.X, area.Y, b.Text, area.Width, catatui.NewStyle())
}

var _ catatui.Widget = Banner{}
```

`catatui.WidgetFunc` adapts a plain function to the interface, and the free
function `catatui.RenderWidget(w, area, buf)` renders one widget inside another
with the same clipping `Frame.RenderWidget` applies.

A `StatefulWidget[S]` renders against caller-owned state that outlives the
widget: a list that remembers its selected row, a viewport that remembers its
scroll offset. The widget is rebuilt every frame; the state is not.

```go
package main

import "github.com/Fiend3d/catatui"

type CounterState struct{ Frames uint64 }

type Counter struct{}

func (Counter) RenderStateful(area catatui.Rect, buf *catatui.Buffer, s *CounterState) {
	s.Frames++
	buf.SetStringn(area.X, area.Y, "frame", area.Width, catatui.NewStyle())
}

func draw(f *catatui.Frame, state *CounterState) {
	catatui.RenderStatefulWidgetOn(f, Counter{}, f.Area(), state)
}
```

### Why RenderStatefulWidget is a free function

In ratatui, `StatefulWidget` declares the state as an associated type and
`Frame::render_stateful_widget` is a method. Go has neither associated types nor
methods with their own type parameters, so the interface is generic over `S` and
rendering goes through two free functions:

| Function | Use |
|---|---|
| `catatui.RenderStatefulWidgetOn(f, w, area, state)` | Render into a `Frame`. |
| `catatui.RenderStatefulWidget(w, area, buf, state)` | Render into a `Buffer`, for composition and tests. |

Both clip `area` to the buffer before calling `RenderStateful`, exactly as
`RenderWidget` does.

**Differences from ratatui.** This generic interface plus free-function pair is
the one place the API shape differs from ratatui for language reasons rather than
by choice. Everything else on this page follows ratatui's names.
