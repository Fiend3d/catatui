// Tests ported from ratatui-widgets/src/fill.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

func TestFillFillsAreaWithSymbolAndStyle(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 3))
	NewFill(".").
		Style(catatui.NewStyle().Fg(catatui.ColorRed)).
		Render(catatui.NewRect(1, 1, 3, 1), buf)

	expected := catatui.NewBufferWithStrings(
		"     ",
		" ... ",
		"     ",
	)
	for x := uint16(1); x <= 3; x++ {
		expected.Get(x, 1).SetStyle(catatui.NewStyle().Fg(catatui.ColorRed))
	}
	catatui.AssertBuffer(t, buf, expected)
}

func TestFillClipsAreaToBuffer(t *testing.T) {
	// The render area extends past the right and bottom edges of the buffer.
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 3, 2))
	NewFill("x").Render(catatui.NewRect(1, 1, 100, 100), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("   ", " xx"))
}

func TestFillRenderFullyOutOfBoundsIsNoop(t *testing.T) {
	buf := catatui.NewBufferWithStrings(repeatLines("xxxxx", 3)...)
	NewFill(".").Render(catatui.NewRect(100, 100, 5, 5), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(repeatLines("xxxxx", 3)...))
}

func TestFillRendersWithOffsetBufferArea(t *testing.T) {
	// A buffer need not start at the origin; the intersection must still work.
	buf := catatui.NewBuffer(catatui.NewRect(2, 2, 2, 2))
	NewFill("#").Render(catatui.NewRect(0, 0, 4, 4), buf)
	expected := catatui.NewBuffer(catatui.NewRect(2, 2, 2, 2))
	for y := uint16(2); y < 4; y++ {
		for x := uint16(2); x < 4; x++ {
			expected.Get(x, y).SetSymbol("#")
		}
	}
	catatui.AssertBuffer(t, buf, expected)
}

// TestFillStyleWithModifier is ratatui's stylize_shorthand_works, written with
// an explicit Style since there is no Stylize trait here.
func TestFillStyleWithModifier(t *testing.T) {
	style := catatui.NewStyle().Fg(catatui.ColorBlue).AddModifier(catatui.ModifierBold)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 2, 1))
	NewFill("*").Style(style).Render(catatui.NewRect(0, 0, 2, 1), buf)
	expected := catatui.NewBufferWithStrings("**")
	for x := uint16(0); x < 2; x++ {
		expected.Get(x, 0).SetStyle(style)
	}
	catatui.AssertBuffer(t, buf, expected)
}

func TestFillAcceptsMultibyteSymbol(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 2, 1))
	NewFill("•").Render(catatui.NewRect(0, 0, 2, 1), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("••"))
}

func TestFillSymbolSetterReplacesSymbol(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 2, 1))
	NewFill("a").Symbol("b").Render(catatui.NewRect(0, 0, 2, 1), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("bb"))
}
