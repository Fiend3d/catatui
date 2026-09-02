// Tests ported from ratatui-widgets/src/block/shadow.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// renderShadow draws a shadow for a 2x2 area at the origin of a 4x4 buffer.
func renderShadow(s Shadow) *catatui.Buffer {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 4, 4))
	s.Render(catatui.NewRect(0, 0, 2, 2), buf)
	return buf
}

func TestShadowOverlayRendersStyleWithoutChangingSymbols(t *testing.T) {
	buf := catatui.NewBufferWithStrings("abcd", "efgh", "ijkl", "mnop")
	shadow := ShadowOverlay().Style(catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorBlue))

	shadow.Render(catatui.NewRect(0, 0, 2, 2), buf)

	if got := buf.Get(2, 1).GetSymbol(); got != "g" {
		t.Errorf("(2, 1) symbol = %q, want g", got)
	}
	if got := buf.Get(1, 2).GetSymbol(); got != "j" {
		t.Errorf("(1, 2) symbol = %q, want j", got)
	}
	if got := buf.Get(2, 2).GetSymbol(); got != "k" {
		t.Errorf("(2, 2) symbol = %q, want k", got)
	}
	if got := buf.Get(2, 1).Fg; got != catatui.ColorRed {
		t.Errorf("(2, 1) fg = %v, want red", got)
	}
	if got := buf.Get(2, 1).Bg; got != catatui.ColorBlue {
		t.Errorf("(2, 1) bg = %v, want blue", got)
	}
	if got := buf.Get(1, 1).Fg; got != catatui.ColorReset {
		t.Errorf("(1, 1) fg = %v, want reset", got)
	}
	if got := buf.Get(1, 1).Bg; got != catatui.ColorReset {
		t.Errorf("(1, 1) bg = %v, want reset", got)
	}
}

func TestShadowSymbolFiltersFillOnlyVisibleShadowCells(t *testing.T) {
	cases := []struct {
		name   string
		shadow Shadow
		symbol string
	}{
		{"symbol", ShadowSymbol("$"), "$"},
		{"block", ShadowBlock(), symbols.ShadeFull},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := renderShadow(c.shadow)
			for _, p := range []catatui.Position{{X: 2, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 2}} {
				if got := buf.Cell(p).GetSymbol(); got != c.symbol {
					t.Errorf("(%d, %d) symbol = %q, want %q", p.X, p.Y, got, c.symbol)
				}
			}
			if got := buf.Get(1, 1).GetSymbol(); got != " " {
				t.Errorf("(1, 1) symbol = %q, want space", got)
			}
		})
	}
}

func TestShadowRenderIsClippedToBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 3, 2))
	shadow := ShadowSymbol("#")

	shadow.Render(catatui.NewRect(0, 0, 2, 1), buf)

	if got := buf.Get(2, 1).GetSymbol(); got != "#" {
		t.Errorf("(2, 1) symbol = %q, want #", got)
	}
}

// plusFilter is the custom effect from ratatui's test.
type plusFilter struct{}

func (plusFilter) Apply(shadowArea, baseArea catatui.Rect, buf *catatui.Buffer) {
	ForEachShadowCell(shadowArea, baseArea, buf, func(x, y uint16, buf *catatui.Buffer) {
		buf.Get(x, y).SetSymbol("+")
	})
}

func TestShadowCustomFilterIsApplied(t *testing.T) {
	buf := renderShadow(NewShadow(plusFilter{}))

	for _, p := range []catatui.Position{{X: 2, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 2}} {
		if got := buf.Cell(p).GetSymbol(); got != "+" {
			t.Errorf("(%d, %d) symbol = %q, want +", p.X, p.Y, got)
		}
	}
}

func TestShadowDimmedFilterDimsBackground(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 4, 4))
	buf.SetStyle(buf.Area, catatui.NewStyle().Bg(catatui.Rgb(100, 120, 140)))
	shadow := NewShadow(NewDimmed())

	shadow.Render(catatui.NewRect(0, 0, 2, 2), buf)

	if !buf.Get(2, 1).Modifier.Contains(catatui.ModifierDim) {
		t.Errorf("(2, 1) should be dimmed")
	}
	if got := buf.Get(2, 1).Bg; got != catatui.Rgb(50, 60, 70) {
		t.Errorf("(2, 1) bg = %v, want rgb(50, 60, 70)", got)
	}
	if got := buf.Get(1, 1).Bg; got != catatui.Rgb(100, 120, 140) {
		t.Errorf("(1, 1) bg = %v, want rgb(100, 120, 140)", got)
	}
	if buf.Get(1, 1).Modifier.Contains(catatui.ModifierDim) {
		t.Errorf("(1, 1) should not be dimmed")
	}
}

// TestBlockShadow checks the shadow is drawn from the block's area, after the
// border and titles, and clipped to the buffer.
func TestBlockShadow(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 8, 5))
	Bordered().Title("Hi").Shadow(ShadowMediumShade()).Render(catatui.NewRect(0, 0, 6, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌Hi──┐  ",
		"│    │▒ ",
		"└────┘▒ ",
		" ▒▒▒▒▒▒ ",
		"        ",
	))
}

// TestBlockShadowOffset checks a custom offset and that the shadow style is
// applied only outside the block.
func TestBlockShadowOffset(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 8, 5))
	style := catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorWhite)
	block := Bordered().Shadow(ShadowDarkShade().Style(style).Offset(catatui.Offset{X: 2, Y: 1}))
	block.Render(catatui.NewRect(0, 0, 6, 3), buf)

	want := catatui.NewBufferWithStrings(
		"┌────┐  ",
		"│    │▓▓",
		"└────┘▓▓",
		"  ▓▓▓▓▓▓",
		"        ",
	)
	want.SetStyle(catatui.NewRect(6, 1, 2, 2), style)
	want.SetStyle(catatui.NewRect(2, 3, 6, 1), style)
	catatui.AssertBuffer(t, buf, want)
}
