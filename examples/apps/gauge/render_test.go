package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the gauges empty, part-filled and full, at sizes from
// nothing to bigger than a screen. Rendering outside the area given panics in
// catatui, so this is what keeps the example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	apps := []*app{
		{},
		{started: true, progressColumns: 40, progress1: 50, progress2: 50.4, progress3: 70.5, progress4: 70.5},
		{started: true, progress1: 100, progress2: 100, progress3: 100, progress4: 100},
	}
	for i, a := range apps {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
			if err != nil {
				t.Fatalf("app %d, %dx%d: %v", i, size[0], size[1], err)
			}
		}
	}
}

// TestNothingMovesUntilStarted checks the gauges sit still until Enter, which
// is what the footer promises.
func TestNothingMovesUntilStarted(t *testing.T) {
	a := &app{}
	for range 10 {
		a.tick(80)
	}
	if a.progressColumns != 0 || a.progress2 != 0 || a.progress3 != 0 {
		t.Errorf("the gauges moved before being started: %+v", a)
	}

	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyEnter})
	a.tick(80)
	if a.progressColumns == 0 {
		t.Errorf("Enter did not start the gauges")
	}
}

// TestGaugesFillAndStop checks every gauge reaches its top and stays there, and
// that none of them runs past it — a ratio over 1 panics.
func TestGaugesFillAndStop(t *testing.T) {
	a := &app{started: true}
	const width = 40
	for range width*2 + 1200 {
		a.tick(width)
	}
	if a.progressColumns != width {
		t.Errorf("filled %d of %d columns", a.progressColumns, width)
	}
	if a.progress1 != 100 {
		t.Errorf("the percentage gauge stopped at %d", a.progress1)
	}
	for i, progress := range []float64{a.progress2, a.progress3, a.progress4} {
		if progress < 0 || progress > 100 {
			t.Errorf("gauge %d is at %v, outside 0 to 100", i+2, progress)
		}
	}
	if a.progress3 != 100 {
		t.Errorf("the unicode gauges stopped at %v", a.progress3)
	}
}

// TestTickIgnoresAZeroWidthTerminal checks the width is never divided by zero,
// which a window dragged shut would otherwise do.
func TestTickIgnoresAZeroWidthTerminal(t *testing.T) {
	a := &app{started: true}
	a.tick(0)
	if a.progress1 != 0 {
		t.Errorf("a zero-width terminal moved the gauge to %d", a.progress1)
	}
}
