package catatui

import (
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// clusterWidth is how far the Windows console advances after printing one
// grapheme cluster.
//
// The console has no notion of a grapheme cluster. Its buffer holds one code
// point per cell and it advances by every one of them, with a floor of one
// column: a mark that Unicode scores zero-width still takes a cell of its own.
// So a cluster costs the sum of its code points' widths, each floored at one,
// and neither uniseg's cluster width nor plain unicode width predicts it.
//
// This is not a preference. Every Indic script the program renders is written
// as a consonant carrying spacing and non-spacing marks, so almost every
// cluster in a line of Devanagari, Bengali, Tamil or Telugu is more than one
// code point. Measuring such a cluster as the one column it is drawn in leaves
// the console's remaining columns unaccounted for, the next cell is written
// into the middle of a glyph the console has already laid down, and the glyph
// is destroyed: हिन्दी came out as न्दी, भारत as रत. The console's own
// arithmetic is the only arithmetic that matters here.
//
// Measured against conhost on Windows 11 build 26100 by printing a string and
// reading the cursor back with GetConsoleScreenBufferInfo. Of a 17k-code-point
// sweep, this rule reproduced the console exactly everywhere except a handful
// of ranges the console scores two columns and uniseg scores one — the emoji
// modifiers U+1F3FB..U+1F3FF, the enclosed alphanumerics U+1F100..U+1F1AC, and
// the combining kana sound marks U+3099..U+309A among them. Text in those is
// still measured short, and a neighbouring glyph can still eat it; complex
// scripts, which are what a text viewer meets in the wild, are exact.
func clusterWidth(cluster string, unisegWidth int) uint16 {
	// A cluster uniseg scores zero-width is invisible: a lone combining mark, a
	// zero-width space, a stray joiner. The Buffer drops those rather than
	// drawing them, so the console never sees them and never advances for them,
	// and the floor below must not resurrect one. The halfwidth katakana sound
	// marks are the exception both policies make: a terminal does draw those.
	if unisegWidth <= 0 && countHalfwidthSoundMarks(cluster) == 0 {
		return 0
	}
	// A cluster of one code point is already measured: the floor is the only
	// correction it can need. This is the path CJK takes, which is most of the
	// non-ASCII a Buffer ever draws.
	if _, n := utf8.DecodeRuneInString(cluster); n == len(cluster) {
		return uint16(min(max(unisegWidth, 1), maxU16))
	}
	var w uint16
	for _, r := range cluster {
		w = SatAdd(w, runeWidth(r))
	}
	return w
}

// runeWidth is the columns the console advances for a single code point, which
// is its unicode width but never less than one.
func runeWidth(r rune) uint16 {
	// Control characters are filtered out before a cluster reaches here, so
	// every ASCII byte left is one column.
	if r < utf8.RuneSelf {
		return 1
	}
	// Measuring the rune on its own, rather than asking uniseg for the width of
	// the cluster it sits in, is the whole point: uniseg applies the emoji and
	// combining rules that the console does not have. The array does not escape
	// StringWidth, so the conversion costs no allocation.
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return uint16(min(max(uniseg.StringWidth(string(buf[:n])), 1), maxU16))
}
