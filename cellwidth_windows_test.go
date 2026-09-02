package catatui

import "testing"

// TestClusterWidthIsThePerCodePointAdvance pins the Windows width policy
// against the console itself.
//
// Every number here was measured, not derived: a probe printed the string with
// WriteConsoleW and read the cursor back with GetConsoleScreenBufferInfo, on
// conhost under Windows 11 build 26100. That is the same text buffer a
// ConPTY-hosted terminal drives, so it is what decides which column the next
// cell lands in.
//
// The shape of it: the console advances per code point, not per cluster, and
// never by less than one column. So a consonant carrying a spacing vowel sign
// is two columns, a conjunct with a virama is three, and no amount of shaping
// in the renderer changes the arithmetic. Measuring these as the single glyph
// they are drawn as is what made the next cell land inside them and destroy
// them.
func TestClusterWidthIsThePerCodePointAdvance(t *testing.T) {
	for _, c := range []struct {
		text string
		want int
	}{
		// Devanagari. Two code points, two columns; six, six.
		{"हि", 2},     // consonant plus vowel sign I
		{"पा", 2},     // consonant plus vowel sign AA
		{"री", 2},     // consonant plus vowel sign II
		{"हिन्दी", 6}, // हि + न् + दी, each of them a pair
		{"क्ष", 3},    // conjunct: consonant, virama, consonant
		{"श्री", 4},
		{"हिन्दी परीक्षण पाठ।", 19},
		{"भारत एक महान देश है।", 20},

		// Tamil, Bengali, Telugu.
		{"தமிழ்", 5},
		{"தமிழ் சோதனை", 11},
		{"বাং", 3},
		{"বাংলা", 5},
		{"পরীক্ষা", 7},
		{"తెలుగు", 6},
		{"నమస్కారం", 8},

		// Thai and Arabic, which are shaped the same way.
		{"ก็", 2},
		{"กำ", 2},
		{"العربية", 7},
		{"مرحبا", 5},

		// Latin: precomposed is one column, decomposed is two, because the
		// console gives the combining mark a cell of its own.
		{"á", 1},
		{"á", 2},

		// CJK and Hangul are unaffected: one code point, its own width.
		{"日", 2},
		{"日本語", 6},
		{"한", 2},
		{"ｱｲｳ", 3},
		{"각", 4}, // decomposed 각: three jamo, four columns

		// Halfwidth katakana with a sound mark: two code points, two columns,
		// which is what the sound-mark correction used to buy by hand.
		{"ﾊﾞ", 2},

		// Emoji sequences are code points like any other to the console.
		{"🇯🇵", 4},
		{"🏳️‍🌈", 5},
		{"👨‍👩‍👧‍👦", 11},

		// Nothing the Buffer refuses to draw is measured, so the floor of one
		// column never resurrects an invisible cluster.
		{"​", 0}, // zero-width space
		{"‍", 0}, // zero-width joiner on its own
		{"a​b", 2},
	} {
		if got := StringWidth(c.text); got != c.want {
			t.Errorf("StringWidth(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}

// TestCombiningKanaMarksTakeACellOfTheirOwn is the same rule applied to the
// cases ratatui's own tests cover, where Unicode says zero and the console says
// otherwise.
func TestCombiningKanaMarksTakeACellOfTheirOwn(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		want uint16
	}{
		{"combining dakuten on halfwidth", "ｶ゙", 2},
		{"combining dakuten on fullwidth", "ガ", 3},
		{"combining handakuten on halfwidth", "ﾊ゚", 2},
		{"combining handakuten on fullwidth", "パ", 3},
	} {
		if got := cellWidth(c.in); got != c.want {
			t.Errorf("%s: cellWidth(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

// TestRuneWidthDoesNotAllocate keeps the per-code-point walk off the heap. It
// runs once per cluster of every non-ASCII row drawn, so an allocation here is
// one per glyph per frame.
func TestRuneWidthDoesNotAllocate(t *testing.T) {
	s := "हिन्दी परीक्षण पाठ। தமிழ் சோதனை"
	got := testing.AllocsPerRun(100, func() {
		var w int
		for g := range AllGraphemes(s) {
			w += int(g.Width)
		}
		_ = w
	})
	if got != 0 {
		t.Errorf("measuring Indic text allocated %.1f times per run, want 0", got)
	}
}
