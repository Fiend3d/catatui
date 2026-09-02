# Documentation

The API reference is on [pkg.go.dev]. These guides are the prose around it: what
the pieces are for, which rules matter, and where catatui deliberately differs
from ratatui. Each page says at the top what it covers and who it is for, and
ends with the differences from ratatui in that area.

| Guide | Covers |
|---|---|
| [Rendering](concepts/rendering.md) | The immediate-mode cycle, `Frame`, `Buffer`, `Cell`, wide graphemes, the diff, writing a widget |
| [Layout](concepts/layout.md) | `Rect`, the six constraints, seven flex modes, spacing, nesting, recipes |
| [Widgets](concepts/widgets.md) | The catalogue, builder conventions, blocks and inner areas, widget state |
| [Text and style](concepts/text-and-style.md) | `Color`, `Modifier`, `Style` patching, `Span`/`Line`/`Text`, how styles layer |
| [Events](concepts/events.md) | `EventReader`, the `Event` kinds, the drain-the-queue loop, app structure |
| [Terminal](concepts/terminal.md) | `term.Init` and its options, raw mode, viewports, `InsertBefore`, panic recovery |
| [Testing](concepts/testing.md) | Rendering into a buffer, `AssertBuffer`, driving an app through `TestBackend` |

New to the library? Read [rendering](concepts/rendering.md) and
[layout](concepts/layout.md), in that order, then pick the page for whatever you
are building. The [examples](../examples) are a runnable program per widget.

Every whole-file Go snippet on these pages is compiled by
`tools/check_doc_snippets.py`, so nothing here can drift from the API without
the check failing.

[pkg.go.dev]: https://pkg.go.dev/github.com/Fiend3d/catatui
