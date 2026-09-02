// Port of the Theme in examples/apps/flex/src/main.rs @ ratatui-v0.30.2

package main

import (
	"os"
	"strconv"

	"github.com/Fiend3d/catatui"
)

// palette holds the colours the demo uses. ratatui takes these from its
// tailwind palette, which catatui does not have; the values are the same
// tailwind shades, named in the comments.
var theme = newTheme()

type themeColors struct {
	minBg         catatui.Color
	maxBg         catatui.Color
	lengthBg      catatui.Color
	percentageBg  catatui.Color
	ratioBg       catatui.Color
	fillBg        catatui.Color
	descriptionFg catatui.Color
	// tab is indexed by tab, so the titles keep their colours in tab order.
	tab []catatui.Color
}

// newTheme picks true colour where the terminal can be trusted with it, and
// the nearest 256-colour indexes where it cannot.
func newTheme() themeColors {
	color := func(trueColor uint32, indexed uint8) catatui.Color {
		if isTrueColorSupported() {
			return catatui.RgbFromU32(trueColor)
		}
		return catatui.Indexed(indexed)
	}

	return themeColors{
		minBg:         color(0x1e3a8a, 24),  // blue 900
		maxBg:         color(0x1e40af, 25),  // blue 800
		lengthBg:      color(0x334155, 67),  // slate 700
		percentageBg:  color(0x1e293b, 18),  // slate 800
		ratioBg:       color(0x0f172a, 17),  // slate 900
		fillBg:        color(0x020617, 16),  // slate 950
		descriptionFg: color(0x94a3b8, 109), // slate 400
		tab: []catatui.Color{
			tabLegacy:       color(0xfb923c, 173), // orange 400
			tabStart:        color(0x38bdf8, 74),  // sky 400
			tabCenter:       color(0x7dd3fc, 116), // sky 300
			tabEnd:          color(0xbae6fd, 152), // sky 200
			tabSpaceAround:  color(0x6366f1, 68),  // indigo 500
			tabSpaceEvenly:  color(0x818cf8, 104), // indigo 400
			tabSpaceBetween: color(0xa5b4fc, 146), // indigo 300
		},
	}
}

// colorForConstraint gives each kind of constraint its own colour, which is
// what makes the illustrations readable at a glance.
func colorForConstraint(constraint catatui.Constraint) catatui.Color {
	switch constraint.Kind() {
	case catatui.ConstraintMin:
		return theme.minBg
	case catatui.ConstraintMax:
		return theme.maxBg
	case catatui.ConstraintLength:
		return theme.lengthBg
	case catatui.ConstraintPercentage:
		return theme.percentageBg
	case catatui.ConstraintRatio:
		return theme.ratioBg
	default:
		return theme.fillBg
	}
}

// isTrueColorSupported reports whether the terminal can show 24-bit colour.
//
// Only one terminal is known not to: Apple's Terminal.app before version 2.15
// (build 465). Everything else is assumed to manage.
func isTrueColorSupported() bool {
	if os.Getenv("TERM_PROGRAM") != "Apple_Terminal" {
		return true
	}
	version, err := strconv.Atoi(os.Getenv("TERM_PROGRAM_VERSION"))
	if err != nil {
		return true
	}
	return version >= 465
}
