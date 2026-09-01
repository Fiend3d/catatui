// Port of ratatui-core/src/style.rs @ ratatui-v0.30.2

package catatui

import (
	"strings"
)

// Modifier is a set of text attributes, held as a bitflag exactly as ratatui's
// bitflags-generated Modifier is.
//
// Note that not every terminal supports every modifier, and some that do
// support them render several identically.
type Modifier uint16

const (
	ModifierBold Modifier = 1 << iota
	ModifierDim
	ModifierItalic
	ModifierUnderlined
	ModifierSlowBlink
	ModifierRapidBlink
	ModifierReversed
	ModifierHidden
	ModifierCrossedOut

	// ModifierNone is the empty set.
	ModifierNone Modifier = 0
	// ModifierAll is every modifier at once. Style.Reset uses it as the
	// sub_modifier so that a reset clears whatever was set beneath it.
	ModifierAll Modifier = 1<<9 - 1
)

// Contains reports whether every modifier in other is set in m.
func (m Modifier) Contains(other Modifier) bool { return m&other == other }

// Intersects reports whether any modifier in other is set in m.
func (m Modifier) Intersects(other Modifier) bool { return m&other != 0 }

// Insert returns m with every modifier in other added.
func (m Modifier) Insert(other Modifier) Modifier { return m | other }

// Remove returns m with every modifier in other cleared.
func (m Modifier) Remove(other Modifier) Modifier { return m &^ other }

// IsEmpty reports whether no modifiers are set.
func (m Modifier) IsEmpty() bool { return m == 0 }

var modifierNames = []struct {
	bit  Modifier
	name string
}{
	{ModifierBold, "BOLD"},
	{ModifierDim, "DIM"},
	{ModifierItalic, "ITALIC"},
	{ModifierUnderlined, "UNDERLINED"},
	{ModifierSlowBlink, "SLOW_BLINK"},
	{ModifierRapidBlink, "RAPID_BLINK"},
	{ModifierReversed, "REVERSED"},
	{ModifierHidden, "HIDDEN"},
	{ModifierCrossedOut, "CROSSED_OUT"},
}

// String formats the set the way the bitflags crate does, e.g. "BOLD | ITALIC"
// or "NONE" when empty.
func (m Modifier) String() string {
	if m == 0 {
		return "NONE"
	}
	var parts []string
	for _, n := range modifierNames {
		if m&n.bit != 0 {
			parts = append(parts, n.name)
		}
	}
	return strings.Join(parts, " | ")
}

// Style is a set of optional styling attributes applied to a run of text.
//
// A Style is a *diff* against whatever is already in the cell, not an absolute
// description of it. An unset color or an absent modifier leaves the underlying
// value alone; this is what lets styles be layered with Patch. Use ResetStyle
// for a Style that overrides everything.
//
// Style is comparable, so styles can be compared with == and used as map keys.
//
//	style := NewStyle().Fg(ColorRed).Bg(ColorBlack).AddModifier(ModifierBold)
//
// Deviation from ratatui: Rust exposes both public fields (`style.fg`) and
// same-named builder methods (`style.fg(color)`). Go allows only one, so the
// fields are unexported, the builders keep ratatui's names, and readers use the
// Get* accessors.
type Style struct {
	fg             Color
	bg             Color
	underlineColor Color
	addModifier    Modifier
	subModifier    Modifier
}

// NewStyle returns the empty style, which changes nothing. It is identical to
// the zero value and exists for symmetry with ratatui's Style::new().
func NewStyle() Style { return Style{} }

// ResetStyle returns a Style that resets every property: all three colors are
// set to ColorReset and every modifier is cleared.
func ResetStyle() Style {
	return Style{
		fg:             ColorReset,
		bg:             ColorReset,
		underlineColor: ColorReset,
		subModifier:    ModifierAll,
	}
}

// Fg returns a copy of s with the foreground color set.
func (s Style) Fg(c Color) Style { s.fg = c; return s }

// Bg returns a copy of s with the background color set.
func (s Style) Bg(c Color) Style { s.bg = c; return s }

// UnderlineColor returns a copy of s with the underline color set. Only some
// terminals support colouring the underline separately from the text.
func (s Style) UnderlineColor(c Color) Style { s.underlineColor = c; return s }

// AddModifier returns a copy of s with the given modifiers added, and removed
// from the set of modifiers s clears.
func (s Style) AddModifier(m Modifier) Style {
	s.subModifier = s.subModifier.Remove(m)
	s.addModifier = s.addModifier.Insert(m)
	return s
}

// RemoveModifier returns a copy of s with the given modifiers cleared, and
// removed from the set of modifiers s adds.
func (s Style) RemoveModifier(m Modifier) Style {
	s.addModifier = s.addModifier.Remove(m)
	s.subModifier = s.subModifier.Insert(m)
	return s
}

// GetFg returns the foreground color, which may be unset.
func (s Style) GetFg() Color { return s.fg }

// GetBg returns the background color, which may be unset.
func (s Style) GetBg() Color { return s.bg }

// GetUnderlineColor returns the underline color, which may be unset.
func (s Style) GetUnderlineColor() Color { return s.underlineColor }

// GetAddModifier returns the modifiers this style turns on.
func (s Style) GetAddModifier() Modifier { return s.addModifier }

// GetSubModifier returns the modifiers this style turns off.
func (s Style) GetSubModifier() Modifier { return s.subModifier }

// Patch layers other on top of s and returns the result. Set properties of
// other win; unset ones leave s's values in place.
//
// Patching is associative: patching with the result of patching two styles
// together is the same as patching with each in turn.
func (s Style) Patch(other Style) Style {
	if other.fg.IsSet() {
		s.fg = other.fg
	}
	if other.bg.IsSet() {
		s.bg = other.bg
	}
	if other.underlineColor.IsSet() {
		s.underlineColor = other.underlineColor
	}
	s.addModifier = s.addModifier.Remove(other.subModifier).Insert(other.addModifier)
	s.subModifier = s.subModifier.Remove(other.addModifier).Insert(other.subModifier)
	return s
}

// IsEmpty reports whether the style changes nothing.
func (s Style) IsEmpty() bool { return s == Style{} }
