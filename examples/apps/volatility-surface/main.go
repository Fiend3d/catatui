// Command volatility-surface draws a 3D wireframe in the terminal: an implied
// volatility surface you can turn, tilt and zoom while it moves.
//
//	go run ./examples/apps/volatility-surface
//
// Arrows or h/j/k/l rotate, z and x zoom, p cycles the colour ramp, space
// pauses, ctrl-r resets the surface, and q quits.
//
// There is no 3D anywhere in catatui, and none is needed. The projection is two
// rotation matrices and a division — a point twice as far away is drawn half as
// far from the centre — and the result goes onto a Canvas in braille, which
// packs eight dots into every cell and is what makes the wireframe read as
// curves rather than as stairs. See surface3d.go, which is the whole of it.
//
// The surface itself is made up rather than priced: volatility.go builds the
// shape a real one has — the skew, the smile, the wings — and ripples it.
//
// Port of examples/apps/volatility-surface @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run redraws fifty times a second, which is what animates the surface, and
// takes keys as they arrive.
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

	for !a.quit {
		if err := terminal.Draw(a.draw); err != nil {
			return err
		}
		a.fps.tick(time.Now())

		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			a.handle(ev)
		case <-ticker.C:
			if !a.paused {
				a.engine.update()
			}
		}
	}
	return nil
}

// app is the view, the surface being viewed, and whether it is moving.
type app struct {
	paused  bool
	quit    bool
	fps     fpsCounter
	engine  *volatilityEngine
	surface surface3D
}

func newApp() *app {
	return &app{
		fps:     fpsCounter{lastInstant: time.Now()},
		engine:  newVolatilityEngine(),
		surface: newSurface3D(),
	}
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape):
		a.quit = true
	case ev.IsRune(' '):
		a.paused = !a.paused
	case ev.IsCtrl('r'):
		a.engine.reset()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.surface.rotateX(0.1)
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.surface.rotateX(-0.1)
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.surface.rotateZ(0.1)
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.surface.rotateZ(-0.1)
	case ev.IsRune('z'):
		a.surface.zoomBy(1.1)
	case ev.IsRune('x'):
		a.surface.zoomBy(0.9)
	case ev.IsRune('p'):
		a.surface.cyclePalette()
	}
}

// draw lays out a header, the surface, and the keys along the bottom.
func (a *app) draw(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Min(0),
		catatui.Length(1),
	).Split(f.Area())

	f.RenderWidget(a.header(), rows[0])
	catatui.RenderStatefulWidgetOn(f, newVolatilitySurface(a.engine.getSurface()), rows[1], &a.surface)
	f.RenderWidget(footer(), rows[2])
}

func (a *app) header() widgets.Paragraph {
	status := "Live"
	if a.paused {
		status = "Paused"
	}
	title := fmt.Sprintf("volatility-surface - Status: %s, Palette: %s, FPS: %d",
		status, a.surface.palette, a.fps.rate())

	return widgets.NewParagraph(title).
		Centered().
		Style(catatui.NewStyle().Bg(tailwind.Slate.C100).Fg(tailwind.Slate.C800))
}

func footer() widgets.Paragraph {
	key := catatui.NewStyle().Fg(catatui.ColorCyan)
	line := catatui.NewLine(
		catatui.NewStyledSpan("↑↓←→/hjkl", key), catatui.NewSpan(" Rotate | "),
		catatui.NewStyledSpan("zx", key), catatui.NewSpan(" Zoom | "),
		catatui.NewStyledSpan("p", key), catatui.NewSpan(" Palette | "),
		catatui.NewStyledSpan("space", key), catatui.NewSpan(" Pause | "),
		catatui.NewStyledSpan("ctrl-r", key), catatui.NewSpan(" Reset | "),
		catatui.NewStyledSpan("q", key), catatui.NewSpan(" Quit"),
	)
	return widgets.NewParagraphFromText(catatui.NewText(line)).
		Centered().
		Style(catatui.NewStyle().Fg(tailwind.Slate.C400).Bg(tailwind.Slate.C950))
}

// fpsCounter counts the frames drawn and works out the rate once a second.
type fpsCounter struct {
	frameCount  int
	lastInstant time.Time
	fps         float64
}

func (c *fpsCounter) tick(now time.Time) {
	c.frameCount++
	if elapsed := now.Sub(c.lastInstant); elapsed >= time.Second {
		c.fps = float64(c.frameCount) / elapsed.Seconds()
		c.frameCount = 0
		c.lastInstant = now
	}
}

// rate is the frame rate as a whole number, which is all the header has room
// for.
func (c *fpsCounter) rate() int { return int(c.fps) }
