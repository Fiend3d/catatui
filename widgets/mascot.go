// Port of ratatui-widgets/src/mascot.rs @ ratatui-v0.30.2

package widgets

import (
	"strings"

	"github.com/Fiend3d/catatui"
)

// ratatuiMascot is the mascot drawn two art rows per cell row, so each cell
// combines an upper and a lower half. The letters are placeholders that pick a
// color: h is the hat, e is the eye, and the shade characters are the terminal.
const ratatuiMascot = `               hhh
             hhhhhh
            hhhhhhh
           hhhhhhhh
          hhhhhhhhh
         hhhhhhhhhh
        hhhhhhhhhhhh
        hhhhhhhhhhhhh
        hhhhhhhhhhhhh     ██████
         hhhhhhhhhhh    ████████
              hhhhh ███████████
               hhh ██ee████████
                h █████████████
            ████ █████████████
           █████████████████
           ████████████████
           ████████████████
            ███ ██████████
          ▒▒    █████████
         ▒░░▒   █████████
        ▒░░░░▒ ██████████
       ▒░░▓░░░▒ █████████
      ▒░░▓▓░░░░▒ ████████
     ▒░░░░░░░░░░▒ ██████████
    ▒░░░░░░░░░░░░▒ ██████████
   ▒░░░░░░░▓▓░░░░░▒ █████████
  ▒░░░░░░░░░▓▓░░░░░▒ ████  ███
 ▒░░░░░░░░░░░░░░░░░░▒ ██   ███
▒░░░░░░░░░░░░░░░░░░░░▒ █   ███
▒░░░░░░░░░░░░░░░░░░░░░▒   ███
 ▒░░░░░░░░░░░░░░░░░░░░░▒ ███
  ▒░░░░░░░░░░░░░░░░░░░░░▒ █`

// The characters of the mascot art and what they stand for.
const (
	mascotEmpty      = ' '
	mascotRat        = '█'
	mascotHat        = 'h'
	mascotEye        = 'e'
	mascotTerm       = '░'
	mascotTermBorder = '▒'
	mascotTermCursor = '▓'
)

// MascotEyeColor is the state of the mascot's eye.
type MascotEyeColor uint8

const (
	// MascotEyeDefault is the default eye color.
	MascotEyeDefault MascotEyeColor = iota
	// MascotEyeRed is the red eye color, used for blinking.
	MascotEyeRed
)

// RatatuiMascot draws the Ratatui mascot: a rat in a hat sitting on a
// terminal. It takes 32x16 cells and is drawn with half block characters.
//
//	f.RenderWidget(widgets.NewRatatuiMascot().SetEye(widgets.MascotEyeRed), area)
type RatatuiMascot struct {
	eyeState MascotEyeColor
	// ratColor is the color of the rat.
	ratColor catatui.Color
	// ratEyeColor is the color of the rat's eye.
	ratEyeColor catatui.Color
	// ratEyeBlink is the color of the rat's eye when blinking.
	ratEyeBlink catatui.Color
	// hatColor is the color of the rat's hat.
	hatColor catatui.Color
	// termColor is the color of the terminal.
	termColor catatui.Color
	// termBorderColor is the color of the terminal border.
	termBorderColor catatui.Color
	// termCursorColor is the color of the terminal cursor.
	termCursorColor catatui.Color
}

// NewRatatuiMascot returns the mascot in its default colors with the eye open.
func NewRatatuiMascot() RatatuiMascot {
	return RatatuiMascot{
		ratColor:        catatui.Indexed(252), // light_gray #d0d0d0
		hatColor:        catatui.Indexed(231), // white #ffffff
		ratEyeColor:     catatui.Indexed(236), // dark_charcoal #303030
		ratEyeBlink:     catatui.Indexed(196), // red #ff0000
		termColor:       catatui.Indexed(232), // vampire_black #080808
		termBorderColor: catatui.Indexed(237), // gray  #808080
		termCursorColor: catatui.Indexed(248), // dark_gray #a8a8a8
		eyeState:        MascotEyeDefault,
	}
}

// SetEye returns a copy of m with the eye state set (open or blinking).
func (m RatatuiMascot) SetEye(eye MascotEyeColor) RatatuiMascot { m.eyeState = eye; return m }

// GetEye returns the eye state.
func (m RatatuiMascot) GetEye() MascotEyeColor { return m.eyeState }

// colorFor returns the color an art character stands for, or the unset color
// for characters that have none.
func (m RatatuiMascot) colorFor(c rune) catatui.Color {
	switch c {
	case mascotRat:
		return m.ratColor
	case mascotHat:
		return m.hatColor
	case mascotEye:
		if m.eyeState == MascotEyeRed {
			return m.ratEyeBlink
		}
		return m.ratEyeColor
	case mascotTerm:
		return m.termColor
	case mascotTermCursor:
		return m.termCursorColor
	case mascotTermBorder:
		return m.termBorderColor
	default:
		return catatui.Color{}
	}
}

// Render draws the mascot using half block characters.
//
// The colors are hard-coded in the widget; the eye color depends on whether it
// is open or blinking.
func (m RatatuiMascot) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}

	lines := strings.Split(ratatuiMascot, "\n")
	for y := 0; y+1 < len(lines); y += 2 {
		line1, line2 := []rune(lines[y]), []rune(lines[y+1])
		for x := 0; x < len(line1) && x < len(line2); x++ {
			ch1, ch2 := line1[x], line2[x]
			cx := catatui.SatAdd(area.Left(), uint16(x))
			cy := catatui.SatAdd(area.Top(), uint16(y/2))

			// Check if coordinates are within the buffer area
			if cx >= area.Right() || cy >= area.Bottom() {
				continue
			}

			cell := buf.Get(cx, cy)
			// Given two cells which make up the top and bottom of the
			// character, the foreground color should be the non-space,
			// non-terminal one.
			var fg, bg catatui.Color
			switch {
			case ch1 == mascotEmpty && ch2 == mascotEmpty:
			case ch2 == mascotEmpty:
				fg = m.colorFor(ch1)
			case ch1 == mascotEmpty:
				fg = m.colorFor(ch2)
			case ch1 == mascotTerm && ch2 == mascotTermBorder:
				fg, bg = m.colorFor(mascotTermBorder), m.colorFor(mascotTerm)
			case ch1 == mascotTerm:
				fg, bg = m.colorFor(ch2), m.colorFor(mascotTerm)
			case ch2 == mascotTerm:
				fg, bg = m.colorFor(ch1), m.colorFor(mascotTerm)
			default:
				fg, bg = m.colorFor(ch1), m.colorFor(ch2)
			}
			// The symbol makes the empty space or terminal background the
			// empty part of the block.
			var symbol rune
			switch {
			case ch1 == mascotEmpty && ch2 == mascotEmpty:
			case ch1 == mascotTerm && ch2 == mascotTerm:
				symbol = mascotEmpty
			case ch2 == mascotEmpty || ch2 == mascotTerm:
				symbol = '▀'
			case ch1 == mascotEmpty || ch1 == mascotTerm:
				symbol = '▄'
			case ch1 == ch2:
				symbol = '█'
			default:
				symbol = '▀'
			}
			if fg.IsSet() {
				cell.Fg = fg
			}
			if bg.IsSet() {
				cell.Bg = bg
			}
			if symbol != 0 {
				cell.SetChar(symbol)
			}
		}
	}
}

var _ catatui.Widget = RatatuiMascot{}
