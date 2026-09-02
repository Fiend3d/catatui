# Widgets

This page covers the widget library: what is in it, the builder conventions
every widget follows, how blocks frame other widgets, how the widgets that
remember something keep that state outside themselves, and what to reach for
when nothing in the library fits. It is for anyone building a screen out of
ready-made parts. Writing your own widget is covered in
[rendering](rendering.md); this page is about using the ones that exist.

Everything here lives in `github.com/Fiend3d/catatui/widgets`, and every widget
has a runnable program in [examples](../../examples).

## The catalogue

| Widget | What it is for | Example |
|---|---|---|
| `Block` | Borders, titles, padding and an optional shadow around anything | [block](../../examples/block) |
| `Paragraph` | Styled text with wrapping, scrolling and alignment | [paragraph](../../examples/paragraph) |
| `Clear` | Blanks an area, so a popup does not show what is under it | [shadow](../../examples/shadow) |
| `Fill` | Fills an area with one symbol | — |
| `List` | A scrolling list with a selection | [list](../../examples/list) |
| `Table` | Columns sized by `Constraint`, with row, column and cell selection | [table](../../examples/table) |
| `Tabs` | A row of titles with one highlighted | [tabs](../../examples/tabs) |
| `Scrollbar` | A track and thumb along any edge | [scrollbar](../../examples/scrollbar) |
| `Gauge`, `LineGauge` | A progress bar, in a block or on a single row | [gauge](../../examples/gauge), [line-gauge](../../examples/line-gauge) |
| `BarChart` | Bars in either direction, grouped or not | [barchart](../../examples/barchart), [barchart-grouped](../../examples/barchart-grouped) |
| `Sparkline` | One row of history, no axes | [sparkline](../../examples/sparkline) |
| `Chart` | Scatter, line, bar and area plots on shared axes | [chart](../../examples/chart) |
| `Canvas` | A coordinate space drawn with braille or block markers | [canvas](../../examples/canvas) |
| `Monthly` | A month, with styles per date | [calendar](../../examples/calendar) |
| `Shadow` | A shadow cast behind a popup | [shadow](../../examples/shadow) |
| `CatatuiLogo`, `RatatuiMascot` | The logo and the rat | [logo](../../examples/logo) |

`Chart` and `BarChart` cover different jobs despite the names: a bar chart draws
one bar per labelled category, a chart plots datasets against numeric axes.

## Builders return copies

Every widget is a value, and every builder returns a modified copy rather than
consuming the receiver. A partly built widget is therefore reusable:

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func panes() (widgets.Block, widgets.Block) {
	base := widgets.Bordered().Padding(widgets.HorizontalPadding(1))

	cyan := base.BorderStyle(catatui.NewStyle().Fg(catatui.ColorCyan))
	red := base.BorderStyle(catatui.NewStyle().Fg(catatui.ColorRed))
	return cyan, red // base is unchanged and still usable
}
```

Fields are unexported behind ratatui's builder names, so the reader is prefixed
with `Get`: `Style` sets a style, `GetStyle` returns it. Rust allows a field and
a method to share a name; Go does not.

## Blocks and inner areas

Widgets that can be framed take a `Block`, and draw into whatever it leaves:

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func framed(f *catatui.Frame, area catatui.Rect) {
	f.RenderWidget(
		widgets.NewParagraph("hello").
			Block(widgets.Bordered().Title("catatui")),
		area)
}
```

To draw something the block does not know about, render the block and ask it
for the inner area. This is the pattern for anything hand-drawn, and for
widgets whose `Block` you would otherwise have to thread through:

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func manual(f *catatui.Frame, area catatui.Rect) {
	block := widgets.Bordered().Title("Files")
	inner := block.Inner(area)
	f.RenderWidget(block, area)

	buf := f.Buffer()
	if !inner.IsEmpty() {
		buf.SetStringn(inner.X, inner.Y, "drawn by hand", inner.Width,
			catatui.NewStyle())
	}
}
```

`Inner` accounts for the borders that are drawn, the padding, and a title row
even where there is no top border.

### Collapsing borders

Two blocks that touch normally draw two lines. `MergeBorders` combines the
characters instead, so neighbours share one:

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

func collapsed(f *catatui.Frame, area catatui.Rect) {
	// Overlap(1) makes the panes share a row; the merge turns the two
	// borders that land there into one.
	rows := catatui.VerticalLayout(catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Overlap(1)).Split(area)

	for _, row := range rows {
		f.RenderWidget(widgets.Bordered().MergeBorders(symbols.MergeExact), row)
	}
}
```

`symbols.MergeExact` merges only where a single character shows both lines;
`symbols.MergeFuzzy` also handles the combinations Unicode is missing by moving
one of them to a nearby line style. The block drawn last wins, so render the
focused pane after the others. See the
[collapsed-borders](../../examples/collapsed-borders) example.

## Widgets that remember: state

Widgets hold no state between frames. The ones that have to remember
something — which row is selected, how far a view has scrolled — take it from
the caller:

| Widget | State | What the state holds |
|---|---|---|
| `List` | `ListState` | Selected index and scroll offset |
| `Table` | `TableState` | Selected row, column and cell, and offset |
| `Scrollbar` | `ScrollbarState` | Content length, position and viewport length |

The state lives in your application struct and outlives the widget:

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

type app struct {
	items []string
	list  widgets.ListState
}

func (a *app) handle(ev term.Event) {
	switch {
	case ev.IsKey(term.KeyDown):
		a.list.SelectNext()
	case ev.IsKey(term.KeyUp):
		a.list.SelectPrevious()
	}
}

func (a *app) draw(f *catatui.Frame) {
	list := widgets.NewListFromStrings(a.items...).
		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
		HighlightSymbol("> ")

	catatui.RenderStatefulWidget(list, f.Area(), f.Buffer(), &a.list)
}
```

The list is rebuilt every frame; the selection is not. Rendering also writes
back to the state — the scroll offset is decided during rendering, because only
then is the available height known — which is why the state is passed as a
pointer.

`ListState` and `TableState` come in two flavours of the same operations:
`WithSelected` returns a copy, for building an initial value, while `Select`,
`SelectNext` and the rest mutate through a pointer, for handling a key.

## Sizing

Widgets take whatever area you give them and never draw outside it. An area too
small for the content is not an error: the widget draws what fits and stops. A
zero-sized area draws nothing.

Three widgets can tell you how much room they want, which is what you need when
a layout should be sized to its content rather than the other way round:

| Method | Returns |
|---|---|
| `Tabs.Width()` | Columns the titles, dividers and padding need |
| `Monthly.Width()`, `Monthly.Height()` | The size of a month, headers included |
| `Paragraph.LineCount(width)` | Lines the text wraps to at that width |

`Paragraph.LineCount` is the one to reach for when a popup should be as tall as
its message:

```go
package main

import "github.com/Fiend3d/catatui/widgets"

func messageHeight(text string, width uint16) uint16 {
	p := widgets.NewParagraph(text).Wrap(widgets.Wrap{Trim: true})
	lines := p.LineCount(width)
	return uint16(min(lines, 0xFFFF))
}
```

## Styles

A widget's own `Style` is applied to its whole area first, and everything drawn
into that area layers on top of it: a `Line`'s style over the widget's, a
`Span`'s over the line's. So a paragraph's style is a background for its text
rather than an override of it. [Text and style](text-and-style.md) has the full
order of precedence.

Highlight styles are patched over the item's own style rather than replacing it,
which is why a yellow item stays yellow under a reversed selection.

## When nothing fits

The library is not the API. `Frame.Buffer()` gives you the buffer, and anything
you can express as cells you can draw directly — see
[drawing directly](rendering.md#drawing-directly-with-framebuffer). Custom
widgets are just a `Render` method, and they compose with the built-in ones
because `Block.Inner` and the layout functions do not care who draws into the
area they return.

**Differences from ratatui.** The widget set matches `ratatui-widgets` 0.3.2,
with these deviations: stateful widgets render through the free function
`catatui.RenderStatefulWidget` rather than a method, because Go methods cannot
have type parameters; builders return copies rather than consuming the receiver;
readers are prefixed with `Get`; and the logo widget spells catatui rather than
ratatui. `Block.MergeBorders` and the border-merging rules behind it are ported
from `ratatui-core`'s `symbols::merge`, and reproduce ratatui's own golden
output for every pair of border types.
