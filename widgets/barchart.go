// Port of ratatui-widgets/src/barchart.rs @ ratatui-v0.30.2

package widgets

import (
	"math/bits"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// BarChart draws values as bars, vertically by default or horizontally with
// Direction.
//
//	┌─────────────────────────────────┐
//	│                             ████│
//	│                        ▅▅▅▅ ████│
//	│            ▇▇▇▇        ████ ████│
//	│     ▄▄▄▄   ████ ████   ████ ████│
//	│▆10▆ █20█   █50█ █40█   █60█ █90█│
//	│ B1   B2     B1   B2     B1   B2 │
//	│ Group1      Group2      Group3  │
//	└─────────────────────────────────┘
//
// Bars come in groups (see BarGroup); each call to Data adds one group, and
// DataPairs is the shorthand for a group of plainly labelled bars. Bars can be
// styled for the whole chart (BarStyle, ValueStyle, LabelStyle) or one at a
// time (Bar.Style and friends), with the bar's own style layered on top.
//
//	chart := widgets.NewBarChart().
//		Block(widgets.Bordered().Title("BarChart")).
//		BarWidth(3).
//		BarGap(1).
//		GroupGap(3).
//		BarStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
//		DataPairs(widgets.BarPair{Label: "B0", Value: 0}, widgets.BarPair{Label: "B1", Value: 2}).
//		Data(widgets.NewBarGroup(
//			widgets.BarWithLabel(catatui.LineFromString("A"), 10),
//			widgets.BarWithLabel(catatui.LineFromString("B"), 20)).
//			Label(catatui.LineFromString("Group 2"))).
//		Max(4)
//
// Use NewBarChart, not the zero value: the zero value has a bar width of 0 and
// draws nothing.
type BarChart struct {
	block      Block
	hasBlock   bool
	barWidth   uint16
	barGap     uint16
	groupGap   uint16
	barSet     symbols.LevelSet
	barStyle   catatui.Style
	valueStyle catatui.Style
	labelStyle catatui.Style
	style      catatui.Style
	data       []BarGroup
	max        uint64
	hasMax     bool
	direction  catatui.Direction
}

// NewBarChart returns an empty vertical chart with bars one cell wide, one
// cell apart, and no gap between groups.
func NewBarChart() BarChart {
	return BarChart{
		barWidth:  1,
		barGap:    1,
		barSet:    symbols.BarNineLevels,
		direction: catatui.Vertical,
	}
}

// VerticalBarChart returns a vertical chart holding one group of bars. An
// empty group is dropped.
func VerticalBarChart(bars ...Bar) BarChart {
	c := NewBarChart()
	c.data = nonEmptyGroups([]BarGroup{NewBarGroup(bars...)})
	return c
}

// HorizontalBarChart returns a horizontal chart holding one group of bars.
// An empty group is dropped.
func HorizontalBarChart(bars ...Bar) BarChart {
	c := NewBarChart()
	c.data = nonEmptyGroups([]BarGroup{NewBarGroup(bars...)})
	c.direction = catatui.Horizontal
	return c
}

// GroupedBarChart returns a vertical chart holding the given groups. Empty
// groups are dropped.
func GroupedBarChart(groups ...BarGroup) BarChart {
	c := NewBarChart()
	c.data = nonEmptyGroups(groups)
	return c
}

// nonEmptyGroups returns the groups that have at least one bar.
func nonEmptyGroups(groups []BarGroup) []BarGroup {
	var out []BarGroup
	for _, g := range groups {
		if len(g.bars) != 0 {
			out = append(out, g)
		}
	}
	return out
}

// Data returns a copy of c with a group of bars added. A group with no bars
// is ignored, label and all.
func (c BarChart) Data(group BarGroup) BarChart {
	if len(group.bars) == 0 {
		return c
	}
	c.data = append(append([]BarGroup(nil), c.data...), group)
	return c
}

// DataPairs returns a copy of c with a group of plainly labelled bars added.
// It is Data(BarGroupFromPairs(pairs...)).
func (c BarChart) DataPairs(pairs ...BarPair) BarChart {
	return c.Data(BarGroupFromPairs(pairs...))
}

// Block returns a copy of c drawn inside the given block.
func (c BarChart) Block(b Block) BarChart { c.block, c.hasBlock = b, true; return c }

// Max returns a copy of c where a bar needs the given value to reach full
// length. Without it, the largest value in the data is full length.
func (c BarChart) Max(max uint64) BarChart { c.max, c.hasMax = max, true; return c }

// BarStyle returns a copy of c with the default style for bars, beneath each
// bar's own style.
func (c BarChart) BarStyle(s catatui.Style) BarChart { c.barStyle = s; return c }

// BarWidth returns a copy of c with bars the given number of cells wide. For
// horizontal bars this is their height. It is also the width a bar's label is
// truncated to. The default is 1; a width of 0 draws nothing.
func (c BarChart) BarWidth(width uint16) BarChart { c.barWidth = width; return c }

// BarGap returns a copy of c with the given number of cells between bars. The
// default is 1.
func (c BarChart) BarGap(gap uint16) BarChart { c.barGap = gap; return c }

// GroupGap returns a copy of c with the given number of cells between groups.
// The default is 0, which in a horizontal chart also leaves no row for group
// labels.
func (c BarChart) GroupGap(gap uint16) BarChart { c.groupGap = gap; return c }

// BarSet returns a copy of c drawing bars with the given level set. The
// default is symbols.BarNineLevels; vertical charts want a bar set and
// horizontal charts a block set.
func (c BarChart) BarSet(set symbols.LevelSet) BarChart { c.barSet = set; return c }

// ValueStyle returns a copy of c with the default style for the values drawn
// in bars, beneath each bar's own value style.
func (c BarChart) ValueStyle(s catatui.Style) BarChart { c.valueStyle = s; return c }

// LabelStyle returns a copy of c with the default style for bar and group
// labels, beneath each label's own style.
func (c BarChart) LabelStyle(s catatui.Style) BarChart { c.labelStyle = s; return c }

// Style returns a copy of c with a style applied to the whole area, beneath
// everything the chart draws.
func (c BarChart) Style(s catatui.Style) BarChart { c.style = s; return c }

// Direction returns a copy of c drawing its bars along the given axis.
// catatui.Vertical, the default, grows bars upward with labels under them;
// catatui.Horizontal grows them rightward with labels to their left.
func (c BarChart) Direction(d catatui.Direction) BarChart { c.direction = d; return c }

// labelInfo says which label rows a vertical chart draws.
type labelInfo struct {
	groupLabelVisible bool
	barLabelVisible   bool
	// height is the number of label rows: 0, 1 or 2.
	height uint16
}

// groupTicks returns the length of every visible bar in ticks, eight to a
// cell, one slice per visible group. availableSpace is the width (vertical) or
// height (horizontal) bars are laid out along, and barMaxLength the length of
// a bar at the maximum value. A group that does not fit whole is cut to the
// bars that do, and nothing after it is drawn.
func (c BarChart) groupTicks(availableSpace, barMaxLength uint16) [][]uint64 {
	maxValue := c.maximumDataValue()
	space := availableSpace
	var out [][]uint64
	for _, group := range c.data {
		if space == 0 {
			break
		}
		nBars := uint16(min(len(group.bars), 0xFFFF))
		groupWidth := catatui.SatAdd(catatui.SatMul(nBars, c.barWidth),
			catatui.SatMul(catatui.SatSub(nBars, 1), c.barGap))

		var n uint16
		if space > groupWidth {
			space = catatui.SatSub(space, catatui.SatAdd(groupWidth, catatui.SatAdd(c.groupGap, c.barGap)))
			n = nBars
		} else {
			maxBars := catatui.SatAdd(space, c.barGap) / catatui.SatAdd(c.barWidth, c.barGap)
			if maxBars == 0 {
				break
			}
			space = 0
			n = maxBars
		}

		visible := group.bars[:min(int(n), len(group.bars))]
		ticks := make([]uint64, 0, len(visible))
		for _, bar := range visible {
			ticks = append(ticks, scaleTicks(bar.value, maxValue, barMaxLength))
		}
		out = append(out, ticks)
	}
	return out
}

// scaleTicks converts a value to a bar length in ticks, in exact integer
// arithmetic so that values near the top of uint64 keep their precision.
func scaleTicks(value, maxValue uint64, maxLength uint16) uint64 {
	maxTicks := uint64(maxLength) * 8
	hi, lo := bits.Mul64(value, maxTicks)
	if hi >= maxValue {
		// The quotient would not fit in 64 bits, so it is past maxTicks.
		return maxTicks
	}
	ticks, _ := bits.Div64(hi, lo, maxValue)
	return min(ticks, maxTicks)
}

// labelInfo decides which label rows fit in the given height under the bars.
// Bar labels take precedence over group labels when there is only one row.
func (c BarChart) labelInfo(availableHeight uint16) labelInfo {
	if availableHeight == 0 {
		return labelInfo{}
	}

	barLabelVisible := false
	for _, g := range c.data {
		for _, b := range g.bars {
			if b.hasLabel {
				barLabelVisible = true
			}
		}
	}
	if availableHeight == 1 && barLabelVisible {
		return labelInfo{barLabelVisible: true, height: 1}
	}

	groupLabelVisible := false
	for _, g := range c.data {
		if g.hasLabel {
			groupLabelVisible = true
		}
	}
	var height uint16
	if groupLabelVisible {
		height++
	}
	if barLabelVisible {
		height++
	}
	return labelInfo{
		groupLabelVisible: groupLabelVisible,
		barLabelVisible:   barLabelVisible,
		height:            height,
	}
}

func (c BarChart) renderHorizontal(buf *catatui.Buffer, area catatui.Rect) {
	// The label column is as wide as the longest bar label.
	var labelSize uint16
	for _, g := range c.data {
		for _, b := range g.bars {
			if b.hasLabel {
				labelSize = catatui.MaxU16(labelSize, lineWidth(b.label))
			}
		}
	}

	labelX := area.X
	var margin uint16
	if labelSize != 0 {
		margin = 1
	}
	barsArea := area
	barsArea.X = catatui.SatAdd(area.X, catatui.SatAdd(labelSize, margin))
	barsArea.Width = catatui.SatSub(catatui.SatSub(area.Width, labelSize), margin)

	groupTicks := c.groupTicks(barsArea.Height, barsArea.Width)

	// Draw every visible bar with its label and value.
	barY := barsArea.Top()
	for gi, ticksVec := range groupTicks {
		group := c.data[gi]
		for bi, ticks := range ticksVec {
			bar := group.bars[bi]
			barLength := uint16(ticks / 8)
			barStyle := c.barStyle.Patch(bar.style)

			for y := uint16(0); y < c.barWidth; y++ {
				rowY := catatui.SatAdd(barY, y)
				for x := uint16(0); x < barsArea.Width; x++ {
					symbol := c.barSet.Empty
					if x < barLength {
						symbol = c.barSet.Full
					}
					buf.Get(catatui.SatAdd(barsArea.Left(), x), rowY).SetSymbol(symbol).SetStyle(barStyle)
				}
			}

			valueArea := barsArea
			valueArea.Y = catatui.SatAdd(barY, c.barWidth>>1)

			if bar.hasLabel {
				buf.SetLine(labelX, valueArea.Top(), bar.label, labelSize)
			}

			bar.renderValueWithDifferentStyles(buf, valueArea, int(barLength), c.valueStyle, c.barStyle)

			barY = catatui.SatAdd(barY, catatui.SatAdd(c.barGap, c.barWidth))
		}

		// With no group gap there is no row for the group label. The label is
		// also skipped once it would fall below the visible area.
		labelY := catatui.SatSub(barY, c.barGap)
		if c.groupGap > 0 && labelY < barsArea.Bottom() {
			labelRect := barsArea
			labelRect.Y = labelY
			group.renderLabel(buf, labelRect, c.labelStyle)
			barY = catatui.SatAdd(barY, c.groupGap)
		}
	}
}

func (c BarChart) renderVertical(buf *catatui.Buffer, area catatui.Rect) {
	info := c.labelInfo(catatui.SatSub(area.Height, 1))

	barsArea := area
	barsArea.Height = catatui.SatSub(area.Height, info.height)

	groupTicks := c.groupTicks(barsArea.Width, barsArea.Height)
	c.renderVerticalBars(barsArea, buf, groupTicks)
	c.renderLabelsAndValues(area, buf, info, groupTicks)
}

// levelSymbol is the character for a cell of a bar with the given number of
// ticks left to draw.
func levelSymbol(set symbols.LevelSet, ticks uint64) string {
	switch ticks {
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

// renderVerticalBars draws the bars themselves, bottom row up, without labels
// or values.
func (c BarChart) renderVerticalBars(area catatui.Rect, buf *catatui.Buffer, groupTicks [][]uint64) {
	barX := area.Left()
	for gi, ticksVec := range groupTicks {
		group := c.data[gi]
		for bi, ticks := range ticksVec {
			bar := group.bars[bi]
			barStyle := c.barStyle.Patch(bar.style)
			for j := int(area.Height) - 1; j >= 0; j-- {
				symbol := levelSymbol(c.barSet, ticks)
				for x := uint16(0); x < c.barWidth; x++ {
					buf.Get(catatui.SatAdd(barX, x), catatui.SatAdd(area.Top(), uint16(j))).
						SetSymbol(symbol).SetStyle(barStyle)
				}
				if ticks < 8 {
					ticks = 0
				} else {
					ticks -= 8
				}
			}
			barX = catatui.SatAdd(barX, catatui.SatAdd(c.barGap, c.barWidth))
		}
		barX = catatui.SatAdd(barX, c.groupGap)
	}
}

// maximumDataValue is the value a full-length bar stands for: Max if set,
// otherwise the largest value in the data, and never less than 1.
func (c BarChart) maximumDataValue() uint64 {
	m := c.max
	if !c.hasMax {
		m = 0
		for _, g := range c.data {
			gm, _ := g.max()
			m = max(m, gm)
		}
	}
	return max(m, 1)
}

// renderLabelsAndValues draws the values into the bottom row of the bars,
// then the bar labels and group labels in the rows below.
func (c BarChart) renderLabelsAndValues(area catatui.Rect, buf *catatui.Buffer, info labelInfo, groupTicks [][]uint64) {
	barX := area.Left()
	barY := catatui.SatSub(catatui.SatSub(area.Bottom(), info.height), 1)
	for gi, ticksVec := range groupTicks {
		group := c.data[gi]
		if len(group.bars) == 0 {
			continue
		}
		// The group label goes under the bars, or under the bar labels.
		if info.groupLabelVisible {
			labelMaxWidth := catatui.SatSub(
				catatui.SatMul(uint16(min(len(ticksVec), 0xFFFF)), catatui.SatAdd(c.barWidth, c.barGap)),
				c.barGap)
			groupArea := catatui.Rect{
				X:      barX,
				Y:      catatui.SatSub(area.Bottom(), 1),
				Width:  labelMaxWidth,
				Height: 1,
			}
			group.renderLabel(buf, groupArea, c.labelStyle)
		}

		for bi, ticks := range ticksVec {
			bar := group.bars[bi]
			if info.barLabelVisible {
				bar.renderLabel(buf, c.barWidth, barX, catatui.SatAdd(barY, 1), c.labelStyle)
			}
			bar.renderValue(buf, c.barWidth, barX, barY, c.valueStyle, ticks)
			barX = catatui.SatAdd(barX, catatui.SatAdd(c.barGap, c.barWidth))
		}
		barX = catatui.SatAdd(barX, c.groupGap)
	}
}

// Render draws the chart.
func (c BarChart) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, c.style)

	inner := area
	if c.hasBlock {
		c.block.Render(area, buf)
		inner = c.block.Inner(area)
	}
	if inner.IsEmpty() || len(c.data) == 0 || c.barWidth == 0 {
		return
	}

	switch c.direction {
	case catatui.Horizontal:
		c.renderHorizontal(buf, inner)
	default:
		c.renderVertical(buf, inner)
	}
}

var _ catatui.Widget = BarChart{}
