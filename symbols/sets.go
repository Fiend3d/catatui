// Symbol sets used by the gauge, bar chart, sparkline, scrollbar and canvas
// widgets.
//
// Port of ratatui-core/src/symbols/{bar,block,scrollbar,marker,shade}.rs @ ratatui-v0.30.2

package symbols

// LevelSet is a set of nine characters that fill a cell in eighths, used to
// draw a bar whose length is not a whole number of cells. Bar sets grow from
// the bottom up; block sets grow from the left in.
type LevelSet struct {
	Full          string
	SevenEighths  string
	ThreeQuarters string
	FiveEighths   string
	Half          string
	ThreeEighths  string
	OneQuarter    string
	OneEighth     string
	Empty         string
}

// The standard level sets. The nine-level sets are the defaults; the
// three-level sets only use full, half and empty and look right in fonts that
// lack the finer block elements.
var (
	// BarNineLevels is the full set of vertical bar segments.
	BarNineLevels = LevelSet{
		Full: BarFull, SevenEighths: BarSevenEighths, ThreeQuarters: BarThreeQuarters,
		FiveEighths: BarFiveEighths, Half: BarHalf, ThreeEighths: BarThreeEighths,
		OneQuarter: BarOneQuarter, OneEighth: BarOneEighth, Empty: BarEmpty,
	}

	// BarThreeLevels rounds every vertical bar to full, half or empty.
	BarThreeLevels = LevelSet{
		Full: BarFull, SevenEighths: BarFull, ThreeQuarters: BarHalf,
		FiveEighths: BarHalf, Half: BarHalf, ThreeEighths: BarHalf,
		OneQuarter: BarHalf, OneEighth: BarEmpty, Empty: BarEmpty,
	}

	// BlockNineLevels is the full set of horizontal block segments.
	BlockNineLevels = LevelSet{
		Full: BlockFull, SevenEighths: BlockSevenEighths, ThreeQuarters: BlockThreeQuarters,
		FiveEighths: BlockFiveEighths, Half: BlockHalf, ThreeEighths: BlockThreeEighths,
		OneQuarter: BlockOneQuarter, OneEighth: BlockOneEighth, Empty: BlockEmpty,
	}

	// BlockThreeLevels rounds every horizontal block to full, half or empty.
	BlockThreeLevels = LevelSet{
		Full: BlockFull, SevenEighths: BlockFull, ThreeQuarters: BlockHalf,
		FiveEighths: BlockHalf, Half: BlockHalf, ThreeEighths: BlockHalf,
		OneQuarter: BlockHalf, OneEighth: BlockEmpty, Empty: BlockEmpty,
	}
)

// ScrollbarSet is the four characters a Scrollbar draws with.
type ScrollbarSet struct {
	Track string
	Thumb string
	Begin string
	End   string
}

// The standard scrollbar sets.
var (
	// ScrollbarDoubleVertical is a double-line track with triangle arrows.
	ScrollbarDoubleVertical = ScrollbarSet{Track: DoubleVertical, Thumb: BlockFull, Begin: "▲", End: "▼"}
	// ScrollbarDoubleHorizontal is a double-line track with triangle arrows.
	ScrollbarDoubleHorizontal = ScrollbarSet{Track: DoubleHorizontal, Thumb: BlockFull, Begin: "◄", End: "►"}
	// ScrollbarVertical is a single-line track with arrow glyphs.
	ScrollbarVertical = ScrollbarSet{Track: Vertical, Thumb: BlockFull, Begin: "↑", End: "↓"}
	// ScrollbarHorizontal is a single-line track with arrow glyphs.
	ScrollbarHorizontal = ScrollbarSet{Track: Horizontal, Thumb: BlockFull, Begin: "←", End: "→"}
)

// Shade characters, from empty to solid.
const (
	ShadeEmpty  = " "
	ShadeLight  = "░"
	ShadeMedium = "▒"
	ShadeDark   = "▓"
	ShadeFull   = "█"
)

// MarkerKind selects how a canvas or chart plots its points.
type MarkerKind uint8

const (
	// MarkerDot draws one • per cell, and is the default.
	MarkerDot MarkerKind = iota
	// MarkerBlock draws one █ per cell.
	MarkerBlock
	// MarkerBar draws one ▄ per cell.
	MarkerBar
	// MarkerBraille uses braille patterns for a 2x4 grid of dots per cell.
	MarkerBraille
	// MarkerHalfBlock uses half blocks for a 1x2 grid per cell, each half
	// with its own color.
	MarkerHalfBlock
	// MarkerQuadrant uses quadrant blocks for a 2x2 grid per cell.
	MarkerQuadrant
	// MarkerSextant uses sextant blocks for a 2x3 grid per cell.
	MarkerSextant
	// MarkerOctant uses octant blocks for a 2x4 grid per cell.
	MarkerOctant
	// MarkerCustom draws the rune in Marker.Rune once per cell.
	MarkerCustom
)

// Marker is a MarkerKind plus the rune used by MarkerCustom.
type Marker struct {
	Kind MarkerKind
	Rune rune
}

// Predefined markers, so a caller can write symbols.Braille rather than build
// the struct.
var (
	Dot       = Marker{Kind: MarkerDot}
	Block     = Marker{Kind: MarkerBlock}
	Bar       = Marker{Kind: MarkerBar}
	Braille   = Marker{Kind: MarkerBraille}
	HalfBlock = Marker{Kind: MarkerHalfBlock}
	Quadrant  = Marker{Kind: MarkerQuadrant}
	Sextant   = Marker{Kind: MarkerSextant}
	Octant    = Marker{Kind: MarkerOctant}
)

// Custom returns a marker that draws the given rune once per cell.
func Custom(r rune) Marker { return Marker{Kind: MarkerCustom, Rune: r} }

// String names the marker kind the way ratatui's Display does.
func (m Marker) String() string {
	switch m.Kind {
	case MarkerDot:
		return "Dot"
	case MarkerBlock:
		return "Block"
	case MarkerBar:
		return "Bar"
	case MarkerBraille:
		return "Braille"
	case MarkerHalfBlock:
		return "HalfBlock"
	case MarkerQuadrant:
		return "Quadrant"
	case MarkerSextant:
		return "Sextant"
	case MarkerOctant:
		return "Octant"
	case MarkerCustom:
		return "Custom"
	}
	return "Unknown"
}
