package main

import (
	"math"
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the charts at the start and after a while of scrolling, at
// sizes from nothing to bigger than a screen. Rendering outside the area given
// panics in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, ticks := range []int{0, 1, 50} {
		a := newApp()
		for range ticks {
			a.onTick()
		}
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("%d ticks, %dx%d: %v", ticks, size[0], size[1], err)
			}
			if err := terminal.Draw(a.render); err != nil {
				t.Fatalf("%d ticks, %dx%d: %v", ticks, size[0], size[1], err)
			}
		}
	}
}

// TestTickKeepsTheWindowOverTheData checks the window and the points move
// together: the data is dropped from the front and added at the back at the
// same rate the window slides, so the wave stays put on screen.
func TestTickKeepsTheWindowOverTheData(t *testing.T) {
	a := newApp()
	before1, before2 := len(a.data1), len(a.data2)

	for range 100 {
		a.onTick()
	}

	if len(a.data1) != before1 || len(a.data2) != before2 {
		t.Errorf("the data grew to %d and %d points, want %d and %d",
			len(a.data1), len(a.data2), before1, before2)
	}
	if got, want := a.window, [2]float64{100, 120}; got != want {
		t.Errorf("the window is %v, want %v", got, want)
	}
	// The window has to stay over points that exist, or the chart draws empty.
	//
	// The tolerance is not decoration: the signal's x is accumulated an
	// interval at a time, so after a few hundred points it is a fraction of a
	// millionth off the whole number the window counts in.
	const tolerance = 1e-6
	first, last := a.data1[0][0], a.data1[len(a.data1)-1][0]
	if a.window[0] < first-tolerance || a.window[1] > last+tolerance {
		t.Errorf("the window %v is outside the data, which runs %v to %v", a.window, first, last)
	}
}

// TestSinSignalFollowsTheWave checks the signal is a sine of the right size and
// steps along by its interval.
func TestSinSignalFollowsTheWave(t *testing.T) {
	s := sinSignal{interval: 0.2, period: 3.0, scale: 18.0}
	points := s.take(50)

	for i, p := range points {
		wantX := float64(i) * 0.2
		if math.Abs(p[0]-wantX) > 1e-9 {
			t.Fatalf("point %d is at x=%v, want %v", i, p[0], wantX)
		}
		wantY := math.Sin(wantX/3.0) * 18.0
		if math.Abs(p[1]-wantY) > 1e-9 {
			t.Fatalf("point %d is at y=%v, want %v", i, p[1], wantY)
		}
		if p[1] < -18.0 || p[1] > 18.0 {
			t.Fatalf("point %d is at y=%v, outside the scale", i, p[1])
		}
	}
}

// TestNumberFormatsLikeRust checks the moving axis labels come out as whole
// numbers where they are whole, which is what Rust's {} prints for an f64.
func TestNumberFormatsLikeRust(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{0, "0"}, {20, "20"}, {10.5, "10.5"}, {-3, "-3"},
	} {
		if got := number(tc.in); got != tc.want {
			t.Errorf("number(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
