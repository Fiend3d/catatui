// Tests ported from ratatui-widgets/src/tabs.rs and ratatui/tests/widgets_tabs.rs
// @ ratatui-v0.30.2

package widgets

import (
	"reflect"
	"testing"

	"github.com/Fiend3d/catatui"
)

// tabsCase renders tabs into an empty buffer covering area and compares it to
// the expected buffer, which is ratatui's test_case helper.
func tabsCase(t *testing.T, tabs Tabs, area catatui.Rect, expected *catatui.Buffer) {
	t.Helper()
	buf := catatui.NewBuffer(area)
	tabs.Render(area, buf)
	catatui.AssertBuffer(t, buf, expected)
}

func TestTabsNew(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4")
	wantTitles := []catatui.Line{
		catatui.LineFromString("Tab1"),
		catatui.LineFromString("Tab2"),
		catatui.LineFromString("Tab3"),
		catatui.LineFromString("Tab4"),
	}
	if !reflect.DeepEqual(tabs.titles, wantTitles) {
		t.Errorf("titles = %+v, want %+v", tabs.titles, wantTitles)
	}
	if tabs.hasBlock {
		t.Error("a new Tabs should have no block")
	}
	if i, ok := tabs.Selected(); !ok || i != 0 {
		t.Errorf("Selected() = %d, %v; want 0, true", i, ok)
	}
	if tabs.style != catatui.NewStyle() {
		t.Errorf("style = %+v, want empty", tabs.style)
	}
	if tabs.highlightStyle != defaultHighlightStyle {
		t.Errorf("highlightStyle = %+v, want reversed", tabs.highlightStyle)
	}
	if tabs.divider != catatui.NewSpan("│") {
		t.Errorf("divider = %+v, want │", tabs.divider)
	}
	if !reflect.DeepEqual(tabs.paddingLeft, catatui.LineFromString(" ")) ||
		!reflect.DeepEqual(tabs.paddingRight, catatui.LineFromString(" ")) {
		t.Errorf("padding = %+v / %+v, want a space on each side", tabs.paddingLeft, tabs.paddingRight)
	}
}

// TestTabsDefault is ratatui's default: NewTabs with no titles has nothing
// selected but keeps the standard divider and padding.
func TestTabsDefault(t *testing.T) {
	tabs := NewTabs()
	if len(tabs.titles) != 0 {
		t.Errorf("titles = %+v, want none", tabs.titles)
	}
	if _, ok := tabs.Selected(); ok {
		t.Error("an empty Tabs should have no selection")
	}
	if tabs.highlightStyle != defaultHighlightStyle {
		t.Errorf("highlightStyle = %+v, want reversed", tabs.highlightStyle)
	}
	if tabs.divider != catatui.NewSpan("│") {
		t.Errorf("divider = %+v, want │", tabs.divider)
	}
	if !reflect.DeepEqual(tabs.paddingLeft, catatui.LineFromString(" ")) ||
		!reflect.DeepEqual(tabs.paddingRight, catatui.LineFromString(" ")) {
		t.Errorf("padding = %+v / %+v, want a space on each side", tabs.paddingLeft, tabs.paddingRight)
	}
}

func TestTabsSelect(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4")
	if i, ok := tabs.Select(2).Selected(); !ok || i != 2 {
		t.Errorf("Select(2).Selected() = %d, %v; want 2, true", i, ok)
	}
	if _, ok := tabs.SelectNone().Selected(); ok {
		t.Error("SelectNone().Selected() should report no selection")
	}
	if i, ok := tabs.Select(1).Selected(); !ok || i != 1 {
		t.Errorf("Select(1).Selected() = %d, %v; want 1, true", i, ok)
	}
	// The original is untouched by the builders.
	if i, ok := tabs.Selected(); !ok || i != 0 {
		t.Errorf("original Selected() = %d, %v; want 0, true", i, ok)
	}
}

func TestTabsSelectBeforeTitles(t *testing.T) {
	tabs := NewTabs().Select(1).Titles("Tab1", "Tab2")
	if i, ok := tabs.Selected(); !ok || i != 1 {
		t.Errorf("Selected() = %d, %v; want 1, true", i, ok)
	}
}

// TestTabsTitlesClampsSelection pins the rest of ratatui's titles(): a
// selection past the new titles is clamped, and no titles means no selection.
func TestTabsTitlesClampsSelection(t *testing.T) {
	if i, ok := NewTabs().Select(5).Titles("Tab1", "Tab2").Selected(); !ok || i != 1 {
		t.Errorf("Select(5).Titles(2 titles).Selected() = %d, %v; want 1, true", i, ok)
	}
	if _, ok := NewTabs("Tab1").Titles().Selected(); ok {
		t.Error("Titles() with no titles should clear the selection")
	}
	if i, ok := NewTabs().Titles("Tab1").Selected(); !ok || i != 0 {
		t.Errorf("Titles() on an unselected Tabs should select the first: got %d, %v", i, ok)
	}
}

func TestTabsRenderNew(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4")
	expected := catatui.NewBufferWithStrings(" Tab1 │ Tab2 │ Tab3 │ Tab4    ")
	// first tab selected
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderNoPadding(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Padding("", "")
	expected := catatui.NewBufferWithStrings("Tab1│Tab2│Tab3│Tab4           ")
	expected.SetStyle(catatui.NewRect(0, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderLeftPadding(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").PaddingLeft("---")
	expected := catatui.NewBufferWithStrings("---Tab1 │---Tab2 │---Tab3 │---Tab4      ")
	expected.SetStyle(catatui.NewRect(3, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 40, 1), expected)
}

func TestTabsRenderRightPadding(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").PaddingRight("++")
	expected := catatui.NewBufferWithStrings(" Tab1++│ Tab2++│ Tab3++│ Tab4++         ")
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 40, 1), expected)
}

func TestTabsRenderMorePadding(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Padding("---", "++")
	expected := catatui.NewBufferWithStrings("---Tab1++│---Tab2++│---Tab3++│")
	expected.SetStyle(catatui.NewRect(3, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderWithBlock(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Block(Bordered().Title("Tabs"))
	expected := catatui.NewBufferWithStrings(
		"┌Tabs────────────────────────┐",
		"│ Tab1 │ Tab2 │ Tab3 │ Tab4  │",
		"└────────────────────────────┘",
	)
	expected.SetStyle(catatui.NewRect(2, 1, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 3), expected)
}

func TestTabsRenderStyle(t *testing.T) {
	red := catatui.NewStyle().Fg(catatui.ColorRed)
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Style(red)
	expected := catatui.NewBufferWithLines(
		catatui.LineFromStyledString(" Tab1 │ Tab2 │ Tab3 │ Tab4    ", red))
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), defaultHighlightStyle.Fg(catatui.ColorRed))
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderSelect(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4")
	reversed := catatui.NewStyle().AddModifier(catatui.ModifierReversed)
	area := catatui.NewRect(0, 0, 30, 1)

	// first tab selected
	expected := catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewSpan(" "),
		catatui.NewStyledSpan("Tab1", reversed),
		catatui.NewSpan(" │ Tab2 │ Tab3 │ Tab4    "),
	))
	tabsCase(t, tabs.Select(0), area, expected)

	// second tab selected
	expected = catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewSpan(" Tab1 │ "),
		catatui.NewStyledSpan("Tab2", reversed),
		catatui.NewSpan(" │ Tab3 │ Tab4    "),
	))
	tabsCase(t, tabs.Select(1), area, expected)

	// last tab selected
	expected = catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewSpan(" Tab1 │ Tab2 │ Tab3 │ "),
		catatui.NewStyledSpan("Tab4", reversed),
		catatui.NewSpan("    "),
	))
	tabsCase(t, tabs.Select(3), area, expected)

	// out of bounds selects no tab
	expected = catatui.NewBufferWithStrings(" Tab1 │ Tab2 │ Tab3 │ Tab4    ")
	tabsCase(t, tabs.Select(4), area, expected)

	// deselect
	expected = catatui.NewBufferWithStrings(" Tab1 │ Tab2 │ Tab3 │ Tab4    ")
	tabsCase(t, tabs.SelectNone(), area, expected)
}

func TestTabsRenderStyleAndSelected(t *testing.T) {
	red := catatui.NewStyle().Fg(catatui.ColorRed)
	underlined := catatui.NewStyle().AddModifier(catatui.ModifierUnderlined)
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").
		Style(red).
		HighlightStyle(underlined).
		Select(0)
	expected := catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewStyledSpan(" ", red),
		catatui.NewStyledSpan("Tab1", red.AddModifier(catatui.ModifierUnderlined)),
		catatui.NewStyledSpan(" │ Tab2 │ Tab3 │ Tab4    ", red),
	))
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderDivider(t *testing.T) {
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Divider("--")
	expected := catatui.NewBufferWithStrings(" Tab1 -- Tab2 -- Tab3 -- Tab4 ")
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), defaultHighlightStyle)
	tabsCase(t, tabs, catatui.NewRect(0, 0, 30, 1), expected)
}

func TestTabsRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Select(1).Divider("|")
	// This must not panic, even though the buffer is too small for the tabs.
	tabs.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestTabsRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	tabs := NewTabs("Tab1", "Tab2", "Tab3", "Tab4").Select(1).Divider("|")
	// This must not panic, even though the buffer has zero size.
	tabs.Render(buf.Area, buf)
}

func TestTabsWidthBasic(t *testing.T) {
	tabs := NewTabs("A", "BB", "CCC")
	if got, want := tabs.Width(), catatui.StringWidth(" A │ BB │ CCC "); got != want {
		t.Errorf("Width() = %d, want %d", got, want)
	}
}

func TestTabsWidthNoPadding(t *testing.T) {
	tabs := NewTabs("A", "BB", "CCC").Padding("", "")
	if got, want := tabs.Width(), catatui.StringWidth("A│BB│CCC"); got != want {
		t.Errorf("Width() = %d, want %d", got, want)
	}
}

func TestTabsWidthCustomDividerAndPadding(t *testing.T) {
	tabs := NewTabs("A", "BB", "CCC").Divider("--").Padding("X", "YY")
	if got, want := tabs.Width(), catatui.StringWidth("XAYY--XBBYY--XCCCYY"); got != want {
		t.Errorf("Width() = %d, want %d", got, want)
	}
}

func TestTabsWidthEmptyTitles(t *testing.T) {
	if got := NewTabs().Width(); got != 0 {
		t.Errorf("Width() = %d, want 0", got)
	}
}

// TestTabsWidthCJK is ratatui's unicode_width_cjk pair. catatui has one width
// function, so the check is simply that Width agrees with StringWidth of what
// gets rendered.
func TestTabsWidthCJK(t *testing.T) {
	tabs := NewTabs("你", "好", "世界")
	if got, want := tabs.Width(), catatui.StringWidth(" 你 │ 好 │ 世界 "); got != want {
		t.Errorf("Width() = %d, want %d", got, want)
	}
	tabs = NewTabs("你", "好", "世界").Divider("分").Padding("左", "右")
	if got, want := tabs.Width(), catatui.StringWidth("左你右分左好右分左世界右"); got != want {
		t.Errorf("Width() = %d, want %d", got, want)
	}
}

// The two cases below come from ratatui/tests/widgets_tabs.rs.

func TestTabsShouldNotPanicOnNarrowAreas(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	NewTabs("Tab1", "Tab2").Render(catatui.NewRect(0, 0, 1, 1), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestTabsShouldTruncateTheLastItem(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 1))
	NewTabs("Tab1", "Tab2").Render(catatui.NewRect(0, 0, 9, 1), buf)
	expected := catatui.NewBufferWithStrings(" Tab1 │ T ")
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), catatui.NewStyle().AddModifier(catatui.ModifierReversed))
	catatui.AssertBuffer(t, buf, expected)
}
