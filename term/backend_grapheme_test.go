package term

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// Test the native Buffer -> Diff -> VT path: no application-side cell writes
// or width overrides should be necessary to preserve Indic text on redraw.
func TestIndicSelectionRedrawsKeepCompleteSymbols(t *testing.T) {
	for _, text := range []string{"हिन्दी", "বাংলা পরীক্ষা লেখা।", "আমার নাম রবি।", "க்ஷி ஶ்ரீ ஸ்ரீ"} {
		prev := catatui.NewBuffer(catatui.NewRect(0, 0, 40, 1))
		prev.SetString(0, 0, strings.Repeat("i", 40), catatui.NewStyle())
		glyphs := catatui.Graphemes(text)
		for step := 0; step <= len(glyphs)*2; step++ {
			selected := step
			if selected > len(glyphs) {
				selected = len(glyphs)*2 - step
			}
			next := catatui.NewBuffer(prev.Area)
			var x uint16
			for i, g := range glyphs {
				style := catatui.NewStyle().Bg(catatui.ColorBlack)
				if i < selected {
					style = style.Bg(catatui.ColorBlue)
				}
				x, _ = next.SetStringn(x, 0, g.Symbol, 40-x, style)
			}
			next.SetStringn(x, 0, strings.Repeat(" ", int(40-x)), 40-x, catatui.NewStyle().Bg(catatui.ColorBlack))
			updates := prev.Diff(next)
			output := render(t, updates)
			var col uint16
			for i, g := range glyphs {
				for _, pc := range updates {
					if pc.X >= col && pc.X < col+g.Width {
						if pc.X != col || pc.Cell.GetSymbol() != g.Symbol || !strings.Contains(output, g.Symbol) {
							t.Fatalf("%q step %d: split %q at %d, VT %q", text, step, g.Symbol, pc.X, output)
						}
						if (pc.Cell.Bg == catatui.ColorBlue) != (i < selected) {
							t.Fatal("incorrect selection style")
						}
					}
				}
				col += g.Width
			}
			if step == 0 {
				// An initial redraw must overwrite all old text, including the
				// columns after each spacing mark and the tail of the row.
				var end uint16
				for _, pc := range updates {
					if pc.X != end {
						t.Fatalf("%q: stale column at %d before %d", text, end, pc.X)
					}
					end += pc.Cell.Width()
				}
				if end != 40 {
					t.Fatalf("%q: redraw covered %d cells", text, end)
				}
			}
			prev = next
		}
	}
}
