// Command sparkline shows catatui's Sparkline widget, including an animated
// sine wave.
//
//	go run ./examples/sparkline
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/sparkline.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// frameRate is how often the animated sparkline is redrawn.
const frameRate = time.Second / 60

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run redraws on a timer, since the sine wave moves with the frame count, and
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

	ticker := time.NewTicker(frameRate)
	defer ticker.Stop()

	for {
		if err := terminal.Draw(render); err != nil {
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

// render draws a fixed sparkline and an animated one.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(1), catatui.Max(2), catatui.Fill(1), catatui.Fill(1),
	).Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(catatui.NewLine(
		catatui.NewStyledSpan("Sparkline Widget",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered(), rows[0])

	renderSparkline(f, rows[1])
	renderSineWave(f, rows[2], f.Count())
}

// renderSparkline draws a repeating saw wave, left to right.
func renderSparkline(f *catatui.Frame, area catatui.Rect) {
	pattern := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	data := make([]uint64, 0, len(pattern)*int(area.Width))
	for range int(area.Width) {
		data = append(data, pattern...)
	}

	sparkline := widgets.NewSparkline().
		Data(data...).
		Max(10).
		Direction(widgets.RenderLeftToRight).
		Style(catatui.NewStyle().Fg(catatui.ColorCyan))
	f.RenderWidget(sparkline, area)
}

// renderSineWave draws a sine wave whose phase follows the frame count, so it
// scrolls as the frames go by.
func renderSineWave(f *catatui.Frame, area catatui.Rect, frameCount uint64) {
	phaseShift := float64(frameCount) * 0.2
	data := make([]uint64, area.Width)
	for i := range data {
		angle := float64(i)*0.5 + phaseShift
		data[i] = uint64(math.Round((math.Sin(angle)*3 + 3) * 10))
	}

	sparkline := widgets.NewSparkline().
		Data(data...).
		Max(100).
		Direction(widgets.RenderRightToLeft).
		Style(catatui.NewStyle().Fg(catatui.ColorMagenta).Bg(catatui.ColorBlack)).
		AbsentValueStyle(catatui.NewStyle().Fg(catatui.ColorRed)).
		AbsentValueSymbol(symbols.ShadeFull)
	f.RenderWidget(sparkline, area)
}
