package main

import (
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, size := range sizes {
		a := newApp()
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
	}
}

// TestTheGreenSquareRecordsWhereItDrew checks the pointer receiver does what
// the example is about: the position is worked out during Render and readable
// afterwards, and it is the right-hand end of the area.
func TestTheGreenSquareRecordsWhereItDrew(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 2))
	square := &rightAlignedSquare{}
	square.Render(catatui.NewRect(4, 1, 12, 1), buf)

	if got, want := square.lastPosition, (catatui.Position{X: 12, Y: 1}); got != want {
		t.Errorf("the square recorded %+v, want %+v", got, want)
	}
	if got := buf.Get(12, 1); got.Fg != catatui.ColorGreen {
		t.Errorf("nothing green was drawn at the recorded position: %+v", got)
	}
	if got := buf.Get(11, 1); got.Fg == catatui.ColorGreen {
		t.Errorf("the square drew a column to the left of where it says it is")
	}
}

// TestTheGreenSquareShrinksToFit checks a square wider than the area is cut to
// it rather than drawn past the left edge, which SatSub would otherwise wrap
// around to the far side of the screen.
func TestTheGreenSquareShrinksToFit(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 1))
	square := &rightAlignedSquare{}
	square.Render(catatui.NewRect(6, 0, 2, 1), buf)

	if got, want := square.lastPosition, (catatui.Position{X: 6, Y: 0}); got != want {
		t.Errorf("the square recorded %+v, want %+v", got, want)
	}
	if got := buf.Get(5, 0); got.Fg == catatui.ColorGreen {
		t.Errorf("the square drew outside the area it was given")
	}
}

// TestTheContainerDrawsEveryChild checks each widget in the container gets its
// own four columns, in order.
func TestTheContainerDrawsEveryChild(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 12, 1))
	squares{children: []catatui.Widget{redSquare{}, blueSquare{}}}.Render(buf.Area, buf)

	for x, want := range map[uint16]catatui.Color{
		0: catatui.ColorRed, 3: catatui.ColorRed,
		4: catatui.ColorBlue, 7: catatui.ColorBlue,
	} {
		if got := buf.Get(x, 0).Fg; got != want {
			t.Errorf("column %d is %v, want %v", x, got, want)
		}
	}
	if got := buf.Get(8, 0).GetSymbol(); got != " " {
		t.Errorf("the container drew %q past its two squares", got)
	}
}

// TestTheTimerCountsUp checks the timer reads its own state rather than a
// number fixed when it was built.
func TestTheTimerCountsUp(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 1))
	timer{start: time.Now().Add(-90 * time.Second)}.Render(buf.Area, buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Elapsed: 90.0s      "))
}
