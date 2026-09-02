# Layout

This page covers how catatui divides the screen: the `Rect` type and its
saturating arithmetic, the six kinds of `Constraint`, the seven `Flex` modes,
spacing and overlap, margins, nesting, and a handful of recipes. It is for anyone
placing widgets on screen, which is everyone. The constraint solver is a port of
ratatui's Cassowary-based layout, and its results are checked against 677 cases
translated from ratatui's own tests.

## Rect and saturating uint16 math

A `Rect` is a position plus a size, all `uint16`, in terminal cells. The right
and bottom edges are exclusive: `Right()` is the first column past the rect, and
`Bottom()` is the first row below it.

| Method | Meaning |
|---|---|
| `NewRect(x, y, w, h)` | Constructor that clamps `w` and `h` so the edges stay within `uint16`. |
| `Left()`, `Top()` | Same as `X` and `Y`. |
| `Right()`, `Bottom()` | One past the last column and row. |
| `Area()` | Cell count, as `uint32` because a full-size rect overflows `uint16`. |
| `IsEmpty()` | True when either dimension is zero. |
| `Contains(p)` | Whether `p` is inside; left and top edges count, right and bottom do not. |
| `Inner(m)` / `Outer(m)` | Shrink or grow by a `Margin` on every side. |
| `Intersection(r)` / `Union(r)` | Overlap, or bounding rect. |
| `Intersects(r)` | Whether the two overlap. |
| `Clamp(r)` | Slide and shrink to fit inside `r`, keeping as much size as possible. |
| `Offset(o)` | Move by a signed `Offset`, clamped to the coordinate space. |
| `Resize(s)` | Change size, keeping position. |
| `Rows()`, `Columns()`, `Positions()` | Enumerate one-cell-high rows, one-cell-wide columns, or every cell. |

All the arithmetic behind these saturates rather than wraps, and the helpers are
exported because every widget author needs them: subtracting a border from an
area that is too small must give zero, not 65535.

| Function | Result |
|---|---|
| `SatAdd(a, b)` | `a + b`, clamped at 65535. |
| `SatSub(a, b)` | `a - b`, clamped at 0. |
| `SatMul(a, b)` | `a * b`, clamped at 65535. |
| `MinU16(a, b)`, `MaxU16(a, b)` | Smaller or larger of two. |
| `ClampU16(v, lo, hi)` | `v` confined to `[lo, hi]`. |

```go
package main

import "github.com/Fiend3d/catatui"

func lastRow(area catatui.Rect) catatui.Rect {
	return catatui.Rect{
		X:      area.X,
		Y:      catatui.SatSub(area.Bottom(), 1),
		Width:  area.Width,
		Height: catatui.MinU16(area.Height, 1),
	}
}
```

`Clamp` and `Intersection` differ in a way that matters for popups: `Intersection`
keeps the rect's position and cuts off whatever hangs outside, while `Clamp`
moves the rect back inside first.

## Constraints

A `Layout` splits an area into one segment per `Constraint`. Constraints are
requests, not guarantees. When they conflict, the solver gives way in a fixed
order of priority, from strongest to weakest: `Min` and `Max`, then `Length`,
then `Percentage`, then `Ratio`, then `Fill`.

| Constraint | Request |
|---|---|
| `Length(n)` | Exactly `n` cells. |
| `Min(n)` | At least `n` cells. Outside `FlexLegacy` it also grows into leftover space like a `Fill(1)`. |
| `Max(n)` | At most `n` cells; it prefers to be exactly `n` if it can. |
| `Percentage(p)` | `p` percent of the area, rounded to the nearest cell. |
| `Ratio(num, den)` | `num/den` of the area. A zero denominator is treated as one. |
| `Fill(scale)` | Whatever is left, shared among the `Fill` constraints in proportion to `scale`. |

The zero `Constraint` value is `Min(0)`; always build constraints with the
constructors.

Here is how a few pairs land in a 20-column area under the default flex mode.
Each letter is one column belonging to that segment and a dot is an empty
column.

```
Min(5)    + Length(6)    |AAAAAAAAAAAAAABBBBBB|   Min grows to fill
Max(5)    + Length(6)    |AAAAABBBBBB.........|   Max stops at 5
Fill(1)   + Fill(2)      |AAAAAAABBBBBBBBBBBBB|   7 : 13, closest to 1 : 2
Pct(25)   + Fill(1)      |AAAAABBBBBBBBBBBBBBB|
Ratio(1,3)+ Fill(1)      |AAAAAAABBBBBBBBBBBBB|
Length(12)+ Min(10)      |AAAAAAAAAABBBBBBBBBB|   Min outranks Length
```

The last row shows priority at work: 12 + 10 does not fit in 20, and `Min` is
stronger than `Length`, so `Length(12)` is the one that shrinks.

```go
package main

import "github.com/Fiend3d/catatui"

func split(area catatui.Rect) (header, body, footer catatui.Rect) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Fill(1),
		catatui.Length(1),
	).Split(area)
	return rows[0], rows[1], rows[2]
}
```

`Split` always returns exactly `len(constraints)` rects, some of which may be
empty when the area is too small.

## Building a Layout

`VerticalLayout(...)` and `HorizontalLayout(...)` are the usual constructors.
`NewLayout()` gives ratatui's default, which is vertical with no constraints.
Every builder method returns a modified copy, so a `Layout` can be stored in a
variable and reused across frames.

| Builder | Effect |
|---|---|
| `Direction(d)`, `Vertical()`, `Horizontal()` | Which axis to split along. |
| `Constraints(...)` | Replace the constraint list. |
| `Margin(n)`, `HorizontalMargin(n)`, `VerticalMargin(n)` | Inset before splitting. |
| `Flex(f)` | What to do with leftover space. |
| `Spacing(s)` | Gap or overlap between segments. |

The direction enum is `catatui.Vertical` and `catatui.Horizontal`. `Vertical` is
the zero value because it is ratatui's default direction.

## Flex modes

Once every constraint is satisfied there may be space left over. `Flex` decides
where it goes. The diagrams below all split the same constraints,
`Length(4), Length(6)`, across 20 columns.

```
FlexStart         |AAAABBBBBB..........|   excess at the end (default)
FlexLegacy        |AAAABBBBBBBBBBBBBBBB|   excess given to the last segment
FlexEnd           |..........AAAABBBBBB|   excess at the start
FlexCenter        |.....AAAABBBBBB.....|   excess split between both ends
FlexSpaceBetween  |AAAA..........BBBBBB|   excess between segments only
FlexSpaceEvenly   |...AAAA....BBBBBB...|   equal gaps everywhere
FlexSpaceAround   |...AAAA.....BBBBBB..|   ends get half a gap
```

With three segments the difference between the last three modes is clearer:

```
                   Length(3) x 3 in 20 columns
FlexSpaceBetween  |AAA......BBB.....CCC|
FlexSpaceEvenly   |...AAA...BBB..CCC...|
FlexSpaceAround   |..AAA....BBB...CCC..|
```

`FlexLegacy` is how tui-rs and early ratatui behaved: the last constraint
absorbs the excess. It is also the only mode in which `Min(n)` does not grow.
In every other mode, segments weakly prefer to be equal in size, which only
matters when nothing stronger decides.

```go
package main

import "github.com/Fiend3d/catatui"

func centeredButtons(area catatui.Rect) []catatui.Rect {
	return catatui.HorizontalLayout(
		catatui.Length(10),
		catatui.Length(10),
	).Flex(catatui.FlexCenter).Spacing(catatui.Space(2)).Split(area)
}
```

## Spacing and Overlap

`Spacing` is the gap between adjacent segments. `Space(n)` separates them by `n`
cells; `Overlap(n)` makes neighbours share `n` cells, which is how adjacent
blocks are drawn with a single collapsed border.

```
Length(4), Length(6) with Space(2)
FlexStart         |AAAA..BBBBBB........|
FlexCenter        |....AAAA..BBBBBB....|
FlexSpaceBetween  |AAAA..........BBBBBB|   spacing is a minimum here

Length(8), Length(8), Length(6) with Overlap(1)
FlexStart         |AAAAAAA#BBBBBB#CCCCC|   # is a shared column
```

In the `Space*` modes the spacing is a lower bound on the gaps; the excess is
still spread out. In `FlexStart`, `FlexEnd`, `FlexCenter` and `FlexLegacy` the
gap is exactly the spacing.

## Split versus SplitWithSpacers

`Split` returns the segments. `SplitWithSpacers` returns the segments and the
gaps between them as a second slice. There is always one more spacer than there
are segments: one before the first segment, one between each adjacent pair, and
one after the last. Spacers can be empty.

```go
package main

import "github.com/Fiend3d/catatui"

func gaps(area catatui.Rect) (segments, spacers []catatui.Rect) {
	return catatui.HorizontalLayout(
		catatui.Length(4),
		catatui.Length(6),
	).Spacing(catatui.Space(2)).SplitWithSpacers(area)
}
```

For a 20-column area that returns segments at `x=0 w=4` and `x=6 w=6`, and
spacers at `x=0 w=0`, `x=4 w=2` and `x=12 w=8`. Widgets that draw separators in
the gaps use the spacers.

## Margins

`Margin(n)` insets the area on every side before it is split; the horizontal
and vertical variants inset one axis. A margin larger than the area produces an
empty inner rect, and then every segment is empty.

```go
package main

import "github.com/Fiend3d/catatui"

func padded(area catatui.Rect) catatui.Rect {
	return catatui.VerticalLayout(catatui.Fill(1)).Margin(2).Split(area)[0]
}
```

For a 20x5 area this yields `x=2 y=2 w=16 h=1`. For a plain inset without a
layout, `area.Inner(catatui.NewMargin(2, 2))` does the same thing.

## Nesting layouts

There is no layout tree. To subdivide, split an area, then split one of the
resulting rects again. This is cheap enough to do every frame; the hello example
does exactly this.

```go
package main

import "github.com/Fiend3d/catatui"

type panes struct {
	title, sidebar, main, status catatui.Rect
}

func layoutPanes(area catatui.Rect) panes {
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Fill(1),
		catatui.Length(1),
	).Split(area)
	cols := catatui.HorizontalLayout(
		catatui.Percentage(30),
		catatui.Fill(1),
	).Split(rows[1])
	return panes{title: rows[0], sidebar: cols[0], main: cols[1], status: rows[2]}
}
```

## Recipes

These are the layouts most applications start from. Each function returns rects
you then render into.

### Header, body, footer

A fixed-height header and footer with the body taking the rest.

```go
package main

import "github.com/Fiend3d/catatui"

func headerBodyFooter(area catatui.Rect) (header, body, footer catatui.Rect) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Min(0),
		catatui.Length(3),
	).Split(area)
	return rows[0], rows[1], rows[2]
}
```

`Min(0)` and `Fill(1)` behave the same here; both grow into the leftover space.

### Sidebar and main pane

A sidebar that is never narrower than 20 columns beside a main pane that takes
the rest.

```go
package main

import "github.com/Fiend3d/catatui"

func sidebarMain(area catatui.Rect) (sidebar, main catatui.Rect) {
	cols := catatui.HorizontalLayout(
		catatui.Min(20),
		catatui.Fill(3),
	).Split(area)
	return cols[0], cols[1]
}
```

Because `Min` participates in fill distribution outside `FlexLegacy`, the
sidebar here grows at one part to the main pane's three once both minimums are
met. Use `Length(20)` instead for a sidebar that never grows.

### Centered popup

Center a rect of a given size, using `FlexCenter` on both axes.

```go
package main

import "github.com/Fiend3d/catatui"

func popup(area catatui.Rect, width, height uint16) catatui.Rect {
	row := catatui.VerticalLayout(catatui.Length(height)).
		Flex(catatui.FlexCenter).Split(area)[0]
	return catatui.HorizontalLayout(catatui.Length(width)).
		Flex(catatui.FlexCenter).Split(row)[0]
}
```

Percentages work as well: `Percentage(60)` in place of `Length(width)` gives a
popup that scales with the terminal. Render a `widgets.Bordered()` block into the
result after everything else, so it sits on top.

**Differences from ratatui.** The constraint semantics, strengths and flex modes
are ratatui's. Two things differ. First, the solver iterates in a deterministic
order, whereas the Rust solver scans hash maps; a degenerate layout therefore
comes out the same on every run in catatui instead of varying. Second, `Layout`
builders return copies rather than consuming the receiver, so a layout value
stays usable after you derive another from it.
