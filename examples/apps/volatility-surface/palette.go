// Port of the palette half of examples/apps/volatility-surface/src/display
// @ ratatui-v0.30.2

package main

import "github.com/Fiend3d/catatui"

// palette is one of the colour ramps the surface can be drawn in, in the order
// the p key cycles through them.
type palette int

const (
	paletteViridis palette = iota
	palettePlasma
	palettePhosphor
	paletteFear
	paletteCalm
	paletteInferno
)

// paletteCount is how many there are, which is what makes the cycle wrap.
const paletteCount = paletteInferno + 1

func (p palette) String() string {
	switch p {
	case paletteViridis:
		return "viridis"
	case palettePlasma:
		return "plasma"
	case palettePhosphor:
		return "phosphor"
	case paletteFear:
		return "fear"
	case paletteCalm:
		return "calm"
	default:
		return "inferno"
	}
}

// next is the palette after this one, wrapping round at the end.
func (p palette) next() palette { return (p + 1) % paletteCount }

// color returns the colour at t, which is clamped to 0..1.
//
// The three named colormaps are tables of 256 stops (see colormaps.go); the
// other three are two-colour ramps written out here, as ratatui's example
// builds them. Either way the stops are mixed in sRGB, which is what colorgrad
// does by default.
func (p palette) color(t float64) catatui.Color {
	t = min(max(t, 0), 1)
	switch p {
	case paletteViridis:
		return sampleStops(viridisStops[:], t)
	case palettePlasma:
		return sampleStops(plasmaStops[:], t)
	case paletteInferno:
		return sampleStops(infernoStops[:], t)
	case palettePhosphor:
		return lerpColor(0x002800, 0x26ff3f, t)
	case paletteFear:
		return lerpColor(0x7f0000, 0xff4c00, t)
	default:
		return lerpColor(0x334c7f, 0x66ccff, t)
	}
}

// sampleStops reads a colormap at t, mixing the two stops it falls between.
func sampleStops(stops []uint32, t float64) catatui.Color {
	position := t * float64(len(stops)-1)
	low := int(position)
	if low >= len(stops)-1 {
		return catatui.RgbFromU32(stops[len(stops)-1])
	}
	return lerpColor(stops[low], stops[low+1], position-float64(low))
}

// lerpColor mixes two packed 0xRRGGBB colours, channel by channel.
func lerpColor(from, to uint32, t float64) catatui.Color {
	r := lerpChannel(from>>16, to>>16, t)
	g := lerpChannel(from>>8, to>>8, t)
	b := lerpChannel(from, to, t)
	return catatui.Rgb(r, g, b)
}

func lerpChannel(from, to uint32, t float64) uint8 {
	a, b := float64(from&0xff), float64(to&0xff)
	return uint8(a + (b-a)*t + 0.5)
}
