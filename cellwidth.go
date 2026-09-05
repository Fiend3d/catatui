package catatui

import (
	"iter"
	"runtime"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

// WidthPolicy selects how grapheme clusters advance the terminal cursor.
type WidthPolicy uint8

const (
	// UnicodeWidth uses uniseg's widths, with halfwidth sound marks counted.
	UnicodeWidth WidthPolicy = iota
	// WindowsTerminalWidth implements Windows Terminal's grapheme mode:
	// spacing marks consume cells and Unicode clusters advance at most two.
	WindowsTerminalWidth
)

// DefaultWidthPolicy is shared by cells, buffers, text and widgets. It defaults
// to WindowsTerminalWidth on Windows and UnicodeWidth elsewhere. Applications
// targeting another terminal may set it during startup, before measuring text
// or creating buffers. Do not change it while buffers or concurrent readers
// are in use; their existing column positions would become invalid.
var DefaultWidthPolicy = func() WidthPolicy {
	if runtime.GOOS == "windows" {
		return WindowsTerminalWidth
	}
	return UnicodeWidth
}()

const (
	halfwidthKatakanaVoicedSoundMark     = 'ﾞ'
	halfwidthKatakanaSemiVoicedSoundMark = 'ﾟ'
)

func countHalfwidthSoundMarks(s string) uint16 {
	var n uint16
	for _, r := range s {
		if r == halfwidthKatakanaVoicedSoundMark || r == halfwidthKatakanaSemiVoicedSoundMark {
			n = SatAdd(n, 1)
		}
	}
	return n
}

func (p WidthPolicy) capWidth(w uint16) uint16 {
	if p == WindowsTerminalWidth {
		return min(w, 2)
	}
	return w
}

// segmentWidth measures one uniseg segment before Indic joining. Grapheme
// break properties are not width properties: Bengali AA is Extend but takes
// one column in Windows Terminal. See its CodepointWidthDetector.cpp.
func (p WidthPolicy) segmentWidth(s string, unisegWidth int) uint16 {
	w := SatAdd(uint16(min(max(unisegWidth, 0), maxU16)), countHalfwidthSoundMarks(s))
	if p == WindowsTerminalWidth && w < 2 {
		for _, r := range s {
			if unicode.Is(unicode.Mc, r) && uniseg.StringWidth(string(r)) == 0 {
				w = SatAdd(w, 1)
				if r == 0x302E || r == 0x302F { // wide Hangul tone marks
					w = SatAdd(w, 1)
				}
			}
		}
	}
	return p.capWidth(w)
}

// cellWidth also accepts symbols containing multiple clusters. Do not cap
// such an entire string as one cluster.
func cellWidth(s string) uint16 {
	if len(s) == 1 && !isControlRune(rune(s[0])) {
		return 1
	}
	return stringWidth(s)
}

// GraphemeWidth measures a single uniseg segment with its stepper-provided
// width. For full Indic conjunct boundaries, use SegmentGraphemes instead.
func GraphemeWidth(cluster string, unisegWidth int) int {
	return int(DefaultWidthPolicy.segmentWidth(cluster, unisegWidth))
}

// Grapheme is a complete drawing unit and its terminal advance. Indic
// conjuncts and the Tamil ksha/sri ligatures are kept together for shaping.
type Grapheme struct {
	Symbol string
	Width  uint16
}

// SegmentGraphemes preserves every byte of s, including tabs, controls and
// zero-width clusters. Controls have width zero. Layout callers can expand
// tabs themselves and retain byte offsets by summing len(g.Symbol).
//
// Segmentation adds Unicode 15.1 GB9c to uniseg's boundaries and tailors Tamil
// ksha/sri ligatures. The same units are used by Buffer.SetStringn and the diff.
func SegmentGraphemes(s string) iter.Seq[Grapheme] {
	return DefaultWidthPolicy.SegmentGraphemes(s)
}

// SegmentGraphemes segments and measures using p without changing the default.
func (p WidthPolicy) SegmentGraphemes(s string) iter.Seq[Grapheme] {
	return func(yield func(Grapheme) bool) {
		state, start, end := -1, 0, 0
		var width uint16
		prev, rest := "", s
		for len(rest) > 0 {
			var g string
			var w int
			g, rest, w, state = uniseg.FirstGraphemeClusterInString(rest, state)
			indic := joinsConjunct(prev, g)
			tamil := joinsTamilLigature(prev, g)
			if end > start && !indic && !tamil {
				if !yield(Grapheme{Symbol: s[start:end], Width: width}) {
					return
				}
				start, width = end, 0
			}
			end += len(g)
			if !containsControl(g) {
				width = SatAdd(width, p.segmentWidth(g, w))
			}
			if indic {
				width = p.capWidth(width)
			}
			// Tamil tailoring combines multiple terminal clusters, so its
			// widths add without capping the whole tailored unit.
			prev = g
		}
		if end > start {
			yield(Grapheme{Symbol: s[start:end], Width: width})
		}
	}
}

// AllGraphemes yields only drawable clusters, without allocating a slice.
func AllGraphemes(s string) iter.Seq[Grapheme] {
	return func(yield func(Grapheme) bool) {
		for g := range SegmentGraphemes(s) {
			if g.Width > 0 && !yield(g) {
				return
			}
		}
	}
}

// Graphemes collects the drawable clusters used by Buffer.SetStringn.
func Graphemes(s string) []Grapheme {
	out := make([]Grapheme, 0, len(s))
	for g := range AllGraphemes(s) {
		out = append(out, g)
	}
	return out
}

func containsControl(s string) bool { return strings.ContainsFunc(s, isControlRune) }

func isControlRune(r rune) bool { return r < 0x20 || (r >= 0x7F && r <= 0x9F) }

// StringWidth is the terminal advance under DefaultWidthPolicy, matching
// cells, Buffer.SetStringn, spans and widgets. Controls occupy no columns.
func StringWidth(s string) int { return DefaultWidthPolicy.StringWidth(s) }

// StringWidth measures s under p without changing the default policy.
func (p WidthPolicy) StringWidth(s string) int {
	var w int
	for g := range p.SegmentGraphemes(s) {
		w += int(g.Width)
	}
	return w
}

func stringWidth(s string) uint16 { return uint16(min(StringWidth(s), maxU16)) }
