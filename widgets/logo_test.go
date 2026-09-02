// Tests ported from ratatui-widgets/src/logo.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

func TestRatatuiLogoNewSize(t *testing.T) {
	for _, size := range []RatatuiLogoSize{RatatuiLogoTiny, RatatuiLogoSmall} {
		if got := NewRatatuiLogo(size).GetSize(); got != size {
			t.Errorf("NewRatatuiLogo(%v).GetSize() = %v", size, got)
		}
	}
}

func TestRatatuiLogoDefaultIsTiny(t *testing.T) {
	if got := (RatatuiLogo{}).GetSize(); got != RatatuiLogoTiny {
		t.Errorf("zero logo size = %v, want tiny", got)
	}
}

func TestRatatuiLogoSetSizeToSmall(t *testing.T) {
	if got := (RatatuiLogo{}).Size(RatatuiLogoSmall).GetSize(); got != RatatuiLogoSmall {
		t.Errorf("Size(small) = %v, want small", got)
	}
}

func TestRatatuiLogoTinyConstant(t *testing.T) {
	if got := TinyRatatuiLogo().GetSize(); got != RatatuiLogoTiny {
		t.Errorf("TinyRatatuiLogo().GetSize() = %v, want tiny", got)
	}
}

func TestRatatuiLogoSmallConstant(t *testing.T) {
	if got := SmallRatatuiLogo().GetSize(); got != RatatuiLogoSmall {
		t.Errorf("SmallRatatuiLogo().GetSize() = %v, want small", got)
	}
}

func TestRatatuiLogoRenderTiny(t *testing.T) {
	buf := renderToBuffer(TinyRatatuiLogo(), 15, 2)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"▛▚▗▀▖▜▘▞▚▝▛▐ ▌▌",
		"▛▚▐▀▌▐ ▛▜ ▌▝▄▘▌",
	))
}

func TestRatatuiLogoRenderSmall(t *testing.T) {
	buf := renderToBuffer(SmallRatatuiLogo(), 27, 2)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"█▀▀▄ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █",
		"█▀▀▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █",
	))
}

func TestRatatuiLogoRenderInMinimalBuffer(t *testing.T) {
	cases := []struct {
		name string
		size RatatuiLogoSize
		want string
	}{
		{"tiny", RatatuiLogoTiny, "▛"},
		{"small", RatatuiLogoSmall, "█"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// This should not panic, even if the buffer is too small to
			// render the logo.
			buf := renderToBuffer(NewRatatuiLogo(c.size), 1, 1)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want))
		})
	}
}

func TestRatatuiLogoRenderInZeroSizeBuffer(t *testing.T) {
	for _, size := range []RatatuiLogoSize{RatatuiLogoTiny, RatatuiLogoSmall} {
		// This should not panic, even if the buffer has zero size.
		renderToBuffer(NewRatatuiLogo(size), 0, 0)
	}
}
