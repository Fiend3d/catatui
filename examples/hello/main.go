// Command hello is a minimal catatui application: a constraint-driven layout,
// direct buffer drawing, and an event loop that blocks when idle.
//
//	go run ./examples/hello
//
// Press q or ctrl+c to quit, arrow keys to move the marker, and click anywhere
// with the mouse.
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
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

	buf := f.Buffer()
	title, body, status := rows[0], rows[1], rows[2]

	fill(buf, title, catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorWhite))
	buf.SetStringn(title.X+1, title.Y, " catatui — ratatui in Go ", title.Width,
		catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorWhite).AddModifier(catatui.ModifierBold))

	// Split the body in two to show the layout solver doing real work.
	cols := catatui.HorizontalLayout(
		catatui.Percentage(40),
		catatui.Fill(1),
	).Spacing(catatui.Space(1)).Split(body)

	drawBox(buf, cols[0], "Layout")
	lines := []string{
		"Percentage(40) | Fill(1)",
		"",
		fmt.Sprintf("left:  %dx%d", cols[0].Width, cols[0].Height),
		fmt.Sprintf("right: %dx%d", cols[1].Width, cols[1].Height),
		"",
		fmt.Sprintf("clicks: %d", a.clicks),
	}
	for i, s := range lines {
		y := cols[0].Y + 1 + uint16(i)
		if y >= cols[0].Bottom()-1 {
			break
		}
		buf.SetStringn(cols[0].X+2, y, s, saturatingSub(cols[0].Width, 4), catatui.NewStyle())
	}

	drawBox(buf, cols[1], "Canvas")
	inner := cols[1].Inner(catatui.NewMargin(1, 1))
	if inner.Contains(catatui.Position{X: a.markerX, Y: a.markerY}) {
		buf.SetString(a.markerX, a.markerY, "@",
			catatui.NewStyle().Fg(catatui.ColorYellow).AddModifier(catatui.ModifierBold))
	} else if !inner.IsEmpty() {
		// Keep the marker reachable after a resize shrinks the pane.
		buf.SetString(inner.X, inner.Y, "@", catatui.NewStyle().Fg(catatui.ColorDarkGray))
	}
	buf.SetStringn(inner.X, inner.Bottom()-1,
		"arrows move · click to place · q quits", inner.Width,
		catatui.NewStyle().Fg(catatui.ColorDarkGray))

	fill(buf, status, catatui.NewStyle().Bg(catatui.ColorDarkGray))
	buf.SetStringn(status.X+1, status.Y, a.lastEvent, saturatingSub(status.Width, 2),
		catatui.NewStyle().Bg(catatui.ColorDarkGray).Fg(catatui.ColorWhite))
}

// fill paints a solid background over an area.
func fill(buf *catatui.Buffer, area catatui.Rect, style catatui.Style) {
	for y := area.Top(); y < area.Bottom(); y++ {
		for x := area.Left(); x < area.Right(); x++ {
			buf.Get(x, y).SetSymbol(" ").SetStyle(style)
		}
	}
}

// drawBox draws a single-line border with a title. The widget layer will make
// this a one-liner; until then it shows what drawing straight into the buffer
// looks like, which is how nezumi uses ratatui.
func drawBox(buf *catatui.Buffer, area catatui.Rect, title string) {
	if area.Width < 2 || area.Height < 2 {
		return
	}
	style := catatui.NewStyle().Fg(catatui.ColorCyan)
	right, bottom := area.Right()-1, area.Bottom()-1

	for x := area.Left() + 1; x < right; x++ {
		buf.Get(x, area.Top()).SetSymbol("─").SetStyle(style)
		buf.Get(x, bottom).SetSymbol("─").SetStyle(style)
	}
	for y := area.Top() + 1; y < bottom; y++ {
		buf.Get(area.Left(), y).SetSymbol("│").SetStyle(style)
		buf.Get(right, y).SetSymbol("│").SetStyle(style)
	}
	buf.Get(area.Left(), area.Top()).SetSymbol("┌").SetStyle(style)
	buf.Get(right, area.Top()).SetSymbol("┐").SetStyle(style)
	buf.Get(area.Left(), bottom).SetSymbol("└").SetStyle(style)
	buf.Get(right, bottom).SetSymbol("┘").SetStyle(style)

	buf.SetStringn(area.X+2, area.Y, " "+title+" ", saturatingSub(area.Width, 4),
		style.AddModifier(catatui.ModifierBold))
}
