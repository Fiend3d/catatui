package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws an empty canvas and one with a stroke on it, at sizes from
// nothing to bigger than a screen. Points can be reported outside the window,
// so this is also what checks they are kept inside it: rendering outside the
// area given panics in catatui.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}

	drawn := &app{currentColor: catatui.ColorWhite, hasMouse: true,
		mousePosition: catatui.Position{X: 150, Y: 40}}
	drawn.handle(mouse(term.MouseDown, 2, 2))
	drawn.handle(mouse(term.MouseDrag, 30, 18))
	// Well outside anything but the largest of the sizes above.
	drawn.handle(mouse(term.MouseDrag, 150, 40))

	for i, a := range []*app{{currentColor: catatui.ColorWhite}, drawn} {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
			if err := terminal.Draw(a.render); err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
		}
	}
}

func mouse(kind term.MouseKind, x, y uint16) term.Event {
	return term.Event{Kind: term.EventMouse, MouseKind: kind, Button: term.MouseButtonLeft, X: x, Y: y}
}

// TestBresenhamJoinsTheEnds checks the line runs from one end to the other with
// no gaps, which is the whole reason it is here: a drag reports where the
// pointer landed, not the cells it crossed.
func TestBresenhamJoinsTheEnds(t *testing.T) {
	cases := []struct{ x0, y0, x1, y1 uint16 }{
		{0, 0, 0, 0},   // a single cell
		{0, 0, 10, 0},  // horizontal
		{0, 0, 0, 10},  // vertical
		{0, 0, 10, 10}, // exactly diagonal
		{0, 0, 10, 3},  // shallow
		{0, 0, 3, 10},  // steep
		{10, 10, 0, 0}, // backwards
		{10, 0, 0, 7},  // backwards and down
	}
	for _, tc := range cases {
		start := catatui.Position{X: tc.x0, Y: tc.y0}
		end := catatui.Position{X: tc.x1, Y: tc.y1}
		line := bresenham(start, end)

		if line[0] != start {
			t.Errorf("%v to %v starts at %v", start, end, line[0])
		}
		if last := line[len(line)-1]; last != end {
			t.Errorf("%v to %v ends at %v", start, end, last)
		}
		for i := 1; i < len(line); i++ {
			dx := abs(int(line[i].X) - int(line[i-1].X))
			dy := abs(int(line[i].Y) - int(line[i-1].Y))
			if dx > 1 || dy > 1 || (dx == 0 && dy == 0) {
				t.Errorf("%v to %v jumps from %v to %v", start, end, line[i-1], line[i])
			}
		}
	}
}

// TestDragWithoutAPointDrawsNothing checks a drag before any click is ignored
// rather than reading off the end of the empty list of points.
func TestDragWithoutAPointDrawsNothing(t *testing.T) {
	a := &app{currentColor: catatui.ColorWhite}
	a.handle(mouse(term.MouseDrag, 5, 5))
	if len(a.points) != 0 {
		t.Errorf("a drag with nothing drawn yet left %d points", len(a.points))
	}
	if !a.hasMouse || a.mousePosition != (catatui.Position{X: 5, Y: 5}) {
		t.Errorf("the pointer was not followed: %+v", a.mousePosition)
	}
}

// TestSpaceChangesTheColor checks space picks a new colour to draw with.
func TestSpaceChangesTheColor(t *testing.T) {
	a := &app{currentColor: catatui.ColorWhite}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: ' '})
	if _, _, _, ok := a.currentColor.RGB(); !ok {
		t.Errorf("space left the colour as %v, want an RGB one", a.currentColor)
	}
}
