// Port of the Theme in examples/apps/flex/src/main.rs @ ratatui-v0.30.2

package main

import (
	"os"
	"strconv"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
)

// theme holds the colours the demo uses, taken from the tailwind palette as
// ratatui's does.
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
	// The palette is 24-bit; where that is not available the nearest indexed
	// colour stands in.
	color := func(trueColor catatui.Color, indexed uint8) catatui.Color {
		if isTrueColorSupported() {
			return trueColor
		}
		return catatui.Indexed(indexed)
	}

	return themeColors{
		minBg:         color(tailwind.Blue.C900, 24),
		maxBg:         color(tailwind.Blue.C800, 25),
		lengthBg:      color(tailwind.Slate.C700, 67),
		percentageBg:  color(tailwind.Slate.C800, 18),
		ratioBg:       color(tailwind.Slate.C900, 17),
		fillBg:        color(tailwind.Slate.C950, 16),
		descriptionFg: color(tailwind.Slate.C400, 109),
		tab: []catatui.Color{
			tabLegacy:       color(tailwind.Orange.C400, 173),
			tabStart:        color(tailwind.Sky.C400, 74),
			tabCenter:       color(tailwind.Sky.C300, 116),
			tabEnd:          color(tailwind.Sky.C200, 152),
			tabSpaceAround:  color(tailwind.Indigo.C500, 68),
			tabSpaceEvenly:  color(tailwind.Indigo.C400, 104),
			tabSpaceBetween: color(tailwind.Indigo.C300, 146),
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
