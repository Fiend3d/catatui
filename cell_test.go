// Tests ported from ratatui-core/src/buffer/cell.rs @ ratatui-v0.30.2

package catatui

import (
	"testing"

	"github.com/Fiend3d/catatui/symbols"
)

func TestCellMergeSymbol(t *testing.T) {
	cases := []struct {
		name     string
		cell     Cell
		symbol   string
		strategy symbols.MergeStrategy
		want     string
	}{
		{
			name:     "exact merge of two corners",
			cell:     NewCell("┘"),
			symbol:   "┏",
			strategy: symbols.MergeExact,
			want:     "╆",
		},
		{
			name:     "fuzzy merge drops the rounded corner",
			cell:     NewCell("╭"),
			symbol:   "┘",
			strategy: symbols.MergeFuzzy,
			want:     "┼",
		},
		{
			name:     "replace keeps the new symbol",
			cell:     NewCell("┘"),
			symbol:   "┏",
			strategy: symbols.MergeReplace,
			want:     "┏",
		},
		{
			name:     "a cell with no symbol takes the new one",
			cell:     EmptyCell(),
			symbol:   "┏",
			strategy: symbols.MergeExact,
			want:     "┏",
		},
		{
			name:     "a border does not overwrite text",
			cell:     NewCell("a"),
			symbol:   "┏",
			strategy: symbols.MergeExact,
			want:     "a",
		},
		{
			name:     "text overwrites a border",
			cell:     NewCell("┏"),
			symbol:   "a",
			strategy: symbols.MergeExact,
			want:     "a",
		},
		{
			name:     "a cell holding a space is not empty",
			cell:     NewCell(" "),
			symbol:   "┏",
			strategy: symbols.MergeExact,
			want:     " ",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cell := c.cell
			if got := cell.MergeSymbol(c.symbol, c.strategy).GetSymbol(); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestCellMergeSymbolKeepsStyle(t *testing.T) {
	cell := NewCell("┘")
	cell.SetStyle(NewStyle().Fg(ColorRed))
	cell.MergeSymbol("┏", symbols.MergeExact)

	if got := cell.GetSymbol(); got != "╆" {
		t.Errorf("symbol = %q, want ╆", got)
	}
	if got := cell.Fg; got != ColorRed {
		t.Errorf("fg = %v, want red", got)
	}
}
