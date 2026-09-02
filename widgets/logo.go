// Port of ratatui-widgets/src/logo.rs @ ratatui-v0.30.2
//
// Deviation from ratatui: the logo spells catatui rather than ratatui. The
// letterforms are ratatui's own 4x4 block font, with its R replaced by a C
// drawn in the same style, so a program can show what it is built with without
// claiming to be the Rust library.

package widgets

import "github.com/Fiend3d/catatui"

// CatatuiLogoSize is the size of a CatatuiLogo.
type CatatuiLogoSize uint8

const (
	// CatatuiLogoTiny is the default size of the logo (2x15 characters).
	//
	//	▞▀▗▀▖▜▘▞▚▝▛▐ ▌▌
	//	▚▄▐▀▌▐ ▛▜ ▌▝▄▘▌
	CatatuiLogoTiny CatatuiLogoSize = iota
	// CatatuiLogoSmall is a slightly larger version of the logo (2x27
	// characters).
	//
	//	▄▀▀▀ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █
	//	▀▄▄▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █
	CatatuiLogoSmall
)

// The logo art. The tiny form packs each letter into a 2x2 block of quadrant
// characters and the small form into 4x2, which is why the same letters look
// squarer at the larger size.
const (
	catatuiLogoTiny  = "▞▀▗▀▖▜▘▞▚▝▛▐ ▌▌\n▚▄▐▀▌▐ ▛▜ ▌▝▄▘▌"
	catatuiLogoSmall = "▄▀▀▀ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █\n▀▄▄▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █"
)

// String returns the logo art for the size.
func (s CatatuiLogoSize) String() string {
	switch s {
	case CatatuiLogoSmall:
		return catatuiLogoSmall
	default:
		return catatuiLogoTiny
	}
}

// CatatuiLogo draws the catatui logo.
//
// The logo takes up two lines of text and comes in two sizes, tiny and small.
// It may be used in an application's help or about screen to show that it is
// powered by catatui.
//
//	f.RenderWidget(widgets.TinyCatatuiLogo(), area)
//
// Renders:
//
//	▞▀▗▀▖▜▘▞▚▝▛▐ ▌▌
//	▚▄▐▀▌▐ ▛▜ ▌▝▄▘▌
type CatatuiLogo struct {
	size CatatuiLogoSize
}

// NewCatatuiLogo returns a logo of the given size.
func NewCatatuiLogo(size CatatuiLogoSize) CatatuiLogo { return CatatuiLogo{size: size} }

// TinyCatatuiLogo returns the tiny logo, which is the default.
func TinyCatatuiLogo() CatatuiLogo { return NewCatatuiLogo(CatatuiLogoTiny) }

// SmallCatatuiLogo returns the small logo.
func SmallCatatuiLogo() CatatuiLogo { return NewCatatuiLogo(CatatuiLogoSmall) }

// Size returns a copy of l with the given size.
func (l CatatuiLogo) Size(size CatatuiLogoSize) CatatuiLogo { l.size = size; return l }

// GetSize returns the logo's size.
func (l CatatuiLogo) GetSize() CatatuiLogoSize { return l.size }

// Render draws the logo.
func (l CatatuiLogo) Render(area catatui.Rect, buf *catatui.Buffer) {
	catatui.TextFromString(l.size.String()).Render(area, buf)
}

var _ catatui.Widget = CatatuiLogo{}
