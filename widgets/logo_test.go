// Tests ported from ratatui-widgets/src/logo.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

func TestCatatuiLogoNewSize(t *testing.T) {
	for _, size := range []CatatuiLogoSize{CatatuiLogoTiny, CatatuiLogoSmall} {
		if got := NewCatatuiLogo(size).GetSize(); got != size {
			t.Errorf("NewCatatuiLogo(%v).GetSize() = %v", size, got)
		}
	}
}

func TestCatatuiLogoDefaultIsTiny(t *testing.T) {
	if got := (CatatuiLogo{}).GetSize(); got != CatatuiLogoTiny {
		t.Errorf("zero logo size = %v, want tiny", got)
	}
}

func TestCatatuiLogoSetSizeToSmall(t *testing.T) {
	if got := (CatatuiLogo{}).Size(CatatuiLogoSmall).GetSize(); got != CatatuiLogoSmall {
		t.Errorf("Size(small) = %v, want small", got)
	}
}

func TestCatatuiLogoTinyConstant(t *testing.T) {
	if got := TinyCatatuiLogo().GetSize(); got != CatatuiLogoTiny {
		t.Errorf("TinyCatatuiLogo().GetSize() = %v, want tiny", got)
	}
}

func TestCatatuiLogoSmallConstant(t *testing.T) {
	if got := SmallCatatuiLogo().GetSize(); got != CatatuiLogoSmall {
		t.Errorf("SmallCatatuiLogo().GetSize() = %v, want small", got)
	}
}

func TestCatatuiLogoRenderTiny(t *testing.T) {
	buf := renderToBuffer(TinyCatatuiLogo(), 15, 2)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"▞▀▗▀▖▜▘▞▚▝▛▐ ▌▌",
		"▚▄▐▀▌▐ ▛▜ ▌▝▄▘▌",
	))
}

func TestCatatuiLogoRenderSmall(t *testing.T) {
	buf := renderToBuffer(SmallCatatuiLogo(), 27, 2)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"▄▀▀▀ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █",
		"▀▄▄▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █",
	))
}

func TestCatatuiLogoRenderInMinimalBuffer(t *testing.T) {
	cases := []struct {
		name string
		size CatatuiLogoSize
		want string
	}{
		{"tiny", CatatuiLogoTiny, "▞"},
		{"small", CatatuiLogoSmall, "▄"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// This should not panic, even if the buffer is too small to
			// render the logo.
			buf := renderToBuffer(NewCatatuiLogo(c.size), 1, 1)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want))
		})
	}
}

func TestCatatuiLogoRenderInZeroSizeBuffer(t *testing.T) {
	for _, size := range []CatatuiLogoSize{CatatuiLogoTiny, CatatuiLogoSmall} {
		// This should not panic, even if the buffer has zero size.
		renderToBuffer(NewCatatuiLogo(size), 0, 0)
	}
}
