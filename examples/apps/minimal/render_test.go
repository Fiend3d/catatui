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

// TestGreetingIsDrawn checks the one thing the program does, and that it starts
// in the top-left corner of the area it is given.
func TestGreetingIsDrawn(t *testing.T) {
	backend := catatui.NewTestBackend(20, 3)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminal.Draw(render); err != nil {
		t.Fatal(err)
	}
	catatui.AssertBuffer(t, backend.Buffer(), catatui.NewBufferWithStrings(
		"Hello World!        ",
		"                    ",
		"                    ",
	))
}
