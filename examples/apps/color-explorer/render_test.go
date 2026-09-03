package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {120, 49}, {200, 60}}
	for _, size := range sizes {
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		if err := terminal.Draw(render); err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
	}
}

// TestGridsHoldEveryColor checks each grid has exactly as many cells as there
// are colours to put in it, at any size. The loops index the cells directly, so
// a grid one short would be an out-of-range panic on some terminal or other.
func TestGridsHoldEveryColor(t *testing.T) {
	for _, size := range [][2]uint16{{0, 0}, {1, 1}, {40, 12}, {120, 49}, {200, 60}} {
		area := catatui.NewRect(0, 0, size[0], size[1])

		if got := len(namedColorCells(area)); got != len(namedColors) {
			t.Errorf("%dx%d: %d cells for %d named colours", size[0], size[1], got, len(namedColors))
		}
		if got := len(cubeCells(area, area)); got != 216 {
			t.Errorf("%dx%d: %d cells for the 216 colours of the cube", size[0], size[1], got)
		}
	}
}

// TestNamedColorsAreNamed checks every colour in the table prints as a name
// rather than a number, which is what the labels rely on.
func TestNamedColorsAreNamed(t *testing.T) {
	for i, color := range namedColors {
		name := color.String()
		if name == "" || name == "Unset" {
			t.Errorf("named colour %d prints as %q", i, name)
		}
	}
	if got, want := catatui.ColorReset.String(), "Reset"; got != want {
		t.Errorf("the reset colour prints as %q, want %q", got, want)
	}
}
