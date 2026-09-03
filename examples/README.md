# Examples

Each directory is a self-contained program: whole applications under
[apps](apps), one program per widget under [widgets](widgets). Run one with:

```sh
go run ./examples/apps/<name>
go run ./examples/widgets/<name>
```

The widget examples draw a single frame and quit on the first key, and so do
the two applications that are really one picture; the rest keep running. Each
says what keys it responds to in the table below and in the doc comment at the
top of its `main.go`.

Nothing is shared between examples on purpose: each carries its own terminal
setup and event loop so that it can be read start to finish, and pasted into an
application as it stands. That is ratatui's convention for its own examples, and
the reason the same twenty lines of scaffolding appear in every one of them.

## Applications

Whole programs, ported from ratatui's [examples/apps]:

| Example | Shows | Keys |
|---|---|---|
| [calendar-explorer](apps/calendar-explorer) | A year of `Monthly` calendars with the holidays marked, in each of the widget's styles | `s`, `n`/`p`, arrows or `h`/`j`/`k`/`l`, `q` |
| [canvas](apps/canvas) | Four canvases at once: a world map, a scratchpad, a bouncing ball and a ruler, in every marker | Enter, arrows or `h`/`j`/`k`/`l`, mouse, `q` |
| [chart](apps/chart) | A scrolling pair of sine waves beside a bar, a line and a scatter chart | `q` |
| [color-explorer](apps/color-explorer) | Every colour catatui can name or number, as text and as backgrounds | any key quits |
| [constraints](apps/constraints) | What each kind of `Constraint` gives way to, a tab per kind | arrows or `h`/`j`/`k`/`l`, `g`/`G`, `q` |
| [custom-widget](apps/custom-widget) | Writing a widget from scratch: a button drawn cell by cell, driven by the mouse | arrows or `h`/`l`, Space, mouse, `q` |
| [demo](apps/demo) | The original tui-rs demo: three tabs of gauges, lists, charts, a table and a world map, animating on a tick | arrows or `h`/`j`/`k`/`l`, `t`, `q` |
| [flex](apps/flex) | What each `Flex` mode does to the same constraints, side by side | arrows, `g`/`G`, `-`/`+`, `q` |
| [gauge](apps/gauge) | Four gauges filling at once: percentage against ratio, whole cells against eighths | Enter, `q` |
| [hello](apps/hello) | Layout, direct buffer drawing, and an event loop that blocks when idle | arrows, mouse, `q` |
| [inline](apps/inline) | An inline viewport, and `InsertBefore` pushing finished lines into the scrollback | `q` |
| [input-form](apps/input-form) | Moving the focus between fields, and putting the cursor where the focused one wants it | Tab, Enter, Esc |
| [modifiers](apps/modifiers) | Every modifier against five foreground and background colours, so you can see which your terminal renders | any key quits |
| [mouse-drawing](apps/mouse-drawing) | Drawing with the mouse, joining up a drag with Bresenham's line algorithm | mouse, Space, `q` |
| [panic](apps/panic) | What a panic does to the terminal with `term.RecoverAndRestore` deferred, and without it | `p`, `e`, `h`, `q` |
| [popup](apps/popup) | A popup drawn over the rest of the UI, with `Clear` behind it | `p`, `q` |
| [release-header](apps/release-header) | A fixed viewport, the logo over a gradient, and two blocks merging their borders | any key quits |
| [scrollbar](apps/scrollbar) | Four `Scrollbar`s over the same text, differing in arrows, track and thumb | arrows or `h`/`j`/`k`/`l`, `q` |
| [table](apps/table) | An interactive `Table` with a scrollbar, row, column and cell selection, and swappable colours | arrows or `h`/`j`/`k`/`l`, Shift with either, `q` |
| [todo-list](apps/todo-list) | A list whose items are selected and ticked off, with the selection kept in a `ListState` | arrows or `h`/`j`/`k`/`l`, `g`/`G`, Enter, `q` |
| [tracing](apps/tracing) | Logging to a file with `log/slog` while the terminal is given over to the UI | `q` |
| [user-input](apps/user-input) | Typing into an input box: editing modes, and where the cursor goes | `e`, Esc, Enter, arrows, `q` |
| [weather](apps/weather) | A day of temperatures as a `BarChart`, coloured yellow through red | any key quits |

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
