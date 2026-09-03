// Command release-header draws the banner that goes with a catatui release:
// the logo over a rainbow shadow, the version, and a menu of the packages.
//
//	go run ./examples/apps/release-header
//	go run ./examples/apps/release-header -version 0.2.0 -name Ragu
//
// Any key quits.
//
// It draws into a fixed 68x16 viewport rather than the whole terminal, so the
// banner comes out the same size wherever it is run and can be screenshotted
// straight into a README. The two menu blocks overlap by a row and merge their
// borders where they meet.
//
// ratatui hard-codes its version and release name and edits them each time;
// flags save editing the file.
//
// Port of examples/apps/release-header @ ratatui-v0.30.2
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// The menus: what you import to use catatui, and what comes with it.
var (
	mainDishes = []string{"> catatui", "> catatui/widgets", "> catatui/symbols"}
	pairings   = []string{"> catatui/term", "> palette/tailwind", "> palette/material"}
)

// The banner's own colours, which do not follow the terminal's palette: the
// point of it is to look the same everywhere.
var (
	fgColor         = catatui.Rgb(246, 214, 187) // #F6D6BB
	bgColor         = catatui.Rgb(20, 20, 50)    // #141432
	menuBorderColor = catatui.Rgb(255, 255, 160) // #FFFFA0
)

func main() {
	version := flag.String("version", "0.1.0", "the version to put under the logo")
	name := flag.String("name", "Ratatouille", "the release name to put beside it")
	flag.Parse()

	if err := run(*version, *name); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(version, name string) error {
	defer term.RecoverAndRestore()

	// A fixed viewport is a region of the terminal the size of the banner, so
	// the drawing does not stretch to whatever window it was run in.
	terminal, restore, err := term.Init(
		term.WithViewport(catatui.FixedViewport(catatui.NewRect(0, 0, 68, 16))))
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, version, name) }); err != nil {
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

// render paints the background and centres the logo and the menu on it.
func render(f *catatui.Frame, version, name string) {
	area := f.Area()
	f.Buffer().SetStyle(area, catatui.NewStyle().Fg(fgColor).Bg(bgColor))

	const (
		logoWidth = 29
		menuWidth = 23
		padding   = 2 // between the logo and the menu
		// Two bordered blocks overlapping by a row: two borders each, less the
		// one they share.
		menuBorders = 3
	)
	height := uint16(len(mainDishes)+len(pairings)) + menuBorders
	width := uint16(logoWidth + menuWidth + padding)

	center := centered(area, catatui.Length(width), catatui.Length(height))
	columns := catatui.HorizontalLayout(
		catatui.Length(logoWidth), catatui.Length(padding), catatui.Length(menuWidth),
	).Split(center)

	renderLogo(f, columns[0], version, name)
	renderMenu(f, columns[2])
}

// centered cuts a rect of the given size out of the middle of area, which is
// what ratatui's Rect::centered does: one centred layout per axis.
func centered(area catatui.Rect, horizontal, vertical catatui.Constraint) catatui.Rect {
	row := catatui.VerticalLayout(vertical).Flex(catatui.FlexCenter).Split(area)[0]
	return catatui.HorizontalLayout(horizontal).Flex(catatui.FlexCenter).Split(row)[0]
}

// renderLogo draws the logo, the glow behind it, and the version under it.
//
// The glow is the logo's top row drawn six times over, each row in a colour
// from a gradient: a block per letter sets the foreground, and the logo drawn
// over it keeps that colour, because the logo's own text carries no style.
func renderLogo(f *catatui.Frame, area catatui.Rect, version, name string) {
	area = area.Inner(catatui.Margin{Horizontal: 1})

	rows := catatui.VerticalLayout(
		catatui.Length(6), catatui.Length(2), catatui.Length(1),
	).Flex(catatui.FlexEnd).Split(area)
	shadowArea, logoArea, versionArea := rows[0], rows[1], rows[2]

	// The widths of the seven letters of the wordmark, which the small logo
	// draws in 27 columns.
	letters := []catatui.Constraint{
		catatui.Length(5), catatui.Length(4), catatui.Length(4), catatui.Length(4),
		catatui.Length(4), catatui.Length(5), catatui.Length(1),
	}

	for row, rowArea := range shadowArea.Rows() {
		for letter, letterArea := range catatui.HorizontalLayout(letters...).Split(rowArea) {
			color := gradientColor(rainbow(letter), row)
			f.RenderWidget(widgets.NewBlock().Style(catatui.NewStyle().Fg(color)), letterArea)
		}
		// One row is not two, so this draws the logo's top line only.
		f.RenderWidget(widgets.SmallCatatuiLogo(), rowArea)
	}

	f.RenderWidget(widgets.NewBlock().Style(catatui.NewStyle().Fg(fgColor)), logoArea)
	f.RenderWidget(widgets.SmallCatatuiLogo(), logoArea)
	f.RenderWidget(
		catatui.LineFromStyledString(fmt.Sprintf("v%s %q", version, name),
			catatui.NewStyle().AddModifier(catatui.ModifierDim)),
		versionArea)
}

// renderMenu draws the two blocks, overlapping by a row so that the border
// between them is drawn once rather than twice.
func renderMenu(f *catatui.Frame, area catatui.Rect) {
	rows := catatui.VerticalLayout(
		catatui.Length(uint16(len(mainDishes))+2),
		catatui.Length(uint16(len(pairings))+2),
	).Spacing(catatui.Overlap(1)).Split(area)

	renderMenuBlock(f, rows[0], "Main Courses", mainDishes)
	renderMenuBlock(f, rows[1], "Pairings", pairings)
}

func renderMenuBlock(f *catatui.Frame, area catatui.Rect, title string, items []string) {
	// Fuzzy merging is what turns the two rounded corners meeting in the
	// middle into a single tee.
	block := widgets.Bordered().
		BorderType(widgets.BorderRounded).
		BorderStyle(catatui.NewStyle().Fg(menuBorderColor)).
		Padding(widgets.HorizontalPadding(1)).
		MergeBorders(symbols.MergeFuzzy).
		Title(title)

	lines := make([]catatui.Line, len(items))
	for i, item := range items {
		lines[i] = catatui.LineFromString(item)
	}
	f.RenderWidget(
		widgets.NewParagraphFromText(catatui.NewText(lines...)).Block(block),
		area)
}

// rainbow is which colour of the seven a letter takes.
type rainbowColor int

const (
	red rainbowColor = iota
	orange
	yellow
	green
	blue
	indigo
	violet
)

// rainbow maps a letter's position to its colour, so the wordmark reads
// ROYGBIV from left to right.
func rainbow(letter int) rainbowColor { return rainbowColor(letter % 7) }

// The gradients the glow is mixed from, one entry per row: dimmest at the top,
// brightest at the bottom, where the logo sits.
var (
	redGradient     = [6]int{41, 43, 50, 68, 104, 156}
	greenGradient   = [6]int{24, 30, 41, 65, 105, 168}
	blueGradient    = [6]int{55, 57, 62, 78, 113, 166}
	ambientGradient = [6]int{17, 18, 20, 25, 40, 60}
)

// gradientColor mixes one letter's colour for one row of the glow.
func gradientColor(c rainbowColor, row int) catatui.Color {
	if row < 0 || row >= len(ambientGradient) {
		return bgColor
	}
	ambient := ambientGradient[row]
	r, g, b := redGradient[row], greenGradient[row], blueGradient[row]
	// The blue saturates towards the top, which is what gives the glow its
	// twilight cast rather than leaving it grey.
	blueSat := min(ambient*(6-row), 255)

	switch c {
	case red:
		return catatui.Rgb(u8(r), u8(ambient), u8(blueSat))
	case orange:
		return catatui.Rgb(u8(r), u8(g/2), u8(blueSat))
	case yellow:
		return catatui.Rgb(u8(r), u8(g), u8(blueSat))
	case green:
		return catatui.Rgb(u8(ambient), u8(g), u8(blueSat))
	case blue:
		return catatui.Rgb(u8(ambient), u8(ambient), u8(max(b, blueSat)))
	case indigo:
		return catatui.Rgb(u8(b), u8(ambient), u8(max(b, blueSat)))
	default:
		return catatui.Rgb(u8(r), u8(ambient), u8(max(b, blueSat)))
	}
}

// u8 clamps a channel into the byte a colour is made of.
func u8(v int) uint8 { return uint8(min(max(v, 0), 255)) }
