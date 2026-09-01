// Port of ratatui-core/src/style/color.rs @ ratatui-v0.30.2
//
// Deviation from ratatui: Rust models "no color set" as Option<Color> inside
// Style, and Color::default() is Color::Reset. Go has no Option, and making
// Style hold *Color would cost an allocation per style and break comparability.
// Instead the Color zero value is a distinct "unset" state, and ColorReset is
// an explicit variant. So:
//
//	var c Color        // unset - a Style with this fg inherits the fg beneath it
//	c := ColorReset    // explicit reset - clears the terminal's color
//
// This is the only place the two concepts differ from Rust; everywhere else
// unset behaves exactly like Rust's None and ColorReset like Some(Reset).

package catatui

import (
	"fmt"
	"strconv"
	"strings"
)

type colorKind uint8

const (
	colorUnset colorKind = iota // zero value: no color specified
	colorReset
	colorNamed
	colorIndexed
	colorRgb
)

// Color is an ANSI color: a named 4-bit color, an 8-bit indexed color, a 24-bit
// RGB color, an explicit reset, or unset.
//
// Color is comparable, so colors and the Styles containing them can be compared
// with == and used as map keys.
type Color struct {
	kind colorKind
	a    uint8 // named ordinal, or indexed value, or red
	g    uint8
	b    uint8
}

// The named ANSI colors. The ordinals match ratatui's enum order, which is also
// the order the ANSI SGR codes fall in: 30..37 then 90..97.
var (
	ColorReset = Color{kind: colorReset}

	ColorBlack        = Color{kind: colorNamed, a: 0}
	ColorRed          = Color{kind: colorNamed, a: 1}
	ColorGreen        = Color{kind: colorNamed, a: 2}
	ColorYellow       = Color{kind: colorNamed, a: 3}
	ColorBlue         = Color{kind: colorNamed, a: 4}
	ColorMagenta      = Color{kind: colorNamed, a: 5}
	ColorCyan         = Color{kind: colorNamed, a: 6}
	ColorGray         = Color{kind: colorNamed, a: 7}
	ColorDarkGray     = Color{kind: colorNamed, a: 8}
	ColorLightRed     = Color{kind: colorNamed, a: 9}
	ColorLightGreen   = Color{kind: colorNamed, a: 10}
	ColorLightYellow  = Color{kind: colorNamed, a: 11}
	ColorLightBlue    = Color{kind: colorNamed, a: 12}
	ColorLightMagenta = Color{kind: colorNamed, a: 13}
	ColorLightCyan    = Color{kind: colorNamed, a: 14}
	ColorWhite        = Color{kind: colorNamed, a: 15}
)

// Rgb returns a 24-bit true color. Only terminals with true color support will
// display it correctly.
func Rgb(r, g, b uint8) Color { return Color{kind: colorRgb, a: r, g: g, b: b} }

// Indexed returns an 8-bit color from the 256-color palette.
func Indexed(i uint8) Color { return Color{kind: colorIndexed, a: i} }

// RgbFromU32 returns an RGB color from a 0x00RRGGBB value.
func RgbFromU32(u uint32) Color {
	return Rgb(uint8(u>>16), uint8(u>>8), uint8(u))
}

// IsSet reports whether the color is anything other than the zero value.
// An unset color leaves whatever color is already in the cell untouched.
func (c Color) IsSet() bool { return c.kind != colorUnset }

// IsReset reports whether the color is an explicit ColorReset.
func (c Color) IsReset() bool { return c.kind == colorReset }

// Named reports whether c is one of the sixteen named ANSI colors, returning its
// ordinal (0..15) if so.
func (c Color) Named() (ordinal uint8, ok bool) { return c.a, c.kind == colorNamed }

// Index reports whether c is an 8-bit indexed color, returning its index if so.
func (c Color) Index() (index uint8, ok bool) { return c.a, c.kind == colorIndexed }

// RGB reports whether c is a 24-bit color, returning its components if so.
func (c Color) RGB() (r, g, b uint8, ok bool) { return c.a, c.g, c.b, c.kind == colorRgb }

// String formats the color the way ratatui's Display impl does: named colors by
// name, RGB as "#RRGGBB", indexed as a decimal number. The result round-trips
// through ParseColor.
func (c Color) String() string {
	switch c.kind {
	case colorUnset:
		return "Unset"
	case colorReset:
		return "Reset"
	case colorNamed:
		return colorNames[c.a]
	case colorIndexed:
		return strconv.Itoa(int(c.a))
	case colorRgb:
		return fmt.Sprintf("#%02X%02X%02X", c.a, c.g, c.b)
	}
	return "Unset"
}

var colorNames = [16]string{
	"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "Gray",
	"DarkGray", "LightRed", "LightGreen", "LightYellow", "LightBlue",
	"LightMagenta", "LightCyan", "White",
}

var namedByString = map[string]Color{
	"reset": ColorReset, "black": ColorBlack, "red": ColorRed,
	"green": ColorGreen, "yellow": ColorYellow, "blue": ColorBlue,
	"magenta": ColorMagenta, "cyan": ColorCyan, "gray": ColorGray,
	"darkgray": ColorDarkGray, "lightred": ColorLightRed,
	"lightgreen": ColorLightGreen, "lightyellow": ColorLightYellow,
	"lightblue": ColorLightBlue, "lightmagenta": ColorLightMagenta,
	"lightcyan": ColorLightCyan, "white": ColorWhite,
}

// colorNormalizer mirrors ratatui's chain of replacements in FromStr. The order
// is significant: "bright"->"light" must run before the "light*" rules, and
// "grey"/"silver"->"gray" before "lightgray"->"white".
var colorNormalizer = strings.NewReplacer(
	" ", "", "-", "", "_", "",
)

// ParseColor parses a color name, a decimal 8-bit index, or a "#RRGGBB" hex
// string. It accepts the same spellings ratatui does: "bright" and "light"
// prefixes, "grey" and "gray", "silver", and space, dash or underscore
// separators.
func ParseColor(s string) (Color, error) {
	n := colorNormalizer.Replace(strings.ToLower(s))
	n = strings.ReplaceAll(n, "bright", "light")
	n = strings.ReplaceAll(n, "grey", "gray")
	n = strings.ReplaceAll(n, "silver", "gray")
	n = strings.ReplaceAll(n, "lightblack", "darkgray")
	n = strings.ReplaceAll(n, "lightwhite", "white")
	n = strings.ReplaceAll(n, "lightgray", "white")

	if c, ok := namedByString[n]; ok {
		return c, nil
	}
	// Parsing of index and hex uses the original string, as ratatui does.
	if i, err := strconv.ParseUint(s, 10, 8); err == nil {
		return Indexed(uint8(i)), nil
	}
	if c, ok := parseHexColor(s); ok {
		return c, nil
	}
	return Color{}, fmt.Errorf("catatui: failed to parse color %q", s)
}

func parseHexColor(s string) (Color, bool) {
	if len(s) != 7 || s[0] != '#' {
		return Color{}, false
	}
	// Parsed as three independent byte pairs, exactly as Rust's
	// u8::from_str_radix chunking does, so odd inputs are accepted or
	// rejected identically.
	var rgb [3]uint8
	for i := range rgb {
		v, err := strconv.ParseUint(s[1+i*2:3+i*2], 16, 8)
		if err != nil {
			return Color{}, false
		}
		rgb[i] = uint8(v)
	}
	return Rgb(rgb[0], rgb[1], rgb[2]), true
}
