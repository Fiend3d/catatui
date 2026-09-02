//go:build !windows

package catatui

import "testing"

// TestClusterWidthIsUnicodeWidth pins the width policy everywhere but Windows:
// a cluster is as wide as Unicode says, with the halfwidth katakana sound marks
// corrected back to a column.
//
// The Windows console has a policy of its own, measured in
// cellwidth_windows_test.go, because it advances per code point rather than per
// cluster. Where the two tables cover the same string they are expected to
// differ; that is the point of splitting them.
func TestClusterWidthIsUnicodeWidth(t *testing.T) {
	for _, c := range []struct {
		text string
		want int
	}{
		// A spacing combining mark is scored a column of its own, so a
		// consonant carrying one measures two.
		{"हि", 2},
		{"हिन्दी", 5}, // हि(2) + न्(1) + दी(2)
		{"தமிழ்", 4},
		{"বাংলা", 3},

		// A non-spacing mark is scored nothing, so these stay at their base.
		{"á", 1},
		{"á", 1},
		{"ก็", 1},

		// CJK, Hangul and emoji sequences are one cluster of their own width.
		{"日", 2},
		{"日本語", 6},
		{"한글", 4},
		{"ｱ", 1},
		{"ア", 2},
		{"ﾊﾞ", 2}, // halfwidth katakana plus a sound mark, drawn in two columns
		{"🇯🇵", 2},
		{"🏳️‍🌈", 2},
		{"👨‍👩‍👧‍👦", 2},

		// Clusters the Buffer refuses to draw measure nothing.
		{"​", 0},
		{"a​b", 2},
	} {
		if got := StringWidth(c.text); got != c.want {
			t.Errorf("StringWidth(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}

// TestCombiningKanaMarksAreZeroWidth is ratatui's own expectation: the true
// combining sound marks get no special handling, unlike their halfwidth
// namesakes.
func TestCombiningKanaMarksAreZeroWidth(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		want uint16
	}{
		{"combining dakuten on halfwidth", "ｶ゙", 1},
		{"combining dakuten on fullwidth", "ガ", 2},
		{"combining handakuten on halfwidth", "ﾊ゚", 1},
		{"combining handakuten on fullwidth", "パ", 2},
	} {
		if got := cellWidth(c.in); got != c.want {
			t.Errorf("%s: cellWidth(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}
