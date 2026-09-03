package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws every marker at sizes from nothing to bigger than a screen,
// with points drawn on the scratchpad from well outside it. Rendering outside
// the area given panics in catatui, so this is what keeps the example honest
// when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, marker := range markerCycle {
		a := newApp()
		a.marker = marker
		a.x, a.y = 30, 45
		a.points = []catatui.Position{{X: 0, Y: 0}, {X: 5, Y: 5}, {X: 400, Y: 300}}
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("marker %v, %dx%d: %v", marker, size[0], size[1], err)
			}
			if err := terminal.Draw(a.render); err != nil {
				t.Fatalf("marker %v, %dx%d: %v", marker, size[0], size[1], err)
			}
		}
	}
}

// TestMarkerCycleReturnsToTheStart checks Enter reaches every marker and comes
// back round, and that the cycle holds each of them once.
func TestMarkerCycleReturnsToTheStart(t *testing.T) {
	a := newApp()
	seen := map[symbols.Marker]bool{a.marker: true}
	for range len(markerCycle) {
		a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
		seen[a.marker] = true
	}
	if len(seen) != len(markerCycle) {
		t.Errorf("Enter reached %d of the %d markers", len(seen), len(markerCycle))
	}
	if a.marker != markerCycle[0] {
		t.Errorf("a full cycle ended on %v, want it back at %v", a.marker, markerCycle[0])
	}

	// A marker that is not in the cycle at all starts it again from the front,
	// rather than leaving Enter doing nothing.
	a.marker = symbols.Custom('?')
	a.cycleMarker()
	if a.marker != markerCycle[0] {
		t.Errorf("an unknown marker cycled to %v, want %v", a.marker, markerCycle[0])
	}
}

// TestBallStaysInThePlayground checks the ball bounces rather than escaping,
// however long it is left running.
func TestBallStaysInThePlayground(t *testing.T) {
	a := newApp()
	left, right := float64(a.playground.Left()), float64(a.playground.Right())
	top, bottom := float64(a.playground.Top()), float64(a.playground.Bottom())

	for range 10000 {
		a.onTick()
		// The bounce happens on the tick after the edge is crossed, so the
		// ball may be one step over it; anything more has escaped.
		if a.ball.X-a.ball.Radius < left-1 || a.ball.X+a.ball.Radius > right+1 {
			t.Fatalf("the ball left the playground sideways at x=%v", a.ball.X)
		}
		if a.ball.Y-a.ball.Radius < top-1 || a.ball.Y+a.ball.Radius > bottom+1 {
			t.Fatalf("the ball left the playground vertically at y=%v", a.ball.Y)
		}
	}
}

// TestDraggingCollectsPoints checks a drag records where the pointer went and
// a press on its own does not.
func TestDraggingCollectsPoints(t *testing.T) {
	a := newApp()
	a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseDown, X: 1, Y: 1})
	if len(a.points) != 0 {
		t.Errorf("a press recorded %d points, want none until the drag", len(a.points))
	}
	if !a.isDrawing {
		t.Errorf("a press did not start drawing")
	}

	a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseDrag, X: 3, Y: 4})
	if len(a.points) != 1 || a.points[0] != (catatui.Position{X: 3, Y: 4}) {
		t.Errorf("the drag recorded %v", a.points)
	}

	a.handle(term.Event{Kind: term.EventMouse, MouseKind: term.MouseUp, X: 3, Y: 4})
	if a.isDrawing {
		t.Errorf("letting go did not stop drawing")
	}
}

// labelCell renders the app and reports the column and row the "You are here"
// label starts at, or -1, -1 if it is not on screen.
func labelCell(t *testing.T, a *app) (int, int) {
	t.Helper()
	backend := catatui.NewTestBackend(80, 24)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminal.Draw(a.render); err != nil {
		t.Fatal(err)
	}
	buf := backend.Buffer()
	for y := uint16(0); y+1 < 24; y++ {
		for x := uint16(0); x+2 < 80; x++ {
			if buf.Get(x, y).GetSymbol() == "Y" && buf.Get(x+1, y).GetSymbol() == "o" &&
				buf.Get(x+2, y).GetSymbol() == "u" {
				return int(x), int(y)
			}
		}
	}
	return -1, -1
}

// TestMovingTheLabelMovesItACell checks a movement key moves the label one cell
// of the map, in both directions.
//
// This is the one place the example departs from ratatui's, which moves by a
// flat 1.0 — one degree. The map is 360 degrees across and 180 down, so on an
// 80x24 terminal that is a tenth of a cell sideways and a twenty-second of one
// down: ten presses to shift a column and twenty-three to shift a row, which
// reads as the keys doing nothing.
func TestMovingTheLabelMovesItACell(t *testing.T) {
	for _, tc := range []struct {
		key            rune
		wantDX, wantDY int
	}{
		{'l', 1, 0},
		{'h', -1, 0},
		{'j', 0, 1},
		{'k', 0, -1},
	} {
		a := newApp()
		// The loop draws before it waits for a key, so the pane size the step
		// is scaled to is known by the time one arrives.
		x0, y0 := labelCell(t, a)
		if x0 < 0 {
			t.Fatal("the label was not drawn to begin with")
		}

		for presses := 1; presses <= 3; presses++ {
			a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: tc.key})
			x, y := labelCell(t, a)
			wantX, wantY := x0+tc.wantDX*presses, y0+tc.wantDY*presses
			if x != wantX || y != wantY {
				t.Errorf("%d presses of %q put the label at %d,%d, want %d,%d",
					presses, tc.key, x, y, wantX, wantY)
			}
		}
	}
}

// TestTheLabelStaysOnTheMap checks holding a key down stops the label at the
// edge rather than walking it off, since a canvas draws no label outside its
// bounds and a cell-sized step reaches the edge in a few presses.
func TestTheLabelStaysOnTheMap(t *testing.T) {
	for _, key := range []rune{'h', 'j', 'k', 'l'} {
		a := newApp()
		labelCell(t, a)
		for range 200 {
			a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: key})
		}
		if a.x < -180 || a.x > 180 {
			t.Errorf("%q took the label to x=%v, off the map", key, a.x)
		}
		if a.y < -90 || a.y > 90 {
			t.Errorf("%q took the label to y=%v, off the map", key, a.y)
		}
	}
}

// TestMovingBeforeTheFirstFrameDoesNotDivideByZero checks a key that somehow
// arrives before anything has been drawn still moves the label, rather than
// scaling it by a pane of no size.
func TestMovingBeforeTheFirstFrameDoesNotDivideByZero(t *testing.T) {
	a := newApp()
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'l'})
	if a.x != 1.0 {
		t.Errorf("moving with no pane yet gave x=%v, want the fallback step of 1", a.x)
	}
}
