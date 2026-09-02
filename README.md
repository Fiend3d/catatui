# catatui

A Go port of [ratatui](https://ratatui.rs) — immediate-mode terminal UI built on
a cell-addressed buffer, a constraint-solved layout, and a diffing terminal.

```go
import (
    "github.com/Fiend3d/catatui"
    "github.com/Fiend3d/catatui/term"
    "github.com/Fiend3d/catatui/widgets"
)

terminal, restore, err := term.Init(term.WithMouse())
if err != nil {
    return err
}
defer restore()

terminal.Draw(func(f *catatui.Frame) {
    rows := catatui.VerticalLayout(
        catatui.Length(3),
        catatui.Fill(1),
    ).Split(f.Area())

    f.RenderWidget(
        widgets.NewParagraph("hello").
            Block(widgets.Bordered().Title("catatui")).
            Wrap(widgets.Wrap{Trim: true}),
        rows[0])
})
```

Run `go run ./examples/hello` for a working application.

## Why

Go's existing TUI libraries take a different shape: Bubble Tea is the Elm
architecture composing styled strings, tview is retained-mode widgets on tcell.
Neither gives you ratatui's model — you draw into a `Buffer` of cells, a
Cassowary solver decides where things go, and the `Terminal` writes only what
changed.

This is a **near-literal port**, not a reinterpretation. Names, structure and
semantics follow ratatui 0.30.x (`ratatui-core` 0.1.2, `ratatui-widgets` 0.3.2),
and every ported file names its Rust source in a header comment.

## Status

Working and tested:

| Area | What is there |
|---|---|
| Style | `Color`, `Modifier`, `Style`, patching |
| Geometry | `Rect`, `Position`, `Size`, `Margin`, `Offset`, saturating `uint16` math |
| Text | `Span`, `Line`, `Text`, grapheme-cluster measurement |
| Buffer | `Cell`, `Buffer`, `SetSpan`/`SetLine`/`SetStringn`, frame diffing |
| Layout | `Constraint`, seven `Flex` modes, `Spacing`, `Split`, `SplitWithSpacers` |
| Terminal | `Widget`, `StatefulWidget`, `Frame`, double-buffered `Terminal`, `Viewport` |
| Backends | `catatui/term` (Windows + Unix), `TestBackend` |
| Terminal control | raw mode, alt screen, mouse, bracketed paste, focus, cursor shape |
| Symbols | box-drawing, block, bar, braille and half-block characters, border merging |
| Widgets | the whole `ratatui-widgets` library, listed below |

The widget library is complete: `Block` (borders, titles, padding, shadow,
merged borders), `Paragraph`, `List`, `Table`, `Tabs`, `Gauge`, `LineGauge`,
`Scrollbar`, `BarChart`, `Sparkline`, `Chart`, `Canvas` (with the line,
rectangle, circle, points and world-map shapes), `Calendar`, `Clear`, `Fill`,
`CatatuiLogo` and `RatatuiMascot`.

Adjacent blocks can collapse their borders into single box-drawing characters,
as ratatui does. `Block.MergeBorders` picks the strategy, `Cell.MergeSymbol`
does the work, and `symbols.MergeStrategy` holds the rules:

```go
widgets.Bordered().Render(catatui.NewRect(0, 0, 5, 5), buf)
widgets.Bordered().
    MergeBorders(symbols.MergeExact).
    Render(catatui.NewRect(4, 0, 5, 5), buf) // the shared edge becomes ┬ │ ┴
```

## Fidelity

The port is checked against ratatui's own tests rather than against
expectations written from scratch:

- **677 layout cases** translated mechanically from ratatui's `rstest` tables by
  `tools/gen_layout_tests.py`, covering every constraint type, every flex mode,
  spacing and overlap. Regenerate with
  `python tools/gen_layout_tests.py && gofmt -w layout_cases_test.go`.
- **kasuari's quadrilateral test**, which reproduces the Rust solver's exact
  values for a mixed system of required, weighted and inequality constraints.
- **ratatui's buffer, style, color, rect and widget tests**, ported by hand —
  both the unit tests next to each widget and the integration tests under
  `ratatui/tests`, including the word-wrap cases, which pin down whitespace
  handling at wrap points.
- **ratatui's `cell_width` cases**, which double as a conformance check that Go's
  `uniseg` agrees with Rust's `unicode-width`.
- **ratatui's three border-merging golden files**, copied verbatim into
  `widgets/testdata/block`. Each is the 43x1000 buffer ratatui's own test draws
  for all 100 pairs of border types meeting four different ways, and catatui
  reproduces all three exactly.

The rules that matter most, and that the tests pin down:

- A grapheme cluster wider than the space left is **dropped, not clipped**, and
  drawing stops there. The continuation columns of a wide cluster are **blanked,
  not styled**.
- There is exactly **one width function** in the tree, exported as
  `catatui.StringWidth`, and a test asserts that what it measures is exactly the
  number of columns `SetString` fills. Disagreeing width implementations are what
  made rows drift in the Go program that ratatui's author replaced.
- Coordinates are `uint16` with saturating arithmetic, as in ratatui.

## Deliberate deviations from ratatui

1. **`StatefulWidget` is generic** and renders through the free function
   `RenderStatefulWidget`, because Go methods cannot have type parameters.
2. **Unset colors use `Color`'s zero value** rather than `Option[Color]`, keeping
   `Style` comparable and allocation-free. `ColorReset` remains a distinct,
   explicit variant.
3. **Builders return copies of value types** rather than consuming the receiver,
   so a pre-builder value stays usable.
4. **Struct fields are unexported behind ratatui's builder names.** Rust allows a
   field and a method to share a name (`style.fg` and `style.fg(c)`); Go does
   not, so readers use `GetFg` and the builder keeps the ratatui name.
5. **One width function everywhere.** ratatui's `Span::width` calls
   `unicode-width` while `Buffer::set_stringn` uses `cell_width`; the two
   disagree on halfwidth katakana sound marks and some emoji, so a ratatui `Line`
   can report a width it does not draw. catatui measures everything the same way.
6. **The layout solver iterates in a deterministic order.** The Rust solver scans
   hash maps where the pivot choice is free; Go randomizes map order per run,
   which would make a degenerate layout come out differently on each launch.

## Dependencies

`github.com/rivo/uniseg`, `golang.org/x/sys`, `golang.org/x/term`. That is all.

## Examples

`examples/` holds one runnable program per widget, ported from ratatui's own
widget examples, plus `hello`, which shows the layout solver, direct buffer
drawing and an event loop that blocks when idle:

```sh
go run ./examples/hello
go run ./examples/chart
go run ./examples/collapsed-borders
```

See [examples/README.md](examples/README.md) for the full list and the keys
each one responds to.

## Development

```sh
go test ./...          # the full suite, examples included
go run ./examples/hello
```

`_ref/` holds read-only checkouts of ratatui and kasuari at the versions being
ported. It is gitignored; recreate it with:

```sh
git clone --depth 1 --branch ratatui-v0.30.2 https://github.com/ratatui/ratatui.git _ref/ratatui
git clone --depth 1 --branch v0.4.11 https://github.com/ratatui/kasuari.git _ref/kasuari
```
