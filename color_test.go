// Tests ported from ratatui-core/src/style/color.rs @ ratatui-v0.30.2

package catatui

import "testing"

func TestColorFromU32(t *testing.T) {
	cases := []struct {
		in   uint32
		want Color
	}{
		{0x000000, Rgb(0, 0, 0)},
		{0xFF0000, Rgb(255, 0, 0)},
		{0x00FF00, Rgb(0, 255, 0)},
		{0x0000FF, Rgb(0, 0, 255)},
		{0xFFFFFF, Rgb(255, 255, 255)},
	}
	for _, c := range cases {
		if got := RgbFromU32(c.in); got != c.want {
			t.Errorf("RgbFromU32(%#06x) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseColor(t *testing.T) {
	cases := []struct {
		in   string
		want Color
	}{
		// rgb and indexed
		{"#FF0000", Rgb(255, 0, 0)},
		{"10", Indexed(10)},

		// named
		{"reset", ColorReset},
		{"black", ColorBlack},
		{"red", ColorRed},
		{"green", ColorGreen},
		{"yellow", ColorYellow},
		{"blue", ColorBlue},
		{"magenta", ColorMagenta},
		{"cyan", ColorCyan},
		{"gray", ColorGray},
		{"darkgray", ColorDarkGray},
		{"lightred", ColorLightRed},
		{"lightgreen", ColorLightGreen},
		{"lightyellow", ColorLightYellow},
		{"lightblue", ColorLightBlue},
		{"lightmagenta", ColorLightMagenta},
		{"lightcyan", ColorLightCyan},
		{"white", ColorWhite},

		// aliases
		{"lightblack", ColorDarkGray},
		{"lightwhite", ColorWhite},
		{"lightgray", ColorWhite},

		// silver = grey = gray
		{"grey", ColorGray},
		{"silver", ColorGray},

		// spaces are ignored
		{"light black", ColorDarkGray},
		{"light white", ColorWhite},
		{"light gray", ColorWhite},

		// dashes are ignored
		{"light-black", ColorDarkGray},
		{"light-white", ColorWhite},
		{"light-gray", ColorWhite},

		// underscores are ignored
		{"light_black", ColorDarkGray},
		{"light_white", ColorWhite},
		{"light_gray", ColorWhite},

		// bright = light
		{"bright-black", ColorDarkGray},
		{"bright-white", ColorWhite},
		{"brightblack", ColorDarkGray},
		{"brightwhite", ColorWhite},

		// case is ignored
		{"LightGreen", ColorLightGreen},
		{"Magenta", ColorMagenta},
	}
	for _, c := range cases {
		got, err := ParseColor(c.in)
		if err != nil {
			t.Errorf("ParseColor(%q) returned error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseColor(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseColorInvalid(t *testing.T) {
	bad := []string{
		"invalid_color",  // not a color string
		"abcdef0",        // 7 chars is not a color
		" bcdefa",        // doesn't start with a '#'
		"#abcdef00",      // too many chars
		"#1\U0001F980 2", // len 7 but not on char boundaries; must not panic
		"resets",         // typo
		"lightblackk",    // typo
	}
	for _, s := range bad {
		if got, err := ParseColor(s); err == nil {
			t.Errorf("ParseColor(%q) = %v, want error", s, got)
		}
	}
}

func TestColorString(t *testing.T) {
	cases := []struct {
		in   Color
		want string
	}{
		{ColorBlack, "Black"},
		{ColorRed, "Red"},
		{ColorGreen, "Green"},
		{ColorYellow, "Yellow"},
		{ColorBlue, "Blue"},
		{ColorMagenta, "Magenta"},
		{ColorCyan, "Cyan"},
		{ColorGray, "Gray"},
		{ColorDarkGray, "DarkGray"},
		{ColorLightRed, "LightRed"},
		{ColorLightGreen, "LightGreen"},
		{ColorLightYellow, "LightYellow"},
		{ColorLightBlue, "LightBlue"},
		{ColorLightMagenta, "LightMagenta"},
		{ColorLightCyan, "LightCyan"},
		{ColorWhite, "White"},
		{Indexed(10), "10"},
		{Rgb(255, 0, 0), "#FF0000"},
		{ColorReset, "Reset"},
		{Color{}, "Unset"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("Color.String() = %q, want %q", got, c.want)
		}
	}
}

// TestColorRoundTrip checks that every value String() produces parses back to
// the same color, which is what makes Color usable in config files.
func TestColorRoundTrip(t *testing.T) {
	all := []Color{
		ColorReset, ColorBlack, ColorRed, ColorGreen, ColorYellow, ColorBlue,
		ColorMagenta, ColorCyan, ColorGray, ColorDarkGray, ColorLightRed,
		ColorLightGreen, ColorLightYellow, ColorLightBlue, ColorLightMagenta,
		ColorLightCyan, ColorWhite, Indexed(0), Indexed(42), Indexed(255),
		Rgb(0, 0, 0), Rgb(1, 2, 3), Rgb(255, 255, 255),
	}
	for _, c := range all {
		got, err := ParseColor(c.String())
		if err != nil {
			t.Errorf("ParseColor(%q) returned error %v", c.String(), err)
			continue
		}
		if got != c {
			t.Errorf("round trip of %v via %q gave %v", c, c.String(), got)
		}
	}
}

// TestColorUnsetIsDistinctFromReset guards the one deliberate deviation from
// ratatui: the zero value means "no color specified", not Reset.
func TestColorUnsetIsDistinctFromReset(t *testing.T) {
	var zero Color
	if zero == ColorReset {
		t.Fatal("zero Color must not equal ColorReset")
	}
	if zero.IsSet() {
		t.Error("zero Color must report IsSet() == false")
	}
	if !ColorReset.IsSet() {
		t.Error("ColorReset must report IsSet() == true")
	}
	if !ColorReset.IsReset() {
		t.Error("ColorReset must report IsReset() == true")
	}
	if zero.IsReset() {
		t.Error("zero Color must report IsReset() == false")
	}
}
