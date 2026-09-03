// Command line-gauge shows catatui's LineGauge widget filling up over time.
//
//	go run ./examples/widgets/line-gauge
//
// Press space to start or stop, r to reset, q to quit.
//
// Port of ratatui-widgets/examples/line-gauge.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// tickRate is how often the progress advances by one column.
const tickRate = time.Second / 20

// appState is whether the progress is advancing.
type appState int

const (
	stateStart appState = iota
	stateStop
	stateQuit
)

// app is the whole example: three gauges showing one shared progress value.
type app struct {
	state           appState
	progressColumns uint16
	progress        float64
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run ticks the progress along and redraws until the user quits.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	a := &app{}
	for a.state != stateQuit {
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
			size, err := terminal.Size()
			if err != nil {
				return err
			}
			a.update(size.Width)
		}
	}
	return nil
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune(' '):
		a.toggleStart()
	case ev.IsRune('r'):
		a.reset()
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.state = stateQuit
	}
}

// update advances the progress by one column of the terminal.
func (a *app) update(terminalWidth uint16) {
	if a.state != stateStart || terminalWidth == 0 {
		return
	}
	a.progressColumns = min(a.progressColumns+1, terminalWidth)
	a.progress = float64(a.progressColumns) / float64(terminalWidth)
}

// toggleStart starts the progress if it is stopped, and stops it if not.
func (a *app) toggleStart() {
	if a.state == stateStart {
		a.state = stateStop
	} else {
		a.state = stateStart
	}
}

// reset empties the gauges and stops.
func (a *app) reset() {
	a.progress = 0
	a.progressColumns = 0
	a.state = stateStop
}

// render draws the header and the three gauges.
func (a *app) render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(3), catatui.Min(0)).Split(f.Area())
	gauges := catatui.VerticalLayout(catatui.Length(2), catatui.Length(2), catatui.Length(2)).
		Split(rows[1])

	renderHeader(f, rows[0])
	a.renderGauge(f, gauges[0], "", "", 149, 58)
	a.renderGauge(f, gauges[1], "⣿", "⣿", 45, 24)
	a.renderGauge(f, gauges[2], "▰", "▱", 75, 25)
}

// renderHeader draws the title and the key bindings.
func renderHeader(f *catatui.Frame, area catatui.Rect) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Min(1)).Split(area)

	f.RenderWidget(
		widgets.NewParagraph("LineGauge Example").
			Style(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
			Centered(),
		rows[0])
	f.RenderWidget(
		widgets.NewParagraph(
			"(press 'SPACE' to start/stop progress, 'r' to reset progress, 'q' to quit)").
			Centered(),
		rows[1])
}

// renderGauge draws one gauge in the given palette. Empty symbols keep the
// gauge's own defaults.
func (a *app) renderGauge(f *catatui.Frame, area catatui.Rect, filled, unfilled string, filledColor, unfilledColor uint8) {
	gauge := widgets.NewLineGauge().
		FilledStyle(catatui.NewStyle().Fg(catatui.Indexed(filledColor))).
		UnfilledStyle(catatui.NewStyle().Fg(catatui.Indexed(unfilledColor))).
		Ratio(a.progress)
	if filled != "" {
		gauge = gauge.FilledSymbol(filled).UnfilledSymbol(unfilled)
	}
	f.RenderWidget(gauge, area)
}
