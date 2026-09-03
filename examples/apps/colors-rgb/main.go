// Command colors-rgb fills the terminal with every colour it can show, and
// animates them.
//
//	go run ./examples/apps/colors-rgb
//
// Any key quits. It needs a terminal with 24-bit colour; one limited to 256
// will show bands instead of a smooth sweep.
//
// Each row of characters is two rows of pixels, drawn with the upper half block
// ▀: the foreground paints the top half and the background the bottom, so a
// terminal cell holds two colours. The colours themselves come from Okhsv (see
// okhsv.go), which keeps a hue sweep at an even lightness.
//
// The two widgets here update themselves while they draw, on a pointer
// receiver: the fps counter to count a frame, the colour field to cache the
// grid it just worked out and to move it along by one column. That is the
// alternative to StatefulWidget for state only the widget cares about — see
// advanced-widget-impl, which is about the choice.
//
// Port of examples/apps/colors-rgb @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"time"

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

// run redraws about sixty times a second, which is what animates the colours,
// and quits on the first key.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	a := &app{fps: fpsWidget{lastInstant: time.Now()}}

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

// app is a title bar with the frame rate at the end of it, over the colours.
type app struct {
	fps    fpsWidget
	colors colorsWidget
}

func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Min(0)).Split(area)
	cols := catatui.HorizontalLayout(catatui.Min(0), catatui.Length(8)).Split(rows[0])

	catatui.LineFromString("colors-rgb example. Press any key to quit").
		Centered().
		Render(cols[0], buf)
	a.fps.Render(cols[1], buf)
	a.colors.Render(rows[1], buf)
}

// fpsWidget counts the frames it is drawn in and shows how many there were in
// the last second.
type fpsWidget struct {
	frameCount  int
	lastInstant time.Time
	fps         float64
	hasFPS      bool
}

func (w *fpsWidget) Render(area catatui.Rect, buf *catatui.Buffer) {
	w.tick(time.Now())
	if !w.hasFPS {
		return
	}
	catatui.NewSpan(fmt.Sprintf("%.1f fps", w.fps)).Render(area, buf)
}

// tick counts a frame, and works the rate out once a second.
//
// It waits for two frames as well as a second, so that a machine too slow to
// draw twice a second reports nothing rather than a number made of one sample.
func (w *fpsWidget) tick(now time.Time) {
	w.frameCount++
	elapsed := now.Sub(w.lastInstant)
	if elapsed > time.Second && w.frameCount > 2 {
		w.fps = float64(w.frameCount) / elapsed.Seconds()
		w.hasFPS = true
		w.frameCount = 0
		w.lastInstant = now
	}
}

// colorsWidget draws the colour field, and holds the grid it drew last so that
// it is not recomputed for every frame. Okhsv is not expensive, but it is a
// cube root and two transcendentals per pixel, and the answer only changes when
// the terminal is resized.
type colorsWidget struct {
	// colors is indexed [row][column] and is twice as tall as the area, since
	// each row of cells holds two rows of pixels.
	colors [][]catatui.Color

	// frameCount shifts the grid sideways, which is the whole animation.
	frameCount int
}

func (w *colorsWidget) Render(area catatui.Rect, buf *catatui.Buffer) {
	if area.IsEmpty() {
		return
	}
	w.setupColors(area.AsSize())

	width := int(area.Width)
	for xi := range width {
		x := area.X + uint16(xi)
		// Shifting the column index by the frame count is what moves the
		// colours; the grid itself never changes.
		shifted := (xi + w.frameCount) % width
		for yi := range int(area.Height) {
			y := area.Y + uint16(yi)
			buf.Get(x, y).
				SetSymbol(symbols.HalfBlockUpper).
				SetFg(w.colors[yi*2][shifted]).
				SetBg(w.colors[yi*2+1][shifted])
		}
	}
	w.frameCount++
}

// setupColors fills the grid, unless it already fits the size given.
//
// The hue runs across the terminal and the value up it, at the highest
// saturation Okhsv has, so the top row is the full sweep of the gamut boundary
// and it fades to black at the bottom.
func (w *colorsWidget) setupColors(size catatui.Size) {
	width := int(size.Width)
	height := int(size.Height) * 2 // two rows of pixels per row of cells
	if len(w.colors) == height && (height == 0 || len(w.colors[0]) == width) {
		return
	}

	w.colors = make([][]catatui.Color, height)
	for y := range height {
		row := make([]catatui.Color, width)
		for x := range width {
			hue := float64(x) * 360 / float64(width)
			value := float64(height-y) / float64(height)
			row[x] = catatui.Rgb(okhsvToRGB(hue, maxSaturationOkhsv, value))
		}
		w.colors[y] = row
	}
}

// maxSaturationOkhsv is the top of Okhsv's saturation axis, which is where
// every colour here is taken from.
const maxSaturationOkhsv = 1.0
