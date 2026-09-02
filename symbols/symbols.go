// Package symbols holds the Unicode characters catatui draws with: box-drawing
// lines, block elements, braille dots and bar segments.
//
// Port of ratatui-core/src/symbols/*.rs @ ratatui-v0.30.2
//
// The package has no dependencies and holds only constants, so both the core
// package and the widgets can use it without a cycle.
package symbols

// Box-drawing characters, by weight.
const (
	Vertical     = "│"
	Horizontal   = "─"
	TopRight     = "┐"
	TopLeft      = "┌"
	BottomRight  = "┘"
	BottomLeft   = "└"
	VerticalLeft = "┤"
	// VerticalRight is a vertical line with a branch to the right.
	VerticalRight  = "├"
	HorizontalDown = "┬"
	HorizontalUp   = "┴"
	Cross          = "┼"

	RoundedTopLeft     = "╭"
	RoundedTopRight    = "╮"
	RoundedBottomLeft  = "╰"
	RoundedBottomRight = "╯"

	DoubleVertical       = "║"
	DoubleHorizontal     = "═"
	DoubleTopRight       = "╗"
	DoubleTopLeft        = "╔"
	DoubleBottomRight    = "╝"
	DoubleBottomLeft     = "╚"
	DoubleVerticalLeft   = "╣"
	DoubleVerticalRight  = "╠"
	DoubleHorizontalDown = "╦"
	DoubleHorizontalUp   = "╩"
	DoubleCross          = "╬"

	ThickVertical       = "┃"
	ThickHorizontal     = "━"
	ThickTopRight       = "┓"
	ThickTopLeft        = "┏"
	ThickBottomRight    = "┛"
	ThickBottomLeft     = "┗"
	ThickVerticalLeft   = "┫"
	ThickVerticalRight  = "┣"
	ThickHorizontalDown = "┳"
	ThickHorizontalUp   = "┻"
	ThickCross          = "╋"

	// The dashed lines come only as verticals and horizontals; there are no
	// dashed corners or junctions.
	LightDoubleDashVertical      = "╎"
	HeavyDoubleDashVertical      = "╏"
	LightTripleDashVertical      = "┆"
	HeavyTripleDashVertical      = "┇"
	LightQuadrupleDashVertical   = "┊"
	HeavyQuadrupleDashVertical   = "┋"
	LightDoubleDashHorizontal    = "╌"
	HeavyDoubleDashHorizontal    = "╍"
	LightTripleDashHorizontal    = "┄"
	HeavyTripleDashHorizontal    = "┅"
	LightQuadrupleDashHorizontal = "┈"
	HeavyQuadrupleDashHorizontal = "┉"
)

// Block elements, used for gauges, bars and sparklines.
const (
	BlockFull          = "█"
	BlockSevenEighths  = "▉"
	BlockThreeQuarters = "▊"
	BlockFiveEighths   = "▋"
	BlockHalf          = "▌"
	BlockThreeEighths  = "▍"
	BlockOneQuarter    = "▎"
	BlockOneEighth     = "▏"
	BlockEmpty         = " "
)

// Bar elements, drawn from the bottom up rather than the left in.
const (
	BarFull          = "█"
	BarSevenEighths  = "▇"
	BarThreeQuarters = "▆"
	BarFiveEighths   = "▅"
	BarHalf          = "▄"
	BarThreeEighths  = "▃"
	BarOneQuarter    = "▂"
	BarOneEighth     = "▁"
	BarEmpty         = " "
)

// Half blocks, used by the half-block canvas marker.
const (
	HalfBlockUpper = "▀"
	HalfBlockLower = "▄"
	HalfBlockFull  = "█"
)

// Dots used as markers.
const (
	DotFull   = "•"
	DotMedium = "·"
)

// LineSet is a full set of box-drawing characters, including the tee and cross
// pieces needed to join lines.
type LineSet struct {
	Vertical       string
	Horizontal     string
	TopRight       string
	TopLeft        string
	BottomRight    string
	BottomLeft     string
	VerticalLeft   string
	VerticalRight  string
	HorizontalDown string
	HorizontalUp   string
	Cross          string
}

// The standard line sets.
var (
	// LineNormal is the single-width box-drawing set, and the default.
	LineNormal = LineSet{
		Vertical: Vertical, Horizontal: Horizontal,
		TopRight: TopRight, TopLeft: TopLeft,
		BottomRight: BottomRight, BottomLeft: BottomLeft,
		VerticalLeft: VerticalLeft, VerticalRight: VerticalRight,
		HorizontalDown: HorizontalDown, HorizontalUp: HorizontalUp,
		Cross: Cross,
	}

	// LineRounded is LineNormal with curved corners.
	LineRounded = LineSet{
		Vertical: Vertical, Horizontal: Horizontal,
		TopRight: RoundedTopRight, TopLeft: RoundedTopLeft,
		BottomRight: RoundedBottomRight, BottomLeft: RoundedBottomLeft,
		VerticalLeft: VerticalLeft, VerticalRight: VerticalRight,
		HorizontalDown: HorizontalDown, HorizontalUp: HorizontalUp,
		Cross: Cross,
	}

	// LineDouble is the double-line set.
	LineDouble = LineSet{
		Vertical: DoubleVertical, Horizontal: DoubleHorizontal,
		TopRight: DoubleTopRight, TopLeft: DoubleTopLeft,
		BottomRight: DoubleBottomRight, BottomLeft: DoubleBottomLeft,
		VerticalLeft: DoubleVerticalLeft, VerticalRight: DoubleVerticalRight,
		HorizontalDown: DoubleHorizontalDown, HorizontalUp: DoubleHorizontalUp,
		Cross: DoubleCross,
	}

	// LineThick is the heavy-weight set.
	LineThick = LineSet{
		Vertical: ThickVertical, Horizontal: ThickHorizontal,
		TopRight: ThickTopRight, TopLeft: ThickTopLeft,
		BottomRight: ThickBottomRight, BottomLeft: ThickBottomLeft,
		VerticalLeft: ThickVerticalLeft, VerticalRight: ThickVerticalRight,
		HorizontalDown: ThickHorizontalDown, HorizontalUp: ThickHorizontalUp,
		Cross: ThickCross,
	}

	// The dashed sets are LineNormal or LineThick with the straight runs
	// dashed; their corners and junctions stay solid, because Unicode has no
	// dashed ones.

	// LineLightDoubleDashed is LineNormal with two-dash straights.
	LineLightDoubleDashed = dashedLineSet(LineNormal, LightDoubleDashVertical, LightDoubleDashHorizontal)
	// LineHeavyDoubleDashed is LineThick with two-dash straights.
	LineHeavyDoubleDashed = dashedLineSet(LineThick, HeavyDoubleDashVertical, HeavyDoubleDashHorizontal)
	// LineLightTripleDashed is LineNormal with three-dash straights.
	LineLightTripleDashed = dashedLineSet(LineNormal, LightTripleDashVertical, LightTripleDashHorizontal)
	// LineHeavyTripleDashed is LineThick with three-dash straights.
	LineHeavyTripleDashed = dashedLineSet(LineThick, HeavyTripleDashVertical, HeavyTripleDashHorizontal)
	// LineLightQuadrupleDashed is LineNormal with four-dash straights.
	LineLightQuadrupleDashed = dashedLineSet(LineNormal, LightQuadrupleDashVertical, LightQuadrupleDashHorizontal)
	// LineHeavyQuadrupleDashed is LineThick with four-dash straights.
	LineHeavyQuadrupleDashed = dashedLineSet(LineThick, HeavyQuadrupleDashVertical, HeavyQuadrupleDashHorizontal)
)

// dashedLineSet copies a line set with the vertical and horizontal replaced by
// their dashed forms.
func dashedLineSet(base LineSet, vertical, horizontal string) LineSet {
	base.Vertical = vertical
	base.Horizontal = horizontal
	return base
}

// BorderSet is the eight characters a Block needs to draw its border. It names
// the two verticals and two horizontals separately so that a border can be
// heavier on one side than another.
type BorderSet struct {
	TopLeft          string
	TopRight         string
	BottomLeft       string
	BottomRight      string
	VerticalLeft     string
	VerticalRight    string
	HorizontalTop    string
	HorizontalBottom string
}

// BorderFromLineSet builds a BorderSet from a LineSet, using the same character
// on both verticals and both horizontals.
func BorderFromLineSet(s LineSet) BorderSet {
	return BorderSet{
		TopLeft: s.TopLeft, TopRight: s.TopRight,
		BottomLeft: s.BottomLeft, BottomRight: s.BottomRight,
		VerticalLeft: s.Vertical, VerticalRight: s.Vertical,
		HorizontalTop: s.Horizontal, HorizontalBottom: s.Horizontal,
	}
}

// The standard border sets.
var (
	// BorderPlain is the single-width border, and the default.
	BorderPlain = BorderFromLineSet(LineNormal)
	// BorderRounded has curved corners.
	BorderRounded = BorderFromLineSet(LineRounded)
	// BorderDouble is drawn with double lines.
	BorderDouble = BorderFromLineSet(LineDouble)
	// BorderThick is drawn with heavy lines.
	BorderThick = BorderFromLineSet(LineThick)

	// BorderLightDoubleDashed has two-dash sides and plain corners.
	BorderLightDoubleDashed = BorderFromLineSet(LineLightDoubleDashed)
	// BorderHeavyDoubleDashed has heavy two-dash sides and thick corners.
	BorderHeavyDoubleDashed = BorderFromLineSet(LineHeavyDoubleDashed)
	// BorderLightTripleDashed has three-dash sides and plain corners.
	BorderLightTripleDashed = BorderFromLineSet(LineLightTripleDashed)
	// BorderHeavyTripleDashed has heavy three-dash sides and thick corners.
	BorderHeavyTripleDashed = BorderFromLineSet(LineHeavyTripleDashed)
	// BorderLightQuadrupleDashed has four-dash sides and plain corners.
	BorderLightQuadrupleDashed = BorderFromLineSet(LineLightQuadrupleDashed)
	// BorderHeavyQuadrupleDashed has heavy four-dash sides and thick corners.
	BorderHeavyQuadrupleDashed = BorderFromLineSet(LineHeavyQuadrupleDashed)

	// BorderQuadrantOutside draws the border in the outer half of each cell,
	// so adjacent blocks appear to share a single crisp line.
	BorderQuadrantOutside = BorderSet{
		TopLeft: "▛", TopRight: "▜", BottomLeft: "▙", BottomRight: "▟",
		VerticalLeft: "▌", VerticalRight: "▐",
		HorizontalTop: "▀", HorizontalBottom: "▄",
	}

	// BorderQuadrantInside draws the border in the inner half of each cell.
	BorderQuadrantInside = BorderSet{
		TopLeft: "▗", TopRight: "▖", BottomLeft: "▝", BottomRight: "▘",
		VerticalLeft: "▐", VerticalRight: "▌",
		HorizontalTop: "▄", HorizontalBottom: "▀",
	}

	// BorderFull fills every border cell solidly.
	BorderFull = BorderSet{
		TopLeft: BlockFull, TopRight: BlockFull,
		BottomLeft: BlockFull, BottomRight: BlockFull,
		VerticalLeft: BlockFull, VerticalRight: BlockFull,
		HorizontalTop: BlockFull, HorizontalBottom: BlockFull,
	}

	// BorderEmpty draws nothing but still reserves the space, which is a
	// convenient way to pad a widget by one cell.
	BorderEmpty = BorderSet{
		TopLeft: " ", TopRight: " ", BottomLeft: " ", BottomRight: " ",
		VerticalLeft: " ", VerticalRight: " ",
		HorizontalTop: " ", HorizontalBottom: " ",
	}
)

// Braille cell geometry. A braille cell packs a 2x4 grid of dots into one
// character, giving a canvas four times the vertical resolution of half blocks.
const (
	BrailleBlank  = '⠀'
	BrailleWidth  = 2
	BrailleHeight = 4
)

// BrailleDots maps a (row, column) within a braille cell to the bit that lights
// it, so a canvas can OR dots together into a single rune.
var BrailleDots = [4][2]rune{
	{0x0001, 0x0008},
	{0x0002, 0x0010},
	{0x0004, 0x0020},
	{0x0040, 0x0080},
}
