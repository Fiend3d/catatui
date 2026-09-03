// Command hello is a minimal catatui application: a constraint-driven layout,
// direct buffer drawing, and an event loop that blocks when idle.
//
//	go run ./examples/apps/hello
//
// Press q or ctrl+c to quit, arrow keys to move the marker, and click anywhere
// with the mouse.
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

type app struct {
	markerX, markerY uint16
	lastEvent        string
	clicks           int
	quit             bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// RecoverAndRestore puts the terminal back if anything below panics, so a
	// crash leaves a usable shell and a readable stack trace.
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{markerX: 2, markerY: 1, lastEvent: "waiting..."}

	for !a.quit {
		if err := terminal.Draw(a.draw); err != nil {
			return err
		}

		// Block until something happens. An idle UI costs no CPU at all, which
		// is the main reason to prefer an event-driven loop over a tick rate.
		ev, ok := <-events.Events()
		if !ok {
			break
		}
		a.handle(ev)

		// Drain whatever else is already queued before drawing again. A fast
		// wheel spin or a mouse drag produces dozens of events, and collapsing
		// them into one frame is what keeps the UI feeling immediate.
		for drained := true; drained; {
			select {
			case ev, ok := <-events.Events():
				if !ok {
					a.quit = true
					drained = false
					break
				}
				a.handle(ev)
			default:
				drained = false
			}
		}
	}
	return events.Err()
}

func (a *app) handle(ev term.Event) {
	switch ev.Kind {
	case term.EventKey:
		switch {
		case ev.IsRune('q'), ev.IsCtrl('c'), ev.IsKey(term.KeyEscape):
			a.quit = true
		case ev.IsKey(term.KeyLeft):
			a.markerX = saturatingSub(a.markerX, 1)
		case ev.IsKey(term.KeyRight):
			a.markerX++
		case ev.IsKey(term.KeyUp):
			a.markerY = saturatingSub(a.markerY, 1)
		case ev.IsKey(term.KeyDown):
			a.markerY++
		}
		if ev.Key == term.KeyRune {
			a.lastEvent = fmt.Sprintf("key %q %s", ev.Rune, ev.Mods)
		} else {
			a.lastEvent = fmt.Sprintf("key %d %s", ev.Key, ev.Mods)
		}
	case term.EventMouse:
		if ev.MouseKind == term.MouseDown {
			a.markerX, a.markerY = ev.X, ev.Y
			a.clicks++
		}
		a.lastEvent = fmt.Sprintf("mouse %d at (%d, %d)", ev.MouseKind, ev.X, ev.Y)
	case term.EventResize:
		a.lastEvent = fmt.Sprintf("resize to %dx%d", ev.Size.Width, ev.Size.Height)
	}
}

func saturatingSub(a, b uint16) uint16 {
	if a < b {
		return 0
	}
	return a - b
}

func (a *app) draw(f *catatui.Frame) {
	// A title bar, a body that takes whatever is left, and a status line.
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Fill(1),
		catatui.Length(1),
	).Split(f.Area())
	title, body, status := rows[0], rows[1], rows[2]

	f.RenderWidget(
		widgets.NewParagraph(" catatui — ratatui in Go ").
			Style(catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorWhite).
				AddModifier(catatui.ModifierBold)),
		title)

	// Split the body in two to show the layout solver doing real work.
	cols := catatui.HorizontalLayout(
		catatui.Percentage(40),
		catatui.Fill(1),
	).Spacing(catatui.Space(1)).Split(body)

	f.RenderWidget(
		widgets.NewParagraph(fmt.Sprintf(
			"Percentage(40) | Fill(1)\n\nleft:  %dx%d\nright: %dx%d\n\nclicks: %d",
			cols[0].Width, cols[0].Height, cols[1].Width, cols[1].Height, a.clicks)).
			Block(widgets.Bordered().
				Title("Layout").
				BorderStyle(catatui.NewStyle().Fg(catatui.ColorCyan)).
				Padding(widgets.HorizontalPadding(1))).
			Wrap(widgets.Wrap{Trim: false}),
		cols[0])

	canvas := widgets.Bordered().
		Title("Canvas").
		BorderStyle(catatui.NewStyle().Fg(catatui.ColorCyan))
	inner := canvas.Inner(cols[1])
	f.RenderWidget(canvas, cols[1])

	// The marker is drawn straight into the buffer, which is how a program with
	// its own rendering needs uses catatui.
	buf := f.Buffer()
	if inner.Contains(catatui.Position{X: a.markerX, Y: a.markerY}) {
		buf.SetString(a.markerX, a.markerY, "@",
			catatui.NewStyle().Fg(catatui.ColorYellow).AddModifier(catatui.ModifierBold))
	} else if !inner.IsEmpty() {
		// Keep the marker reachable after a resize shrinks the pane.
		buf.SetString(inner.X, inner.Y, "@", catatui.NewStyle().Fg(catatui.ColorDarkGray))
	}
	if inner.Height > 1 {
		f.RenderWidget(
			widgets.NewParagraph("arrows move · click to place · q quits").
				Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)),
			catatui.Rect{X: inner.X, Y: inner.Bottom() - 1, Width: inner.Width, Height: 1})
	}

	f.RenderWidget(
		widgets.NewParagraph(" "+a.lastEvent).
			Style(catatui.NewStyle().Bg(catatui.ColorDarkGray).Fg(catatui.ColorWhite)),
		status)
}
