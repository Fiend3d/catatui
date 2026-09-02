// Port of ratatui-widgets/src/chart.rs @ ratatui-v0.30.2

package widgets

import (
	"fmt"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// --- Axis ----------------------------------------------------------------

// Axis is the X or Y axis of a Chart.
//
// An axis can have a title, drawn at the end of the axis — the right for an X
// axis, the top for a Y axis — and bounds, which decide what part of the data
// is visible: points outside them are not drawn. Labels are placed along the
// axis, left to right for X and bottom to top for Y.
//
//	axis := widgets.NewAxis().
//		Title("X Axis").
//		Style(catatui.NewStyle().Fg(catatui.ColorGray)).
//		Bounds([2]float64{0, 50}).
//		Labels("0", "25", "50")
//
// As in ratatui, at least two labels are needed for any of them to be drawn,
// and more than three puts the middle ones at slightly wrong positions.
type Axis struct {
	title           catatui.Line
	hasTitle        bool
	bounds          [2]float64
	labels          []catatui.Line
	style           catatui.Style
	labelsAlignment catatui.Alignment
}

// NewAxis returns an axis with no title, no labels and bounds of [0, 0], which
// is ratatui's Axis::default().
func NewAxis() Axis { return Axis{} }

// Title returns a copy of a with the given title.
func (a Axis) Title(title string) Axis { return a.TitleLine(catatui.LineFromString(title)) }

// TitleLine is Title for an already-styled line.
func (a Axis) TitleLine(title catatui.Line) Axis { a.title, a.hasTitle = title, true; return a }

// Bounds returns a copy of a with the min and max value on this axis.
func (a Axis) Bounds(bounds [2]float64) Axis { a.bounds = bounds; return a }

// Labels returns a copy of a with the labels drawn along the axis: left to
// right on an X axis, bottom to top on a Y axis.
func (a Axis) Labels(labels ...string) Axis {
	lines := make([]catatui.Line, len(labels))
	for i, s := range labels {
		lines[i] = catatui.LineFromString(s)
	}
	return a.LabelLines(lines...)
}

// LabelLines is Labels for already-styled lines. An alignment set on a label is
// ignored; the axis decides how labels are aligned.
func (a Axis) LabelLines(labels ...catatui.Line) Axis { a.labels = labels; return a }

// Style returns a copy of a with the style the axis line itself is drawn in.
func (a Axis) Style(s catatui.Style) Axis { a.style = s; return a }

// LabelsAlignment returns a copy of a with the alignment of its labels. On a Y
// axis the labels are aligned within the area to the left of the axis; on an X
// axis only the first label is affected, and it is aligned relative to the Y
// axis. AlignmentNone means left, which is ratatui's default.
func (a Axis) LabelsAlignment(alignment catatui.Alignment) Axis {
	a.labelsAlignment = alignment
	return a
}

// GetTitle returns the title and whether there is one.
func (a Axis) GetTitle() (catatui.Line, bool) { return a.title, a.hasTitle }

// GetBounds returns the axis bounds.
func (a Axis) GetBounds() [2]float64 { return a.bounds }

// GetLabels returns the axis labels.
func (a Axis) GetLabels() []catatui.Line { return a.labels }

// GetStyle returns the axis style.
func (a Axis) GetStyle() catatui.Style { return a.style }

// GetLabelsAlignment returns the alignment of the axis labels.
func (a Axis) GetLabelsAlignment() catatui.Alignment { return a.labelsAlignment }

// --- GraphType -----------------------------------------------------------

// GraphType is how a Dataset is drawn.
type GraphType uint8

const (
	// GraphTypeScatter draws each point on its own, and is the default.
	GraphTypeScatter GraphType = iota
	// GraphTypeLine draws a line between each point and the next, in the
	// order the points are given, so lines can run either way.
	GraphTypeLine
	// GraphTypeBar draws a line from the X axis up to each point.
	GraphTypeBar
	// GraphTypeArea draws a line between the points like GraphTypeLine and
	// fills the area between it and Dataset.FillToY.
	GraphTypeArea
)

// String returns the graph type's name.
func (g GraphType) String() string {
	switch g {
	case GraphTypeLine:
		return "Line"
	case GraphTypeBar:
		return "Bar"
	case GraphTypeArea:
		return "Area"
	default:
		return "Scatter"
	}
}

// ParseGraphType parses "Scatter", "Line", "Bar" or "Area".
func ParseGraphType(s string) (GraphType, error) {
	switch s {
	case "Scatter":
		return GraphTypeScatter, nil
	case "Line":
		return GraphTypeLine, nil
	case "Bar":
		return GraphTypeBar, nil
	case "Area":
		return GraphTypeArea, nil
	}
	return GraphTypeScatter, fmt.Errorf("catatui: unknown graph type %q", s)
}

// --- LegendPosition ------------------------------------------------------

// LegendPosition is where a Chart puts its legend.
//
// Deviation from ratatui: the variants are ordered so that the zero value is
// TopRight, which is ratatui's default, rather than following the Rust enum's
// declaration order.
type LegendPosition uint8

const (
	// LegendPositionTopRight puts the legend in the top-right corner. This
	// is the default.
	LegendPositionTopRight LegendPosition = iota
	// LegendPositionTop centers the legend on the top edge.
	LegendPositionTop
	// LegendPositionTopLeft puts the legend in the top-left corner.
	LegendPositionTopLeft
	// LegendPositionLeft centers the legend on the left edge.
	LegendPositionLeft
	// LegendPositionRight centers the legend on the right edge.
	LegendPositionRight
	// LegendPositionBottom centers the legend on the bottom edge.
	LegendPositionBottom
	// LegendPositionBottomRight puts the legend in the bottom-right corner.
	LegendPositionBottomRight
	// LegendPositionBottomLeft puts the legend in the bottom-left corner.
	LegendPositionBottomLeft
)

// String returns the position's name.
func (p LegendPosition) String() string {
	switch p {
	case LegendPositionTop:
		return "Top"
	case LegendPositionTopLeft:
		return "TopLeft"
	case LegendPositionLeft:
		return "Left"
	case LegendPositionRight:
		return "Right"
	case LegendPositionBottom:
		return "Bottom"
	case LegendPositionBottomRight:
		return "BottomRight"
	case LegendPositionBottomLeft:
		return "BottomLeft"
	default:
		return "TopRight"
	}
}

// layout places a legend of the given size inside area, nudging it out of the
// way of the axis titles. It reports false when there is not enough height for
// both the legend and the titles.
func (p LegendPosition) layout(area catatui.Rect, legendWidth, legendHeight, xTitleWidth, yTitleWidth uint16) (catatui.Rect, bool) {
	heightMargin := int32(area.Height) - int32(legendHeight)
	if xTitleWidth != 0 {
		heightMargin--
	}
	if yTitleWidth != 0 {
		heightMargin--
	}
	if heightMargin < 0 {
		return catatui.Rect{}, false
	}

	var x, y uint16
	switch p {
	case LegendPositionTop:
		dx := catatui.SatSub(area.Width, legendWidth) / 2
		x = catatui.SatAdd(area.Left(), dx)
		y = area.Top()
		if uint32(area.Left())+uint32(yTitleWidth) > uint32(dx) {
			y = catatui.SatAdd(area.Top(), 1)
		}
	case LegendPositionTopLeft:
		x = area.Left()
		y = area.Top()
		if yTitleWidth != 0 {
			y = catatui.SatAdd(area.Top(), 1)
		}
	case LegendPositionLeft:
		x = area.Left()
		y = catatui.SatAdd(area.Top(), centredLegendOffset(area, legendHeight, xTitleWidth, yTitleWidth))
	case LegendPositionRight:
		x = catatui.SatSub(area.Right(), legendWidth)
		y = catatui.SatAdd(area.Top(), centredLegendOffset(area, legendHeight, xTitleWidth, yTitleWidth))
	case LegendPositionBottom:
		x = catatui.SatAdd(area.Left(), catatui.SatSub(area.Width, legendWidth)/2)
		y = catatui.SatSub(area.Bottom(), legendHeight)
		if uint32(x)+uint32(legendWidth) > uint32(catatui.SatSub(area.Right(), xTitleWidth)) {
			y = catatui.SatSub(y, 1)
		}
	case LegendPositionBottomRight:
		x = catatui.SatSub(area.Right(), legendWidth)
		y = catatui.SatSub(area.Bottom(), legendHeight)
		if xTitleWidth != 0 {
			y = catatui.SatSub(y, 1)
		}
	case LegendPositionBottomLeft:
		x = area.Left()
		y = catatui.SatSub(area.Bottom(), legendHeight)
		if uint32(xTitleWidth)+uint32(legendWidth) > uint32(area.Width) {
			y = catatui.SatSub(y, 1)
		}
	default: // LegendPositionTopRight
		x = catatui.SatSub(area.Right(), legendWidth)
		y = area.Top()
		if uint32(legendWidth)+uint32(yTitleWidth) > uint32(area.Width) {
			y = catatui.SatAdd(area.Top(), 1)
		}
	}

	return catatui.NewRect(x, y, legendWidth, legendHeight), true
}

// centredLegendOffset is how far down a vertically centred legend starts, moved
// down a row for a Y axis title and up a row for an X axis title.
func centredLegendOffset(area catatui.Rect, legendHeight, xTitleWidth, yTitleWidth uint16) uint16 {
	y := catatui.SatSub(area.Height, legendHeight) / 2
	if yTitleWidth != 0 {
		y = catatui.SatAdd(y, 1)
	}
	if xTitleWidth != 0 {
		y = catatui.SatSub(y, 1)
	}
	return y
}

// --- Dataset -------------------------------------------------------------

// Dataset is one group of points in a Chart.
//
// The points are (x, y) pairs, and unlike a Rect the Y axis runs bottom to top,
// as in maths. A dataset only appears in the legend when it has a name.
//
//	dataset := widgets.NewDataset().
//		Name("dataset 1").
//		Data([][2]float64{{1, 1}, {5, 5}}).
//		Marker(symbols.Braille).
//		GraphType(widgets.GraphTypeLine).
//		Style(catatui.NewStyle().Fg(catatui.ColorRed))
type Dataset struct {
	name      catatui.Line
	hasName   bool
	data      [][2]float64
	marker    symbols.Marker
	graphType GraphType
	style     catatui.Style
	fillToY   float64
}

// NewDataset returns an empty dataset drawn with the dot marker as a scatter
// plot, which is ratatui's Dataset::default().
func NewDataset() Dataset { return Dataset{} }

// Name returns a copy of d named for the legend. A dataset without a name is
// not listed there.
func (d Dataset) Name(name string) Dataset { return d.NameLine(catatui.LineFromString(name)) }

// NameLine is Name for an already-styled line. The dataset's own style is
// patched over the name's when the legend is drawn.
func (d Dataset) NameLine(name catatui.Line) Dataset { d.name, d.hasName = name, true; return d }

// Data returns a copy of d holding the given points. The slice is used as
// given, not copied.
func (d Dataset) Data(data [][2]float64) Dataset { d.data = data; return d }

// Marker returns a copy of d drawn with the given marker. Braille needs a font
// with the Unicode braille patterns.
func (d Dataset) Marker(m symbols.Marker) Dataset { d.marker = m; return d }

// GraphType returns a copy of d drawn as a scatter, line, bar or area plot.
func (d Dataset) GraphType(g GraphType) Dataset { d.graphType = g; return d }

// Style returns a copy of d with the style used for its legend entry and its
// points. The legend uses the whole style, the points only its foreground.
func (d Dataset) Style(s catatui.Style) Dataset { d.style = s; return d }

// FillToY returns a copy of d with the Y coordinate a GraphTypeArea plot fills
// down (or up) to. The default is 0.
func (d Dataset) FillToY(y float64) Dataset { d.fillToY = y; return d }

// GetName returns the dataset name and whether there is one.
func (d Dataset) GetName() (catatui.Line, bool) { return d.name, d.hasName }

// GetData returns the dataset's points.
func (d Dataset) GetData() [][2]float64 { return d.data }

// GetMarker returns the marker the dataset is drawn with.
func (d Dataset) GetMarker() symbols.Marker { return d.marker }

// GetGraphType returns how the dataset is drawn.
func (d Dataset) GetGraphType() GraphType { return d.graphType }

// GetStyle returns the dataset style.
func (d Dataset) GetStyle() catatui.Style { return d.style }

// GetFillToY returns the level a GraphTypeArea plot fills to.
func (d Dataset) GetFillToY() float64 { return d.fillToY }

// --- Chart ---------------------------------------------------------------

// chartLayout is where each part of the chart goes, once the area has been
// divided up. Parts that do not fit are left absent.
type chartLayout struct {
	// titleX is where the X axis title starts.
	titleX    catatui.Position
	hasTitleX bool
	// titleY is where the Y axis title starts.
	titleY    catatui.Position
	hasTitleY bool
	// labelX is the row the X axis labels are drawn on.
	labelX    uint16
	hasLabelX bool
	// labelY is the column the Y axis labels start at.
	labelY    uint16
	hasLabelY bool
	// axisX is the row of the horizontal axis.
	axisX    uint16
	hasAxisX bool
	// axisY is the column of the vertical axis.
	axisY    uint16
	hasAxisY bool
	// legendArea is where the legend goes, borders included.
	legendArea catatui.Rect
	hasLegend  bool
	// graphArea is what is left for the points themselves.
	graphArea catatui.Rect
}

// Chart plots one or more Dataset in a cartesian coordinate system.
//
// Build the datasets first, then the axes, then hand both to the chart. The
// points are drawn on a Canvas, so a dataset's marker decides the resolution:
// braille gives eight subcells per cell, dots one.
//
//	chart := widgets.NewChart(
//		widgets.NewDataset().
//			Name("data").
//			Marker(symbols.Braille).
//			GraphType(widgets.GraphTypeLine).
//			Data([][2]float64{{0, 5}, {1, 6}, {1.5, 6.434}}),
//	).
//		Block(widgets.Bordered().Title("Chart")).
//		XAxis(widgets.NewAxis().Title("X").Bounds([2]float64{0, 10}).Labels("0", "5", "10")).
//		YAxis(widgets.NewAxis().Title("Y").Bounds([2]float64{0, 10}).Labels("0", "5", "10"))
//	f.RenderWidget(chart, area)
//
// Build a Chart with NewChart; the zero value draws no legend, which is
// ratatui's Chart::default().
type Chart struct {
	block              Block
	hasBlock           bool
	xAxis              Axis
	yAxis              Axis
	datasets           []Dataset
	style              catatui.Style
	hiddenLegendWidth  catatui.Constraint
	hiddenLegendHeight catatui.Constraint
	legendPosition     LegendPosition
	hasLegendPosition  bool
}

// NewChart returns a chart of the given datasets, with empty axes, the legend
// in the top-right corner and hidden when it would take more than a quarter of
// the graph in either direction.
func NewChart(datasets ...Dataset) Chart {
	return Chart{
		datasets:           datasets,
		hiddenLegendWidth:  catatui.Ratio(1, 4),
		hiddenLegendHeight: catatui.Ratio(1, 4),
		legendPosition:     LegendPositionTopRight,
		hasLegendPosition:  true,
	}
}

// Block returns a copy of c drawn inside the given block.
func (c Chart) Block(b Block) Chart { c.block, c.hasBlock = b, true; return c }

// Style returns a copy of c with a style applied to the whole area. The styles
// of the axes and the datasets take precedence over it.
func (c Chart) Style(s catatui.Style) Chart { c.style = s; return c }

// XAxis returns a copy of c with the given horizontal axis.
func (c Chart) XAxis(axis Axis) Chart { c.xAxis = axis; return c }

// YAxis returns a copy of c with the given vertical axis.
func (c Chart) YAxis(axis Axis) Chart { c.yAxis = axis; return c }

// HiddenLegendConstraints returns a copy of c with the constraints that decide
// whether the legend is shown: if it is wider than width or taller than height
// allows, it is hidden. catatui.Min always shows it, and the default is a
// quarter of the graph in each direction.
func (c Chart) HiddenLegendConstraints(width, height catatui.Constraint) Chart {
	c.hiddenLegendWidth, c.hiddenLegendHeight = width, height
	return c
}

// LegendPosition returns a copy of c with the legend in the given corner or
// edge. HiddenLegendConstraints can still hide it.
func (c Chart) LegendPosition(p LegendPosition) Chart {
	c.legendPosition, c.hasLegendPosition = p, true
	return c
}

// LegendPositionNone returns a copy of c with no legend at all, whatever
// HiddenLegendConstraints says. It is ratatui's legend_position(None).
func (c Chart) LegendPositionNone() Chart {
	c.legendPosition, c.hasLegendPosition = LegendPositionTopRight, false
	return c
}

// GetDatasets returns the chart's datasets.
func (c Chart) GetDatasets() []Dataset { return c.datasets }

// GetStyle returns the chart style.
func (c Chart) GetStyle() catatui.Style { return c.style }

// GetLegendPosition returns where the legend goes and whether there is one.
func (c Chart) GetLegendPosition() (LegendPosition, bool) {
	return c.legendPosition, c.hasLegendPosition
}

// layout works out where every part of the chart goes. Parts that do not fit
// are dropped, starting with the labels and the axes; it reports false only
// when the area has no room at all.
func (c Chart) layout(area catatui.Rect) (chartLayout, bool) {
	if area.Height == 0 || area.Width == 0 {
		return chartLayout{}, false
	}
	x := area.Left()
	y := catatui.SatSub(area.Bottom(), 1)

	var layout chartLayout

	if len(c.xAxis.labels) > 0 && y > area.Top() {
		layout.labelX, layout.hasLabelX = y, true
		y--
	}

	if len(c.yAxis.labels) > 0 {
		layout.labelY, layout.hasLabelY = x, true
	}
	x = catatui.SatAdd(x, c.maxWidthOfLabelsLeftOfYAxis(area, len(c.yAxis.labels) > 0))

	if len(c.xAxis.labels) > 0 && y > area.Top() {
		layout.axisX, layout.hasAxisX = y, true
		y--
	}

	if len(c.yAxis.labels) > 0 && x+1 < area.Right() {
		layout.axisY, layout.hasAxisY = x, true
		x++
	}

	graphWidth := catatui.SatSub(area.Right(), x)
	graphHeight := catatui.SatAdd(catatui.SatSub(y, area.Top()), 1)
	graphArea := catatui.NewRect(x, area.Top(), graphWidth, graphHeight)
	layout.graphArea = graphArea

	if c.xAxis.hasTitle {
		w := lineWidth(c.xAxis.title)
		if w < graphArea.Width && graphArea.Height > 2 {
			layout.titleX = catatui.Position{X: catatui.SatSub(catatui.SatAdd(x, graphArea.Width), w), Y: y}
			layout.hasTitleX = true
		}
	}

	if c.yAxis.hasTitle {
		w := lineWidth(c.yAxis.title)
		if w+1 < graphArea.Width && graphArea.Height > 2 {
			layout.titleY = catatui.Position{X: x, Y: area.Top()}
			layout.hasTitleY = true
		}
	}

	if c.hasLegendPosition {
		var innerWidth uint16
		var named uint16
		for _, d := range c.datasets {
			if !d.hasName {
				continue
			}
			named++
			innerWidth = catatui.MaxU16(innerWidth, lineWidth(d.name))
		}

		if named > 0 {
			legendWidth := catatui.SatAdd(innerWidth, 2)
			legendHeight := catatui.SatAdd(named, 2)

			maxLegendWidth := catatui.HorizontalLayout(c.hiddenLegendWidth).
				Flex(catatui.FlexStart).Split(graphArea)[0]
			maxLegendHeight := catatui.VerticalLayout(c.hiddenLegendHeight).
				Flex(catatui.FlexStart).Split(graphArea)[0]

			if innerWidth > 0 &&
				legendWidth <= maxLegendWidth.Width &&
				legendHeight <= maxLegendHeight.Height {
				var xTitleWidth, yTitleWidth uint16
				if layout.hasTitleX {
					xTitleWidth = lineWidth(c.xAxis.title)
				}
				if layout.hasTitleY {
					yTitleWidth = lineWidth(c.yAxis.title)
				}
				layout.legendArea, layout.hasLegend = c.legendPosition.layout(
					graphArea, legendWidth, legendHeight, xTitleWidth, yTitleWidth)
			}
		}
	}

	return layout, true
}

// maxWidthOfLabelsLeftOfYAxis is how many columns the Y axis labels — and
// whatever of the first X axis label sticks out to the left of the axis — need,
// capped at a third of the width.
func (c Chart) maxWidthOfLabelsLeftOfYAxis(area catatui.Rect, hasYAxis bool) uint16 {
	var maxWidth uint16
	for _, label := range c.yAxis.labels {
		maxWidth = catatui.MaxU16(maxWidth, lineWidth(label))
	}

	if len(c.xAxis.labels) > 0 {
		firstLabelWidth := lineWidth(c.xAxis.labels[0])
		var widthLeftOfYAxis uint16
		switch c.xAxis.labelsAlignment {
		case catatui.AlignmentCenter:
			widthLeftOfYAxis = firstLabelWidth / 2
		case catatui.AlignmentRight:
			widthLeftOfYAxis = 0
		default:
			// The last character of the label should be below the Y axis
			// when there is one, not to its left.
			var yAxisOffset uint16
			if hasYAxis {
				yAxisOffset = 1
			}
			widthLeftOfYAxis = catatui.SatSub(firstLabelWidth, yAxisOffset)
		}
		maxWidth = catatui.MaxU16(maxWidth, widthLeftOfYAxis)
	}

	// The Y axis labels and the first X axis label get at most a third of the
	// total width between them.
	return catatui.MinU16(maxWidth, area.Width/3)
}

func (c Chart) renderXLabels(buf *catatui.Buffer, layout chartLayout, chartArea, graphArea catatui.Rect) {
	if !layout.hasLabelX {
		return
	}
	y := layout.labelX
	labels := c.xAxis.labels
	if len(labels) < 2 {
		return
	}
	labelsLen := uint16(min(len(labels), 0xFFFF))

	widthBetweenTicks := graphArea.Width / labelsLen

	labelArea := c.firstXLabelArea(y, lineWidth(labels[0]), widthBetweenTicks, chartArea, graphArea)

	// The first label hangs off the left of the axis, so it is aligned the
	// other way round from the rest.
	labelAlignment := catatui.AlignmentRight
	switch c.xAxis.labelsAlignment {
	case catatui.AlignmentCenter:
		labelAlignment = catatui.AlignmentCenter
	case catatui.AlignmentRight:
		labelAlignment = catatui.AlignmentLeft
	}
	renderChartLabel(buf, labels[0], labelArea, labelAlignment)

	for i, label := range labels[1 : len(labels)-1] {
		// One column is added to x (and taken off the width below) so that
		// there is at least one space before each intermediate label.
		x := catatui.SatAdd(graphArea.Left(), catatui.SatAdd(catatui.SatMul(uint16(i+1), widthBetweenTicks), 1))
		area := catatui.NewRect(x, y, catatui.SatSub(widthBetweenTicks, 1), 1)
		renderChartLabel(buf, label, area, catatui.AlignmentCenter)
	}

	x := catatui.SatSub(graphArea.Right(), widthBetweenTicks)
	area := catatui.NewRect(x, y, widthBetweenTicks, 1)
	// The last label is right aligned so that it ends at the edge of the graph.
	renderChartLabel(buf, labels[len(labels)-1], area, catatui.AlignmentRight)
}

// firstXLabelArea is the room the first X axis label gets, which depends on how
// far left of the Y axis it is allowed to reach.
func (c Chart) firstXLabelArea(y, labelWidth, maxWidthAfterYAxis uint16, chartArea, graphArea catatui.Rect) catatui.Rect {
	var minX, maxX uint16
	switch c.xAxis.labelsAlignment {
	case catatui.AlignmentCenter:
		minX = chartArea.Left()
		maxX = catatui.SatAdd(graphArea.Left(), catatui.MinU16(maxWidthAfterYAxis, labelWidth))
	case catatui.AlignmentRight:
		minX = catatui.SatSub(graphArea.Left(), 1)
		maxX = catatui.SatAdd(graphArea.Left(), maxWidthAfterYAxis)
	default:
		minX = chartArea.Left()
		maxX = graphArea.Left()
	}

	return catatui.NewRect(minX, y, catatui.SatSub(maxX, minX), 1)
}

func (c Chart) renderYLabels(buf *catatui.Buffer, layout chartLayout, chartArea, graphArea catatui.Rect) {
	if !layout.hasLabelY {
		return
	}
	x := layout.labelY
	labels := c.yAxis.labels
	if len(labels) < 2 {
		return
	}
	labelsLen := uint16(min(len(labels), 0xFFFF))

	for i, label := range labels {
		dy := uint16(uint32(i) * uint32(catatui.SatSub(graphArea.Height, 1)) / uint32(labelsLen-1))
		if dy < graphArea.Bottom() {
			area := catatui.NewRect(
				x,
				catatui.SatSub(catatui.SatSub(graphArea.Bottom(), 1), dy),
				catatui.SatSub(catatui.SatSub(graphArea.Left(), chartArea.Left()), 1),
				1,
			)
			renderChartLabel(buf, label, area, c.yAxis.labelsAlignment)
		}
	}
}

// renderChartLabel draws a label with the alignment the axis imposes, ignoring
// whatever alignment the label itself carries. AlignmentNone means left, as in
// ratatui, where Alignment has no unset state.
func renderChartLabel(buf *catatui.Buffer, label catatui.Line, area catatui.Rect, alignment catatui.Alignment) {
	if alignment == catatui.AlignmentNone {
		alignment = catatui.AlignmentLeft
	}
	label.Alignment(alignment).Render(area, buf)
}

// Render draws the axes, the labels, the points and the legend.
func (c Chart) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, c.style)

	chartArea := area
	if c.hasBlock {
		c.block.Render(area, buf)
		chartArea = c.block.Inner(area)
	}
	layout, ok := c.layout(chartArea)
	if !ok {
		return
	}
	graphArea := layout.graphArea

	// Sample the style of the whole widget. It is used to reset the style of
	// the cells under the things drawn on top of the graph area — the legend
	// and the axis titles.
	originalStyle := buf.Get(area.Left(), area.Top()).GetStyle()

	c.renderXLabels(buf, layout, chartArea, graphArea)
	c.renderYLabels(buf, layout, chartArea, graphArea)

	if layout.hasAxisX {
		for x := graphArea.Left(); x < graphArea.Right(); x++ {
			buf.Get(x, layout.axisX).SetSymbol(symbols.Horizontal).SetStyle(c.xAxis.style)
		}
	}

	if layout.hasAxisY {
		for y := graphArea.Top(); y < graphArea.Bottom(); y++ {
			buf.Get(layout.axisY, y).SetSymbol(symbols.Vertical).SetStyle(c.yAxis.style)
		}
	}

	if layout.hasAxisX && layout.hasAxisY {
		buf.Get(layout.axisY, layout.axisX).
			SetSymbol(symbols.BottomLeft).
			SetStyle(c.xAxis.style)
	}

	background := c.style.GetBg()
	if !background.IsSet() {
		background = catatui.ColorReset
	}
	NewCanvas().
		BackgroundColor(background).
		XBounds(c.xAxis.bounds).
		YBounds(c.yAxis.bounds).
		Paint(func(ctx *Context) {
			for _, dataset := range c.datasets {
				ctx.Marker(dataset.marker)

				color := dataset.style.GetFg()
				if !color.IsSet() {
					color = catatui.ColorReset
				}
				ctx.Draw(Points{Coords: dataset.data, Color: color})

				switch dataset.graphType {
				case GraphTypeLine:
					for i := 0; i+1 < len(dataset.data); i++ {
						a, b := dataset.data[i], dataset.data[i+1]
						ctx.Draw(NewCanvasLine(a[0], a[1], b[0], b[1], color))
					}
				case GraphTypeBar:
					for _, p := range dataset.data {
						ctx.Draw(NewCanvasLine(p[0], 0, p[0], p[1], color))
					}
				case GraphTypeArea:
					for i := 0; i+1 < len(dataset.data); i++ {
						a, b := dataset.data[i], dataset.data[i+1]
						ctx.Draw(NewFilledLine(a[0], a[1], b[0], b[1], dataset.fillToY, color))
					}
				case GraphTypeScatter:
				}
			}
		}).
		Render(graphArea, buf)

	if layout.hasTitleX {
		title := c.xAxis.title
		pos := layout.titleX
		width := catatui.MinU16(catatui.SatSub(graphArea.Right(), pos.X), lineWidth(title))
		buf.SetStyle(catatui.NewRect(pos.X, pos.Y, width, 1), originalStyle)
		buf.SetLine(pos.X, pos.Y, title, width)
	}

	if layout.hasTitleY {
		title := c.yAxis.title
		pos := layout.titleY
		width := catatui.MinU16(catatui.SatSub(graphArea.Right(), pos.X), lineWidth(title))
		buf.SetStyle(catatui.NewRect(pos.X, pos.Y, width, 1), originalStyle)
		buf.SetLine(pos.X, pos.Y, title, width)
	}

	if layout.hasLegend {
		legendArea := layout.legendArea
		buf.SetStyle(legendArea, originalStyle)
		Bordered().Render(legendArea, buf)

		row := uint16(0)
		for _, dataset := range c.datasets {
			if !dataset.hasName {
				continue
			}
			name := dataset.name.Patch(dataset.style)
			name.Render(catatui.NewRect(
				catatui.SatAdd(legendArea.X, 1),
				catatui.SatAdd(catatui.SatAdd(legendArea.Y, 1), row),
				catatui.SatSub(legendArea.Width, 2),
				1,
			), buf)
			row++
		}
	}
}

var _ catatui.Widget = Chart{}
