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
| Symbols | box-drawing, block, bar, braille and half-block characters |
| Widgets | `Block` (borders, titles, padding), `Paragraph` (wrap, scroll, align) |

Not yet written: the rest of the widget library — `List`, `Table`, `Tabs`,
`Gauge`, `Scrollbar`, `BarChart`, `Sparkline`, `Chart` and `Canvas`. Anything
they would do can be done by drawing into `Frame.Buffer()` directly, which is
how [nezumi](https://github.com/Fiend3d/nezumi) uses ratatui anyway.

## Fidelity

The port is checked against ratatui's own tests rather than against
expectations written from scratch:

- **677 layout cases** translated mechanically from ratatui's `rstest` tables by
  `tools/gen_layout_tests.py`, covering every constraint type, every flex mode,
  spacing and overlap. Regenerate with
  `python tools/gen_layout_tests.py && gofmt -w layout_cases_test.go`.
- **kasuari's quadrilateral test**, which reproduces the Rust solver's exact
  values for a mixed system of required, weighted and inequality constraints.
- **ratatui's buffer, style, color, rect, block and paragraph tests**, ported by
  hand — including the word-wrap cases, which pin down whitespace handling at
  wrap points.
- **ratatui's `cell_width` cases**, which double as a conformance check that Go's
  `uniseg` agrees with Rust's `unicode-width`.

The rules that matter most, and that the tests pin down:

- A grapheme cluster wider than the space left is **dropped, not clipped**, and
  drawing stops there. The continuation columns of a wide cluster are **blanked,
  not styled**.
- There is exactly **one width function** in the tree. Disagreeing width
  implementations are what made rows drift in the Go program that ratatui's
  author replaced.
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

## Development

```sh
go test ./...          # the full suite
go run ./examples/hello
```

`_ref/` holds read-only checkouts of ratatui and kasuari at the versions being
ported. It is gitignored; recreate it with:

```sh
git clone --depth 1 --branch ratatui-v0.30.2 https://github.com/ratatui/ratatui.git _ref/ratatui
git clone --depth 1 --branch v0.4.11 https://github.com/ratatui/kasuari.git _ref/kasuari
```
