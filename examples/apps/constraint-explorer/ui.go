// Port of the rendering half of examples/apps/constraint-explorer
// @ ratatui-v0.30.2

package main

import (
	"fmt"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// The colours of the chrome around the blocks.
var (
	headerColor = tailwind.Slate.C200
	textColor   = tailwind.Slate.C400
	axisColor   = tailwind.Slate.C500

	blockTextColor    = tailwind.Slate.C200
	spacerTextColor   = tailwind.Slate.C500
	spacerBorderColor = tailwind.Slate.C600
)

// color is the colour a kind of constraint is drawn in, unselected. They run by
// the priority the solver gives each kind: Min and Max first, then Length,
// Percentage and Ratio, then Fill last.
func (n constraintName) color() catatui.Color {
	switch n {
	case nameMin:
		return tailwind.Blue.C800
	case nameMax:
		return tailwind.Blue.C900
	case nameLength:
		return tailwind.Slate.C700
	case namePercentage:
		return tailwind.Slate.C800
	case nameRatio:
		return tailwind.Slate.C900
	default:
		return tailwind.Slate.C950
	}
}

// lighterColor is the same block with the selection on it.
func (n constraintName) lighterColor() catatui.Color {
	switch n {
	case nameMin:
		return tailwind.Sky.C600
	case nameMax:
		return tailwind.Sky.C700
	case nameLength:
		return tailwind.Stone.C500
	case namePercentage:
		return tailwind.Stone.C600
	case nameRatio:
		return tailwind.Stone.C700
	default:
		return tailwind.Stone.C800
	}
}

// flexModes are the six panels, in the order they are stacked up.
var flexModes = []catatui.Flex{
	catatui.FlexStart,
	catatui.FlexCenter,
	catatui.FlexEnd,
	catatui.FlexSpaceBetween,
	catatui.FlexSpaceAround,
	catatui.FlexSpaceEvenly,
}

// panelHeight is a label, an axis and four rows of block.
const panelHeight uint16 = 7

// Render draws the header, the keys, the legend and the six panels.
func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(2), // header
		catatui.Length(2), // instructions
		catatui.Length(1), // the legend of what the number keys swap to
		catatui.Length(1), // gap
		catatui.Fill(1),   // the blocks
	).Split(area)

	renderHeader(rows[0], buf)
	renderInstructions(rows[1], buf)
	renderSwapLegend(rows[2], buf)
	a.renderLayoutBlocks(rows[4], buf)
}

func renderHeader(area catatui.Rect, buf *catatui.Buffer) {
	catatui.LineFromStyledString("Constraint Explorer",
		catatui.NewStyle().Fg(headerColor).AddModifier(catatui.ModifierBold)).
		Centered().
		Render(area, buf)
}

func renderInstructions(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("◄ ►: select, ▲ ▼: edit, 1-6: swap, a: add, x: delete, q: quit, +/-: spacing").
		Style(catatui.NewStyle().Fg(textColor)).
		Centered().
		Wrap(widgets.Wrap{Trim: false}).
		Render(area, buf)
}

// renderSwapLegend draws the number keys against the colour each kind is drawn
// in, which is what makes the panels below readable.
func renderSwapLegend(area catatui.Rect, buf *catatui.Buffer) {
	spans := make([]catatui.Span, 0, len(swapOrder)*2)
	for i, name := range swapOrder {
		if i > 0 {
			spans = append(spans, catatui.NewSpan(" "))
		}
		spans = append(spans, catatui.NewStyledSpan(
			fmt.Sprintf("  %d: %s  ", i+1, name),
			catatui.NewStyle().Fg(tailwind.Slate.C200).Bg(name.color())))
	}
	widgets.NewParagraphFromText(catatui.NewText(catatui.NewLine(spans...).Centered())).
		Wrap(widgets.Wrap{Trim: false}).
		Render(area, buf)
}

// renderLayoutBlocks draws the constraints once as a legend, then once per flex
// mode.
func (a *app) renderLayoutBlocks(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(catatui.Length(3), catatui.Fill(1)).
		Spacing(catatui.Space(1)).
		Split(area)

	a.renderUserConstraintsLegend(rows[0], buf)

	panels := make([]catatui.Constraint, len(flexModes))
	for i := range panels {
		panels[i] = catatui.Length(panelHeight)
	}
	areas := catatui.VerticalLayout(panels...).Split(rows[1])
	for i, flex := range flexModes {
		a.renderLayoutBlock(flex, areas[i], buf)
	}
}

// renderUserConstraintsLegend draws the blocks evenly across the width,
// whatever their constraints say, so that the list can be read and edited
// without the solver hiding any of it.
func (a *app) renderUserConstraintsLegend(area catatui.Rect, buf *catatui.Buffer) {
	if len(a.blocks) == 0 {
		return
	}
	constraints := make([]catatui.Constraint, len(a.blocks))
	for i := range constraints {
		constraints[i] = catatui.Fill(1)
	}
	areas := catatui.HorizontalLayout(constraints...).Split(area)

	for i, b := range a.blocks {
		constraintBlock{block: b, selected: a.selected == i, legend: true}.Render(areas[i], buf)
	}
}

// renderLayoutBlock draws one panel: the name of the flex mode, the axis, and
// the blocks with the spacers between them.
func (a *app) renderLayoutBlock(flex catatui.Flex, area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(1), // the name of the mode
		catatui.Max(1),    // the axis, which is the first thing a short panel loses
		catatui.Length(4), // the blocks
	).Split(area)

	if rows[0].Height > 0 {
		catatui.LineFromStyledString("Flex::"+flex.String(),
			catatui.NewStyle().AddModifier(catatui.ModifierBold)).
			Render(rows[0], buf)
	}
	a.axis(area.Width).Render(rows[1], buf)

	if len(a.blocks) == 0 {
		return
	}
	segments, spacers := catatui.HorizontalLayout(a.constraints()...).
		Flex(flex).
		Spacing(a.spacingValue()).
		SplitWithSpacers(rows[2])

	for i, b := range a.blocks {
		constraintBlock{block: b, selected: a.selected == i}.Render(segments[i], buf)
	}
	for _, spacer := range spacers {
		spacerBlock{}.Render(spacer, buf)
	}
}

// spacingValue turns the app's signed spacing into the Spacing the layout
// takes: a positive one separates the blocks, a negative one overlaps them.
func (a *app) spacingValue() catatui.Spacing {
	if a.spacing < 0 {
		return catatui.Overlap(uint16(-int32(a.spacing)))
	}
	return catatui.Space(uint16(a.spacing))
}

// axis draws a bar like <----- 80 px (gap: 2 px) ----->, naming the width the
// blocks are laid out in and what the spacing is doing to it.
func (a *app) axis(width uint16) catatui.Widget {
	var label string
	switch {
	case a.spacing > 0:
		label = fmt.Sprintf("%d px (gap: %d px)", width, a.spacing)
	case a.spacing < 0:
		label = fmt.Sprintf("%d px (overlap: %d px)", width, -int32(a.spacing))
	default:
		label = fmt.Sprintf("%d px", width)
	}
	// Two columns go to the < and > at the ends.
	bar := "<" + center(label, int(catatui.SatSub(width, 2)), '-') + ">"
	return widgets.NewParagraph(bar).
		Style(catatui.NewStyle().Fg(axisColor)).
		Centered()
}

// center pads s out to width on both sides, leaving it alone if it is already
// that long. It is Rust's {:-^width$}.
func center(s string, width int, pad byte) string {
	missing := width - len(s)
	if missing <= 0 {
		return s
	}
	left := missing / 2
	return strings.Repeat(string(pad), left) + s + strings.Repeat(string(pad), missing-left)
}

// constraintBlock draws one constraint as a labelled block, in as much detail
// as the height it is given allows.
type constraintBlock struct {
	block    block
	selected bool

	// legend is set for the row of blocks at the top, which are drawn evenly
	// rather than by their constraints, and so carry the selection colour over
	// the whole block rather than on the border alone.
	legend bool
}

func (c constraintBlock) Render(area catatui.Rect, buf *catatui.Buffer) {
	switch area.Height {
	case 1:
		c.render1px(area, buf)
	case 2:
		c.render2px(area, buf)
	default:
		c.render4px(area, buf)
	}
}

// color is the block's colour, and the lighter one when it is selected.
func (c constraintBlock) color() catatui.Color {
	if c.selected {
		return c.block.name.lighterColor()
	}
	return c.block.name.color()
}

// render1px has room for a colour and nothing else.
func (c constraintBlock) render1px(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewBlock().
		Style(catatui.NewStyle().Fg(blockTextColor).Bg(c.color())).
		Render(area, buf)
}

// render2px has room for the border, drawn in half blocks so that the block
// reads as a solid bar rather than a box.
func (c constraintBlock) render2px(area catatui.Rect, buf *catatui.Buffer) {
	widgets.Bordered().
		BorderSet(symbols.BorderQuadrantOutside).
		BorderStyle(catatui.ResetStyle().Fg(c.color()).AddModifier(catatui.ModifierReversed)).
		Render(area, buf)
}

// render4px has room for the constraint and the width it came to.
func (c constraintBlock) render4px(area catatui.Rect, buf *catatui.Buffer) {
	color := c.color()
	if !c.legend {
		// Down here the blocks are laid out by their own constraints, so the
		// selection goes on the last row alone and the fill stays the colour of
		// the kind.
		color = c.block.name.color()
	}
	style := catatui.NewStyle().Fg(blockTextColor).Bg(color)

	block := widgets.Bordered().
		BorderSet(symbols.BorderQuadrantOutside).
		BorderStyle(catatui.ResetStyle().Fg(color).AddModifier(catatui.ModifierReversed)).
		Style(style)

	widgets.NewParagraph(c.label(area.Width)).
		Centered().
		Style(style).
		Block(block).
		Render(area, buf)

	if !c.legend {
		if rows := area.Rows(); len(rows) > 0 {
			buf.SetStyle(rows[len(rows)-1], catatui.NewStyle().Fg(c.color()))
		}
	}
}

// label is the constraint over the width it came to, with the width dropped
// when the block is too narrow to hold it.
func (c constraintBlock) label(width uint16) string {
	long := fmt.Sprintf("%d px", width)
	short := fmt.Sprintf("%d", width)
	available := int(catatui.SatSub(width, 2)) // the border takes two columns

	widthLabel := ""
	switch {
	case len(long) < available:
		widthLabel = long
	case len(short) < available:
		widthLabel = short
	}
	return c.block.constraint().String() + "\n" + widthLabel
}

// spacerBlock draws the gap between two blocks: corners, the word Spacer, and
// how wide the gap is, as each fits.
type spacerBlock struct{}

func (s spacerBlock) Render(area catatui.Rect, buf *catatui.Buffer) {
	switch area.Height {
	case 0, 1:
		return
	case 2:
		s.renderFrame(area, buf)
	case 3:
		s.renderFrame(area, buf)
		s.renderRow(area, 1, spacerLabel(area.Width), buf)
	default:
		s.renderFrame(area, buf)
		s.renderRow(area, 1, spacerLabel(area.Width), buf)
		s.renderRow(area, 2, widthLabel(area.Width), buf)
	}
}

// renderFrame draws the corners of the gap, or a single line down it when there
// is no room for two corners side by side.
func (s spacerBlock) renderFrame(area catatui.Rect, buf *catatui.Buffer) {
	if area.Width > 1 {
		widgets.Bordered().
			BorderSet(cornersOnly).
			BorderStyle(catatui.NewStyle().Fg(spacerBorderColor)).
			Render(area, buf)
		return
	}
	widgets.NewParagraphFromText(catatui.NewText(
		catatui.LineFromString(""),
		catatui.LineFromString(symbols.Vertical),
		catatui.LineFromString(symbols.Vertical),
		catatui.LineFromString(""),
	)).
		Style(catatui.NewStyle().Fg(spacerBorderColor)).
		Render(area, buf)
}

// renderRow draws one centred line inside the gap.
func (s spacerBlock) renderRow(area catatui.Rect, row int, text string, buf *catatui.Buffer) {
	rows := area.Rows()
	if row >= len(rows) {
		return
	}
	catatui.LineFromStyledString(text, catatui.NewStyle().Fg(spacerTextColor)).
		Centered().
		Render(rows[row], buf)
}

// cornersOnly is a border of four corners and no sides, which is what makes a
// gap look like a gap rather than another block.
var cornersOnly = symbols.BorderSet{
	TopLeft:          symbols.LineNormal.TopLeft,
	TopRight:         symbols.LineNormal.TopRight,
	BottomLeft:       symbols.LineNormal.BottomLeft,
	BottomRight:      symbols.LineNormal.BottomRight,
	VerticalLeft:     " ",
	VerticalRight:    " ",
	HorizontalTop:    " ",
	HorizontalBottom: " ",
}

// spacerLabel names the gap, when it is wide enough to hold the word.
func spacerLabel(width uint16) string {
	if width >= 6 {
		return "Spacer"
	}
	return ""
}

// widthLabel is how wide the gap is, dropped when it does not fit.
func widthLabel(width uint16) string {
	long := fmt.Sprintf("%d px", width)
	short := fmt.Sprintf("%d", width)
	switch {
	case len(long) < int(width):
		return long
	case len(short) < int(width):
		return short
	}
	return ""
}
