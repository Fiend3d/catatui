// Command gauge fills four gauges at once, which is what shows the difference
// between the ways of measuring them.
//
//	go run ./examples/apps/gauge
//
// Press Enter or space to start, q to quit.
//
// The first two gauges track the same value, one as a whole percentage and one
// as a ratio, so the first steps and the second glides. The last two track
// another value, one drawn in whole cells and one in eighths of a cell.
//
// Port of examples/apps/gauge @ ratatui-v0.30.2
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

// A colour per gauge, and one for the labels over them.
var (
	gauge1Color      = tailwind.Red.C800
	gauge2Color      = tailwind.Green.C800
	gauge3Color      = tailwind.Blue.C800
	gauge4Color      = tailwind.Orange.C800
	customLabelColor = tailwind.Slate.C200
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run ticks the gauges along twenty times a second, drawing between ticks.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()

	a := &app{}

	for !a.quit {
		if err := terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) }); err != nil {
			return err
		}

		// Whichever comes first: a key, or the next tick.
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
			a.tick(size.Width)
		}
	}
	return nil
}

// app is how far each gauge has filled, and whether they have been started.
type app struct {
	started bool
	quit    bool

	// progressColumns is counted in columns of the terminal rather than in
	// percent, so that the two gauges below it disagree in a way you can see.
	progressColumns uint16
	progress1       uint16
	progress2       float64
	progress3       float64
	progress4       float64
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune(' '), ev.IsKey(term.KeyEnter):
		a.started = true
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	}
}

// tick moves every gauge on by one step, once the app has been started.
func (a *app) tick(terminalWidth uint16) {
	if !a.started || terminalWidth == 0 {
		return
	}

	// The first two show the difference between a ratio and a percentage
	// measuring the same thing: one rounds to a whole percent, the other does
	// not, so the first gauge lags behind the second.
	a.progressColumns = min(a.progressColumns+1, terminalWidth)
	a.progress1 = a.progressColumns * 100 / terminalWidth
	a.progress2 = float64(a.progressColumns) * 100.0 / float64(terminalWidth)

	// The last two show the difference between a gauge drawn in whole cells
	// and one drawn in eighths, and start part-filled so there is something to
	// see straight away.
	a.progress3 = min(max(a.progress3+0.1, 40.0), 100.0)
	a.progress4 = min(max(a.progress4+0.1, 40.0), 100.0)
}

// Render draws the heading, the four gauges and the footer.
func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(2), catatui.Min(0), catatui.Length(1),
	).Split(area)

	renderHeader(rows[0], buf)
	renderFooter(rows[2], buf)

	gauges := catatui.VerticalLayout(
		catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4),
	).Split(rows[1])

	a.renderGauge1(gauges[0], buf)
	a.renderGauge2(gauges[1], buf)
	a.renderGauge3(gauges[2], buf)
	a.renderGauge4(gauges[3], buf)
}

func renderHeader(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Catatui Gauge Example").
		Centered().
		Style(catatui.NewStyle().
			Fg(customLabelColor).
			AddModifier(catatui.ModifierBold)).
		Render(area, buf)
}

func renderFooter(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Press ENTER to start").
		Centered().
		Style(catatui.NewStyle().
			Fg(customLabelColor).
			AddModifier(catatui.ModifierBold)).
		Render(area, buf)
}

// renderGauge1 measures in whole percent, so it moves a step at a time.
func (a *app) renderGauge1(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewGauge().
		Block(titleBlock("Gauge with percentage")).
		GaugeStyle(catatui.NewStyle().Fg(gauge1Color)).
		Percent(a.progress1).
		Render(area, buf)
}

// renderGauge2 measures the same value as a ratio, and labels it to a tenth.
func (a *app) renderGauge2(area catatui.Rect, buf *catatui.Buffer) {
	label := catatui.NewStyledSpan(fmt.Sprintf("%.1f/100", a.progress2),
		catatui.NewStyle().
			Fg(customLabelColor).
			AddModifier(catatui.ModifierItalic).
			AddModifier(catatui.ModifierBold))

	widgets.NewGauge().
		Block(titleBlock("Gauge with ratio and custom label")).
		GaugeStyle(catatui.NewStyle().Fg(gauge2Color)).
		Ratio(a.progress2/100.0).
		LabelSpan(label).
		Render(area, buf)
}

// renderGauge3 fills a whole cell at a time.
func (a *app) renderGauge3(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewGauge().
		Block(titleBlock("Gauge with ratio (no unicode)")).
		GaugeStyle(catatui.NewStyle().Fg(gauge3Color)).
		Ratio(a.progress3/100.0).
		Label(fmt.Sprintf("%.1f%%", a.progress3)).
		Render(area, buf)
}

// renderGauge4 fills the same value in eighths of a cell, using the block
// characters, so the last cell is partly drawn rather than empty.
func (a *app) renderGauge4(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewGauge().
		Block(titleBlock("Gauge with ratio (unicode)")).
		GaugeStyle(catatui.NewStyle().Fg(gauge4Color)).
		Ratio(a.progress4/100.0).
		Label(fmt.Sprintf("%.1f%%", a.progress3)).
		UseUnicode(true).
		Render(area, buf)
}

// titleBlock is a heading over a gauge, with a blank row above and below it
// rather than a border.
func titleBlock(title string) widgets.Block {
	return widgets.NewBlock().
		Borders(widgets.BordersNone).
		Padding(widgets.VerticalPadding(1)).
		TitleLine(catatui.LineFromString(title).Centered()).
		Style(catatui.NewStyle().Fg(customLabelColor))
}
