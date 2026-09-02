// Port of ratatui-widgets/src/sparkline.rs @ ratatui-v0.30.2

package widgets

import (
	"math/bits"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// RenderDirection is which way a Sparkline lays out its data.
type RenderDirection uint8

const (
	// RenderLeftToRight puts the first value on the left, and is the default.
	RenderLeftToRight RenderDirection = iota
	// RenderRightToLeft puts the first value on the right.
	RenderRightToLeft
)

// String names the direction the way ratatui's Display does.
func (d RenderDirection) String() string {
	switch d {
	case RenderLeftToRight:
		return "LeftToRight"
	case RenderRightToLeft:
		return "RightToLeft"
	}
	return "Unknown"
}

// SparklineBar is one bar of a Sparkline: a value, or the absence of one, plus
// an optional style of its own.
//
// An absent bar is distinct from a bar with value zero: it is drawn with the
// sparkline's absent-value symbol and style over the full height.
type SparklineBar struct {
	value    uint64
	hasValue bool
	style    catatui.Style
	hasStyle bool
}

// NewSparklineBar returns a bar with the given value and no style of its own.
func NewSparklineBar(value uint64) SparklineBar {
	return SparklineBar{value: value, hasValue: true}
}

// AbsentSparklineBar returns a bar with no value.
func AbsentSparklineBar() SparklineBar { return SparklineBar{} }

// Style returns a copy of b with its own style, which is layered over the
// sparkline's style when the bar is drawn.
func (b SparklineBar) Style(style catatui.Style) SparklineBar {
	b.style, b.hasStyle = style, true
	return b
}

// Value returns the bar's value, and whether it has one.
func (b SparklineBar) Value() (uint64, bool) { return b.value, b.hasValue }

// GetStyle returns the bar's own style, and whether it has one.
func (b SparklineBar) GetStyle() (catatui.Style, bool) { return b.style, b.hasStyle }

// Sparkline draws a dataset as a row of bars, one column per value, using the
// eighth-block characters so that a bar's height is resolved to an eighth of
// a cell.
//
//	widgets.NewSparkline().
//		Block(widgets.Bordered().Title("Sparkline")).
//		Data(0, 2, 3, 4, 1, 4, 10).
//		Max(5).
//		Direction(widgets.RenderRightToLeft).
//		Style(catatui.NewStyle().Fg(catatui.ColorRed))
//
// The style's foreground colors the bars and its background everything else.
// A bar given its own style through SparklineBar.Style is drawn in the
// sparkline's style patched with its own. Bars with no value are drawn with
// the absent-value symbol and style over the full height.
type Sparkline struct {
	block                Block
	hasBlock             bool
	style                catatui.Style
	absentValueStyle     catatui.Style
	absentValueSymbol    string
	hasAbsentValueSymbol bool
	data                 []SparklineBar
	max                  uint64
	hasMax               bool
	barSet               symbols.LevelSet
	hasBarSet            bool
	direction            RenderDirection
}

// NewSparkline returns an empty sparkline, drawn left to right with the
// nine-level bar set.
func NewSparkline() Sparkline { return Sparkline{} }

// Block returns a copy of s drawn inside the given block.
func (s Sparkline) Block(b Block) Sparkline { s.block, s.hasBlock = b, true; return s }

// Style returns a copy of s with the given style. Its foreground colors the
// bars and its background everything else.
func (s Sparkline) Style(style catatui.Style) Sparkline { s.style = style; return s }

// AbsentValueStyle returns a copy of s with the style used for bars that have
// no value.
func (s Sparkline) AbsentValueStyle(style catatui.Style) Sparkline {
	s.absentValueStyle = style
	return s
}

// AbsentValueSymbol returns a copy of s with the symbol used for bars that
// have no value. The default is symbols.ShadeEmpty.
func (s Sparkline) AbsentValueSymbol(symbol string) Sparkline {
	s.absentValueSymbol, s.hasAbsentValueSymbol = symbol, true
	return s
}

// Data returns a copy of s showing the given values, one bar each.
func (s Sparkline) Data(values ...uint64) Sparkline {
	bars := make([]SparklineBar, len(values))
	for i, v := range values {
		bars[i] = NewSparklineBar(v)
	}
	s.data = bars
	return s
}

// DataBars returns a copy of s showing the given bars, which may be absent or
// carry their own style.
func (s Sparkline) DataBars(bars ...SparklineBar) Sparkline {
	s.data = append([]SparklineBar(nil), bars...)
	return s
}

// Max returns a copy of s with the value that fills a bar. Without it, the
// largest value in the data does.
func (s Sparkline) Max(max uint64) Sparkline { s.max, s.hasMax = max, true; return s }

// BarSet returns a copy of s drawn with the given characters, such as
// symbols.BarThreeLevels for fonts without the finer block elements.
func (s Sparkline) BarSet(set symbols.LevelSet) Sparkline {
	s.barSet, s.hasBarSet = set, true
	return s
}

// Direction returns a copy of s laid out in the given direction.
func (s Sparkline) Direction(d RenderDirection) Sparkline { s.direction = d; return s }

func (s Sparkline) set() symbols.LevelSet {
	if s.hasBarSet {
		return s.barSet
	}
	return symbols.BarNineLevels
}

func (s Sparkline) absentSymbol() string {
	if s.hasAbsentValueSymbol {
		return s.absentValueSymbol
	}
	return symbols.ShadeEmpty
}

// Render draws the block, if any, then the bars inside it.
func (s Sparkline) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	inner := area
	if s.hasBlock {
		s.block.Render(area, buf)
		inner = s.block.Inner(area)
	}
	s.renderSparkline(inner, buf)
}

func (s Sparkline) renderSparkline(sparkArea catatui.Rect, buf *catatui.Buffer) {
	if sparkArea.IsEmpty() {
		return
	}

	// The maximum height across all bars.
	maxHeight := s.max
	if !s.hasMax {
		found := false
		for _, bar := range s.data {
			if bar.hasValue && (!found || bar.value > maxHeight) {
				maxHeight, found = bar.value, true
			}
		}
		if !found {
			maxHeight = 1
		}
	}

	// The number of bars that fit.
	maxIndex := min(int(sparkArea.Width), len(s.data))

	for i, item := range s.data[:maxIndex] {
		var x uint16
		switch s.direction {
		case RenderRightToLeft:
			x = sparkArea.Right() - uint16(i) - 1
		default:
			x = sparkArea.Left() + uint16(i)
		}

		// A bar with a value is as tall as its value scaled to the area, and
		// its symbol follows from the height left on each row. An absent bar
		// is full height and uses the absent-value symbol on every row.
		var (
			height    uint64
			symbol    string
			hasSymbol bool
			style     catatui.Style
		)
		if item.hasValue {
			height = scaleHeight(item.value, maxHeight, sparkArea.Height)
			style = item.style
		} else {
			height = uint64(sparkArea.Height) * 8
			symbol, hasSymbol = s.absentSymbol(), true
			style = s.absentValueStyle
		}

		// Rows are drawn bottom up, each taking eight ticks off the height.
		for j := int(sparkArea.Height) - 1; j >= 0; j-- {
			sym := symbol
			if !hasSymbol {
				sym = s.symbolForHeight(height)
			}
			if height > 8 {
				height -= 8
			} else {
				height = 0
			}
			buf.Get(x, sparkArea.Top()+uint16(j)).
				SetSymbol(sym).
				SetStyle(s.style.Patch(style))
		}
	}
}

func (s Sparkline) symbolForHeight(height uint64) string {
	set := s.set()
	switch height {
	case 0:
		return set.Empty
	case 1:
		return set.OneEighth
	case 2:
		return set.OneQuarter
	case 3:
		return set.ThreeEighths
	case 4:
		return set.Half
	case 5:
		return set.FiveEighths
	case 6:
		return set.ThreeQuarters
	case 7:
		return set.SevenEighths
	default:
		return set.Full
	}
}

// scaleHeight converts a value to a number of eighth-cell ticks out of the
// area's height, keeping integer precision for values near the top of the
// uint64 range, as ratatui does by computing in u128.
func scaleHeight(value, max uint64, maxHeight uint16) uint64 {
	if max == 0 {
		return 0
	}
	maxTicks := uint64(maxHeight) * 8
	hi, lo := bits.Mul64(value, maxTicks)
	if hi >= max {
		// The quotient would not fit in 64 bits, so it is certainly past the cap.
		return maxTicks
	}
	ticks, _ := bits.Div64(hi, lo, max)
	return min(ticks, maxTicks)
}

var _ catatui.Widget = Sparkline{}
