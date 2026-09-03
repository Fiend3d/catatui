// Command mouse-drawing is a scratchpad you draw on with the mouse.
//
//	go run ./examples/apps/mouse-drawing
//
// Click to place a point, drag to draw a line, space for a new colour at
// random, q or Esc to quit.
//
// Dragging only reports where the pointer is now, not every cell it passed
// over, so the gaps are filled in with Bresenham's line algorithm — otherwise a
// quick stroke comes out as a dotted line.
//
// Port of examples/apps/mouse-drawing @ ratatui-v0.30.2
package main

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	defer term.RecoverAndRestore()

	// Without this the mouse events never arrive and there is nothing to draw
	// with.
	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{currentColor: catatui.ColorWhite}

	for !a.quit {
		if err := terminal.Draw(a.render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		a.handle(ev)
	}
	return nil
}

// point is one cell that has been drawn on, in the colour it was drawn in.
type point struct {
	position catatui.Position
	color    catatui.Color
}

// app is everything drawn so far, where the pointer is, and what colour the
// next stroke will be.
type app struct {
	points        []point
	mousePosition catatui.Position
	hasMouse      bool
	currentColor  catatui.Color
	quit          bool
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	switch ev.Kind {
	case term.EventKey:
		a.handleKey(ev)
	case term.EventMouse:
		a.handleMouse(ev)
	}
}

func (a *app) handleKey(ev term.Event) {
	switch {
	case ev.IsRune(' '):
		a.currentColor = catatui.Rgb(
			uint8(rand.UintN(256)), uint8(rand.UintN(256)), uint8(rand.UintN(256)))
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	}
}

func (a *app) handleMouse(ev term.Event) {
	position := catatui.Position{X: ev.X, Y: ev.Y}
	switch ev.MouseKind {
	case term.MouseDown:
		a.points = append(a.points, point{position, a.currentColor})
	case term.MouseDrag:
		a.drawLine(position)
	}
	a.mousePosition, a.hasMouse = position, true
}

// drawLine fills in the cells between the last point and this one.
func (a *app) drawLine(position catatui.Position) {
	if len(a.points) == 0 {
		return
	}
	start := a.points[len(a.points)-1].position
	for _, p := range bresenham(start, position) {
		a.points = append(a.points, point{p, a.currentColor})
	}
}

// bresenham lists the cells on the line from start to end, both included.
//
// It is the integer line algorithm: step along whichever axis is longer, and
// carry an error term that says when to step along the other one.
func bresenham(start, end catatui.Position) []catatui.Position {
	x0, y0 := int(start.X), int(start.Y)
	x1, y1 := int(end.X), int(end.Y)

	dx, dy := abs(x1-x0), -abs(y1-y0)
	stepX, stepY := 1, 1
	if x0 > x1 {
		stepX = -1
	}
	if y0 > y1 {
		stepY = -1
	}

	var line []catatui.Position
	err := dx + dy
	for {
		line = append(line, catatui.Position{X: uint16(x0), Y: uint16(y0)})
		if x0 == x1 && y0 == y1 {
			return line
		}
		// Doubling the error is what keeps this in whole numbers: it compares
		// the error against dy and dx rather than against half of them.
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += stepX
		}
		if e2 <= dx {
			err += dx
			y0 += stepY
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// render draws what has been drawn, then the pointer over it, then the title
// over that: later widgets land on top of earlier ones.
func (a *app) render(f *catatui.Frame) {
	a.renderPoints(f)
	a.renderMouseCursor(f)
	f.RenderWidget(
		catatui.LineFromString(
			"Mouse Example ('Esc' to quit. Click / drag to draw. 'Space' to change color)").
			Centered(),
		f.Area())
}

func (a *app) renderPoints(f *catatui.Frame) {
	for _, p := range a.points {
		f.RenderWidget(
			catatui.LineFromStyledString(symbols.BlockFull,
				catatui.NewStyle().Fg(p.color)),
			cell(p.position, f.Area()))
	}
}

func (a *app) renderMouseCursor(f *catatui.Frame) {
	if !a.hasMouse {
		return
	}
	f.RenderWidget(
		catatui.LineFromStyledString("╳", catatui.NewStyle().Bg(a.currentColor)),
		cell(a.mousePosition, f.Area()))
}

// cell is the single-cell area at a position, kept inside the screen. A point
// drawn at the edge and then reported outside it would otherwise be drawn out
// of bounds, which panics.
func cell(position catatui.Position, area catatui.Rect) catatui.Rect {
	return catatui.NewRect(position.X, position.Y, 1, 1).Clamp(area)
}
