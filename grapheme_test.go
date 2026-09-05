package catatui

import (
	"slices"
	"strings"
	"testing"
)

func TestIndicWidthPolicies(t *testing.T) {
	for _, tc := range []struct {
		text             string
		unicode, windows int
		clusters         []string
	}{
		{"हिन्दी", 5, 4, []string{"हि", "न्दी"}},
		{"क्षि", 3, 2, []string{"क्षि"}},
		{"লা", 1, 2, []string{"লা"}},
		{"খা", 1, 2, []string{"খা"}},
		{"বাংলা", 3, 4, []string{"বাং", "লা"}},
		{"পरी", 3, 3, []string{"প", "री"}},
		{"க்ஷி", 3, 3, []string{"க்ஷி"}},
		{"ஸ்ரீ", 2, 2, []string{"ஸ்ரீ"}},
		{"ஶ்ரீ", 2, 2, []string{"ஶ்ரீ"}},
		{"கா", 1, 2, []string{"கா"}},
		{"👨‍👩‍👧‍👦", 2, 2, []string{"👨‍👩‍👧‍👦"}},
	} {
		for _, p := range []WidthPolicy{UnicodeWidth, WindowsTerminalWidth} {
			want := tc.unicode
			if p == WindowsTerminalWidth {
				want = tc.windows
			}
			if got := p.StringWidth(tc.text); got != want {
				t.Errorf("policy %d, %q: width %d, want %d", p, tc.text, got, want)
			}
			var symbols []string
			for g := range p.SegmentGraphemes(tc.text) {
				symbols = append(symbols, g.Symbol)
			}
			if !slices.Equal(symbols, tc.clusters) {
				t.Errorf("%q: clusters %q, want %q", tc.text, symbols, tc.clusters)
			}
		}
	}
}

func TestRawGraphemesPreserveOffsets(t *testing.T) {
	const input = "\tला\x00\r\n\u200bக்ஷி\tend"
	var reconstructed strings.Builder
	for g := range SegmentGraphemes(input) {
		reconstructed.WriteString(g.Symbol)
		if containsControl(g.Symbol) && g.Width != 0 {
			t.Errorf("control %q has width %d", g.Symbol, g.Width)
		}
	}
	if reconstructed.String() != input {
		t.Fatal("segmentation lost source bytes")
	}
	if got := TrimLeftColumns("\t\u200bला!", 1); got != "!" {
		t.Errorf("TrimLeftColumns lost source offsets: %q", got)
	}
}

func TestNativeBufferPreservesIndicUnits(t *testing.T) {
	old := DefaultWidthPolicy
	t.Cleanup(func() { DefaultWidthPolicy = old })
	for _, p := range []WidthPolicy{UnicodeWidth, WindowsTerminalWidth} {
		DefaultWidthPolicy = p
		for _, text := range []string{"हिन्दी", "বাংলা পরীক্ষা লেখা।", "க்ஷி ஸ்ரீ"} {
			buf := NewBuffer(NewRect(0, 0, 40, 1))
			x, _ := buf.SetStringn(0, 0, text, 40, NewStyle().Bg(ColorBlue))
			if int(x) != StringWidth(text) {
				t.Fatalf("%q: drawing/measurement disagree", text)
			}
			var col uint16
			for g := range AllGraphemes(text) {
				c := buf.CellAt(col, 0)
				if c.GetSymbol() != g.Symbol || c.Width() != g.Width || c.DiffOption != CellDiffNone {
					t.Errorf("%q: wrong native cell at %d: %+v", text, col, c)
				}
				if g.Width > 1 {
					clipped := NewBuffer(NewRect(0, 0, g.Width-1, 1))
					if end, _ := clipped.SetStringn(0, 0, g.Symbol, g.Width-1, NewStyle()); end != 0 {
						t.Errorf("clipped partial cluster %q", g.Symbol)
					}
				}
				col += g.Width
			}
		}
	}
}

func TestConjunctBoundariesRespectMarks(t *testing.T) {
	for _, tc := range []struct {
		prev, next string
		join       bool
	}{
		{"क्", "ष", true}, {"क्‍", "ष", true}, {"का", "ष", false},
		{"का्", "ष", false}, {"ि", "ष", false}, {"क्", "ि", false}, {"क्‌", "ष", false},
	} {
		if got := joinsConjunct(tc.prev, tc.next); got != tc.join {
			t.Errorf("%q + %q: join=%v", tc.prev, tc.next, got)
		}
	}
}
