// Okhsv to sRGB, after Björn Ottosson's reference implementation, which he
// placed in the public domain: https://bottosson.github.io/posts/colorpicker/
//
// ratatui's example reaches for the `palette` crate; catatui has no colour
// science of its own and no dependency to add one, so the conversion lives
// here, in the one example that needs it.
//
// Okhsv is a hue/saturation/value space built on Oklab, in which a sweep of the
// hue keeps its apparent lightness. The same sweep in HSV does not: its yellows
// come out glaring and its blues murky, because HSV's "value" is the largest
// sRGB channel and nothing about how bright the colour looks.

package main

import "math"

// okhsvToRGB converts a colour in Okhsv to 24-bit sRGB. The hue is in degrees,
// saturation and value in 0..1.
func okhsvToRGB(hue, saturation, value float64) (r, g, b uint8) {
	rf, gf, bf := okhsvToSRGB(hue, saturation, value)
	return quantize(rf), quantize(gf), quantize(bf)
}

// quantize turns a channel in 0..1 into one of the 256 values a terminal can be
// told. Out-of-range values are clamped: a colour on the far side of the gamut
// boundary has no sRGB to be exact about.
func quantize(v float64) uint8 {
	return uint8(math.Round(min(max(v, 0), 1) * 255))
}

// okhsvToSRGB converts Okhsv to sRGB channels in 0..1.
func okhsvToSRGB(hue, saturation, value float64) (r, g, b float64) {
	if value <= 0 {
		return 0, 0, 0
	}

	// The hue as a point on the unit circle of the Oklab a/b plane.
	radians := hue * math.Pi / 180
	aUnit, bUnit := math.Cos(radians), math.Sin(radians)

	// The cusp is the most saturated colour of this hue that sRGB can show. In
	// (S, T) form it gives the two edges of the triangle the gamut would be if
	// it were flat, which is what the next few lines pretend.
	cuspL, cuspC := findCusp(aUnit, bUnit)
	sMax := cuspC / cuspL
	tMax := cuspC / (1 - cuspL)
	const s0 = 0.5
	k := 1 - s0/sMax

	// L and C at the top of the triangle, where value is 1, and then scaled
	// down to the value asked for.
	denominator := s0 + tMax - tMax*k*saturation
	lv := 1 - saturation*s0/denominator
	cv := saturation * tMax * s0 / denominator
	l := value * lv
	c := value * cv

	// Now undo both of the ways the triangle lied. The toe function is the
	// curve Okhsv applies to lightness so that dark colours are spaced the way
	// the eye sees them; toeInv takes L back to Oklab's own lightness.
	lvt := toeInv(lv)
	cvt := cv * lvt / lv

	lNew := toeInv(l)
	if l > 0 {
		c *= lNew / l
	}
	l = lNew

	// The top of the real gamut is curved rather than a straight edge, so
	// scale the whole thing to put the cusp back on the boundary.
	sr, sg, sb := oklabToLinearSRGB(lvt, aUnit*cvt, bUnit*cvt)
	scale := math.Cbrt(1 / max(sr, sg, sb, 0))
	l *= scale
	c *= scale

	lr, lg, lb := oklabToLinearSRGB(l, c*aUnit, c*bUnit)
	return srgbTransfer(lr), srgbTransfer(lg), srgbTransfer(lb)
}

// The toe function and its inverse. Oklab's L is linear in perceived
// lightness except at the dark end, where it disagrees with CIE L*; the toe is
// the correction, and Okhsv works in the corrected lightness.
const (
	toeK1 = 0.206
	toeK2 = 0.03
	toeK3 = (1 + toeK1) / (1 + toeK2)
)

func toeInv(x float64) float64 {
	return (x*x + toeK1*x) / (toeK3 * (x + toeK2))
}

// oklabToLinearSRGB converts Oklab to linear sRGB, which is sRGB before the
// transfer function is applied. Oklab goes through the LMS cone responses,
// which is the cube root in the middle.
func oklabToLinearSRGB(L, a, b float64) (r, g, blue float64) {
	lRoot := L + 0.3963377774*a + 0.2158037573*b
	mRoot := L - 0.1055613458*a - 0.0638541728*b
	sRoot := L - 0.0894841775*a - 1.2914855480*b

	l := lRoot * lRoot * lRoot
	m := mRoot * mRoot * mRoot
	s := sRoot * sRoot * sRoot

	return 4.0767416621*l - 3.3077115913*m + 0.2309699292*s,
		-1.2684380046*l + 2.6097574011*m - 0.3413193965*s,
		-0.0041960863*l - 0.7034186147*m + 1.7076147010*s
}

// srgbTransfer applies the sRGB transfer function, turning a linear channel
// into the encoded one a terminal expects.
func srgbTransfer(x float64) float64 {
	if x <= 0.0031308 {
		return 12.92 * x
	}
	return 1.055*math.Pow(x, 1/2.4) - 0.055
}

// findCusp returns the L and C of the most saturated colour of a hue that still
// fits in sRGB. The hue is given as a point on the unit circle.
func findCusp(a, b float64) (L, C float64) {
	sCusp := maxSaturation(a, b)
	// The saturation above is C/L, so this is the direction of the cusp without
	// its scale. Where the brightest channel reaches 1 is where the colour
	// leaves the gamut, and the cube root is that scale.
	r, g, blue := oklabToLinearSRGB(1, sCusp*a, sCusp*b)
	L = math.Cbrt(1 / max(r, g, blue))
	return L, L * sCusp
}

// maxSaturation returns the largest C/L a hue can have inside sRGB.
//
// The gamut boundary is where one of the three channels crosses zero, and which
// one it is depends on the hue; each gets a polynomial that lands near the
// answer, refined by a single step of Halley's method. The coefficients are
// Ottosson's, fitted to the sRGB primaries.
func maxSaturation(a, b float64) float64 {
	var k0, k1, k2, k3, k4, wl, wm, ws float64
	switch {
	case -1.88170328*a-0.80936493*b > 1: // red goes negative first
		k0, k1, k2, k3, k4 = 1.19086277, 1.76576728, 0.59662641, 0.75515197, 0.56771245
		wl, wm, ws = 4.0767416621, -3.3077115913, 0.2309699292
	case 1.81444104*a-1.19445276*b > 1: // green
		k0, k1, k2, k3, k4 = 0.73956515, -0.45954404, 0.08285427, 0.12541070, 0.14503204
		wl, wm, ws = -1.2684380046, 2.6097574011, -0.3413193965
	default: // blue
		k0, k1, k2, k3, k4 = 1.35733652, -0.00915799, -1.15130210, -0.50559606, 0.00692167
		wl, wm, ws = -0.0041960863, -0.7034186147, 1.7076147010
	}

	s := k0 + k1*a + k2*b + k3*a*a + k4*a*b

	kl := 0.3963377774*a + 0.2158037573*b
	km := -0.1055613458*a - 0.0638541728*b
	ks := -0.0894841775*a - 1.2914855480*b

	lRoot := 1 + s*kl
	mRoot := 1 + s*km
	sRoot := 1 + s*ks

	l := lRoot * lRoot * lRoot
	m := mRoot * mRoot * mRoot
	sc := sRoot * sRoot * sRoot

	// The channel and its first two derivatives with respect to saturation.
	ldS := 3 * kl * lRoot * lRoot
	mdS := 3 * km * mRoot * mRoot
	sdS := 3 * ks * sRoot * sRoot

	ldS2 := 6 * kl * kl * lRoot
	mdS2 := 6 * km * km * mRoot
	sdS2 := 6 * ks * ks * sRoot

	f := wl*l + wm*m + ws*sc
	f1 := wl*ldS + wm*mdS + ws*sdS
	f2 := wl*ldS2 + wm*mdS2 + ws*sdS2

	return s - f*f1/(f1*f1-0.5*f*f2)
}
