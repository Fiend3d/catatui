package main

import (
	"math"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the app at sizes from nothing to bigger than a screen, in
// each of the states the keys can put it in. Rendering outside the area given
// panics in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}

	states := map[string]func(*app){
		"default": func(*app) {},
		"paused":  func(a *app) { a.paused = true },
		"rotated to the limits": func(a *app) {
			for range 30 {
				a.surface.rotateX(0.1)
				a.surface.rotateZ(0.1)
			}
		},
		"zoomed in":  func(a *app) { a.surface.zoom = 3 },
		"zoomed out": func(a *app) { a.surface.zoom = 0.3 },
		"last palette": func(a *app) {
			for range paletteCount - 1 {
				a.surface.cyclePalette()
			}
		},
		"an empty surface": func(a *app) { a.engine.surface = nil },
		"a single point":   func(a *app) { a.engine.surface = [][]float64{{20}} },
		"a flat surface": func(a *app) {
			a.engine.surface = [][]float64{{20, 20, 20}, {20, 20, 20}}
		},
	}

	for name, setup := range states {
		for _, size := range sizes {
			a := newApp()
			setup(a)
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
			if err := terminal.Draw(a.draw); err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
		}
	}
}

// TestTheSurfaceIsDrawn checks something actually lands on the canvas, since
// every other test here would pass just as well on an empty screen.
func TestTheSurfaceIsDrawn(t *testing.T) {
	backend := catatui.NewTestBackend(80, 24)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminal.Draw(newApp().draw); err != nil {
		t.Fatal(err)
	}

	painted := 0
	for _, p := range catatui.NewRect(0, 1, 80, 22).Positions() {
		if backend.Buffer().Get(p.X, p.Y).GetSymbol() != " " {
			painted++
		}
	}
	if painted < 100 {
		t.Errorf("only %d cells of the surface were painted:\n%s", painted, backend.Buffer())
	}
}

// TestTheOriginStaysPut checks the projection against the one point whose
// answer is known: the centre of the surface is the centre of the view, at any
// rotation and any zoom.
func TestTheOriginStaysPut(t *testing.T) {
	s := surface3D{rotationX: 1.2, rotationZ: 2.5, zoom: 2.4}
	x, y := s.project(0, 0, 0)
	if x != 0 || y != 0 {
		t.Errorf("the origin projects to (%v, %v), want the centre", x, y)
	}
}

// TestPerspectiveShrinksWithDistance checks the one thing that makes the
// drawing 3D rather than flat: the far half of the surface is drawn smaller
// than the near half.
func TestPerspectiveShrinksWithDistance(t *testing.T) {
	// Tilted so that the Z axis runs away from the viewer, and a point high up
	// is a point far away.
	s := surface3D{rotationX: 0, rotationZ: 0, zoom: 1}

	near, _ := s.project(1, 0, -1)
	middle, _ := s.project(1, 0, 0)
	far, _ := s.project(1, 0, 1)

	if !(near > middle && middle > far) {
		t.Errorf("the same point drawn at three depths gives %v, %v, %v; want it to shrink as it recedes",
			near, middle, far)
	}
	if middle != 1 {
		t.Errorf("a point on the projection plane moved to %v, want it left alone", middle)
	}
}

// TestRotationsStayInRange checks the tilt stops face on rather than turning
// the surface inside out, and that spinning it wraps round instead of growing
// without bound.
func TestRotationsStayInRange(t *testing.T) {
	s := newSurface3D()
	for range 100 {
		s.rotateX(0.1)
	}
	if s.rotationX != math.Pi/2 {
		t.Errorf("the tilt went to %v, want it stopped at a right angle", s.rotationX)
	}
	for range 200 {
		s.rotateX(-0.1)
	}
	if s.rotationX != -math.Pi/2 {
		t.Errorf("the tilt went to %v, want it stopped at a right angle the other way", s.rotationX)
	}

	for range 200 {
		s.rotateZ(0.1)
		if s.rotationZ < 0 || s.rotationZ > 2*math.Pi {
			t.Fatalf("the spin left the circle at %v", s.rotationZ)
		}
	}
	for range 400 {
		s.rotateZ(-0.1)
		if s.rotationZ < 0 || s.rotationZ > 2*math.Pi {
			t.Fatalf("the spin left the circle at %v", s.rotationZ)
		}
	}
}

// TestZoomStaysInRange checks z and x cannot shrink the surface to nothing or
// blow it up off the screen.
func TestZoomStaysInRange(t *testing.T) {
	s := newSurface3D()
	for range 100 {
		s.zoomBy(1.1)
	}
	if s.zoom != 3 {
		t.Errorf("zooming in reached %v, want it capped at 3", s.zoom)
	}
	for range 200 {
		s.zoomBy(0.9)
	}
	if s.zoom != 0.3 {
		t.Errorf("zooming out reached %v, want it floored at 0.3", s.zoom)
	}
}

// TestThePaletteCycles checks p works its way round all six and back, and that
// each one has a name of its own for the header.
func TestThePaletteCycles(t *testing.T) {
	seen := make(map[string]bool)
	p := paletteViridis
	for range paletteCount {
		if seen[p.String()] {
			t.Fatalf("two palettes are both called %q", p)
		}
		seen[p.String()] = true
		p = p.next()
	}
	if p != paletteViridis {
		t.Errorf("cycling %d times ended at %v, want back at the start", paletteCount, p)
	}
}

// TestTheColormapsMatchTheirSource checks the generated tables against the ends
// of matplotlib's own maps, which is what would catch gen_colormaps.py being
// run against something else.
func TestTheColormapsMatchTheirSource(t *testing.T) {
	for _, tc := range []struct {
		palette    palette
		start, end uint32
	}{
		{paletteViridis, 0x440154, 0xfde725},
		{palettePlasma, 0x0d0887, 0xf0f921},
		{paletteInferno, 0x000004, 0xfcffa4},
		{palettePhosphor, 0x002800, 0x26ff3f},
		{paletteFear, 0x7f0000, 0xff4c00},
		{paletteCalm, 0x334c7f, 0x66ccff},
	} {
		if got := tc.palette.color(0); got != catatui.RgbFromU32(tc.start) {
			t.Errorf("%v starts at %v, want %v", tc.palette, got, catatui.RgbFromU32(tc.start))
		}
		if got := tc.palette.color(1); got != catatui.RgbFromU32(tc.end) {
			t.Errorf("%v ends at %v, want %v", tc.palette, got, catatui.RgbFromU32(tc.end))
		}
	}
}

// TestTheColormapsAreClamped checks a value off either end of a ramp comes back
// as the end of it rather than reading past the table.
func TestTheColormapsAreClamped(t *testing.T) {
	for p := range palette(paletteCount) {
		if p.color(-5) != p.color(0) {
			t.Errorf("%v below zero is not its first colour", p)
		}
		if p.color(5) != p.color(1) {
			t.Errorf("%v above one is not its last colour", p)
		}
	}
}

// TestTheSurfaceHasTheRightShape checks the grid the engine builds, since the
// renderer indexes it as [expiry][strike] and would run off the end if it were
// the other way round.
func TestTheSurfaceHasTheRightShape(t *testing.T) {
	e := newVolatilityEngine()
	surface := e.getSurface()
	if len(surface) != expiryCount {
		t.Fatalf("the surface has %d expiries, want %d", len(surface), expiryCount)
	}
	for i, row := range surface {
		if len(row) != strikeCount {
			t.Fatalf("expiry %d has %d strikes, want %d", i, len(row), strikeCount)
		}
		for j, vol := range row {
			if math.IsNaN(vol) || vol < 5 || vol > 80 {
				t.Fatalf("(%d, %d) is %v, want it clamped between 5 and 80", i, j, vol)
			}
		}
	}
}

// TestTheSurfaceMoves checks a tick changes the surface and that ctrl-r puts
// the clock back, which is the difference between the two.
func TestTheSurfaceMoves(t *testing.T) {
	e := newVolatilityEngine()

	// A copy, since regenerating replaces the rows rather than editing them.
	before := make([][]float64, len(e.getSurface()))
	for i, row := range e.getSurface() {
		before[i] = append([]float64(nil), row...)
	}

	for range 20 {
		e.update()
	}
	if e.time == 0 {
		t.Errorf("twenty ticks left the clock at zero")
	}

	// Not every cell moves: the near, far out-of-the-money corners sit against
	// the clamp whatever the time is.
	moved := 0
	for i, row := range e.getSurface() {
		for j, vol := range row {
			if vol != before[i][j] {
				moved++
			}
		}
	}
	if moved == 0 {
		t.Errorf("twenty ticks left the surface exactly as it was")
	}

	e.reset()
	if e.time != 0 {
		t.Errorf("reset left the clock at %v", e.time)
	}
}

// TestADegenerateSurfaceHasNoNaNs checks the two divisions that have a zero in
// them on a surface with one row, one column, or no range at all. A NaN would
// not panic; it would silently drop points and leave a half-drawn frame.
func TestADegenerateSurfaceHasNoNaNs(t *testing.T) {
	s := newSurface3D()
	for _, data := range [][][]float64{
		{{20}},
		{{20, 30, 40}},
		{{20}, {30}, {40}},
		{{20, 20}, {20, 20}},
	} {
		minVol, maxVol := volatilityRange(data)
		for i, row := range data {
			for j, vol := range row {
				point := s.projectNormalized(
					normalize(j, len(row)),
					normalize(i, len(data)),
					normalizeVol(vol, minVol, maxVol))
				if math.IsNaN(point[0]) || math.IsNaN(point[1]) {
					t.Errorf("%v: (%d, %d) projects to %v", data, i, j, point)
				}
			}
		}
	}
}

// TestTheKeysDoWhatTheFooterSays checks each key against the footer's promise,
// since a key wired to the wrong axis is invisible in a still picture.
func TestTheKeysDoWhatTheFooterSays(t *testing.T) {
	rune := func(r rune) term.Event {
		return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
	}

	a := newApp()
	before := a.surface

	a.handle(rune('k'))
	if a.surface.rotationX <= before.rotationX {
		t.Errorf("k did not tilt the surface up")
	}
	a.handle(rune('j'))
	a.handle(rune('j'))
	if a.surface.rotationX >= before.rotationX {
		t.Errorf("j did not tilt the surface down")
	}

	a = newApp()
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyLeft})
	if a.surface.rotationZ <= before.rotationZ {
		t.Errorf("left did not spin the surface")
	}

	a = newApp()
	a.handle(rune('z'))
	if a.surface.zoom <= before.zoom {
		t.Errorf("z did not zoom in")
	}
	a.handle(rune('x'))
	a.handle(rune('x'))
	if a.surface.zoom >= before.zoom {
		t.Errorf("x did not zoom out")
	}

	a = newApp()
	a.handle(rune('p'))
	if a.surface.palette == before.palette {
		t.Errorf("p did not change the palette")
	}

	a = newApp()
	a.handle(rune(' '))
	if !a.paused {
		t.Errorf("space did not pause")
	}
	a.handle(rune(' '))
	if a.paused {
		t.Errorf("space did not resume")
	}

	a = newApp()
	for range 10 {
		a.engine.update()
	}
	a.handle(term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: 'r', Mods: term.ModCtrl})
	if a.engine.time != 0 {
		t.Errorf("ctrl-r did not reset the surface")
	}

	a = newApp()
	a.handle(rune('q'))
	if !a.quit {
		t.Errorf("q did not quit")
	}
}
