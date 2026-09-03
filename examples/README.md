# Examples

Each directory is a self-contained program: whole applications under
[apps](apps), one program per widget under [widgets](widgets). Run one with:

```sh
go run ./examples/apps/<name>
go run ./examples/widgets/<name>
```

The widget examples draw a single frame and quit on the first key; the
applications keep running. Each says what keys it responds to in the table
below and in the doc comment at the top of its `main.go`.

Nothing is shared between examples on purpose: each carries its own terminal
setup and event loop so that it can be read start to finish, and pasted into an
application as it stands. That is ratatui's convention for its own examples, and
the reason the same twenty lines of scaffolding appear in every one of them.

## Applications

Whole programs, ported from ratatui's [examples/apps]:

| Example | Shows | Keys |
|---|---|---|
| [demo](apps/demo) | The original tui-rs demo: three tabs of gauges, lists, charts, a table and a world map, animating on a tick | arrows or `h`/`j`/`k`/`l`, `t`, `q` |
| [flex](apps/flex) | What each `Flex` mode does to the same constraints, side by side | arrows, `g`/`G`, `-`/`+`, `q` |
| [hello](apps/hello) | Layout, direct buffer drawing, and an event loop that blocks when idle | arrows, mouse, `q` |
| [inline](apps/inline) | An inline viewport, and `InsertBefore` pushing finished lines into the scrollback | `q` |
| [popup](apps/popup) | A popup drawn over the rest of the UI, with `Clear` behind it | `p`, `q` |
| [todo-list](apps/todo-list) | A list whose items are selected and ticked off, with the selection kept in a `ListState` | arrows or `h`/`j`/`k`/`l`, `g`/`G`, Enter, `q` |
| [user-input](apps/user-input) | Typing into an input box: editing modes, and where the cursor goes | `e`, Esc, Enter, arrows, `q` |

## Widgets

One program per widget, ported from [ratatui-widgets/examples]:

| Example | Shows | Keys |
|---|---|---|
| [barchart](widgets/barchart) | `BarChart` drawn both ways up | any key quits |
| [barchart-grouped](widgets/barchart-grouped) | `BarChart` with grouped bars and value labels | any key quits |
| [block](widgets/block) | `Block` borders, styles and border types | any key quits |
| [calendar](widgets/calendar) | `Monthly`, with events and header styles | any key quits |
| [canvas](widgets/canvas) | `Canvas`: a world map, lines, rectangles and points on layers | any key quits |
| [chart](widgets/chart) | `Chart`: a line plot and a filled area plot on shared axes | any key quits |
| [collapsed-borders](widgets/collapsed-borders) | `Block.MergeBorders` and `Overlap` spacing, so neighbours share one border | arrows, `q` |
| [gauge](widgets/gauge) | `Gauge` and `LineGauge` | any key quits |
| [line-gauge](widgets/line-gauge) | `LineGauge` filling up on a timer | space, `r`, `q` |
| [list](widgets/list) | `List` with a moving selection, and a bottom-to-top list | `j`/`k`, arrows, `q` |
| [logo](widgets/logo) | `CatatuiLogo` in an inline viewport, leaving the scrollback alone | any key quits |
| [paragraph](widgets/paragraph) | `Paragraph` alignment, wrapping and styled spans | any key quits |
| [scrollbar](widgets/scrollbar) | `Scrollbar` on both axes, scrolling a paragraph | `h`/`j`/`k`/`l`, arrows, `q` |
| [shadow](widgets/shadow) | `Shadow`: overlay, block, symbol and dimmed, behind a popup | any key quits |
| [sparkline](widgets/sparkline) | `Sparkline`, including an animated sine wave | any key quits |
| [table](widgets/table) | `Table` with row, column and cell selection | `h`/`j`/`k`/`l`, `g`/`G`, `q` |
| [tabs](widgets/tabs) | `Tabs` drawn over the top border of a block | `h`/`l`, arrows, `q` |

Except for `hello`, every example is a port of the file or crate of the same
name in ratatui at `ratatui-v0.30.2`, and names its source in a header comment.
The split follows ratatui, where the two sets live in different crates: several
names, `table` and `gauge` among them, belong to one of each.

Each one also has a `render_test.go` that draws it at sizes from 0x0 up to
200x60. Drawing outside the area a widget is given panics in catatui, so those
tests are what keep the examples working as the library changes.

[ratatui-widgets/examples]: https://github.com/ratatui/ratatui/tree/ratatui-v0.30.2/ratatui-widgets/examples
[examples/apps]: https://github.com/ratatui/ratatui/tree/ratatui-v0.30.2/examples/apps
