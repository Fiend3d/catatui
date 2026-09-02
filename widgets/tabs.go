// Port of ratatui-widgets/src/tabs.rs @ ratatui-v0.30.2

package widgets

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// defaultHighlightStyle is what a selected tab is drawn in unless
// Tabs.HighlightStyle says otherwise: the tab's colors swapped.
var defaultHighlightStyle = catatui.NewStyle().AddModifier(catatui.ModifierReversed)

// Tabs draws a horizontal row of titles with one of them highlighted.
//
// Each title is a Line, so titles can be styled individually. The selected tab
// is chosen with Select and drawn in HighlightStyle; the divider between tabs
// defaults to │ and the padding on either side of each title to one space.
//
//	tabs := widgets.NewTabs("Tab1", "Tab2", "Tab3").
//		Block(widgets.Bordered().Title("Tabs")).
//		HighlightStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
//		Select(1).
//		Divider(symbols.DotFull).
//		Padding("->", "<-")
//	f.RenderWidget(tabs, area)
//
// Build a Tabs with NewTabs or NewTabsFromLines; the zero value has no divider
// and no padding.
type Tabs struct {
	block          Block
	hasBlock       bool
	titles         []catatui.Line
	selected       int
	hasSelected    bool
	style          catatui.Style
	highlightStyle catatui.Style
	divider        catatui.Span
	paddingLeft    catatui.Line
	paddingRight   catatui.Line
}

// NewTabs returns tabs with the given titles. The first tab is selected when
// there is one; with no titles nothing is selected, which is ratatui's
// Tabs::default().
func NewTabs(titles ...string) Tabs {
	lines := make([]catatui.Line, len(titles))
	for i, s := range titles {
		lines[i] = catatui.LineFromString(s)
	}
	return NewTabsFromLines(lines...)
}

// NewTabsFromLines returns tabs whose titles are already-styled lines. The
// first tab is selected when there is one.
func NewTabsFromLines(lines ...catatui.Line) Tabs {
	return Tabs{
		titles:         lines,
		hasSelected:    len(lines) > 0,
		highlightStyle: defaultHighlightStyle,
		divider:        catatui.NewSpan(symbols.Vertical),
		paddingLeft:    catatui.LineFromString(" "),
		paddingRight:   catatui.LineFromString(" "),
	}
}

// Titles returns a copy of t with new titles. A selection that is now out of
// range is clamped to the last tab, and if nothing was selected the first tab
// becomes selected.
func (t Tabs) Titles(titles ...string) Tabs {
	lines := make([]catatui.Line, len(titles))
	for i, s := range titles {
		lines[i] = catatui.LineFromString(s)
	}
	return t.TitleLines(lines...)
}

// TitleLines is Titles for already-styled lines.
func (t Tabs) TitleLines(lines ...catatui.Line) Tabs {
	t.titles = lines
	switch {
	case len(lines) == 0:
		t.selected, t.hasSelected = 0, false
	case t.hasSelected:
		t.selected = min(t.selected, len(lines)-1)
	default:
		t.selected, t.hasSelected = 0, true
	}
	return t
}

// GetTitles returns the titles.
func (t Tabs) GetTitles() []catatui.Line { return t.titles }

// Block returns a copy of t drawn inside the given block.
func (t Tabs) Block(b Block) Tabs { t.block, t.hasBlock = b, true; return t }

// Select returns a copy of t with the tab at index i selected. The first tab
// is index 0. An index past the last tab selects nothing, as in ratatui.
func (t Tabs) Select(i int) Tabs { t.selected, t.hasSelected = i, true; return t }

// SelectNone returns a copy of t with no tab selected, which is ratatui's
// select(None).
func (t Tabs) SelectNone() Tabs { t.selected, t.hasSelected = 0, false; return t }

// Selected returns the selected index and whether there is one.
func (t Tabs) Selected() (int, bool) { return t.selected, t.hasSelected }

// Style returns a copy of t with a style applied to the whole area, beneath
// the titles' own styles.
func (t Tabs) Style(s catatui.Style) Tabs { t.style = s; return t }

// HighlightStyle returns a copy of t with the style the selected tab is drawn
// in, on top of the title's own style.
func (t Tabs) HighlightStyle(s catatui.Style) Tabs { t.highlightStyle = s; return t }

// Divider returns a copy of t with the string drawn between tabs. The default
// is a pipe (│).
func (t Tabs) Divider(divider string) Tabs { t.divider = catatui.NewSpan(divider); return t }

// DividerSpan is Divider for a styled span.
func (t Tabs) DividerSpan(divider catatui.Span) Tabs { t.divider = divider; return t }

// Padding returns a copy of t with what is drawn on either side of every
// title. Both default to a single space; pass "" for none.
func (t Tabs) Padding(left, right string) Tabs {
	t.paddingLeft = catatui.LineFromString(left)
	t.paddingRight = catatui.LineFromString(right)
	return t
}

// PaddingLeft returns a copy of t with what is drawn before every title.
func (t Tabs) PaddingLeft(padding string) Tabs {
	t.paddingLeft = catatui.LineFromString(padding)
	return t
}

// PaddingRight returns a copy of t with what is drawn after every title.
func (t Tabs) PaddingRight(padding string) Tabs {
	t.paddingRight = catatui.LineFromString(padding)
	return t
}

// Width reports how many columns the tabs take when rendered: the titles, the
// dividers and the padding, but not the border of any block.
//
//	widgets.NewTabs("Tab1", "Tab2", "Tab3").Width() // 20: " Tab1 │ Tab2 │ Tab3 "
func (t Tabs) Width() int {
	width := 0
	for _, title := range t.titles {
		width += title.Width()
	}
	n := len(t.titles)
	if n > 0 {
		width += (n - 1) * t.divider.Width()
	}
	width += n * t.paddingLeft.Width()
	width += n * t.paddingRight.Width()
	return width
}

// Render draws the tabs on the first row of the area.
func (t Tabs) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, t.style)
	inner := area
	if t.hasBlock {
		t.block.Render(area, buf)
		inner = t.block.Inner(area)
	}
	t.renderTabs(inner, buf)
}

func (t Tabs) renderTabs(tabsArea catatui.Rect, buf *catatui.Buffer) {
	if tabsArea.IsEmpty() {
		return
	}

	x := tabsArea.Left()
	y := tabsArea.Top()
	for i, title := range t.titles {
		lastTitle := i == len(t.titles)-1
		remaining := catatui.SatSub(tabsArea.Right(), x)
		if remaining == 0 {
			break
		}

		// Left padding.
		x, _ = buf.SetLine(x, y, t.paddingLeft, remaining)
		remaining = catatui.SatSub(tabsArea.Right(), x)
		if remaining == 0 {
			break
		}

		// Title, highlighted over exactly the columns it took.
		nextX, _ := buf.SetLine(x, y, title, remaining)
		if t.hasSelected && i == t.selected {
			buf.SetStyle(catatui.Rect{
				X:      x,
				Y:      y,
				Width:  catatui.SatSub(nextX, x),
				Height: 1,
			}, t.highlightStyle)
		}
		x = nextX
		remaining = catatui.SatSub(tabsArea.Right(), x)
		if remaining == 0 {
			break
		}

		// Right padding.
		x, _ = buf.SetLine(x, y, t.paddingRight, remaining)
		remaining = catatui.SatSub(tabsArea.Right(), x)
		if remaining == 0 || lastTitle {
			break
		}

		x, _ = buf.SetSpan(x, y, t.divider, remaining)
	}
}

var _ catatui.Widget = Tabs{}
