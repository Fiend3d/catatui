// Command advanced-widget-impl shows the ways a type can be a widget, and what
// each one is for.
//
//	go run ./examples/apps/advanced-widget-impl
//
// Any key quits.
//
// ratatui's example is about Rust's three impls — Widget for T, for &T and for
// &mut T — plus WidgetRef for the boxed case. Go's Widget interface has one
// method and no such split, so the same four widgets are here as the four
// shapes a Go widget takes:
//
//   - greeting, a value built for one frame and thrown away.
//   - timer, a value receiver over state that lives longer than the frame.
//   - squares, a container holding widgets of types it does not know.
//   - rightAlignedSquare, a pointer receiver, which is what lets a widget
//     record something it can only work out while drawing — an alternative to
//     StatefulWidget when the state belongs to the widget itself.
//
// Port of examples/apps/advanced-widget-impl @ ratatui-v0.30.2
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

// run redraws fifty times a second, which is what keeps the timer moving, and
// quits on the first key.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(time.Second / 50)
	defer ticker.Stop()

	a := newApp()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) }); err != nil {
			return err
		}
		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			if ev.Kind == term.EventKey {
				return nil
			}
		case <-ticker.C:
		}
	}
}

// app owns the widgets it draws, and is a widget itself.
//
// It renders on a pointer receiver because one of its children does: the green
// square works out where it goes while drawing, and the app reads that back
// afterwards to label it.
type app struct {
	timer       timer
	squares     squares
	greenSquare rightAlignedSquare
}

func newApp() *app {
	return &app{
		timer:   timer{start: time.Now()},
		squares: squares{children: []catatui.Widget{redSquare{}, blueSquare{}}},
	}
}

func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Length(1),
		catatui.Length(2),
		catatui.Length(1),
	).Split(area)
	greetingArea, timerArea, squaresArea, positionArea := rows[0], rows[1], rows[2], rows[3]

	// A widget built for this frame and dropped after it.
	greeting{name: "catatui!"}.Render(greetingArea, buf)

	// A widget over state that outlives the frame.
	a.timer.Render(timerArea, buf)

	// A container of widgets whose types it does not know.
	a.squares.Render(squaresArea, buf)

	// A widget that records where it drew itself, over the same area.
	a.greenSquare.Render(squaresArea, buf)
	position := fmt.Sprintf("Green square is at (%d, %d)",
		a.greenSquare.lastPosition.X, a.greenSquare.lastPosition.Y)
	catatui.NewSpan(position).Render(positionArea, buf)
}

// greeting is built fresh each frame and holds nothing between them. This is
// the ordinary case, and what every widget in the library is.
type greeting struct{ name string }

func (g greeting) Render(area catatui.Rect, buf *catatui.Buffer) {
	catatui.NewSpan("Hello, "+g.name).Render(area, buf)
}

// timer draws how long it has been running. The value receiver copies the
// start time, which is all it needs: the state is read, never written.
type timer struct{ start time.Time }

func (t timer) Render(area catatui.Rect, buf *catatui.Buffer) {
	elapsed := time.Since(t.start).Seconds()
	catatui.NewSpan(fmt.Sprintf("Elapsed: %.1fs", elapsed)).Render(area, buf)
}

// squares lays a list of widgets out left to right, four columns each. It
// stores them as catatui.Widget, so anything that draws itself can go in.
type squares struct{ children []catatui.Widget }

func (s squares) Render(area catatui.Rect, buf *catatui.Buffer) {
	constraints := make([]catatui.Constraint, len(s.children))
	for i := range constraints {
		constraints[i] = catatui.Length(squareWidth)
	}
	areas := catatui.HorizontalLayout(constraints...).Split(area)
	for i, child := range s.children {
		child.Render(areas[i], buf)
	}
}

// redSquare and blueSquare are two types with nothing in common but the Widget
// interface, which is how they can share the container above.
type redSquare struct{}

func (redSquare) Render(area catatui.Rect, buf *catatui.Buffer) {
	fill(area, buf, catatui.ColorRed)
}

type blueSquare struct{}

func (blueSquare) Render(area catatui.Rect, buf *catatui.Buffer) {
	fill(area, buf, catatui.ColorBlue)
}

// squareWidth is how wide each square is drawn.
const squareWidth uint16 = 4

// rightAlignedSquare draws itself against the right edge of whatever area it is
// given, and remembers where that turned out to be.
//
// The pointer receiver is what makes that possible. A widget that has to work
// something out while drawing — where it ended up, so a later click can be
// tested against it — can keep the answer in itself rather than in the separate
// state a StatefulWidget would take.
type rightAlignedSquare struct{ lastPosition catatui.Position }

func (r *rightAlignedSquare) Render(area catatui.Rect, buf *catatui.Buffer) {
	width := catatui.MinU16(squareWidth, area.Width)
	r.lastPosition = catatui.Position{X: catatui.SatSub(area.Right(), width), Y: area.Y}
	fill(catatui.NewRect(r.lastPosition.X, r.lastPosition.Y, width, area.Height),
		buf, catatui.ColorGreen)
}

// fill paints an area solid. widgets.Fill is the library's answer to the helper
// ratatui's example writes out by hand and wishes were a method on Buffer.
func fill(area catatui.Rect, buf *catatui.Buffer, color catatui.Color) {
	widgets.NewFill(symbols.BlockFull).
		Style(catatui.NewStyle().Fg(color)).
		Render(area, buf)
}
