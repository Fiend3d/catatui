# Testing

This page shows how to test catatui code without a terminal: rendering a widget
into a `Buffer`, comparing it against an expected buffer with `AssertBuffer`,
building expectations with `NewBufferWithStrings`, adding expected styles, and
driving a whole application through `TestBackend`. It is for anyone writing
widgets or wanting regression tests for their screens. The approach is the one
ratatui's own test suite uses, and catatui's ported tests use it too.

## Rendering a widget into a Buffer

A `Widget` only needs a `Rect` and a `*Buffer`, so a test creates a buffer of
the size it wants, renders into it, and inspects the cells. No `Terminal` or
backend is involved.

```go
package widgets_test

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func TestBorderedBlock(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 3))
	widgets.Bordered().Title("Hi").Render(buf.Area, buf)

	if got := buf.Get(1, 0).GetSymbol(); got != "H" {
		t.Errorf("cell (1, 0) = %q, want %q", got, "H")
	}
}
```

Rendering through `catatui.RenderWidget(w, area, buf)` rather than calling
`Render` directly adds the same clipping to the buffer that `Frame.RenderWidget`
applies, which is worth using when the area under test is deliberately larger
than the buffer. For a `StatefulWidget`, `catatui.RenderStatefulWidget` takes
the state pointer as well.

## AssertBuffer

Checking individual cells does not scale. `catatui.AssertBuffer(t, got, want)`
compares two whole buffers and, on mismatch, prints both buffers side by side
with every row framed in `|` so that stray spaces are visible, followed by a
per-cell list of differences. It is the counterpart of ratatui's
`assert_eq!(buf, Buffer::with_lines(..))`.

```go
package widgets_test

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func TestBorderedBlockLooksRight(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 3))
	widgets.Bordered().Title("Hi").Render(buf.Area, buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌Hi──────┐",
		"│        │",
		"└────────┘",
	))
}
```

`AssertBuffer` takes a `catatui.TestingT`, a small interface satisfied by
`*testing.T`, so the core package does not import `testing`. The comparison uses
`Cell.Equal`: an empty symbol equals a single space, and an unset color equals
`ColorReset`, so the way a cell was produced does not matter. Styles are compared
in full, so a cell showing the right character in the wrong color fails. The
areas must match too; a buffer of the wrong size fails before any cell is
compared.

`catatui.BufferEqual(got, want)` is the same comparison returning an `error`
instead of failing a test, for use outside a test function.

## NewBufferWithStrings and NewBufferWithLines

`catatui.NewBufferWithStrings(lines...)` builds the expected buffer from one
string per row. Its width is that of the widest row measured with `StringWidth`,
and its height is the number of rows, so the expected buffer must be exactly the
size of the one under test. Rows shorter than the widest are padded with blank
cells.

`catatui.NewBufferWithLines(lines...)` is the same for `catatui.Line` values,
which lets the expectation carry styled spans directly.

```go
package widgets_test

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

func TestStyledLineExpectation(t *testing.T) {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 6, 1))
	catatui.NewLine(catatui.NewSpan("ab"), catatui.NewStyledSpan("cd", bold)).Render(buf.Area, buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		catatui.NewLine(catatui.NewSpan("ab"), catatui.NewStyledSpan("cd", bold), catatui.NewSpan("  ")),
	))
}
```

The trailing two-space span is needed because the buffer under test is six
columns wide and `NewBufferWithLines` sizes the expectation from its content.

## Setting expected styles

For anything beyond a span or two, it is clearer to build the expectation from
plain strings and then paint styles onto regions with `Buffer.SetStyle`, exactly
as ratatui tests call `expected.set_style`. `SetStyle` changes only styling, so
the symbols from the strings are kept.

```go
package widgets_test

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func TestBorderStyle(t *testing.T) {
	cyan := catatui.NewStyle().Fg(catatui.ColorCyan)

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 6, 3))
	widgets.Bordered().BorderStyle(cyan).Render(buf.Area, buf)

	want := catatui.NewBufferWithStrings(
		"┌────┐",
		"│    │",
		"└────┘",
	)
	want.SetStyle(catatui.NewRect(0, 0, 6, 1), cyan)
	want.SetStyle(catatui.NewRect(0, 2, 6, 1), cyan)
	want.SetStyle(catatui.NewRect(0, 1, 1, 1), cyan)
	want.SetStyle(catatui.NewRect(5, 1, 1, 1), cyan)

	catatui.AssertBuffer(t, buf, want)
}
```

Cells can also be edited individually through `want.Get(x, y)`, whose setters
`SetSymbol`, `SetFg`, `SetBg`, `SetStyle` and `SetChar` return the cell pointer
for chaining.

## Testing a whole app with TestBackend

To test the screen an application produces, including layout and every widget it
renders, drive a real `catatui.Terminal` with a `catatui.TestBackend`. The
backend draws into an in-memory buffer that `Buffer()` exposes, and the terminal
diffs and flushes exactly as it would against a real terminal.

```go
package app_test

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

type app struct{ n int }

func (a *app) draw(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Fill(1),
	).Split(f.Area())
	f.RenderWidget(
		widgets.NewParagraph("count: 1").Block(widgets.Bordered()),
		rows[0])
	f.RenderWidget(widgets.NewParagraph("q quits"), rows[1])
}

func TestAppScreen(t *testing.T) {
	backend := catatui.NewTestBackend(12, 4)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}

	a := &app{n: 1}
	if err := terminal.Draw(a.draw); err != nil {
		t.Fatal(err)
	}

	catatui.AssertBuffer(t, backend.Buffer(), catatui.NewBufferWithStrings(
		"┌──────────┐",
		"│count: 1  │",
		"└──────────┘",
		"q quits     ",
	))
	if !backend.CursorHidden() {
		t.Error("cursor should be hidden when the frame did not place it")
	}
}
```

The `app` struct and its `draw` method are the same shape the events page
describes, which is what makes this work: the test calls `draw` through the
terminal without any input plumbing, and `handle` can be tested separately by
constructing `term.Event` values.

Because the backend buffer persists across frames, a second `Draw` after
changing state produces the new screen, and the test can assert it the same
way. `backend.Resize(w, h)` simulates a terminal resize; the next `Draw` picks
it up. `backend.Scrollback()` returns lines pushed off the top by
`Terminal.InsertBefore`, for testing inline-viewport programs.

## Running the tests

Widget tests in this repository sit beside the widgets and use exactly these
helpers.

```sh
go test ./...
go test ./widgets -run TestParagraph -v
```

**Differences from ratatui.** The helpers mirror ratatui's: `NewBufferWithStrings`
is `Buffer::with_lines`, `AssertBuffer` is `assert_eq!` on buffers with the
same side-by-side failure output, and `TestBackend` is ratatui's `TestBackend`.
The one difference is that `AssertBuffer` takes a `TestingT` interface rather
than being a macro, so it can be called from any test framework that provides
`Helper`, `Errorf` and `FailNow`.
