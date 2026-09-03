# Text and style

This page describes how catatui represents colors, text attributes, styles and
styled text, and how the layers combine when a widget draws. It is for anyone
who wants text on screen in something other than the terminal's default colors,
and for widget authors who need to know which style wins when several apply.

## Color

A `Color` is one of five things: a named ANSI color, an 8-bit indexed color, a
24-bit RGB color, an explicit reset, or unset.

| Kind | How to write it | Accessor |
|---|---|---|
| Named | `catatui.ColorRed`, `catatui.ColorLightBlue`, ... (sixteen of them) | `Named()` returns the ordinal 0..15 |
| Indexed | `catatui.Indexed(208)` | `Index()` |
| RGB | `catatui.Rgb(255, 128, 0)` or `catatui.RgbFromU32(0xFF8000)` | `RGB()` |
| Reset | `catatui.ColorReset` | `IsReset()` |
| Unset | the zero value, `catatui.Color{}` | `IsSet()` is false |

The sixteen named colors are `Black`, `Red`, `Green`, `Yellow`, `Blue`,
`Magenta`, `Cyan`, `Gray`, `DarkGray`, `LightRed`, `LightGreen`, `LightYellow`,
`LightBlue`, `LightMagenta`, `LightCyan` and `White`, each prefixed with `Color`.

The distinction between reset and unset is the one to internalize. `ColorReset`
means "use the terminal's default"; it actively overrides whatever color was
there before. Unset means "do not touch"; a style whose foreground is unset
leaves the cell's existing foreground alone. `Cell` colors are always concrete
(a blank cell holds `ColorReset`), while `Style` colors are frequently unset.

`ParseColor` accepts the spellings ratatui does: color names with or without
`light`/`bright`, `grey`/`gray`/`silver`, spaces, dashes or underscores, a
decimal index such as `"208"`, or a hex string such as `"#FF8000"`.
`Color.String()` produces the canonical form, and the two round-trip.

```go
package main

import "github.com/Fiend3d/catatui"

func accent() catatui.Color {
	c, err := catatui.ParseColor("light blue")
	if err != nil {
		return catatui.ColorBlue
	}
	return c
}
```

**Differences from ratatui.** ratatui models "no color" as `Option<Color>` inside
`Style`, and `Color::default()` is `Reset`. catatui uses the `Color` zero value
for unset so that `Style` stays comparable and allocation-free, and `ColorReset`
is a distinct explicit variant. Everywhere else the two behave exactly like
`None` and `Some(Reset)`.

## Palettes

The sixteen named colours are whatever the terminal's theme says they are, which
is fine for accents and awkward for a design that wants specific shades. For
those, `catatui/palette/tailwind` and `catatui/palette/material` hold the
Tailwind CSS and Material design ramps that ratatui ships: one palette per hue,
in shades from `C50` at the lightest to `C950` (tailwind) or `C900` (material)
at the darkest, plus `A100` to `A700` accents on most Material palettes.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
)

func panel() catatui.Style {
	return catatui.NewStyle().
		Fg(tailwind.Slate.C200).
		Bg(tailwind.Slate.C900)
}
```

Shades from one ramp read as a set, which is the point: picking `C700` for a
border and `C900` for the fill behind it gives a contrast that holds up, where
two hand-picked RGB values usually do not.

These are 24-bit colours. A terminal without true colour support approximates
them, so where that matters keep an indexed fallback beside them, as the
[flex example](../../examples/apps/flex) does.

## Modifier

`Modifier` is a bit set of text attributes, matching ratatui's bitflags. Combine
them with `|`.

| Constant | Attribute |
|---|---|
| `ModifierBold` | Bold. |
| `ModifierDim` | Dim or faint. |
| `ModifierItalic` | Italic. |
| `ModifierUnderlined` | Underlined. |
| `ModifierSlowBlink`, `ModifierRapidBlink` | Blink. |
| `ModifierReversed` | Swap foreground and background. |
| `ModifierHidden` | Invisible. |
| `ModifierCrossedOut` | Strikethrough. |
| `ModifierNone` | The empty set. |
| `ModifierAll` | Every bit. |

`Contains`, `Intersects`, `Insert`, `Remove` and `IsEmpty` do what their names
say, and `String()` prints the set as `BOLD | ITALIC` the way the bitflags crate
does. Not every terminal renders every modifier, and some render several the
same way.

## Style and Patch semantics

A `Style` is a diff against whatever is already in a cell, not a full
description of the cell. It holds a foreground, a background, an underline
color, a set of modifiers to add and a set of modifiers to remove, and any of
them may be left unset. `NewStyle()` is the empty style and changes nothing;
`ResetStyle()` sets all three colors to `ColorReset` and removes every modifier.

Builders return copies, so a base style is safe to reuse.

```go
package main

import "github.com/Fiend3d/catatui"

var (
	base     = catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorBlack)
	warning  = base.Fg(catatui.ColorYellow).AddModifier(catatui.ModifierBold)
	disabled = base.AddModifier(catatui.ModifierDim).RemoveModifier(catatui.ModifierBold)
)
```

Readers are `GetFg`, `GetBg`, `GetUnderlineColor`, `GetAddModifier` and
`GetSubModifier`. `AddModifier(m)` also removes `m` from the sub set, and
`RemoveModifier(m)` also removes it from the add set, so a style never both adds
and removes the same bit.

`Patch` layers one style over another and is the operation everything else is
built on. `a.Patch(b)` keeps `a`'s values wherever `b` is unset and takes `b`'s
wherever it is set. For modifiers, `b`'s removals are applied and then `b`'s
additions. Patching is associative: `a.Patch(b).Patch(c)` equals
`a.Patch(b.Patch(c))`.

```go
package main

import "github.com/Fiend3d/catatui"

func example() catatui.Style {
	a := catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorBlue)
	b := catatui.NewStyle().Fg(catatui.ColorGreen).AddModifier(catatui.ModifierBold)
	// fg=Green, bg=Blue, add=BOLD
	return a.Patch(b)
}
```

`Cell.SetStyle` is the same operation applied to a cell: set colors overwrite
the cell's colors, unset ones are kept, and the modifier sets are applied.
`Buffer.SetStyle(area, style)` does that for every cell in an area without
touching the symbols, which is how a widget paints a background under text it
has not drawn yet.

**Differences from ratatui.** Rust lets a field and a method share a name, so
ratatui has both `style.fg` and `style.fg(color)`. Go does not, so the fields are
unexported, the builders keep the ratatui names, and readers are the `Get*`
methods.

## Span, Line and Text

Styled text is a three-level hierarchy. A `Span` is a run of text in one style
and never contains a line break. A `Line` is a row of spans with its own style
and alignment. A `Text` is a block of lines with its own style and alignment.

| Level | Constructors | Own style | Own alignment |
|---|---|---|---|
| `Span` | `NewSpan(s)`, `NewStyledSpan(s, style)` | yes | no |
| `Line` | `NewLine(spans...)`, `LineFromString(s)`, `LineFromStyledString(s, style)` | yes | yes |
| `Text` | `NewText(lines...)`, `TextFromString(s)`, `TextFromStyledString(s, style)` | yes | yes |

`TextFromString` splits on newlines, so a multi-line string becomes several
lines, and a trailing `\r` is stripped from each. `LineFromStyledString` puts
the style on the span, matching ratatui's `Line::styled`, while
`TextFromStyledString` puts it on the text as a whole.

Each level has `Style(s)` to replace its style and `Patch(s)` to layer a style
onto it, plus `Width()` measured with `StringWidth`. `Line` and `Text` add
`Alignment(a)` and the shorthands `Left()`, `Centered()` and `Right()`. Readers
are `GetContent`, `GetSpans`, `GetLines`, `GetStyle` and `GetAlignment`.

All three implement `Widget`, so a `Line` can be rendered straight into a frame.

```go
package main

import "github.com/Fiend3d/catatui"

func statusLine(name string, dirty bool) catatui.Line {
	spans := []catatui.Span{
		catatui.NewStyledSpan(" "+name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
	}
	if dirty {
		spans = append(spans,
			catatui.NewStyledSpan(" [+]", catatui.NewStyle().Fg(catatui.ColorYellow)))
	}
	return catatui.NewLine(spans...).
		Style(catatui.NewStyle().Bg(catatui.ColorDarkGray))
}

func drawStatus(f *catatui.Frame, area catatui.Rect) {
	f.RenderWidget(statusLine("main.go", true), area)
}
```

## Alignment and its fallback rules

`Alignment` has four values: `AlignmentNone`, `AlignmentLeft`,
`AlignmentCenter` and `AlignmentRight`. `AlignmentNone` is the zero value and
means "not specified": the containing widget decides. This stands in for
ratatui's `Option<Alignment>`.

The fallback chain runs from the innermost level outward. A line with its own
alignment uses it. A line with `AlignmentNone` uses its `Text`'s alignment, or
the `Paragraph`'s, or a `Block`'s title alignment, depending on what is rendering
it. If nothing along the chain sets one, the result is left alignment.
`Line.RenderWithFallbackAlignment(area, buf, parent)` is the method container
widgets call to impose their alignment on the lines they hold.

A line wider than its area is truncated on the right, and the alignment decides
what is dropped: a right-aligned line loses its beginning, a centered one loses
half from each side, and a left-aligned one loses its end. Truncation happens on
grapheme boundaries, so a wide character straddling the cut is dropped whole.

**Differences from ratatui.** ratatui's standalone `Alignment` defaults to
`Left`, while catatui's zero value is `AlignmentNone`. Write `AlignmentLeft`
when you mean left rather than relying on the zero value.

## StringWidth

`catatui.StringWidth(s)` returns the number of terminal columns `s` occupies
when drawn. It counts grapheme clusters, treats CJK and emoji as two columns,
gives halfwidth katakana sound marks a column of their own, and skips control
characters and zero-width clusters. What it returns is exactly the number of
columns `Buffer.SetString` fills, and a test enforces that.

Use it for anything that reasons about width outside a buffer: sizing a status
bar, deciding whether a label fits, placing a cursor after typed text.

```go
package main

import "github.com/Fiend3d/catatui"

func truncateLabel(label string, width int) string {
	if catatui.StringWidth(label) <= width {
		return label
	}
	var out string
	used := 0
	for g := range catatui.AllGraphemes(label) {
		if used+int(g.Width) > width {
			break
		}
		out += g.Symbol
		used += int(g.Width)
	}
	return out
}
```

## How styles layer

When a `Paragraph` draws, four styles can apply to one cell. They are applied
from the outside in, each patched over the previous, so the innermost set value
wins and unset values fall through.

| Layer | Set by | Applied how |
|---|---|---|
| 1. Widget style | `Paragraph.Style`, `Block.Style` | `Buffer.SetStyle` over the whole area, before any text |
| 2. Text style | `Text.Style` | Patched under each line's style |
| 3. Line style | `Line.Style` | `Buffer.SetStyle` over the row, then patched under each span |
| 4. Span style | `Span.Style` | Drawn with the text itself |

Concretely, `Buffer.SetLine` draws every span with `line.style.Patch(span.style)`,
and `Paragraph` first replaces each line's style with
`text.style.Patch(line.style)`. A span that sets only a foreground therefore
keeps the line's background, which keeps the paragraph's background, which keeps
the block's background.

```go
package main

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

func layered() widgets.Paragraph {
	line := catatui.NewLine(
		catatui.NewSpan("plain "),
		catatui.NewStyledSpan("bold", catatui.NewStyle().AddModifier(catatui.ModifierBold)),
	).Style(catatui.NewStyle().Fg(catatui.ColorCyan))

	text := catatui.NewText(line).Style(catatui.NewStyle().Bg(catatui.ColorBlack))

	return widgets.NewParagraphFromText(text).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite))
}
```

In that example the `"bold"` span ends up cyan on black and bold: the paragraph's
white foreground is overridden by the line's cyan, the text's black background
survives because nothing inside sets one, and the span adds bold.

One more rule follows from how `SetStringn` works: the continuation cells of a
wide character are reset to blank cells, not given the style being drawn, and
that reset happens after any `Buffer.SetStyle` that painted the area. This is
correct because the terminal draws the whole wide glyph with the attributes of
its first cell, and the frame diff never writes the continuation column while
the wide character is there. Do not read a continuation cell's style back and
expect it to match its neighbour.
