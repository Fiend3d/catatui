package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
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

// TestGridCoversEveryCombination checks the grid has room for each background,
// foreground and modifier exactly once, which is what the index arithmetic in
// render assumes.
func TestGridCoversEveryCombination(t *testing.T) {
	cells := gridCells(catatui.NewRect(0, 0, 80, 60))
	if got, want := len(cells), len(colors)*len(colors)*len(allModifiers); got != want {
		t.Fatalf("grid has %d cells, want %d", got, want)
	}

	seen := make(map[[3]int]bool, len(cells))
	for i := range cells {
		bg := i / (len(colors) * len(allModifiers))
		fg := i / len(allModifiers) % len(colors)
		modifier := i % len(allModifiers)
		key := [3]int{bg, fg, modifier}
		if seen[key] {
			t.Fatalf("cell %d repeats the combination %v", i, key)
		}
		seen[key] = true
	}
}
