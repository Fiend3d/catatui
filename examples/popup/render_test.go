package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example with and without the popup, at sizes from
// nothing to bigger than a screen. Rendering outside the area given panics in
// catatui, so this is what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, showPopup := range []bool{false, true} {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("popup=%v, %dx%d: %v", showPopup, size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { render(f, showPopup) })
			if err != nil {
				t.Fatalf("popup=%v, %dx%d: %v", showPopup, size[0], size[1], err)
			}
		}
	}
}

// TestCenteredStaysInside checks the popup area never leaves the area it is
// centred in, which is what stops it drawing out of bounds on a small screen.
func TestCenteredStaysInside(t *testing.T) {
	for width := uint16(0); width < 40; width++ {
		for height := uint16(0); height < 20; height++ {
			area := catatui.NewRect(3, 2, width, height)
			got := centered(area, catatui.Percentage(60), catatui.Percentage(20))
			if got.Intersection(area) != got {
				t.Errorf("%v centred in %v escapes it", got, area)
			}
		}
	}
}
