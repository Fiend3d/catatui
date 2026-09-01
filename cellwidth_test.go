// Tests ported from ratatui-core/src/buffer/cell_width.rs @ ratatui-v0.30.2
//
// These double as a conformance check that Go's uniseg agrees with Rust's
// unicode-width crate on the cases ratatui depends on. If a Go dependency bump
// ever changes a width, this is where it surfaces.

package catatui

import "testing"

func TestCellWidth(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want uint16
	}{
		{"empty", "", 0},
		{"ascii", "a", 1},
		{"wide char", "あ", 2},

		// The sound marks alone occupy a column.
		{"halfwidth dakuten alone", "ﾞ", 1},
		{"halfwidth handakuten alone", "ﾟ", 1},

		// Halfwidth katakana plus a non-combining sound mark: two columns.
		{"halfwidth katakana with dakuten", "ｶﾞ", 2},      // U+FF76 + U+FF9E
		{"halfwidth katakana with dakuten 2", "ｻﾞ", 2},    // U+FF7B + U+FF9E
		{"halfwidth katakana with handakuten", "ﾊﾟ", 2},   // U+FF8A + U+FF9F
		{"halfwidth katakana with handakuten 2", "ﾋﾟ", 2}, // U+FF8B + U+FF9F

		// Linguistically wrong but valid clusters; the mark still takes a column.
		{"ascii with halfwidth dakuten", "aﾞ", 2},
		{"digit with halfwidth handakuten", "1ﾟ", 2},
		{"hiragana with halfwidth dakuten", "あﾞ", 3},
		{"kanji with halfwidth dakuten", "紅ﾞ", 3},

		// True combining marks get no special handling: they are zero-width.
		{"combining dakuten on halfwidth", "ｶ゙", 1}, // U+FF76 + U+3099
		{"combining dakuten on fullwidth", "ガ", 2}, // U+30AB + U+3099
		{"combining handakuten on halfwidth", "ﾊ゚", 1},
		{"combining handakuten on fullwidth", "パ", 2},

		// Mixed text is unchanged.
		{"halfwidth katakana", "ｶ", 1},
		{"fullwidth katakana", "カ", 2},
	}
	for _, c := range cases {
		if got := cellWidth(c.in); got != c.want {
			t.Errorf("%s: cellWidth(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

func TestStringWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"a", 1},
		{"あ", 2},
		{"ｶ", 1},
		{"カ", 2},
		{"aｶﾞb", 4}, // a(1) + ｶﾞ(2) + b(1)
		{"あｶﾞ", 4},  // あ(2) + ｶﾞ(2)
		{"hello", 5},
		{"", 0},
		// Control characters and zero-width clusters are never drawn, so they
		// occupy nothing.
		{"a\tb", 2},
		{"a\nb", 2},
		{"a​b", 2},
	}
	for _, c := range cases {
		if got := StringWidth(c.in); got != c.want {
			t.Errorf("StringWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestGraphemesFiltersControlAndZeroWidth(t *testing.T) {
	// Control characters are dropped entirely rather than drawn.
	got := Graphemes("a\tb\nc")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("graphemes dropped control chars incorrectly: got %d clusters %+v, want %d", len(got), got, len(want))
	}
	for i, g := range got {
		if g.Symbol != want[i] {
			t.Errorf("cluster %d = %q, want %q", i, g.Symbol, want[i])
		}
	}

	// A base character and its combining mark stay in one cluster of width 1.
	got = Graphemes("é")
	if len(got) != 1 {
		t.Fatalf("combining mark should stay in one cluster, got %+v", got)
	}
	if got[0].Symbol != "é" || got[0].Width != 1 {
		t.Errorf("cluster = %q width %d, want %q width 1", got[0].Symbol, got[0].Width, "é")
	}
}

func TestIsControlRune(t *testing.T) {
	for _, r := range []rune{0x00, 0x1F, 0x7F, 0x80, 0x9F} {
		if !isControlRune(r) {
			t.Errorf("U+%04X should be a control character", r)
		}
	}
	for _, r := range []rune{' ', 'a', 0xA0, 'あ'} {
		if isControlRune(r) {
			t.Errorf("U+%04X should not be a control character", r)
		}
	}
}

// TestStringWidthMatchesWhatIsDrawn is the contract that makes StringWidth
// usable for laying text out by hand: what it measures must be exactly what a
// Buffer puts on screen.
//
// Measuring with one function and drawing with another is what makes rows drift
// out of alignment, so this walks a spread of awkward strings and checks that
// the measured width equals the number of columns SetString actually fills.
func TestStringWidthMatchesWhatIsDrawn(t *testing.T) {
	inputs := []string{
		"",
		"hello",
		"あ",
		"日本語のテキスト",
		"ｶﾞ",           // halfwidth katakana plus a non-combining sound mark
		"ガ",            // fullwidth katakana with a combining mark
		"é",            // base plus combining acute
		"a​b",          // zero-width space in the middle
		"a\tb",         // control character, which is not drawn
		"mixed あ text", // narrow and wide together
	}
	for _, in := range inputs {
		want := StringWidth(in)

		// Draw into a buffer wide enough for anything, then find how far the
		// draw actually reached.
		buf := NewBuffer(NewRect(0, 0, 64, 1))
		nextX, _ := buf.SetString(0, 0, in, NewStyle())

		if got := int(nextX); got != want {
			t.Errorf("StringWidth(%q) = %d, but SetString filled %d columns", in, want, got)
		}
	}
}

// TestStringWidthAgreesWithSpanWidth guards against the two drifting apart, the
// way ratatui's Span::width and cell_width have.
func TestStringWidthAgreesWithSpanWidth(t *testing.T) {
	for _, in := range []string{"", "abc", "あい", "ｶﾞｻﾞ", "a​b", "é"} {
		if got, want := NewSpan(in).Width(), StringWidth(in); got != want {
			t.Errorf("Span(%q).Width() = %d but StringWidth = %d", in, got, want)
		}
		if got, want := LineFromString(in).Width(), StringWidth(in); got != want {
			t.Errorf("Line(%q).Width() = %d but StringWidth = %d", in, got, want)
		}
	}
}
