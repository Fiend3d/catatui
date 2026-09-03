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

// TestBothChildrenAreDrawn checks the container splits the area between the two
// widgets and draws each in its own half.
func TestBothChildrenAreDrawn(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 11, 6))
	stackContainer{
		direction: catatui.Vertical,
		children: []child{
			{greeting{}, catatui.Percentage(50)},
			{farewell{}, catatui.Percentage(50)},
		},
	}.Render(buf.Area, buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌─────────┐",
		"│Hello    │",
		"└─────────┘",
		"┌─────────┐",
		"│Goodbye  │",
		"└─────────┘",
	))
}

// TestAnEmptyContainerDrawsNothing checks the layout is not asked to split an
// area between no constraints, which is what an empty children slice would do.
func TestAnEmptyContainerDrawsNothing(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 6, 2))
	stackContainer{direction: catatui.Horizontal}.Render(buf.Area, buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"      ",
		"      ",
	))
}
