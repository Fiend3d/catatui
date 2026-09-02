// Port of ratatui-widgets/src/gauge.rs @ ratatui-v0.30.2

package widgets

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// Gauge is a horizontal progress bar filling its whole area.
//
// The bar is filled according to Percent or Ratio, and is as wide and as tall
// as the area it is rendered in. The label sits in the center, and defaults to
// the percentage filled. UseUnicode draws the end of the bar with block
// elements for eight extra steps per cell.
//
//	widgets.NewGauge().
//		Block(widgets.Bordered().Title("Progress")).
//		GaugeStyle(catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorBlack)).
//		Percent(20)
//
// See LineGauge for a bar one row high with the label on the left.
type Gauge struct {
	block      Block
	hasBlock   bool
	ratio      float64
	label      catatui.Span
	hasLabel   bool
	useUnicode bool
	style      catatui.Style
	gaugeStyle catatui.Style
}

// NewGauge returns an empty gauge: no block, no progress, no label.
func NewGauge() Gauge { return Gauge{} }

// Block returns a copy of g drawn inside the given block. The bar fills the
// block's inner area; the block's styles do not affect the bar.
func (g Gauge) Block(b Block) Gauge { g.block, g.hasBlock = b, true; return g }

// Percent returns a copy of g filled to the given percentage. It panics if
// percent is over 100, as ratatui does.
func (g Gauge) Percent(percent uint16) Gauge {
	if percent > 100 {
		panic("Percentage should be between 0 and 100 inclusively.")
	}
	g.ratio = float64(percent) / 100
	return g
}

// Ratio returns a copy of g filled to the given fraction, so 0.75 is three
// quarters. It panics if ratio is outside 0..=1, as ratatui does.
func (g Gauge) Ratio(ratio float64) Gauge {
	if !(ratio >= 0 && ratio <= 1) {
		panic("Ratio should be between 0 and 1 inclusively.")
	}
	g.ratio = ratio
	return g
}

// Label returns a copy of g showing the given text in the center of the bar
// instead of the percentage.
func (g Gauge) Label(label string) Gauge {
	return g.LabelSpan(catatui.NewSpan(label))
}

// LabelSpan is Label for a styled span.
func (g Gauge) LabelSpan(label catatui.Span) Gauge {
	g.label, g.hasLabel = label, true
	return g
}

// Style returns a copy of g with a style applied to everything except the bar
// itself: the block, if any, and the background. A style set on the block wins
// over it.
func (g Gauge) Style(s catatui.Style) Gauge { g.style = s; return g }

// GaugeStyle returns a copy of g with the style of the bar. Its foreground is
// the filled part and its background the unfilled part.
func (g Gauge) GaugeStyle(s catatui.Style) Gauge { g.gaugeStyle = s; return g }

// UseUnicode returns a copy of g that draws the end of the bar with block
// elements, giving eight fractional steps per cell.
func (g Gauge) UseUnicode(unicode bool) Gauge { g.useUnicode = unicode; return g }

// Render draws the gauge.
func (g Gauge) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, g.style)
	inner := area
	if g.hasBlock {
		g.block.Render(area, buf)
		inner = g.block.Inner(area)
	}
	g.renderGauge(inner, buf)
}

func (g Gauge) renderGauge(gaugeArea catatui.Rect, buf *catatui.Buffer) {
	if gaugeArea.IsEmpty() {
		return
	}

	buf.SetStyle(gaugeArea, g.gaugeStyle)

	// The label goes in the center of the area, clipped to its width.
	label := catatui.NewSpan(strconv.FormatFloat(math.Round(g.ratio*100), 'f', -1, 64) + "%")
	if g.hasLabel {
		label = g.label
	}
	clampedLabelWidth := catatui.MinU16(gaugeArea.Width, uint16(min(label.Width(), 0xFFFF)))
	labelCol := gaugeArea.Left() + (gaugeArea.Width-clampedLabelWidth)/2
	labelRow := gaugeArea.Top() + gaugeArea.Height/2

	// The bar is filled in proportion to the ratio.
	filledWidth := float64(gaugeArea.Width) * g.ratio
	var end uint16
	if g.useUnicode {
		end = gaugeArea.Left() + uint16(math.Floor(filledWidth))
	} else {
		end = gaugeArea.Left() + uint16(math.Round(filledWidth))
	}
	fg := colorOrReset(g.gaugeStyle.GetFg())
	bg := colorOrReset(g.gaugeStyle.GetBg())
	for y := gaugeArea.Top(); y < gaugeArea.Bottom(); y++ {
		for x := gaugeArea.Left(); x < end; x++ {
			// The filled part is drawn with full blocks, except under the
			// label, where the cell is a space with the colors swapped so the
			// label reads as inverted rather than vanishing into the bar.
			if x < labelCol || x > catatui.SatAdd(labelCol, clampedLabelWidth) || y != labelRow {
				buf.Get(x, y).SetSymbol(symbols.BlockFull).SetFg(fg).SetBg(bg)
			} else {
				buf.Get(x, y).SetSymbol(" ").SetFg(bg).SetBg(fg)
			}
		}
		if g.useUnicode && g.ratio < 1 {
			buf.Get(end, y).SetSymbol(getUnicodeBlock(math.Mod(filledWidth, 1)))
		}
	}
	buf.SetSpan(labelCol, labelRow, label, clampedLabelWidth)
}

// getUnicodeBlock picks the block element for the fractional cell at the end
// of a bar, rounding to the nearest eighth.
func getUnicodeBlock(frac float64) string {
	switch uint16(math.Round(frac * 8)) {
	case 1:
		return symbols.BlockOneEighth
	case 2:
		return symbols.BlockOneQuarter
	case 3:
		return symbols.BlockThreeEighths
	case 4:
		return symbols.BlockHalf
	case 5:
		return symbols.BlockFiveEighths
	case 6:
		return symbols.BlockThreeQuarters
	case 7:
		return symbols.BlockSevenEighths
	case 8:
		return symbols.BlockFull
	default:
		return " "
	}
}

// colorOrReset is ratatui's unwrap_or(Color::Reset) for an optional color.
func colorOrReset(c catatui.Color) catatui.Color {
	if c.IsSet() {
		return c
	}
	return catatui.ColorReset
}

// LineGauge is a progress bar one row high: a label on the left, then a line
// whose filled and unfilled parts are drawn in their own symbols and styles.
//
// Only the width comes from the area; the height is always one row. The label
// defaults to the percentage filled.
//
//	widgets.NewLineGauge().
//		Block(widgets.Bordered().Title("Progress")).
//		FilledStyle(catatui.NewStyle().Fg(catatui.ColorWhite).Bg(catatui.ColorBlack)).
//		FilledSymbol(symbols.ThickHorizontal).
//		Ratio(0.4)
//
// See Gauge for a taller, higher-precision bar with a centered label. Build a
// LineGauge with NewLineGauge; the zero value has no symbols to draw with.
type LineGauge struct {
	block          Block
	hasBlock       bool
	ratio          float64
	label          catatui.Line
	hasLabel       bool
	style          catatui.Style
	filledSymbol   string
	unfilledSymbol string
	filledStyle    catatui.Style
	unfilledStyle  catatui.Style
}

// NewLineGauge returns a line gauge with no progress, drawn with ─ on both the
// filled and unfilled parts.
func NewLineGauge() LineGauge {
	return LineGauge{
		filledSymbol:   symbols.Horizontal,
		unfilledSymbol: symbols.Horizontal,
	}
}

// Block returns a copy of g drawn inside the given block.
func (g LineGauge) Block(b Block) LineGauge { g.block, g.hasBlock = b, true; return g }

// Ratio returns a copy of g filled to the given fraction, so 0.75 is three
// quarters. It panics if ratio is outside 0..=1, as ratatui does.
func (g LineGauge) Ratio(ratio float64) LineGauge {
	if !(ratio >= 0 && ratio <= 1) {
		panic("Ratio should be between 0 and 1 inclusively.")
	}
	g.ratio = ratio
	return g
}

// LineSet returns a copy of g drawing both parts of the line with the set's
// horizontal character.
//
// Deprecated: ratatui deprecated line_set in 0.30.0; use FilledSymbol and
// UnfilledSymbol instead.
func (g LineGauge) LineSet(set symbols.LineSet) LineGauge {
	g.filledSymbol = set.Horizontal
	g.unfilledSymbol = set.Horizontal
	return g
}

// FilledSymbol returns a copy of g drawing the filled part with the given
// symbol.
func (g LineGauge) FilledSymbol(symbol string) LineGauge { g.filledSymbol = symbol; return g }

// UnfilledSymbol returns a copy of g drawing the unfilled part with the given
// symbol.
func (g LineGauge) UnfilledSymbol(symbol string) LineGauge { g.unfilledSymbol = symbol; return g }

// Label returns a copy of g showing the given text on the left instead of the
// percentage.
func (g LineGauge) Label(label string) LineGauge {
	return g.LabelLine(catatui.LineFromString(label))
}

// LabelLine is Label for a styled line.
func (g LineGauge) LabelLine(label catatui.Line) LineGauge {
	g.label, g.hasLabel = label, true
	return g
}

// Style returns a copy of g with a style applied to everything except the bar
// itself: the block, if any, and the background.
func (g LineGauge) Style(s catatui.Style) LineGauge { g.style = s; return g }

// GaugeStyle returns a copy of g with the bar styled the way ratatui did before
// 0.27: the style's foreground colors the filled part and its background
// colors the unfilled part, each drawn as a foreground on a reset background.
//
// Deprecated: ratatui deprecated gauge_style in 0.27.0; use FilledStyle and
// UnfilledStyle instead.
func (g LineGauge) GaugeStyle(s catatui.Style) LineGauge {
	filled := colorOrReset(s.GetFg())
	unfilled := colorOrReset(s.GetBg())
	g.filledStyle = s.Fg(filled).Bg(catatui.ColorReset)
	g.unfilledStyle = s.Fg(unfilled).Bg(catatui.ColorReset)
	return g
}

// FilledStyle returns a copy of g with the style of the filled part.
func (g LineGauge) FilledStyle(s catatui.Style) LineGauge { g.filledStyle = s; return g }

// UnfilledStyle returns a copy of g with the style of the unfilled part.
func (g LineGauge) UnfilledStyle(s catatui.Style) LineGauge { g.unfilledStyle = s; return g }

// Render draws the line gauge on the first row of the area.
func (g LineGauge) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, g.style)
	gaugeArea := area
	if g.hasBlock {
		g.block.Render(area, buf)
		gaugeArea = g.block.Inner(area)
	}
	if gaugeArea.IsEmpty() {
		return
	}

	label := catatui.LineFromString(fmt.Sprintf("%3.0f%%", g.ratio*100))
	if g.hasLabel {
		label = g.label
	}
	col, row := buf.SetLine(gaugeArea.Left(), gaugeArea.Top(), label, gaugeArea.Width)
	start := catatui.SatAdd(col, 1)
	if start >= gaugeArea.Right() {
		return
	}

	end := start + uint16(math.Floor(float64(catatui.SatSub(gaugeArea.Right(), start))*g.ratio))
	for c := start; c < end; c++ {
		buf.Get(c, row).SetSymbol(g.filledSymbol).SetStyle(g.filledStyle)
	}
	for c := end; c < gaugeArea.Right(); c++ {
		buf.Get(c, row).SetSymbol(g.unfilledSymbol).SetStyle(g.unfilledStyle)
	}
}

var (
	_ catatui.Widget = Gauge{}
	_ catatui.Widget = LineGauge{}
)
