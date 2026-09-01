// Port of ratatui-core/src/buffer/cell_width.rs @ ratatui-v0.30.2

package catatui

import (
	"strings"

	"github.com/rivo/uniseg"
)

const (
	halfwidthKatakanaVoicedSoundMark     = 'ﾞ' // ﾞ
	halfwidthKatakanaSemiVoicedSoundMark = 'ﾟ' // ﾟ
)

// cellWidth is the number of terminal columns a grapheme cluster occupies.
//
// This is not plain unicode width. Terminals render the halfwidth katakana
// voiced and semi-voiced sound marks in their own column even though Unicode
// scores them as zero-width combining marks, so each one gets a column added
// back. Every width decision in the library goes through this one function;
// having more than one notion of width is what makes rows drift out of
// alignment.
func cellWidth(s string) uint16 {
	// A single byte is ASCII, and control characters are filtered out before
	// they reach here, so it is always exactly one column.
	if len(s) == 1 {
		return 1
	}
	return SatAdd(uint16(min(uniseg.StringWidth(s), maxU16)), countHalfwidthSoundMarks(s))
}

func countHalfwidthSoundMarks(s string) uint16 {
	var n uint16
	for _, r := range s {
		if r == halfwidthKatakanaVoicedSoundMark || r == halfwidthKatakanaSemiVoicedSoundMark {
			n = SatAdd(n, 1)
		}
	}
	return n
}

// Grapheme is one extended grapheme cluster together with the number of
// terminal columns it occupies.
type Grapheme struct {
	Symbol string
	Width  uint16
}

// Graphemes splits s into extended grapheme clusters paired with their cell
// width, dropping clusters that contain a control character and clusters that
// occupy no columns — exactly the filtering a Buffer applies before drawing.
//
// Widgets that lay text out themselves should measure with this rather than
// with any other width function, so that what they measure is what a Buffer
// will draw.
func Graphemes(s string) []Grapheme {
	out := make([]Grapheme, 0, len(s))
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		c := g.Str()
		if containsControl(c) {
			continue
		}
		w := cellWidth(c)
		if w == 0 {
			continue
		}
		out = append(out, Grapheme{Symbol: c, Width: w})
	}
	return out
}

// containsControl reports whether s holds a Unicode control character. It
// matches Rust's char::is_control, which is the Cc category: U+0000..U+001F,
// U+007F, and U+0080..U+009F.
func containsControl(s string) bool {
	return strings.ContainsFunc(s, isControlRune)
}

func isControlRune(r rune) bool {
	return r < 0x20 || (r >= 0x7F && r <= 0x9F)
}

// displayWidth is the total number of columns s occupies when drawn, counting
// each grapheme cluster once and skipping the clusters a Buffer would not draw.
//
// Deviation from ratatui: Span::width, Line::width and Text::width call
// unicode-width directly, while Buffer::set_stringn measures with cell_width.
// The two disagree on halfwidth katakana sound marks, control characters and
// some emoji sequences, so in ratatui a Line can report a width it does not
// actually draw. catatui measures everything with this one function instead.
// Keeping a single notion of width is the whole point of the rule; koneko's
// four disagreeing width implementations are what made rows drift.
func displayWidth(s string) int {
	var w int
	for _, g := range Graphemes(s) {
		w += int(g.Width)
	}
	return w
}

// stringWidth is displayWidth saturated into a uint16, for the coordinate math.
func stringWidth(s string) uint16 {
	return uint16(min(displayWidth(s), maxU16))
}
