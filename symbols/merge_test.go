// Tests ported from ratatui-core/src/symbols/merge.rs @ ratatui-v0.30.2

package symbols

import "testing"

// allBorderSymbols is every character in the merge table, plus a space and two
// letters, which stand for everything that is not a box-drawing character.
var allBorderSymbols = func() []string {
	symbols := make([]string, 0, len(borderSymbolTable)+3)
	for _, e := range borderSymbolTable {
		symbols = append(symbols, e.symbol)
	}
	return append(symbols, " ", "a", "b")
}()

func TestMergeStrategyString(t *testing.T) {
	cases := []struct {
		strategy MergeStrategy
		want     string
	}{
		{MergeReplace, "Replace"},
		{MergeExact, "Exact"},
		{MergeFuzzy, "Fuzzy"},
	}
	for _, c := range cases {
		if got := c.strategy.String(); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}

func TestParseMergeStrategy(t *testing.T) {
	for _, want := range []MergeStrategy{MergeReplace, MergeExact, MergeFuzzy} {
		got, err := ParseMergeStrategy(want.String())
		if err != nil || got != want {
			t.Errorf("ParseMergeStrategy(%q) = %v, %v; want %v, nil", want, got, err, want)
		}
	}
	if _, err := ParseMergeStrategy(""); err == nil {
		t.Error("ParseMergeStrategy(\"\") = nil error, want an error")
	}
}

func TestMergeReplaceStrategy(t *testing.T) {
	for _, a := range allBorderSymbols {
		for _, b := range allBorderSymbols {
			if got := MergeReplace.Merge(a, b); got != b {
				t.Fatalf("Merge(%q, %q) = %q, want %q", a, b, got, b)
			}
		}
	}
}

func TestMergeExactStrategy(t *testing.T) {
	cases := [][3]string{
		{"┆", "─", "─"},
		{"┏", "┆", "┆"},
		{"╎", "┉", "┉"},
		{"┋", "┋", "┋"},
		{"╷", "╶", "┌"},
		{"╭", "┌", "┌"},
		{"│", "┕", "┝"},
		{"┏", "│", "┝"},
		{"│", "┏", "┢"},
		{"╽", "┕", "┢"},
		{"│", "─", "┼"},
		{"┘", "┌", "┼"},
		{"┵", "┝", "┿"},
		{"│", "━", "┿"},
		{"┵", "╞", "╞"},
		{" ", "╠", " "},
		{"╠", " ", " "},
		{"╎", "╧", "╧"},
		{"╛", "╒", "╪"},
		{"│", "═", "╪"},
		{"╤", "╧", "╪"},
		{"╡", "╞", "╪"},
		{"┌", "╭", "╭"},
		{"┘", "╭", "╭"},
		{"┌", "a", "a"},
		{"a", "╭", "a"},
		{"a", "b", "b"},
	}
	for _, c := range cases {
		if got := MergeExact.Merge(c[0], c[1]); got != c[2] {
			t.Errorf("Merge(%q, %q) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
}

func TestMergeFuzzyStrategy(t *testing.T) {
	cases := [][3]string{
		{"┄", "╴", "─"},
		{"│", "┆", "┆"},
		{" ", "┉", " "},
		{"┋", "┋", "┋"},
		{"╷", "╶", "┌"},
		{"╭", "┌", "┌"},
		{"│", "┕", "┝"},
		{"┏", "│", "┝"},
		{"┏", "┆", "┝"},
		{"│", "┏", "┢"},
		{"╽", "┕", "┢"},
		{"│", "─", "┼"},
		{"┆", "─", "┼"},
		{"┘", "┌", "┼"},
		{"┘", "╭", "┼"},
		{"╎", "┉", "┿"},
		{" ", "╠", " "},
		{"╠", " ", " "},
		{"┵", "╞", "╪"},
		{"╛", "╒", "╪"},
		{"│", "═", "╪"},
		{"╤", "╧", "╪"},
		{"╡", "╞", "╪"},
		{"╎", "╧", "╪"},
		{"┌", "╭", "╭"},
		{"┌", "a", "a"},
		{"a", "╭", "a"},
		{"a", "b", "b"},
	}
	for _, c := range cases {
		if got := MergeFuzzy.Merge(c[0], c[1]); got != c[2] {
			t.Errorf("Merge(%q, %q) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
}

// TestBorderSymbolTableIsABijection pins down what the two lookup maps assume:
// no character appears twice, and no two characters are made of the same four
// segments. Rust gets this from a macro that generates both directions of a
// match from one table; here the maps are built at init, so a duplicate would
// silently make one entry unreachable.
func TestBorderSymbolTableIsABijection(t *testing.T) {
	if len(borderBySymbol) != len(borderSymbolTable) {
		t.Errorf("table has %d entries but only %d distinct symbols",
			len(borderSymbolTable), len(borderBySymbol))
	}
	if len(symbolByBorder) != len(borderSymbolTable) {
		t.Errorf("table has %d entries but only %d distinct segment combinations",
			len(borderSymbolTable), len(symbolByBorder))
	}
	for _, e := range borderSymbolTable {
		parsed, ok := parseBorderSymbol(e.symbol)
		if !ok || parsed != e.border {
			t.Errorf("parseBorderSymbol(%q) = %+v, %v; want %+v, true",
				e.symbol, parsed, ok, e.border)
		}
		symbol, ok := e.border.symbol()
		if !ok || symbol != e.symbol {
			t.Errorf("%+v.symbol() = %q, %v; want %q, true",
				e.border, symbol, ok, e.symbol)
		}
	}
}
