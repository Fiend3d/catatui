// Port of ratatui-widgets/src/block/shadow.rs @ ratatui-v0.30.2

package widgets

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// CellEffect modifies the cells covered by a Shadow once the shadow's style
// has been applied. Implement it to draw a shadow of your own.
//
//	type checker struct{}
//
//	func (checker) Apply(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer) {
//		ForEachShadowCell(shadowArea, baseArea, buf, func(x, y uint16, buf *catatui.Buffer) {
//			if (x+y)%2 == 0 {
//				buf.Get(x, y).SetSymbol("░")
//			}
//		})
//	}
//
//	block := widgets.Bordered().Shadow(widgets.ShadowCustom(checker{}))
type CellEffect interface {
	// Apply changes the cells in shadowArea. baseArea is the area of the
	// widget casting the shadow; cells inside it are normally left alone.
	Apply(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer)
}

// shadowEffectKind tags the built-in effects.
type shadowEffectKind uint8

const (
	// shadowOverlay applies no symbol changes and only keeps the shadow style.
	shadowOverlay shadowEffectKind = iota
	// shadowSymbol fills the shadow area with a single symbol.
	shadowSymbol
	// shadowCustom applies a user-defined shadow effect.
	shadowCustom
)

// Shadow is a configurable shadow drawn behind a Block.
//
// It is drawn in an area offset from the block. Its Style is applied first,
// then an optional cell effect can modify the affected cells, for example by
// filling them with a shading symbol or dimming the existing background.
//
// Built-in presets:
//
//   - ShadowOverlay applies only style
//   - ShadowBlock fills with full block symbols
//   - ShadowLightShade, ShadowMediumShade and ShadowDarkShade fill with shade
//     symbols
//
// A dark shade shadow one cell right and one cell down of a block looks like
// this:
//
//	┌Popup─────┐
//	│content   │▒
//	└──────────┘▒
//	  ▒▒▒▒▒▒▒▒▒▒▒
//
// Shadows are attached with Block.Shadow:
//
//	block := widgets.Bordered().Title("Popup").Shadow(
//		widgets.ShadowDarkShade().
//			Style(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorWhite)).
//			Offset(catatui.Offset{X: 2, Y: 1}))
type Shadow struct {
	kind   shadowEffectKind
	symbol string
	custom CellEffect
	style  catatui.Style
	offset catatui.Offset
}

// ShadowOverlay returns a shadow that only applies style to the offset area,
// leaving the existing cell symbols unchanged. It is the default shadow.
func ShadowOverlay() Shadow {
	return Shadow{kind: shadowOverlay, offset: catatui.Offset{X: 1, Y: 1}}
}

// ShadowBlock returns a shadow filled with full block symbols.
func ShadowBlock() Shadow { return ShadowSymbol(symbols.ShadeFull) }

// ShadowLightShade returns a shadow filled with light shade symbols.
func ShadowLightShade() Shadow { return ShadowSymbol(symbols.ShadeLight) }

// ShadowMediumShade returns a shadow filled with medium shade symbols.
func ShadowMediumShade() Shadow { return ShadowSymbol(symbols.ShadeMedium) }

// ShadowDarkShade returns a shadow filled with dark shade symbols.
func ShadowDarkShade() Shadow { return ShadowSymbol(symbols.ShadeDark) }

// ShadowSymbol returns a shadow filled with the given symbol.
func ShadowSymbol(symbol string) Shadow {
	return Shadow{kind: shadowSymbol, symbol: symbol, offset: catatui.Offset{X: 1, Y: 1}}
}

// ShadowCustom returns a shadow drawn by a custom cell effect.
//
// The effect receives the shadow area, the original block area and the target
// buffer. It is called after the shadow style has been applied.
func ShadowCustom(effect CellEffect) Shadow {
	return Shadow{kind: shadowCustom, custom: effect, offset: catatui.Offset{X: 1, Y: 1}}
}

// NewShadow returns a shadow drawn by a custom cell effect. It is an alias for
// ShadowCustom, matching ratatui's Shadow::new.
func NewShadow(effect CellEffect) Shadow { return ShadowCustom(effect) }

// Style returns a copy of s with the style applied to the shadow area.
func (s Shadow) Style(style catatui.Style) Shadow { s.style = style; return s }

// Offset returns a copy of s displaced by the given offset relative to the
// area casting it. Positive X moves the shadow right, positive Y moves it down.
func (s Shadow) Offset(o catatui.Offset) Shadow { s.offset = o; return s }

// GetStyle returns the style applied to the shadow area.
func (s Shadow) GetStyle() catatui.Style { return s.style }

// GetOffset returns the shadow's displacement.
func (s Shadow) GetOffset() catatui.Offset { return s.offset }

// Render draws the shadow for a widget occupying area: the style is applied to
// every cell of the offset area that is not covered by area, then the effect
// runs.
func (s Shadow) Render(area catatui.Rect, buf *catatui.Buffer) {
	shadowArea := area.Offset(s.offset).Intersection(buf.Area)

	// Apply style
	for y := shadowArea.Top(); y < shadowArea.Bottom(); y++ {
		for x := shadowArea.Left(); x < shadowArea.Right(); x++ {
			if area.Contains(catatui.Position{X: x, Y: y}) {
				continue
			}
			buf.Get(x, y).SetStyle(s.style)
		}
	}

	// Apply effect
	s.applyEffect(shadowArea, area, buf)
}

// applyEffect runs the effect over the shadow area.
func (s Shadow) applyEffect(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer) {
	switch s.kind {
	case shadowOverlay:
	case shadowSymbol:
		ForEachShadowCell(shadowArea, baseArea, buf, func(x, y uint16, buf *catatui.Buffer) {
			buf.Get(x, y).SetSymbol(s.symbol)
		})
	case shadowCustom:
		if s.custom != nil {
			s.custom.Apply(shadowArea, baseArea, buf)
		}
	}
}

// Dimmed is a CellEffect that dims the shadow cells by setting ModifierDim.
//
// If the cell background is RGB, each channel is halved. Otherwise the
// background is replaced with ColorBlack.
type Dimmed struct{}

// Apply implements CellEffect.
func (Dimmed) Apply(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer) {
	ForEachShadowCell(shadowArea, baseArea, buf, func(x, y uint16, buf *catatui.Buffer) {
		cell := buf.Get(x, y)
		cell.Modifier = cell.Modifier.Insert(catatui.ModifierDim)
		if r, g, b, ok := cell.Bg.RGB(); ok {
			cell.Bg = catatui.Rgb(r/2, g/2, b/2)
		} else {
			cell.Bg = catatui.ColorBlack
		}
	})
}

// NewDimmed returns a Dimmed shadow effect. It matches ratatui's dimmed().
func NewDimmed() Dimmed { return Dimmed{} }

// ForEachShadowCell calls f for every cell of shadowArea that is not covered by
// baseArea. Custom effects use it to skip the widget casting the shadow.
func ForEachShadowCell(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer, f func(x, y uint16, buf *catatui.Buffer)) {
	for y := shadowArea.Top(); y < shadowArea.Bottom(); y++ {
		for x := shadowArea.Left(); x < shadowArea.Right(); x++ {
			if baseArea.Contains(catatui.Position{X: x, Y: y}) {
				continue
			}
			f(x, y, buf)
		}
	}
}

var (
	_ catatui.Widget = Shadow{}
	_ CellEffect     = Dimmed{}
)
