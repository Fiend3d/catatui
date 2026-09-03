// Port of examples/apps/volatility-surface/src/display/surface_3d.rs
// @ ratatui-v0.30.2

package main

import (
	"math"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// cameraDistance is how far the camera sits from the projection plane. Bring it
// closer and the perspective exaggerates; push it away and the drawing flattens
// towards an isometric view.
const cameraDistance = 4.0

// surface3D is where the view is from: the two rotations, the zoom and the
// palette. It is the state the widget below draws through, so that the view
// survives from frame to frame while the widget is rebuilt each time.
type surface3D struct {
	rotationX float64
	rotationZ float64
	zoom      float64
	palette   palette
}

func newSurface3D() surface3D {
	return surface3D{
		rotationX: 0.6, // tilted forward, so the surface is seen from above
		rotationZ: 0.3, // and turned slightly, so it is not seen edge on
		zoom:      1.0,
		palette:   palettePlasma,
	}
}

// rotateX tilts the surface towards and away from the viewer, stopping at
// straight down and straight up.
func (s *surface3D) rotateX(delta float64) {
	s.rotationX = min(max(s.rotationX+delta, -math.Pi/2), math.Pi/2)
}

// rotateZ spins the surface about the vertical axis, which wraps round.
func (s *surface3D) rotateZ(delta float64) {
	s.rotationZ += delta
	if s.rotationZ > 2*math.Pi {
		s.rotationZ -= 2 * math.Pi
	}
	if s.rotationZ < 0 {
		s.rotationZ += 2 * math.Pi
	}
}

// zoomBy scales the drawing, within limits that keep it on the screen and
// bigger than a dot.
func (s *surface3D) zoomBy(factor float64) {
	s.zoom = min(max(s.zoom*factor, 0.3), 3.0)
}

// cyclePalette moves to the next colour ramp.
func (s *surface3D) cyclePalette() { s.palette = s.palette.next() }

// project turns a point in 3D space into the 2D one the canvas draws.
//
// This is the whole of the 3D in this example, and it is two rotation matrices
// and a division. Rotating about Z spins the surface; rotating about X tilts
// it. Then the perspective divide: a point twice as far away is drawn half as
// far from the centre, which is what makes the far edge of the surface look
// smaller than the near one.
func (s *surface3D) project(x, y, z float64) (float64, float64) {
	sinX, cosX := math.Sincos(s.rotationX)
	sinZ, cosZ := math.Sincos(s.rotationZ)

	// About the Z axis.
	x1 := x*cosZ - y*sinZ
	y1 := x*sinZ + y*cosZ

	// About the X axis.
	y2 := y1*cosX - z*sinX
	z2 := y1*sinX + z*cosX

	// The perspective divide.
	perspective := cameraDistance / (cameraDistance + z2)
	return x1 * perspective * s.zoom, y2 * perspective * s.zoom
}

// projectNormalized projects a point whose three coordinates are each 0..1,
// which is how the grid indices and the volatility arrive.
func (s *surface3D) projectNormalized(strikeNorm, expiryNorm, volNorm float64) [2]float64 {
	// Centre the surface on the origin, then scale it into world space.
	x := (strikeNorm - 0.5) * 3 // strike: ±1.5 units
	y := (expiryNorm - 0.5) * 3 // expiry: ±1.5 units
	z := (volNorm - 0.5) * 2    // height: ±1 unit, less tall so it is not distorted
	px, py := s.project(x, y, z)
	return [2]float64{px, py}
}

// volatilitySurface draws a grid of volatilities as a 3D wireframe. The view it
// is drawn from is the surface3D state it renders through.
type volatilitySurface struct {
	data [][]float64
}

func newVolatilitySurface(data [][]float64) volatilitySurface {
	return volatilitySurface{data: data}
}

// RenderStateful draws the wireframe.
//
// Deviation from ratatui, which puts this on the state and has the widget
// forward to it: catatui's StatefulWidget is generic over the state, so the
// widget can hold the method itself.
func (v volatilitySurface) RenderStateful(area catatui.Rect, buf *catatui.Buffer, state *surface3D) {
	expiries := len(v.data)
	if expiries == 0 || len(v.data[0]) == 0 {
		return
	}
	strikes := len(v.data[0])
	minVol, maxVol := volatilityRange(v.data)

	widgets.NewCanvas().
		Marker(symbols.Braille). // eight dots to a cell, which is what smooths the curves
		XBounds([2]float64{-2, 2}).
		YBounds([2]float64{-1.5, 1.5}). // narrower, since a cell is taller than it is wide
		Paint(func(ctx *widgets.Context) {
			v.drawStrikeLines(ctx, state, expiries, strikes, minVol, maxVol)
			v.drawExpiryLines(ctx, state, expiries, strikes, minVol, maxVol)
			v.drawPeaks(ctx, state, expiries, strikes, minVol, maxVol)
		}).
		Render(area, buf)
}

// drawStrikeLines draws one line along each expiry, which are the lines running
// across the surface.
func (v volatilitySurface) drawStrikeLines(ctx *widgets.Context, state *surface3D, expiries, strikes int, minVol, maxVol float64) {
	for i, row := range v.data {
		points := make([][2]float64, 0, strikes)
		for j := range strikes {
			points = append(points, state.projectNormalized(
				normalize(j, strikes),
				normalize(i, expiries),
				normalizeVol(row[j], minVol, maxVol)))
		}
		// Kept off the dark end of the ramp, where the lines would be lost
		// against the background.
		shade := float64(i) / float64(expiries)
		drawLineStrip(ctx, points, state.palette.color(min(shade*0.7+0.3, 1)))
	}
}

// drawExpiryLines draws the lines running the other way, up the surface. Only
// every other one, which is enough to read the shape without a thicket.
func (v volatilitySurface) drawExpiryLines(ctx *widgets.Context, state *surface3D, expiries, strikes int, minVol, maxVol float64) {
	for j := 0; j < strikes; j += 2 {
		points := make([][2]float64, 0, expiries)
		for i, row := range v.data {
			points = append(points, state.projectNormalized(
				normalize(j, strikes),
				normalize(i, expiries),
				normalizeVol(row[j], minVol, maxVol)))
		}
		drawLineStrip(ctx, points, state.palette.color(min(normalize(j, strikes)*0.7+0.3, 1)))
	}
}

// drawPeaks marks the highest points, every other one, so that the ridges read
// as ridges where the wireframe alone is ambiguous.
func (v volatilitySurface) drawPeaks(ctx *widgets.Context, state *surface3D, expiries, strikes int, minVol, maxVol float64) {
	var peaks [][2]float64
	for i := 0; i < expiries; i += 2 {
		for j := 0; j < strikes; j += 2 {
			volNorm := normalizeVol(v.data[i][j], minVol, maxVol)
			if volNorm <= 0.7 { // the top 30% only
				continue
			}
			peaks = append(peaks, state.projectNormalized(
				normalize(j, strikes), normalize(i, expiries), volNorm))
		}
	}
	if len(peaks) > 0 {
		ctx.Draw(widgets.NewPoints(peaks, state.palette.color(0.9)))
	}
}

// drawLineStrip joins a run of points up with straight segments.
func drawLineStrip(ctx *widgets.Context, points [][2]float64, color catatui.Color) {
	for i := 1; i < len(points); i++ {
		from, to := points[i-1], points[i]
		ctx.Draw(widgets.NewCanvasLine(from[0], from[1], to[0], to[1], color))
	}
}

// volatilityRange is the lowest and highest value on the surface, which is what
// the heights and the colours are measured against.
func volatilityRange(data [][]float64) (minVol, maxVol float64) {
	minVol, maxVol = math.Inf(1), math.Inf(-1)
	for _, row := range data {
		for _, vol := range row {
			minVol = math.Min(minVol, vol)
			maxVol = math.Max(maxVol, vol)
		}
	}
	return minVol, maxVol
}

// normalize maps an index to 0..1 across n of them.
//
// A single column has no span to spread over, and dividing by n-1 would be a
// division by zero; it goes at the near edge instead. ratatui divides anyway,
// which gives a NaN the canvas silently drops.
func normalize(index, n int) float64 {
	if n < 2 {
		return 0
	}
	return float64(index) / float64(n-1)
}

// normalizeVol maps a volatility to 0..1 across the range on the surface. A
// surface with no range at all sits flat in the middle rather than at a NaN.
func normalizeVol(vol, minVol, maxVol float64) float64 {
	if maxVol == minVol {
		return 0.5
	}
	return (vol - minVol) / (maxVol - minVol)
}
