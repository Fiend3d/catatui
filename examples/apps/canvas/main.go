// Command canvas draws four canvases at once: a world map, a scratchpad, a
// bouncing ball and a row of rectangles.
//
//	go run ./examples/apps/canvas
//
// h/j/k/l or the arrows move the label on the map a cell at a time, Enter
// cycles through the markers a canvas can be drawn with, dragging the mouse
// draws in the top-left pane, q quits.
//
// Every pane is drawn with the same marker, which is what makes the difference
// between them plain: braille packs eight dots into a cell, octants eight
// blocks, a dot just one.
//
// Port of examples/apps/canvas @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run redraws about sixty times a second so the ball moves smoothly, and
// handles whatever arrives in between.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	a := newApp()

	for !a.exit {
		if err := terminal.Draw(a.render); err != nil {
			return err
		}
		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			a.handle(ev)
		case <-ticker.C:
			a.onTick()
		}
	}
	return nil
}

// app is where the marker on the map is, how the ball is moving, what has been
// drawn, and which marker everything is drawn with.
type app struct {
	exit bool

	// x and y move the "You are here" label about the map.
	x, y float64

	// The ball and the box it bounces around inside, in the canvas's own
	// coordinates rather than in cells.
	ball       widgets.Circle
	playground catatui.Rect
	vx, vy     float64

	marker    symbols.Marker
	points    []catatui.Position
	isDrawing bool

	// mapArea is the pane the map was last drawn in, remembered so that a
	// keypress can move the label by a cell of it. See step.
	mapArea catatui.Rect
}

// step is how far one press of a movement key moves the label: one cell of the
// map as it was last drawn, across and down.
//
// ratatui's example moves by a flat 1.0, which is one degree. The map is 360
// degrees across and 180 down, so on an ordinary terminal that is a tenth of a
// cell sideways and a twenty-second of one downwards: ten presses to move a
// column and twenty-three to move a row, which looks like nothing happening at
// all. Scaling to the pane is the one place this example does not follow
// ratatui, and it is why the keys do anything you can see.
func (a *app) step() (x, y float64) {
	// Two columns and rows go to the block's border, and the label is placed
	// against one less than what is left, as the canvas measures it.
	width := float64(catatui.SatSub(a.mapArea.Width, 3))
	height := float64(catatui.SatSub(a.mapArea.Height, 3))
	if width <= 0 || height <= 0 {
		return 1.0, 1.0
	}
	return 360.0 / width, 180.0 / height
}

// move shifts the label by that many cells and keeps it on the map. A canvas
// draws no label outside its bounds, and a cell-sized step reaches the edge in
// a few presses, so without this the label would simply vanish.
func (a *app) move(cellsX, cellsY float64) {
	stepX, stepY := a.step()
	// The label is drawn at (x, -y), so the y bounds apply to it flipped.
	a.x = min(max(a.x+cellsX*stepX, -180.0), 180.0)
	a.y = min(max(a.y+cellsY*stepY, -90.0), 90.0)
}

func newApp() *app {
	return &app{
		ball:       widgets.NewCircle(20.0, 40.0, 10.0, catatui.ColorYellow),
		playground: catatui.NewRect(10, 10, 200, 100),
		vx:         1.0,
		vy:         1.0,
		marker:     symbols.Dot,
	}
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
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.exit = true
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.move(0, 1)
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.move(0, -1)
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.move(1, 0)
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.move(-1, 0)
	case ev.IsKey(term.KeyEnter):
		a.cycleMarker()
	}
}

func (a *app) handleMouse(ev term.Event) {
	switch ev.MouseKind {
	case term.MouseDown:
		a.isDrawing = true
	case term.MouseUp:
		a.isDrawing = false
	case term.MouseDrag:
		a.points = append(a.points, catatui.Position{X: ev.X, Y: ev.Y})
	}
}

// markerCycle is every marker there is, in the order Enter walks through them:
// coarsest first, then finer and finer, and finally a custom character.
var markerCycle = []symbols.Marker{
	symbols.Dot,
	symbols.Braille,
	symbols.Block,
	symbols.HalfBlock,
	symbols.Quadrant,
	symbols.Sextant,
	symbols.Octant,
	symbols.Custom('×'),
	symbols.Bar,
}

// cycleMarker moves to the next marker, and round to the first from the last.
func (a *app) cycleMarker() {
	for i, marker := range markerCycle {
		if marker == a.marker {
			a.marker = markerCycle[(i+1)%len(markerCycle)]
			return
		}
	}
	a.marker = markerCycle[0]
}

// onTick moves the ball, bouncing it off the walls of the playground by
// flipping whichever part of its velocity took it there.
func (a *app) onTick() {
	left, right := float64(a.playground.Left()), float64(a.playground.Right())
	top, bottom := float64(a.playground.Top()), float64(a.playground.Bottom())

	if a.ball.X-a.ball.Radius < left || a.ball.X+a.ball.Radius > right {
		a.vx = -a.vx
	}
	if a.ball.Y-a.ball.Radius < top || a.ball.Y+a.ball.Radius > bottom {
		a.vy = -a.vy
	}
	a.ball.X += a.vx
	a.ball.Y += a.vy
}

// render draws the header and the four canvases in a two by two grid.
func (a *app) render(f *catatui.Frame) {
	header := catatui.NewText(
		catatui.LineFromStyledString("Canvas Example",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.LineFromString("<q> Quit | <enter> Change Marker | <hjkl> Move"),
	).Centered()

	rows := catatui.VerticalLayout(
		catatui.Length(uint16(header.Height())),
		catatui.Fill(1),
		catatui.Fill(1),
	).Split(f.Area())

	f.RenderWidget(widgets.NewParagraphFromText(header), rows[0])

	top := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).Split(rows[1])
	bottom := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).Split(rows[2])
	draw, pong := top[0], top[1]
	worldMap, boxes := bottom[0], bottom[1]

	// Remembered for step, so the movement keys know how big a cell of the map
	// is in degrees.
	a.mapArea = worldMap

	f.RenderWidget(a.mapCanvas(), worldMap)
	f.RenderWidget(a.drawCanvas(draw), draw)
	f.RenderWidget(a.pongCanvas(), pong)
	f.RenderWidget(a.boxesCanvas(boxes), boxes)
}

// mapCanvas is the world with a label on it, in degrees of latitude and
// longitude.
func (a *app) mapCanvas() widgets.Canvas {
	return widgets.NewCanvas().
		Block(widgets.Bordered().Title("World")).
		Marker(a.marker).
		XBounds([2]float64{-180.0, 180.0}).
		YBounds([2]float64{-90.0, 90.0}).
		Paint(func(ctx *widgets.Context) {
			ctx.Draw(widgets.Map{
				Resolution: widgets.MapResolutionHigh,
				Color:      catatui.ColorGreen,
			})
			ctx.Print(a.x, -a.y, catatui.LineFromStyledString("You are here",
				catatui.NewStyle().Fg(catatui.ColorYellow)))
		})
}

// drawCanvas is the scratchpad. Its bounds are the size of the area in cells,
// so a mouse position converts into it by subtracting where the area starts —
// and flipping the y axis, because a canvas counts up from the bottom while a
// terminal counts down from the top.
func (a *app) drawCanvas(area catatui.Rect) widgets.Canvas {
	return widgets.NewCanvas().
		Block(widgets.Bordered().Title("Draw here")).
		Marker(a.marker).
		XBounds([2]float64{0.0, float64(area.Width)}).
		YBounds([2]float64{0.0, float64(area.Height)}).
		Paint(func(ctx *widgets.Context) {
			coords := make([][2]float64, len(a.points))
			for i, p := range a.points {
				coords[i] = [2]float64{
					float64(p.X) - float64(area.Left()),
					float64(area.Bottom()) - float64(p.Y),
				}
			}
			ctx.Draw(widgets.NewPoints(coords, catatui.ColorWhite))
		})
}

// pongCanvas is the ball, drawn in the coordinates its playground is measured
// in.
func (a *app) pongCanvas() widgets.Canvas {
	return widgets.NewCanvas().
		Block(widgets.Bordered().Title("Pong")).
		Marker(a.marker).
		XBounds([2]float64{10.0, 210.0}).
		YBounds([2]float64{10.0, 110.0}).
		Paint(func(ctx *widgets.Context) {
			ctx.Draw(a.ball)
		})
}

// boxesCanvas is two rows of squares growing left to right, with a ruler along
// each edge, which is what shows how far a marker can be trusted.
func (a *app) boxesCanvas(area catatui.Rect) widgets.Canvas {
	right := float64(area.Width)
	top := float64(area.Height)*2.0 - 4.0

	return widgets.NewCanvas().
		Block(widgets.Bordered().Title("Rects")).
		Marker(a.marker).
		XBounds([2]float64{0.0, right}).
		YBounds([2]float64{0.0, top}).
		Paint(func(ctx *widgets.Context) {
			for i := range 12 {
				x := float64(i*i+3*i)/2.0 + 2.0
				ctx.Draw(widgets.NewRectangle(x, 2.0, float64(i), float64(i), catatui.ColorRed))
				ctx.Draw(widgets.NewRectangle(x, 21.0, float64(i), float64(i), catatui.ColorBlue))
			}
			for i := range 100 {
				if i%10 != 0 {
					ctx.Print(float64(i)+1.0, 0.0,
						catatui.LineFromString(fmt.Sprint(i%10)))
				}
				if i%2 == 0 && i%10 != 0 {
					ctx.Print(0.0, float64(i),
						catatui.LineFromString(fmt.Sprint(i%10)))
				}
			}
		})
}
