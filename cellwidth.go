// Port of ratatui-core/src/buffer/cell_width.rs @ ratatui-v0.30.2

package catatui

import (
	"iter"
	"strings"

	"github.com/rivo/uniseg"
)

const (
	halfwidthKatakanaVoicedSoundMark     = 'ﾞ' // ﾞ
	halfwidthKatakanaSemiVoicedSoundMark = 'ﾟ' // ﾟ
)

// countHalfwidthSoundMarks counts the halfwidth katakana voiced and
// semi-voiced sound marks in a cluster.
func countHalfwidthSoundMarks(s string) uint16 {
	var n uint16
	for _, r := range s {
		if r == halfwidthKatakanaVoicedSoundMark || r == halfwidthKatakanaSemiVoicedSoundMark {
			n = SatAdd(n, 1)
		}
	}
	return n
}

// cellWidth is the number of terminal columns a grapheme cluster occupies.
//
// Every width decision in the library goes through here; having more than one
// notion of width is what makes rows drift out of alignment.
func cellWidth(s string) uint16 {
	// A single byte is ASCII, and control characters are filtered out before
	// they reach here, so it is always exactly one column.
	if len(s) == 1 {
		return 1
	}
	return clusterWidth(s, uniseg.StringWidth(s))
}

// clusterWidth applies the width policy to one grapheme cluster, given the
// width uniseg measured for it while segmenting.
//
// The number wanted here is not a property of the text but of the terminal: it
// is how far the cursor moves once the cluster has been printed. Getting it
// wrong in either direction corrupts the frame. Too few columns and the next
// cell is written on top of a glyph the terminal has already laid down, taking
// part of the glyph with it. Too many and the extra columns are never written
// at all — a wide cell covers its own trailing columns, so the diff skips them
// — and whatever the previous frame left there stays on screen.
//
// A terminal that segments text into grapheme clusters, which is what Windows
// Terminal, kitty, WezTerm and VTE do, advances by the cluster's unicode width,
// and that is what uniseg reports: a spacing combining mark adds a column, a
// non-spacing one adds none, and a sequence held together by joiners counts
// once. So the measured width is taken as it comes.
//
// The halfwidth katakana sound marks are the single correction: terminals
// render them in a column of their own even though Unicode scores them as
// zero-width combining marks.
//
// Two things this must not do, both learned against Windows Terminal by
// watching the screen. It must not cap a cluster at its widest rune on the
// theory that one glyph costs one advance: हि is drawn as a single glyph and
// still costs two columns, and measuring it as one wiped out every consonant
// carrying a spacing vowel sign, turning हिन्दी into न्दी. And it must not sum
// the code points the way the legacy console buffer does: that scores a
// zero-width joiner and a combining accent a column each, so a four-person
// family measures eleven where the terminal draws two, and the nine columns
// nobody then writes keep the previous frame's text.
//
// The console API is no guide here. GetConsoleScreenBufferInfo reports the
// per-code-point figure — 11 for that family — even under Windows Terminal,
// whose screen shows 2. The terminal's own buffer is what the user sees, and
// the only way to measure it is to look.
func clusterWidth(cluster string, unisegWidth int) uint16 {
	w := uint16(min(max(unisegWidth, 0), maxU16))
	return SatAdd(w, countHalfwidthSoundMarks(cluster))
}

// GraphemeWidth is the number of terminal columns one grapheme cluster
// occupies, given the width uniseg reported for it while segmenting —
// the third return value of uniseg.FirstGraphemeClusterInString.
//
// It exists for callers that do their own segmentation pass and need each
// cluster's width without paying to segment it a second time, which is what
// cellWidth would cost them. Measuring any other way is how a caller's columns
// come to disagree with the ones a Buffer draws.
func GraphemeWidth(cluster string, unisegWidth int) int {
	return int(clusterWidth(cluster, unisegWidth))
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
	for g := range AllGraphemes(s) {
		out = append(out, g)
	}
	return out
}

// AllGraphemes is Graphemes as an iterator, which costs no allocation at all.
//
// Prefer it on any path that runs per frame. Graphemes has to build a slice
// whose size is proportional to the string, and every drawn row would pay for
// one; the buffer's own drawing goes through here for that reason.
func AllGraphemes(s string) iter.Seq[Grapheme] {
	return func(yield func(Grapheme) bool) {
		state := -1
		for len(s) > 0 {
			var cluster string
			var w int
			cluster, s, w, state = uniseg.FirstGraphemeClusterInString(s, state)
			g, ok := drawable(cluster, w)
			if !ok {
				continue
			}
			if !yield(g) {
				return
			}
		}
	}
}

// drawable applies the same filter and width correction as cellWidth, given a
// cluster and the width uniseg already measured for it while segmenting.
//
// Reusing that width is what makes the iterator cheap: cellWidth would call
// uniseg.StringWidth, which segments the cluster a second time. For a single
// cluster the two produce the same number by construction, and
// TestGraphemeIterationAgreesWithCellWidth pins that down.
func drawable(cluster string, unisegWidth int) (Grapheme, bool) {
	// A single byte is ASCII, so it is one column unless it is a control.
	if len(cluster) == 1 {
		if isControlRune(rune(cluster[0])) {
			return Grapheme{}, false
		}
		return Grapheme{Symbol: cluster, Width: 1}, true
	}
	if containsControl(cluster) {
		return Grapheme{}, false
	}
	w := clusterWidth(cluster, unisegWidth)
	if w == 0 {
		return Grapheme{}, false
	}
	return Grapheme{Symbol: cluster, Width: w}, true
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

// StringWidth is the number of terminal columns s occupies when drawn.
//
// It counts each grapheme cluster once and skips the clusters a Buffer would
// not draw — control characters and zero-width clusters — so what it measures
// is exactly what SetString would put on screen. Use it for anything that has
// to reason about width outside a Buffer: sizing a status bar, deciding how
// much of a label fits, laying out a widget by hand.
//
// This is the library's only notion of width, deliberately. ratatui has two:
// Span::width, Line::width and Text::width call unicode-width directly, while
// Buffer::set_stringn measures with cell_width, and the two disagree on
// halfwidth katakana sound marks, control characters and some emoji sequences.
// A Line can therefore report a width it does not draw. Keeping one function
// is the whole point of the rule; several disagreeing width implementations
// are what made rows drift out of alignment in the program that prompted this
// port.
func StringWidth(s string) int {
	var w int
	for g := range AllGraphemes(s) {
		w += int(g.Width)
	}
	return w
}

// stringWidth is StringWidth saturated into a uint16, for the coordinate math.
func stringWidth(s string) uint16 {
	return uint16(min(StringWidth(s), maxU16))
}
