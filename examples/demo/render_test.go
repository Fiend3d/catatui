package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws every tab at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes underneath it.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, tab := range []int{0, 1, 2} {
		for _, unicode := range []bool{true, false} {
			for _, size := range sizes {
				a := newApp("Catatui Demo", unicode)
				a.tabs.index = tab
				a.onTick()

				terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
				if err != nil {
					t.Fatalf("tab %d, %dx%d: %v", tab, size[0], size[1], err)
				}
				err = terminal.Draw(func(f *catatui.Frame) { render(f, a) })
				if err != nil {
					t.Fatalf("tab %d, %dx%d: %v", tab, size[0], size[1], err)
				}
			}
		}
	}
}

// TestOnTickKeepsSignalLengths checks the windows stay the size they started,
// since a signal that grew every tick would slow the demo down over time.
func TestOnTickKeepsSignalLengths(t *testing.T) {
	a := newApp("Catatui Demo", true)
	sparkline, sin1, sin2 := len(a.sparkline.points), len(a.signals.sin1.points), len(a.signals.sin2.points)
	logs, bars := len(a.logs.items), len(a.barchart)

	for range 50 {
		a.onTick()
	}

	if got := len(a.sparkline.points); got != sparkline {
		t.Errorf("sparkline has %d points, want %d", got, sparkline)
	}
	if got := len(a.signals.sin1.points); got != sin1 {
		t.Errorf("sin1 has %d points, want %d", got, sin1)
	}
	if got := len(a.signals.sin2.points); got != sin2 {
		t.Errorf("sin2 has %d points, want %d", got, sin2)
	}
	if got := len(a.logs.items); got != logs {
		t.Errorf("logs has %d items, want %d", got, logs)
	}
	if got := len(a.barchart); got != bars {
		t.Errorf("barchart has %d bars, want %d", got, bars)
	}
	if a.signals.window != [2]float64{50, 70} {
		t.Errorf("window = %v, want [50 70]", a.signals.window)
	}
}
