package tailwind_test

import (
	"reflect"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
)

// palettes is every ramp in the package, for the checks that have to be
// exhaustive. It is written out by hand rather than generated so that a palette
// dropped by the generator fails a test instead of quietly vanishing.
var palettes = map[string]tailwind.Palette{
	"Slate": tailwind.Slate, "Gray": tailwind.Gray, "Zinc": tailwind.Zinc,
	"Neutral": tailwind.Neutral, "Stone": tailwind.Stone, "Red": tailwind.Red,
	"Orange": tailwind.Orange, "Amber": tailwind.Amber, "Yellow": tailwind.Yellow,
	"Lime": tailwind.Lime, "Green": tailwind.Green, "Emerald": tailwind.Emerald,
	"Teal": tailwind.Teal, "Cyan": tailwind.Cyan, "Sky": tailwind.Sky,
	"Blue": tailwind.Blue, "Indigo": tailwind.Indigo, "Violet": tailwind.Violet,
	"Purple": tailwind.Purple, "Fuchsia": tailwind.Fuchsia, "Pink": tailwind.Pink,
	"Rose": tailwind.Rose,
}

// TestKnownValues checks the shades ratatui's own documentation asserts, which
// is what pins these values to Tailwind's rather than to whatever the generator
// happened to read.
func TestKnownValues(t *testing.T) {
	cases := []struct {
		name    string
		color   catatui.Color
		r, g, b uint8
	}{
		{"Red.C500", tailwind.Red.C500, 239, 68, 68},
		{"Blue.C500", tailwind.Blue.C500, 59, 130, 246},
		{"Slate.C50", tailwind.Slate.C50, 0xf8, 0xfa, 0xfc},
		{"Slate.C950", tailwind.Slate.C950, 0x02, 0x06, 0x17},
		{"Black", tailwind.Black, 0, 0, 0},
		{"White", tailwind.White, 255, 255, 255},
	}
	for _, c := range cases {
		r, g, b, ok := c.color.RGB()
		if !ok || r != c.r || g != c.g || b != c.b {
			t.Errorf("%s = (%d, %d, %d, %v), want (%d, %d, %d, true)",
				c.name, r, g, b, ok, c.r, c.g, c.b)
		}
	}
}

// TestEveryShadeIsSet walks each palette by reflection, so a shade the
// generator failed to write shows up as the zero Color rather than as a colour
// that silently inherits whatever is underneath it.
func TestEveryShadeIsSet(t *testing.T) {
	for name, palette := range palettes {
		value := reflect.ValueOf(palette)
		for i := range value.NumField() {
			shade := value.Field(i).Interface().(catatui.Color)
			if _, _, _, ok := shade.RGB(); !ok {
				t.Errorf("%s.%s is %v, want an RGB colour",
					name, value.Type().Field(i).Name, shade)
			}
		}
	}
}

// TestRampsRunLightToDark checks each ramp gets darker as the numbers get
// bigger, which is the one invariant that would catch shades written out of
// order.
func TestRampsRunLightToDark(t *testing.T) {
	for name, palette := range palettes {
		value := reflect.ValueOf(palette)
		previous := 256.0
		for i := range value.NumField() {
			field := value.Type().Field(i).Name
			shade := value.Field(i).Interface().(catatui.Color)
			l := luminance(t, name+"."+field, shade)
			if l > previous {
				t.Errorf("%s.%s is lighter than the shade before it (%.1f > %.1f)",
					name, field, l, previous)
			}
			previous = l
		}
	}
}

// TestPalettesAreDistinct checks no two ramps are the same, which would mean
// one name had been generated from another's values.
func TestPalettesAreDistinct(t *testing.T) {
	seen := make(map[tailwind.Palette]string, len(palettes))
	for name, palette := range palettes {
		if other, ok := seen[palette]; ok {
			t.Errorf("%s and %s are the same palette", name, other)
		}
		seen[palette] = name
	}
	if len(palettes) != 22 {
		t.Errorf("checked %d palettes, want tailwind's 22", len(palettes))
	}
}

// luminance is the perceived brightness of a colour, 0 to 255.
func luminance(t *testing.T, name string, c catatui.Color) float64 {
	t.Helper()
	r, g, b, ok := c.RGB()
	if !ok {
		t.Fatalf("%s is %v, want an RGB colour", name, c)
	}
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}
