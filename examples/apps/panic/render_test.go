package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example both ways round, at sizes from nothing to
// bigger than a screen. Rendering outside the area given panics in catatui, so
// this is what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, recovery := range []bool{true, false} {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("recovery=%v, %dx%d: %v", recovery, size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { render(f, recovery) })
			if err != nil {
				t.Fatalf("recovery=%v, %dx%d: %v", recovery, size[0], size[1], err)
			}
		}
	}
}
