// Command canvas shows catatui's Canvas widget: a world map with shapes and
// points drawn over it.
//
//	go run ./examples/widgets/canvas
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/canvas.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

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

// run draws the UI and waits for a key.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

// render draws the title and the canvas below it.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())

	f.RenderWidget(title("Canvas Widget"), rows[0])
	renderCanvas(f, rows[1])
}

// renderCanvas draws a world map, then a second layer of shapes over it.
func renderCanvas(f *catatui.Frame, area catatui.Rect) {
	canvas := widgets.NewCanvas().
		XBounds([2]float64{-180, 180}).
		YBounds([2]float64{-90, 90}).
		Marker(symbols.Braille).
		Paint(func(ctx *widgets.Context) {
			ctx.Draw(widgets.Map{
				Resolution: widgets.MapResolutionHigh,
				Color:      catatui.ColorWhite,
			})
			// A new layer starts a fresh grid, so what follows is drawn on
			// top of the map rather than merged into it.
			ctx.Layer()
			ctx.Draw(widgets.NewCanvasLine(0, 10, 10, 10, catatui.ColorBlue))
			ctx.Draw(widgets.Rectangle{
				X: 10, Y: 20, Width: 10, Height: 10,
				Color: catatui.ColorGreen,
			})
			ctx.Draw(widgets.Points{
				Coords: [][2]float64{
					{2.3522, 48.8566},    // Paris
					{-122.3321, 47.6062}, // Seattle
					{-79.3837, 43.6511},  // Toronto
					{32.8597, 39.9334},   // Ankara
				},
				Color: catatui.ColorRed,
			})
		})

	f.RenderWidget(canvas, area)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
