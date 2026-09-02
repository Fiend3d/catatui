// Port of ratatui-widgets/src/barchart/bar_group.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// BarPair is a label and a value, the shorthand for building a group of bars
// when nothing needs individual styling. It stands in for ratatui's
// (&str, u64) pair conversions.
//
//	chart := widgets.NewBarChart().DataPairs(
//		widgets.BarPair{Label: "foo", Value: 1},
//		widgets.BarPair{Label: "bar", Value: 2},
//	)
type BarPair struct {
	Label string
	Value uint64
}

// BarGroup is a set of bars drawn together in a BarChart, with an optional
// label under the group. A chart holds one or more groups; see BarChart.Data.
//
//	group := widgets.NewBarGroup(
//		widgets.BarWithLabel(catatui.LineFromString("Red"), 20),
//		widgets.BarWithLabel(catatui.LineFromString("Blue"), 15),
//	).Label(catatui.LineFromString("Group 1"))
type BarGroup struct {
	label    catatui.Line
	hasLabel bool
	bars     []Bar
}

// NewBarGroup returns a group holding the given bars.
func NewBarGroup(bars ...Bar) BarGroup { return BarGroup{bars: bars} }

// BarGroupWithLabel returns a labelled group holding the given bars.
func BarGroupWithLabel(label catatui.Line, bars ...Bar) BarGroup {
	return BarGroup{label: label, hasLabel: true, bars: bars}
}

// BarGroupFromPairs returns a group with one labelled bar per pair.
func BarGroupFromPairs(pairs ...BarPair) BarGroup {
	bars := make([]Bar, 0, len(pairs))
	for _, p := range pairs {
		bars = append(bars, BarWithLabel(catatui.LineFromString(p.Label), p.Value))
	}
	return BarGroup{bars: bars}
}

// Label returns a copy of g with a label, drawn under the group in a vertical
// chart and after the group's bars in a horizontal one. The label's own
// alignment decides where in the group's width it sits.
func (g BarGroup) Label(label catatui.Line) BarGroup {
	g.label, g.hasLabel = label, true
	return g
}

// Bars returns a copy of g holding the given bars instead of its current ones.
func (g BarGroup) Bars(bars ...Bar) BarGroup {
	g.bars = append([]Bar(nil), bars...)
	return g
}

// max returns the largest bar value in the group, and false if there are no
// bars.
func (g BarGroup) max() (uint64, bool) {
	if len(g.bars) == 0 {
		return 0, false
	}
	var m uint64
	for _, b := range g.bars {
		m = max(m, b.value)
	}
	return m, true
}

// renderLabel draws the group label in area, aligned as the label asks. Only
// the label's own cells get the default label style, not the whole area.
func (g BarGroup) renderLabel(buf *catatui.Buffer, area catatui.Rect, defaultLabelStyle catatui.Style) {
	if !g.hasLabel {
		return
	}
	width := lineWidth(g.label)
	switch g.label.GetAlignment() {
	case catatui.AlignmentCenter:
		area.X = catatui.SatAdd(area.X, catatui.SatSub(area.Width, width)/2)
		area.Width = width
	case catatui.AlignmentRight:
		area.X = catatui.SatAdd(area.X, catatui.SatSub(area.Width, width))
		area.Width = width
	default:
		area.Width = width
	}
	buf.SetStyle(area, defaultLabelStyle)
	g.label.Render(area, buf)
}
