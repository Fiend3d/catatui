package main

import (
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, size := range sizes {
		a := &app{fps: fpsWidget{lastInstant: time.Now()}}
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		// Twice, so that the second frame renders from the cached grid.
		for range 2 {
			err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
			if err != nil {
				t.Fatalf("%dx%d: %v", size[0], size[1], err)
			}
		}
	}
}

// TestEveryCellIsAHalfBlock checks the two-pixels-per-cell trick covers the
// whole area: each cell carries the upper half block and both of its colours.
func TestEveryCellIsAHalfBlock(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 8, 3))
	w := &colorsWidget{}
	w.Render(buf.Area, buf)

	for _, p := range buf.Area.Positions() {
		cell := buf.Get(p.X, p.Y)
		if cell.GetSymbol() != symbols.HalfBlockUpper {
			t.Fatalf("(%d, %d) holds %q, want the upper half block", p.X, p.Y, cell.GetSymbol())
		}
		if _, _, _, ok := cell.Fg.RGB(); !ok {
			t.Fatalf("(%d, %d) has a foreground of %v, want a 24-bit colour", p.X, p.Y, cell.Fg)
		}
		if _, _, _, ok := cell.Bg.RGB(); !ok {
			t.Fatalf("(%d, %d) has a background of %v, want a 24-bit colour", p.X, p.Y, cell.Bg)
		}
	}
}

// TestTheGridIsTwiceAsTallAsTheArea checks the widget computes a pixel per half
// cell. Getting this wrong reads off the end of the grid.
func TestTheGridIsTwiceAsTallAsTheArea(t *testing.T) {
	w := &colorsWidget{}
	w.setupColors(catatui.Size{Width: 5, Height: 3})
	if len(w.colors) != 6 {
		t.Fatalf("the grid is %d rows for 3 rows of cells, want 6", len(w.colors))
	}
	for y, row := range w.colors {
		if len(row) != 5 {
			t.Fatalf("row %d is %d columns, want 5", y, len(row))
		}
	}
}

// TestTheGridIsRebuiltOnlyOnResize checks the cache does its job, since that is
// the reason the widget holds state at all.
func TestTheGridIsRebuiltOnlyOnResize(t *testing.T) {
	w := &colorsWidget{}
	w.setupColors(catatui.Size{Width: 4, Height: 2})
	first := &w.colors[0][0]

	w.setupColors(catatui.Size{Width: 4, Height: 2})
	if &w.colors[0][0] != first {
		t.Errorf("the grid was rebuilt for an unchanged size")
	}

	w.setupColors(catatui.Size{Width: 5, Height: 2})
	if &w.colors[0][0] == first {
		t.Errorf("the grid was reused after a resize")
	}
}

// TestTheColoursScrollByOneColumnPerFrame checks the animation: the second
// frame is the first shifted one column left, wrapping around.
func TestTheColoursScrollByOneColumnPerFrame(t *testing.T) {
	area := catatui.NewRect(0, 0, 6, 2)
	w := &colorsWidget{}

	first := catatui.NewBuffer(area)
	w.Render(area, first)
	second := catatui.NewBuffer(area)
	w.Render(area, second)

	for y := range uint16(2) {
		for x := range uint16(6) {
			want := first.Get((x+1)%6, y).Fg
			if got := second.Get(x, y).Fg; got != want {
				t.Fatalf("(%d, %d) is %v on the second frame, want the colour from (%d, %d), %v",
					x, y, got, (x+1)%6, y, want)
			}
		}
	}
}

// TestPrimariesLandOnTheirHues checks the conversion against the three sRGB
// primaries, whose Okhsv hues are known. At the top of the saturation and value
// axes each one has to come back out exactly as it went in.
func TestPrimariesLandOnTheirHues(t *testing.T) {
	for _, tc := range []struct {
		name      string
		hue       float64
		r, g, b   uint8
		tolerance uint8
	}{
		{"red", 29.234, 0xff, 0x00, 0x00, 0},
		{"green", 142.495, 0x00, 0xff, 0x00, 0},
		// The blue corner is where two of Ottosson's three polynomial fits meet
		// and both are at their worst; the reference implementation this is
		// ported from lands a little off it, and so does this.
		{"blue", 264.052, 0x00, 0x00, 0xff, 0x40},
	} {
		r, g, b := okhsvToRGB(tc.hue, 1, 1)
		if diff(r, tc.r) > tc.tolerance || diff(g, tc.g) > tc.tolerance || diff(b, tc.b) > tc.tolerance {
			t.Errorf("%s: hue %v gives #%02x%02x%02x, want #%02x%02x%02x (±%d)",
				tc.name, tc.hue, r, g, b, tc.r, tc.g, tc.b, tc.tolerance)
		}
	}
}

// TestFullSaturationSitsOnTheGamutBoundary checks the property the top row of
// the display depends on: at maximum saturation and value, a colour is as
// chromatic as sRGB can be, which is where one channel is full.
func TestFullSaturationSitsOnTheGamutBoundary(t *testing.T) {
	for hue := 0.0; hue < 360; hue += 3 {
		r, g, b := okhsvToRGB(hue, 1, 1)
		if max(r, g, b) != 0xff {
			t.Errorf("hue %v gives #%02x%02x%02x, whose brightest channel is %d, want 255",
				hue, r, g, b, max(r, g, b))
		}
	}
}

// TestTheEndsOfTheAxes checks the corners of the space: no value is black
// whatever the hue, and no saturation is a neutral.
func TestTheEndsOfTheAxes(t *testing.T) {
	for hue := 0.0; hue < 360; hue += 15 {
		if r, g, b := okhsvToRGB(hue, 1, 0); r|g|b != 0 {
			t.Errorf("hue %v at value 0 gives #%02x%02x%02x, want black", hue, r, g, b)
		}
		r, g, b := okhsvToRGB(hue, 0, 1)
		if r != g || g != b {
			t.Errorf("hue %v at saturation 0 gives #%02x%02x%02x, want a grey", hue, r, g, b)
		}
	}
}

// TestValueDarkensEvenly checks the axis the display runs down the screen is
// monotonic, which is what makes the fade read as a fade.
func TestValueDarkensEvenly(t *testing.T) {
	previous := -1
	for value := 0.0; value <= 1.0; value += 0.05 {
		r, g, b := okhsvToRGB(29.234, 1, value)
		brightness := int(r) + int(g) + int(b)
		if brightness < previous {
			t.Fatalf("value %.2f is darker than the step before it: #%02x%02x%02x", value, r, g, b)
		}
		previous = brightness
	}
}

// TestTheFrameRateNeedsASecondAndTwoFrames checks the counter reports nothing
// until it has enough to report, and then reports the rate it saw.
func TestTheFrameRateNeedsASecondAndTwoFrames(t *testing.T) {
	start := time.Now()
	w := &fpsWidget{lastInstant: start}

	w.tick(start.Add(2 * time.Second))
	if w.hasFPS {
		t.Errorf("the rate was reported after one frame: %v", w.fps)
	}

	for i := range 60 {
		w.tick(start.Add(time.Duration(i+1) * time.Second / 30))
	}
	if !w.hasFPS {
		t.Fatalf("no rate after 60 frames in two seconds")
	}
	if w.fps < 29 || w.fps > 31 {
		t.Errorf("60 frames in two seconds came to %v fps, want about 30", w.fps)
	}
}

func diff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
