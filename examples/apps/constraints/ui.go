// Port of the rendering half of examples/apps/constraints @ ratatui-v0.30.2

package main

import (
	"fmt"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// Each example is drawn four rows high, with nothing between them.
const (
	illustrationHeight uint16 = 4
	spacerHeight       uint16 = 0
	exampleHeight             = illustrationHeight + spacerHeight
)

// The colours run by the priority the solver gives each kind: Min and Max
// first, then Length, Percentage and Ratio, then Fill last.
var (
	minColor        = tailwind.Blue.C900
	maxColor        = tailwind.Blue.C800
	lengthColor     = tailwind.Slate.C700
	percentageColor = tailwind.Slate.C800
	ratioColor      = tailwind.Slate.C900
	fillColor       = tailwind.Slate.C950
	titleColor      = tailwind.Slate.C200
)

// tab is one kind of constraint, in the order the tabs appear.
type tab int

const (
	tabMin tab = iota
	tabMax
	tabLength
	tabPercentage
	tabRatio
	tabFill
)

// tabs is every tab with its title colour and the cases it stacks up. Each
// case is one row of constraints laid out across the width.
var tabs = []struct {
	name     string
	color    catatui.Color
	examples [][]catatui.Constraint
}{
	tabMin: {"Min", minColor, [][]catatui.Constraint{
		{catatui.Percentage(100), catatui.Min(0)},
		{catatui.Percentage(100), catatui.Min(20)},
		{catatui.Percentage(100), catatui.Min(40)},
		{catatui.Percentage(100), catatui.Min(60)},
		{catatui.Percentage(100), catatui.Min(80)},
	}},
	tabMax: {"Max", maxColor, [][]catatui.Constraint{
		{catatui.Percentage(0), catatui.Max(0)},
		{catatui.Percentage(0), catatui.Max(20)},
		{catatui.Percentage(0), catatui.Max(40)},
		{catatui.Percentage(0), catatui.Max(60)},
		{catatui.Percentage(0), catatui.Max(80)},
	}},
	tabLength: {"Length", lengthColor, [][]catatui.Constraint{
		{catatui.Length(20), catatui.Length(20)},
		{catatui.Length(20), catatui.Min(20)},
		{catatui.Length(20), catatui.Max(20)},
	}},
	tabPercentage: {"Percentage", percentageColor, [][]catatui.Constraint{
		{catatui.Percentage(75), catatui.Fill(0)},
		{catatui.Percentage(25), catatui.Fill(0)},
		{catatui.Percentage(50), catatui.Min(20)},
		{catatui.Percentage(0), catatui.Max(0)},
		{catatui.Percentage(0), catatui.Fill(0)},
	}},
	tabRatio: {"Ratio", ratioColor, [][]catatui.Constraint{
		{catatui.Ratio(1, 2), catatui.Ratio(1, 2)},
		{catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4), catatui.Ratio(1, 4)},
		{catatui.Ratio(1, 2), catatui.Ratio(1, 3), catatui.Ratio(1, 4)},
		{catatui.Ratio(1, 2), catatui.Percentage(25), catatui.Length(10)},
	}},
	tabFill: {"Fill", fillColor, [][]catatui.Constraint{
		{catatui.Fill(1), catatui.Fill(2), catatui.Fill(3)},
		{catatui.Fill(1), catatui.Percentage(50), catatui.Fill(1)},
	}},
}

// Render draws the tab bar, the axis and the examples under it.
func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Length(3),
		catatui.Fill(0),
	).Split(area)

	a.renderTabs(rows[0], buf)
	renderAxis(rows[1], buf)
	a.renderDemo(rows[2], buf)
}

func (a *app) renderTabs(area catatui.Rect, buf *catatui.Buffer) {
	titles := make([]catatui.Line, len(tabs))
	for i := range tabs {
		titles[i] = catatui.LineFromStyledString("  "+tabs[i].name+"  ",
			catatui.NewStyle().Fg(titleColor).Bg(tabs[i].color))
	}

	block := widgets.NewBlock().
		TitleLine(catatui.LineFromStyledString("Constraints ",
			catatui.NewStyle().AddModifier(catatui.ModifierBold))).
		TitleLine(catatui.LineFromString(
			" Use h l or ◄ ► to change tab and j k or ▲ ▼  to scroll"))

	widgets.NewTabsFromLines(titles...).
		Block(block).
		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
		Select(int(a.selectedTab)).
		Divider(" ").
		Padding("", "").
		Render(area, buf)
}

// renderAxis draws a bar like <----- 80 px ----->, measuring the width the
// examples are laid out in.
func renderAxis(area catatui.Rect, buf *catatui.Buffer) {
	label := fmt.Sprintf("%d px", area.Width)
	// ratatui pads to the width less half the label, which is what leaves room
	// for the arrows at the ends and keeps the dashes even.
	bar := centerPad(label, int(catatui.SatSub(area.Width, uint16(len(label)/2))), '-')

	widgets.NewParagraph("<"+bar+">").
		Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)).
		Centered().
		Block(widgets.NewBlock().Padding(widgets.TopPadding(1))).
		Render(area, buf)
}

// centerPad centres s in a field of the given width, padded with fill, which is
// the Rust format {:-^width$}.
func centerPad(s string, width int, fill rune) string {
	pad := width - catatui.StringWidth(s)
	if pad <= 0 {
		return s
	}
	left := pad / 2
	return strings.Repeat(string(fill), left) + s + strings.Repeat(string(fill), pad-left)
}

// renderDemo draws the examples into a buffer of their own and copies the part
// that is scrolled into view, which is far simpler than teaching every example
// where the viewport starts.
func (a *app) renderDemo(area catatui.Rect, buf *catatui.Buffer) {
	if area.IsEmpty() {
		return
	}

	examples := tabs[a.selectedTab].examples
	// The extra area.Height keeps the last example whole when the scroll is at
	// its end.
	height := uint16(len(examples))*exampleHeight + area.Height
	demoArea := catatui.NewRect(0, 0, area.Width, height)
	demoBuf := catatui.NewBuffer(demoArea)

	scrollbarNeeded := a.scrollOffset != 0 || height > area.Height
	contentArea := demoArea
	if scrollbarNeeded {
		contentArea.Width = catatui.SatSub(contentArea.Width, 1)
	}
	renderExamples(examples, contentArea, demoBuf)

	// Cells are row-major, so skipping whole rows is skipping width*offset of
	// them.
	skip := int(area.Width) * int(a.scrollOffset)
	visible := demoBuf.Content
	if skip < len(visible) {
		visible = visible[skip:]
	} else {
		visible = nil
	}
	for i, cell := range visible {
		if uint32(i) >= area.Area() {
			break
		}
		x := uint16(i) % area.Width
		y := uint16(i) / area.Width
		*buf.Get(area.X+x, area.Y+y) = cell
	}

	if scrollbarNeeded {
		state := widgets.NewScrollbarState(int(a.maxScrollOffset())).
			Position(int(a.scrollOffset))
		catatui.RenderStatefulWidget(
			widgets.NewScrollbar(widgets.ScrollbarVerticalRight),
			area.Intersection(buf.Area), buf, &state)
	}
}

// renderExamples stacks the tab's cases, one four-row band each.
func renderExamples(examples [][]catatui.Constraint, area catatui.Rect, buf *catatui.Buffer) {
	heights := make([]catatui.Constraint, len(examples))
	for i := range heights {
		heights[i] = catatui.Length(exampleHeight)
	}
	areas := catatui.VerticalLayout(heights...).Flex(catatui.FlexStart).Split(area)

	for i, constraints := range examples {
		renderExample(constraints, areas[i], buf)
	}
}

// renderExample lays the constraints out across the width and labels what each
// one asked for and what it got.
func renderExample(constraints []catatui.Constraint, area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(illustrationHeight),
		catatui.Length(spacerHeight),
	).Split(area)

	blocks := catatui.HorizontalLayout(constraints...).Split(rows[0])
	for i, constraint := range constraints {
		illustration(constraint, blocks[i].Width).Render(blocks[i], buf)
	}
}

// illustration is one block: what the constraint asked for, above how many
// columns it was given.
func illustration(constraint catatui.Constraint, width uint16) widgets.Paragraph {
	color := colorForConstraint(constraint)
	block := widgets.Bordered().
		BorderSet(symbols.BorderQuadrantOutside).
		BorderStyle(catatui.ResetStyle().Fg(color).AddModifier(catatui.ModifierReversed)).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite).Bg(color))

	text := fmt.Sprintf("%v\n%d px", constraint, width)
	return widgets.NewParagraph(text).Centered().Block(block)
}

// colorForConstraint gives each kind its own colour, which is what makes the
// illustrations readable at a glance.
func colorForConstraint(constraint catatui.Constraint) catatui.Color {
	switch constraint.Kind() {
	case catatui.ConstraintMin:
		return minColor
	case catatui.ConstraintMax:
		return maxColor
	case catatui.ConstraintLength:
		return lengthColor
	case catatui.ConstraintPercentage:
		return percentageColor
	case catatui.ConstraintRatio:
		return ratioColor
	default:
		return fillColor
	}
}
