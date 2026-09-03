// Command flex shows what each Flex mode does to the same set of constraints.
//
//	go run ./examples/apps/flex
//
// Left and right change the flex mode, up and down scroll, - and + change the
// spacing between the segments, q quits.
//
// Port of examples/apps/flex @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// example is one row of the demo: a set of constraints, and a note about what
// they do to each other.
type example struct {
	description string
	constraints []catatui.Constraint
}

// exampleData is every case the demo walks through, in order. The descriptions
// are ratatui's, and are what the example is really for: they say which
// constraint wins when two of them cannot both be satisfied.
var exampleData = []example{
	{"Min(u16) takes any excess space always", []catatui.Constraint{
		catatui.Length(10), catatui.Min(10), catatui.Max(10),
		catatui.Percentage(10), catatui.Ratio(1, 10),
	}},
	{"Fill(u16) takes any excess space always", []catatui.Constraint{
		catatui.Length(20), catatui.Percentage(20), catatui.Ratio(1, 5), catatui.Fill(1),
	}},
	{"Here's all constraints in one line", []catatui.Constraint{
		catatui.Length(10), catatui.Min(10), catatui.Max(10),
		catatui.Percentage(10), catatui.Ratio(1, 10), catatui.Fill(1),
	}},
	{"", []catatui.Constraint{catatui.Max(50), catatui.Min(50)}},
	{"", []catatui.Constraint{catatui.Max(20), catatui.Length(10)}},
	{"", []catatui.Constraint{catatui.Max(20), catatui.Length(10)}},
	{"Min grows always but also allows Fill to grow", []catatui.Constraint{
		catatui.Percentage(50), catatui.Fill(1), catatui.Fill(2), catatui.Min(50),
	}},
	{"In `Legacy`, the last constraint of lowest priority takes excess space",
		[]catatui.Constraint{catatui.Length(20), catatui.Length(20), catatui.Percentage(20)}},
	{"", []catatui.Constraint{catatui.Length(20), catatui.Percentage(20), catatui.Length(20)}},
	{"A lowest priority constraint will be broken before a high priority constraint",
		[]catatui.Constraint{catatui.Ratio(1, 4), catatui.Percentage(20)}},
	{"`Length` is higher priority than `Percentage`",
		[]catatui.Constraint{catatui.Percentage(20), catatui.Length(10)}},
	{"`Min/Max` is higher priority than `Length`",
		[]catatui.Constraint{catatui.Length(10), catatui.Max(20)}},
	{"", []catatui.Constraint{catatui.Length(100), catatui.Min(20)}},
	{"`Length` is higher priority than `Min/Max`",
		[]catatui.Constraint{catatui.Max(20), catatui.Length(10)}},
	{"", []catatui.Constraint{catatui.Min(20), catatui.Length(90)}},
	{"Fill is the lowest priority and will fill any excess space",
		[]catatui.Constraint{catatui.Fill(1), catatui.Ratio(1, 4)}},
	{"Fill can be used to scale proportionally with other Fill blocks",
		[]catatui.Constraint{catatui.Fill(1), catatui.Percentage(20), catatui.Fill(2)}},
	{"", []catatui.Constraint{catatui.Ratio(1, 3), catatui.Percentage(20), catatui.Ratio(2, 3)}},
	{"Legacy will stretch the last lowest priority constraint\nStretch will only stretch equal weighted constraints",
		[]catatui.Constraint{catatui.Length(20), catatui.Length(15)}},
	{"", []catatui.Constraint{catatui.Percentage(20), catatui.Length(15)}},
	{"`Fill(u16)` fills up excess space, but is lower priority to spacers.\ni.e. Fill will only have widths in Flex::Stretch and Flex::Legacy",
		[]catatui.Constraint{catatui.Fill(1), catatui.Fill(1)}},
	{"", []catatui.Constraint{catatui.Length(20), catatui.Length(20)}},
	{"When not using `Flex::Stretch` or `Flex::Legacy`,\n`Min(u16)` and `Max(u16)` collapse to their lowest values",
		[]catatui.Constraint{catatui.Min(20), catatui.Max(20)}},
	{"", []catatui.Constraint{catatui.Max(20)}},
	{"", []catatui.Constraint{
		catatui.Min(20), catatui.Max(20), catatui.Length(20), catatui.Length(20),
	}},
	{"", []catatui.Constraint{catatui.Fill(0), catatui.Fill(0)}},
	{"`Fill(1)` can be to scale with respect to other `Fill(2)`",
		[]catatui.Constraint{catatui.Fill(1), catatui.Fill(2)}},
	{"", []catatui.Constraint{
		catatui.Fill(1), catatui.Min(10), catatui.Max(10), catatui.Fill(2),
	}},
	{"`Fill(0)` collapses if there are other non-zero `Fill(_)`\nconstraints. e.g. `[Fill(0), Fill(0), Fill(1)]`:",
		[]catatui.Constraint{catatui.Fill(0), catatui.Fill(0), catatui.Fill(1)}},
}

// tab is one of the flex modes, in the order they appear across the top.
type tab int

const (
	tabLegacy tab = iota
	tabStart
	tabCenter
	tabEnd
	tabSpaceAround
	tabSpaceEvenly
	tabSpaceBetween
)

// tabs lists every mode with the name shown on its title and the Flex it
// selects.
var tabs = []struct {
	name string
	flex catatui.Flex
}{
	tabLegacy:       {"Legacy", catatui.FlexLegacy},
	tabStart:        {"Start", catatui.FlexStart},
	tabCenter:       {"Center", catatui.FlexCenter},
	tabEnd:          {"End", catatui.FlexEnd},
	tabSpaceAround:  {"SpaceAround", catatui.FlexSpaceAround},
	tabSpaceEvenly:  {"SpaceEvenly", catatui.FlexSpaceEvenly},
	tabSpaceBetween: {"SpaceBetween", catatui.FlexSpaceBetween},
}

// app is the whole demo: which mode is shown, how far it is scrolled, and how
// much spacing the layouts are given.
type app struct {
	selectedTab  tab
	scrollOffset uint16
	spacing      uint16
	quit         bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// descriptionHeight is how many rows a description takes: one per line, none
// when there is nothing to say.
func descriptionHeight(description string) uint16 {
	if description == "" {
		return 0
	}
	return uint16(strings.Count(description, "\n") + 1)
}

// exampleHeight is the height of every example stacked together. Each takes
// four rows plus however many its description needs.
func exampleHeight() uint16 {
	var height uint16
	for _, e := range exampleData {
		height = catatui.SatAdd(height, descriptionHeight(e.description)+4)
	}
	return height
}

// maxScrollOffset stops the scroll at the point where the last example is fully
// on screen.
func maxScrollOffset() uint16 {
	last := exampleData[len(exampleData)-1]
	return catatui.SatSub(exampleHeight(), descriptionHeight(last.description)+4)
}

// Render draws the tab bar, the axis and the examples under it. The app is a
// widget itself, which is how ratatui's example is written.
func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(3),
		catatui.Length(1),
		catatui.Fill(0),
	).Split(area)

	a.renderTabs(rows[0], buf)
	scrollNeeded := a.renderDemo(rows[2], buf)

	// The axis measures the space the examples actually got, which is one
	// column narrower when the scrollbar is there.
	axisWidth := rows[1].Width
	if scrollNeeded {
		axisWidth = catatui.SatSub(axisWidth, 1)
	}
	renderAxis(axisWidth, a.spacing).Render(rows[1], buf)
}

func (a *app) renderTabs(area catatui.Rect, buf *catatui.Buffer) {
	titles := make([]catatui.Line, len(tabs))
	for i := range tabs {
		titles[i] = catatui.LineFromStyledString(" "+tabs[i].name+" ",
			catatui.NewStyle().Fg(theme.tab[i]).Bg(catatui.ColorBlack))
	}

	block := widgets.NewBlock().
		TitleLine(catatui.LineFromStyledString("Flex Layouts ",
			catatui.NewStyle().AddModifier(catatui.ModifierBold))).
		TitleLine(catatui.LineFromString(
			" Use ◄ ► to change tab, ▲ ▼  to scroll, - + to change spacing "))

	widgets.NewTabsFromLines(titles...).
		Block(block).
		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
		Select(int(a.selectedTab)).
		Divider(" ").
		Padding("", "").
		Render(area, buf)
}

// renderAxis draws a bar like <----- 80 px (gap: 2 px) ----->.
func renderAxis(width, spacing uint16) widgets.Paragraph {
	label := fmt.Sprintf("%d px", width)
	if spacing != 0 {
		label = fmt.Sprintf("%d px (gap: %d px)", width, spacing)
	}
	// Two columns go to the < and > at the ends.
	bar := centerPad(label, int(catatui.SatSub(width, 2)), '-')
	return widgets.NewParagraph("<" + bar + ">").
		Style(catatui.NewStyle().Fg(catatui.ColorDarkGray)).
		Centered()
}

// centerPad centres s in a field of the given width, padded with fill, which is
// Rust's {:-^width$} format.
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
// where the viewport starts. It reports whether a scrollbar was needed.
func (a *app) renderDemo(area catatui.Rect, buf *catatui.Buffer) bool {
	if area.IsEmpty() {
		return false
	}

	height := exampleHeight()
	demoArea := catatui.NewRect(0, 0, area.Width, height)
	demoBuf := catatui.NewBuffer(demoArea)

	scrollbarNeeded := a.scrollOffset != 0 || height > area.Height
	contentArea := demoArea
	if scrollbarNeeded {
		contentArea.Width = catatui.SatSub(contentArea.Width, 1)
	}
	a.renderExamples(contentArea, demoBuf)

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
		state := widgets.NewScrollbarState(int(maxScrollOffset())).
			Position(int(a.scrollOffset))
		catatui.RenderStatefulWidget(
			widgets.NewScrollbar(widgets.ScrollbarVerticalRight),
			area.Intersection(buf.Area), buf, &state)
	}
	return scrollbarNeeded
}

// renderExamples stacks every example, each in the currently selected mode.
func (a *app) renderExamples(area catatui.Rect, buf *catatui.Buffer) {
	heights := make([]catatui.Constraint, len(exampleData))
	for i, e := range exampleData {
		heights[i] = catatui.Length(descriptionHeight(e.description) + 4)
	}
	areas := catatui.VerticalLayout(heights...).Flex(catatui.FlexStart).Split(area)

	for i, e := range exampleData {
		a.renderExample(e, areas[i], buf)
	}
}

// renderExample draws one row: the description, then the segments the
// constraints produce, with the spacers between them labelled.
func (a *app) renderExample(e example, area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(descriptionHeight(e.description)),
		catatui.Fill(0),
	).Split(area)

	if e.description != "" {
		lines := make([]catatui.Line, 0, 4)
		for _, line := range strings.Split(e.description, "\n") {
			lines = append(lines, catatui.LineFromStyledString("// "+line,
				catatui.NewStyle().
					Fg(theme.descriptionFg).
					AddModifier(catatui.ModifierItalic)))
		}
		widgets.NewParagraphFromText(catatui.NewText(lines...)).Render(rows[0], buf)
	}

	blocks, spacers := catatui.HorizontalLayout(e.constraints...).
		Flex(tabs[a.selectedTab].flex).
		Spacing(catatui.Space(a.spacing)).
		SplitWithSpacers(rows[1])

	for i, block := range blocks {
		illustration(e.constraints[i], block.Width).Render(block, buf)
	}
	for _, spacer := range spacers {
		renderSpacer(spacer, buf)
	}
}

// illustration is the box drawn for one constraint: what it asked for, and how
// many columns it got.
func illustration(constraint catatui.Constraint, width uint16) widgets.Paragraph {
	color := colorForConstraint(constraint)
	block := widgets.Bordered().
		BorderSet(symbols.BorderQuadrantOutside).
		BorderStyle(catatui.ResetStyle().
			Fg(color).
			AddModifier(catatui.ModifierReversed)).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite).Bg(color))

	text := fmt.Sprintf("%s\n%d px", constraint, width)
	return widgets.NewParagraph(text).Centered().Block(block)
}

// renderSpacer draws the gap between two segments, with its width in it when
// there is room. A gap one column wide gets a bare line instead of a box.
func renderSpacer(spacer catatui.Rect, buf *catatui.Buffer) {
	darkGray := catatui.ResetStyle().Fg(catatui.ColorDarkGray)

	if spacer.Width > 1 {
		// Corners only: the sides would read as segment borders.
		cornersOnly := symbols.BorderSet{
			TopLeft:      symbols.TopLeft,
			TopRight:     symbols.TopRight,
			BottomLeft:   symbols.BottomLeft,
			BottomRight:  symbols.BottomRight,
			VerticalLeft: " ", VerticalRight: " ",
			HorizontalTop: " ", HorizontalBottom: " ",
		}
		widgets.Bordered().
			BorderSet(cornersOnly).
			BorderStyle(darkGray).
			Render(spacer, buf)
	} else {
		widgets.NewParagraphFromText(catatui.NewText(
			catatui.LineFromString(""),
			catatui.LineFromString("│"),
			catatui.LineFromString("│"),
			catatui.LineFromString(""),
		)).Style(darkGray).Render(spacer, buf)
	}

	label := ""
	switch {
	case spacer.Width > 4:
		label = fmt.Sprintf("%d px", spacer.Width)
	case spacer.Width > 2:
		label = strconv.Itoa(int(spacer.Width))
	}
	text := catatui.NewText(
		catatui.LineFromString(""),
		catatui.LineFromString(""),
		catatui.LineFromStyledString(label, darkGray),
	)
	widgets.NewParagraphFromText(text).
		Style(darkGray).
		Alignment(catatui.AlignmentCenter).
		Render(spacer, buf)
}

var _ catatui.Widget = (*app)(nil)
