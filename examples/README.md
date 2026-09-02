# Examples

Each directory is a self-contained program. Run one with:

```sh
go run ./examples/<name>
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
| [demo](demo) | The original tui-rs demo: three tabs of gauges, lists, charts, a table and a world map, animating on a tick | arrows or `h`/`j`/`k`/`l`, `t`, `q` |
| [flex](flex) | What each `Flex` mode does to the same constraints, side by side | arrows, `g`/`G`, `-`/`+`, `q` |
| [hello](hello) | Layout, direct buffer drawing, and an event loop that blocks when idle | arrows, mouse, `q` |

## Widgets

One program per widget, ported from [ratatui-widgets/examples]:

| Example | Shows | Keys |
|---|---|---|
| [barchart](barchart) | `BarChart` drawn both ways up | any key quits |
| [barchart-grouped](barchart-grouped) | `BarChart` with grouped bars and value labels | any key quits |
| [block](block) | `Block` borders, styles and border types | any key quits |
| [calendar](calendar) | `Monthly`, with events and header styles | any key quits |
| [canvas](canvas) | `Canvas`: a world map, lines, rectangles and points on layers | any key quits |
| [chart](chart) | `Chart`: a line plot and a filled area plot on shared axes | any key quits |
| [collapsed-borders](collapsed-borders) | `Block.MergeBorders` and `Overlap` spacing, so neighbours share one border | arrows, `q` |
| [gauge](gauge) | `Gauge` and `LineGauge` | any key quits |
| [line-gauge](line-gauge) | `LineGauge` filling up on a timer | space, `r`, `q` |
| [list](list) | `List` with a moving selection, and a bottom-to-top list | `j`/`k`, arrows, `q` |
| [logo](logo) | `CatatuiLogo` in an inline viewport, leaving the scrollback alone | any key quits |
| [paragraph](paragraph) | `Paragraph` alignment, wrapping and styled spans | any key quits |
| [scrollbar](scrollbar) | `Scrollbar` on both axes, scrolling a paragraph | `h`/`j`/`k`/`l`, arrows, `q` |
| [shadow](shadow) | `Shadow`: overlay, block, symbol and dimmed, behind a popup | any key quits |
| [sparkline](sparkline) | `Sparkline`, including an animated sine wave | any key quits |
| [table](table) | `Table` with row, column and cell selection | `h`/`j`/`k`/`l`, `g`/`G`, `q` |
| [tabs](tabs) | `Tabs` drawn over the top border of a block | `h`/`l`, arrows, `q` |

Except for `hello`, every example is a port of the file or crate of the same
name in ratatui at `ratatui-v0.30.2`, and names its source in a header comment.

Each one also has a `render_test.go` that draws it at sizes from 0x0 up to
200x60. Drawing outside the area a widget is given panics in catatui, so those
tests are what keep the examples working as the library changes.

[ratatui-widgets/examples]: https://github.com/ratatui/ratatui/tree/ratatui-v0.30.2/ratatui-widgets/examples
[examples/apps]: https://github.com/ratatui/ratatui/tree/ratatui-v0.30.2/examples/apps
