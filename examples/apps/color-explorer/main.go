// Command color-explorer draws every colour catatui can name or number, so you
// can see what your terminal actually makes of them.
//
//	go run ./examples/apps/color-explorer
//
// Any key quits.
//
// The named colours come out however the terminal's palette says; the 256
// indexed ones are fixed, apart from the first sixteen, which are the named
// ones again. Give it a tall window: the whole thing wants 49 rows.
//
// Port of examples/apps/color-explorer @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

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

// namedColors is every colour catatui has a name for, in the order the
// terminal numbers them.
var namedColors = []catatui.Color{
	catatui.ColorBlack,
	catatui.ColorRed,
	catatui.ColorGreen,
	catatui.ColorYellow,
	catatui.ColorBlue,
	catatui.ColorMagenta,
	catatui.ColorCyan,
	catatui.ColorGray,
	catatui.ColorDarkGray,
	catatui.ColorLightRed,
	catatui.ColorLightGreen,
	catatui.ColorLightYellow,
	catatui.ColorLightBlue,
	catatui.ColorLightMagenta,
	catatui.ColorLightCyan,
	catatui.ColorWhite,
}

// backdrops are the five backgrounds each set of named colours is shown
// against, and the five foregrounds they are written in.
var backdrops = []catatui.Color{
	catatui.ColorReset,
	catatui.ColorBlack,
	catatui.ColorDarkGray,
	catatui.ColorGray,
	catatui.ColorWhite,
}

func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(30), // the named colours, five backgrounds and five foregrounds
		catatui.Length(17), // 0-231 of the indexed ones
		catatui.Length(2),  // 232-255, the greys
	).Split(f.Area())

	renderNamedColors(f, rows[0])
	renderIndexedColors(f, rows[1])
	renderIndexedGrayscale(f, rows[2])
}

// renderNamedColors shows the sixteen named colours as text on each backdrop,
// and then as backdrops themselves.
func renderNamedColors(f *catatui.Frame, area catatui.Rect) {
	heights := make([]catatui.Constraint, 10)
	for i := range heights {
		heights[i] = catatui.Length(3)
	}
	rows := catatui.VerticalLayout(heights...).Split(area)

	for i, bg := range backdrops {
		renderForegroundNamedColors(f, bg, rows[i])
	}
	for i, fg := range backdrops {
		renderBackgroundNamedColors(f, fg, rows[len(backdrops)+i])
	}
}

// renderForegroundNamedColors writes each name in its own colour on the given
// background.
func renderForegroundNamedColors(f *catatui.Frame, bg catatui.Color, area catatui.Rect) {
	block := titleBlock(fmt.Sprintf("Foreground colors on %v background", bg))
	inner := block.Inner(area)
	f.RenderWidget(block, area)

	for i, cell := range namedColorCells(inner) {
		fg := namedColors[i]
		f.RenderWidget(
			widgets.NewParagraph(fg.String()).
				Style(catatui.NewStyle().Fg(fg).Bg(bg)),
			cell)
	}
}

// renderBackgroundNamedColors writes each name on its own colour in the given
// foreground.
func renderBackgroundNamedColors(f *catatui.Frame, fg catatui.Color, area catatui.Rect) {
	block := titleBlock(fmt.Sprintf("Background colors with %v foreground", fg))
	inner := block.Inner(area)
	f.RenderWidget(block, area)

	for i, cell := range namedColorCells(inner) {
		bg := namedColors[i]
		f.RenderWidget(
			widgets.NewParagraph(bg.String()).
				Style(catatui.NewStyle().Fg(fg).Bg(bg)),
			cell)
	}
}

// namedColorCells cuts an area into the sixteen cells one row of named colours
// needs: two rows of eight.
func namedColorCells(area catatui.Rect) []catatui.Rect {
	cells := make([]catatui.Rect, 0, len(namedColors))
	for _, row := range catatui.VerticalLayout(catatui.Length(1), catatui.Length(1)).Split(area) {
		cells = append(cells, catatui.HorizontalLayout(repeat(catatui.Ratio(1, 8), 8)...).Split(row)...)
	}
	return cells
}

// renderIndexedColors shows 0 to 231: the sixteen named ones along the top,
// then the 6x6x6 colour cube in blocks that keep its shape.
func renderIndexedColors(f *catatui.Frame, area catatui.Rect) {
	block := titleBlock("Indexed colors")
	inner := block.Inner(area)
	f.RenderWidget(block, area)

	rows := catatui.VerticalLayout(
		catatui.Length(1), // 0-15
		catatui.Length(1), // blank
		catatui.Min(6),    // 16-123
		catatui.Length(1), // blank
		catatui.Min(6),    // 124-231
		catatui.Length(1), // blank
	).Split(inner)

	widths := make([]catatui.Constraint, 16)
	for i := range widths {
		widths[i] = catatui.Length(5)
	}
	for i, cell := range catatui.HorizontalLayout(widths...).Split(rows[0]) {
		// Index 0 is black, which needs a lighter backdrop to be readable.
		bg := catatui.ColorBlack
		if i < 1 {
			bg = catatui.ColorDarkGray
		}
		color := catatui.Indexed(uint8(i))
		f.RenderWidget(swatch(fmt.Sprintf("%02d", i), color, bg, "██"), cell)
	}

	// The cube is 216 colours that read best as six blocks of six by six: two
	// bands of three columns, each column six rows of six.
	cells := cubeCells(rows[2], rows[4])
	for i := 16; i <= 231; i++ {
		color := catatui.Indexed(uint8(i))
		f.RenderWidget(
			swatch(fmt.Sprintf("%03d", i), color, catatui.ColorReset, "."),
			cells[i-16])
	}
}

// cubeCells lays the colour cube out over the two bands, in the order the
// indexes run.
func cubeCells(bands ...catatui.Rect) []catatui.Rect {
	columns := repeat(catatui.Length(27), 3)
	rows := repeat(catatui.Length(1), 6)
	entries := repeat(catatui.Min(4), 6)

	cells := make([]catatui.Rect, 0, 216)
	for _, band := range bands {
		for _, column := range catatui.HorizontalLayout(columns...).Split(band) {
			for _, row := range catatui.VerticalLayout(rows...).Split(column) {
				cells = append(cells, catatui.HorizontalLayout(entries...).Split(row)...)
			}
		}
	}
	return cells
}

// renderIndexedGrayscale shows 232 to 255, the greys at the end of the table.
func renderIndexedGrayscale(f *catatui.Frame, area catatui.Rect) {
	cells := make([]catatui.Rect, 0, 24)
	for _, row := range catatui.VerticalLayout(catatui.Length(1), catatui.Length(1)).Split(area) {
		cells = append(cells, catatui.HorizontalLayout(repeat(catatui.Length(6), 12)...).Split(row)...)
	}

	for i := 232; i <= 255; i++ {
		// The dark end of the ramp needs a light backdrop to be readable, and
		// the light end a dark one.
		bg := catatui.ColorBlack
		if i < 244 {
			bg = catatui.ColorGray
		}
		f.RenderWidget(
			swatch(fmt.Sprintf("%03d", i), catatui.Indexed(uint8(i)), bg, "██"),
			cells[i-232])
	}
}

// swatch is a label in the colour, followed by a block of the colour itself:
// the label says what a terminal makes of it as a foreground, the block as a
// background.
func swatch(label string, color, bg catatui.Color, block string) widgets.Paragraph {
	return widgets.NewParagraphFromText(catatui.NewText(catatui.NewLine(
		catatui.NewStyledSpan(label, catatui.NewStyle().Fg(color).Bg(bg)),
		catatui.NewStyledSpan(block, catatui.NewStyle().Fg(color).Bg(color)),
	)))
}

// titleBlock is the centred heading over each section, drawn on a rule.
func titleBlock(title string) widgets.Block {
	return widgets.NewBlock().
		Borders(widgets.BordersTop).
		TitleAlignment(catatui.AlignmentCenter).
		BorderStyle(catatui.NewStyle().Fg(catatui.ColorDarkGray)).
		TitleStyle(catatui.ResetStyle()).
		Title(title)
}

// repeat is n of the same constraint, which several of the grids here want.
func repeat(constraint catatui.Constraint, n int) []catatui.Constraint {
	constraints := make([]catatui.Constraint, n)
	for i := range constraints {
		constraints[i] = constraint
	}
	return constraints
}
