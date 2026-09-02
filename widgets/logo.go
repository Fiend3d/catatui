// Port of ratatui-widgets/src/logo.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// RatatuiLogoSize is the size of a RatatuiLogo.
type RatatuiLogoSize uint8

const (
	// RatatuiLogoTiny is the default size of the logo (2x15 characters).
	//
	//	▛▚▗▀▖▜▘▞▚▝▛▐ ▌▌
	//	▛▚▐▀▌▐ ▛▜ ▌▝▄▘▌
	RatatuiLogoTiny RatatuiLogoSize = iota
	// RatatuiLogoSmall is a slightly larger version of the logo (2x27
	// characters).
	//
	//	█▀▀▄ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █
	//	█▀▀▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █
	RatatuiLogoSmall
)

// The logo art, exactly as ratatui draws it.
const (
	ratatuiLogoTiny  = "▛▚▗▀▖▜▘▞▚▝▛▐ ▌▌\n▛▚▐▀▌▐ ▛▜ ▌▝▄▘▌"
	ratatuiLogoSmall = "█▀▀▄ ▄▀▀▄▝▜▛▘▄▀▀▄▝▜▛▘█  █ █\n█▀▀▄ █▀▀█ ▐▌ █▀▀█ ▐▌ ▀▄▄▀ █"
)

// String returns the logo art for the size.
func (s RatatuiLogoSize) String() string {
	switch s {
	case RatatuiLogoSmall:
		return ratatuiLogoSmall
	default:
		return ratatuiLogoTiny
	}
}

// RatatuiLogo draws the Ratatui logo.
//
// The logo takes up two lines of text and comes in two sizes, tiny and small.
// It may be used in an application's help or about screen to show that it is
// powered by ratatui.
//
//	f.RenderWidget(widgets.TinyRatatuiLogo(), area)
//
// Renders:
//
//	▛▚▗▀▖▜▘▞▚▝▛▐ ▌▌
//	▛▚▐▀▌▐ ▛▜ ▌▝▄▘▌
type RatatuiLogo struct {
	size RatatuiLogoSize
}

// NewRatatuiLogo returns a logo of the given size.
func NewRatatuiLogo(size RatatuiLogoSize) RatatuiLogo { return RatatuiLogo{size: size} }

// TinyRatatuiLogo returns the tiny logo, which is the default.
func TinyRatatuiLogo() RatatuiLogo { return NewRatatuiLogo(RatatuiLogoTiny) }

// SmallRatatuiLogo returns the small logo.
func SmallRatatuiLogo() RatatuiLogo { return NewRatatuiLogo(RatatuiLogoSmall) }

// Size returns a copy of l with the given size.
func (l RatatuiLogo) Size(size RatatuiLogoSize) RatatuiLogo { l.size = size; return l }

// GetSize returns the logo's size.
func (l RatatuiLogo) GetSize() RatatuiLogoSize { return l.size }

// Render draws the logo.
func (l RatatuiLogo) Render(area catatui.Rect, buf *catatui.Buffer) {
	catatui.TextFromString(l.size.String()).Render(area, buf)
}

var _ catatui.Widget = RatatuiLogo{}
