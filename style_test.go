// Tests ported from ratatui-core/src/style.rs @ ratatui-v0.30.2

package catatui

import "testing"

// patchStyles is ratatui's fixture for the associativity property below.
var patchStyles = []Style{
	NewStyle(),
	NewStyle().Fg(ColorYellow),
	NewStyle().Bg(ColorYellow),
	NewStyle().AddModifier(ModifierBold),
	NewStyle().RemoveModifier(ModifierBold),
	NewStyle().AddModifier(ModifierItalic),
	NewStyle().RemoveModifier(ModifierItalic),
	NewStyle().AddModifier(ModifierItalic | ModifierBold),
	NewStyle().RemoveModifier(ModifierItalic | ModifierBold),
}

// TestCombinedPatchGivesSameResultAsIndividualPatch is the associativity
// property that makes style layering safe: collapsing a chain of styles into
// one must equal applying them one at a time.
func TestCombinedPatchGivesSameResultAsIndividualPatch(t *testing.T) {
	for _, a := range patchStyles {
		for _, b := range patchStyles {
			for _, c := range patchStyles {
				for _, d := range patchStyles {
					seq := NewStyle().Patch(a).Patch(b).Patch(c).Patch(d)
					combined := NewStyle().Patch(a.Patch(b.Patch(c.Patch(d))))
					if seq != combined {
						t.Fatalf("patch is not associative:\n  sequential = %+v\n  combined   = %+v\n  a=%+v b=%+v c=%+v d=%+v",
							seq, combined, a, b, c, d)
					}
				}
			}
		}
	}
}

func TestCombineIndividualModifiers(t *testing.T) {
	mods := []Modifier{
		ModifierBold, ModifierDim, ModifierItalic, ModifierUnderlined,
		ModifierSlowBlink, ModifierRapidBlink, ModifierReversed,
		ModifierHidden, ModifierCrossedOut,
	}
	for _, m := range mods {
		s := ResetStyle().Patch(NewStyle().AddModifier(m))
		if !s.GetAddModifier().Contains(m) {
			t.Errorf("%v: add_modifier should contain the modifier", m)
		}
		if s.GetSubModifier().Contains(m) {
			t.Errorf("%v: sub_modifier should not contain the modifier", m)
		}
	}
}

func TestModifierString(t *testing.T) {
	cases := []struct {
		in   Modifier
		want string
	}{
		{ModifierNone, "NONE"},
		{ModifierBold, "BOLD"},
		{ModifierDim, "DIM"},
		{ModifierItalic, "ITALIC"},
		{ModifierUnderlined, "UNDERLINED"},
		{ModifierSlowBlink, "SLOW_BLINK"},
		{ModifierRapidBlink, "RAPID_BLINK"},
		{ModifierReversed, "REVERSED"},
		{ModifierHidden, "HIDDEN"},
		{ModifierCrossedOut, "CROSSED_OUT"},
		{ModifierBold | ModifierItalic, "BOLD | ITALIC"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("Modifier(%d).String() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestModifierAll(t *testing.T) {
	for _, n := range modifierNames {
		if !ModifierAll.Contains(n.bit) {
			t.Errorf("ModifierAll should contain %s", n.name)
		}
	}
	if got, want := ModifierAll, Modifier(0b1_1111_1111); got != want {
		t.Errorf("ModifierAll = %b, want %b", got, want)
	}
}

func TestStyleReset(t *testing.T) {
	s := ResetStyle()
	if s.GetFg() != ColorReset || s.GetBg() != ColorReset || s.GetUnderlineColor() != ColorReset {
		t.Errorf("ResetStyle should set all three colors to ColorReset, got %+v", s)
	}
	if !s.GetAddModifier().IsEmpty() {
		t.Error("ResetStyle should add no modifiers")
	}
	if s.GetSubModifier() != ModifierAll {
		t.Error("ResetStyle should subtract all modifiers")
	}
}

// TestPatchOverridesAndPreserves is ratatui's documented patch behaviour: a set
// property in the patch wins, an unset one leaves the base alone.
func TestPatchOverridesAndPreserves(t *testing.T) {
	base := NewStyle().Fg(ColorBlue)

	if got, want := base.Patch(NewStyle().Fg(ColorRed)), NewStyle().Fg(ColorRed); got != want {
		t.Errorf("set fg should override: got %+v, want %+v", got, want)
	}
	// An empty patch must not clear anything.
	if got := base.Patch(NewStyle()); got != base {
		t.Errorf("empty patch should preserve: got %+v, want %+v", got, base)
	}
	// fg and bg from different styles combine.
	got := NewStyle().Fg(ColorYellow).Patch(NewStyle().Bg(ColorRed))
	want := NewStyle().Fg(ColorYellow).Bg(ColorRed)
	if got != want {
		t.Errorf("fg and bg should combine: got %+v, want %+v", got, want)
	}
}

// TestAddThenRemoveModifier checks the add/sub bookkeeping: adding then
// removing a modifier must leave it only in sub, not in add.
func TestAddThenRemoveModifier(t *testing.T) {
	s := NewStyle().AddModifier(ModifierBold | ModifierItalic).RemoveModifier(ModifierItalic)
	if !s.GetAddModifier().Contains(ModifierBold) {
		t.Error("BOLD should still be added")
	}
	if s.GetAddModifier().Contains(ModifierItalic) {
		t.Error("ITALIC should no longer be added")
	}
	if !s.GetSubModifier().Contains(ModifierItalic) {
		t.Error("ITALIC should be subtracted")
	}
}

func TestStyleIsEmpty(t *testing.T) {
	if !NewStyle().IsEmpty() {
		t.Error("NewStyle should be empty")
	}
	if NewStyle().Fg(ColorRed).IsEmpty() {
		t.Error("a style with a foreground should not be empty")
	}
	// The deliberate deviation: an explicit reset is *not* the empty style.
	if ResetStyle().IsEmpty() {
		t.Error("ResetStyle should not be empty")
	}
}
