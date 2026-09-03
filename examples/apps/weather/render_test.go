package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	// Fixed readings rather than newTemperatures', so a failure is the same
	// failure twice running: both ends of the colour range and the middle.
	temperatures := make([]uint8, 24)
	for i := range temperatures {
		temperatures[i] = uint8(50 + i*39/23)
	}

	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, size := range sizes {
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, temperatures) }); err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
	}
}

// TestTemperatureStyleSpansTheRange checks the colour runs from yellow to red
// across the readings newTemperatures makes, and that neither end wraps around
// — the green channel is a uint8, so a reading under 50 would underflow it.
func TestTemperatureStyleSpansTheRange(t *testing.T) {
	previous := 256
	for value := 50; value < 90; value++ {
		fg := temperatureStyle(uint8(value)).GetFg()
		_, green, _, ok := fg.RGB()
		if !ok {
			t.Fatalf("%d degrees is not an RGB colour", value)
		}
		if int(green) >= previous {
			t.Errorf("%d degrees is greener than the reading below it (%d, was %d)",
				value, green, previous)
		}
		previous = int(green)
	}
	if got := len(newTemperatures()); got != 24 {
		t.Errorf("made up %d readings, want one an hour", got)
	}
	for _, value := range newTemperatures() {
		if value < 50 || value >= 90 {
			t.Errorf("made up a reading of %d, outside the range the colours cover", value)
		}
	}
}
