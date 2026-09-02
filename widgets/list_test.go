// Tests ported from ratatui-widgets/src/list.rs, list/item.rs, list/state.rs
// and list/rendering.rs @ ratatui-v0.30.2

package widgets

import (
	"math"
	"reflect"
	"testing"

	"github.com/Fiend3d/catatui"
)

// renderListStateful draws a list into a fresh buffer of the given size with
// the given state, the counterpart of ratatui's stateful_widget helper.
func renderListStateful(list List, state *ListState, width, height uint16) *catatui.Buffer {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, height))
	list.RenderStateful(buf.Area, buf, state)
	return buf
}

// styledLine builds a single-span line in one style, standing in for
// ratatui's `"text".red()` in buffer expectations.
func styledLine(s string, style catatui.Style) catatui.Line {
	return catatui.NewLine(catatui.NewStyledSpan(s, style))
}

func assertSelected(t *testing.T, state ListState, want int, wantOk bool) {
	t.Helper()
	got, ok := state.Selected()
	if ok != wantOk || (ok && got != want) {
		t.Errorf("Selected() = (%d, %v), want (%d, %v)", got, ok, want, wantOk)
	}
}

var (
	bold   = catatui.NewStyle().AddModifier(catatui.ModifierBold)
	italic = catatui.NewStyle().AddModifier(catatui.ModifierItalic)
	red    = catatui.NewStyle().Fg(catatui.ColorRed)
	yellow = catatui.NewStyle().Fg(catatui.ColorYellow)
	onBlue = catatui.NewStyle().Bg(catatui.ColorBlue)
)

// --- list.rs -------------------------------------------------------------

func TestListFromStrings(t *testing.T) {
	collected := NewListFromStrings("Item0", "Item1", "Item2")
	expected := NewList(NewListItem("Item0"), NewListItem("Item1"), NewListItem("Item2"))
	if !reflect.DeepEqual(collected, expected) {
		t.Errorf("NewListFromStrings differs from NewList of the same items")
	}
}

func TestListNoStyle(t *testing.T) {
	list := NewList(NewListItemFromText(catatui.TextFromString("Item 1"))).
		HighlightSymbol(">>").
		HighlightSpacing(HighlightSpacingAlways)
	var state ListState
	buf := renderListStateful(list, &state, 10, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("  Item 1  "))
}

func TestListStyledText(t *testing.T) {
	text := catatui.TextFromString("Item 1").Style(bold)
	list := NewList(NewListItemFromText(text)).
		HighlightSymbol(">>").
		HighlightSpacing(HighlightSpacingAlways)
	var state ListState
	buf := renderListStateful(list, &state, 10, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewSpan("  "),
		catatui.NewStyledSpan("Item 1  ", bold),
	)))
}

func TestListStyledListItem(t *testing.T) {
	item := NewListItemFromText(catatui.TextFromString("Item 1")).Style(italic)
	list := NewList(item).
		HighlightSymbol(">>").
		HighlightSpacing(HighlightSpacingAlways)
	var state ListState
	buf := renderListStateful(list, &state, 10, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(styledLine("  Item 1  ", italic)))
}

func TestListStyledTextAndListItem(t *testing.T) {
	text := catatui.TextFromString("Item 1").Style(bold)
	item := NewListItemFromText(text).Style(italic)
	list := NewList(item).
		HighlightSymbol(">>").
		HighlightSpacing(HighlightSpacingAlways)
	var state ListState
	buf := renderListStateful(list, &state, 10, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewStyledSpan("  ", italic),
		catatui.NewStyledSpan("Item 1  ", bold.Patch(italic)),
	)))
}

func TestListStyledHighlight(t *testing.T) {
	text := catatui.TextFromString("Item 1").Style(bold)
	item := NewListItemFromText(text).Style(italic)
	state := NewListState().WithSelected(0)
	list := NewList(item).
		HighlightSymbol(">>").
		HighlightStyle(red)
	buf := renderListStateful(list, &state, 10, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(catatui.NewLine(
		catatui.NewStyledSpan(">>", italic.Patch(red)),
		catatui.NewStyledSpan("Item 1  ", bold.Patch(italic).Patch(red)),
	)))
}

func TestListStyleInheritance(t *testing.T) {
	items := []ListItem{
		NewListItemFromText(catatui.TextFromString("Item 1")),                           // no style
		NewListItemFromText(catatui.TextFromStyledString("Item 2", bold)),               // affects only the text
		NewListItemFromText(catatui.TextFromString("Item 3")).Style(italic),             // affects the entire line
		NewListItemFromText(catatui.TextFromStyledString("Item 4", bold)).Style(italic), // bold text, italic line
		NewListItemFromText(catatui.TextFromStyledString("Item 5", bold)).Style(italic), // same but highlighted
	}
	state := NewListState().WithSelected(4)
	list := NewList(items...).
		HighlightSymbol(">>").
		HighlightStyle(red).
		Style(onBlue)
	buf := renderListStateful(list, &state, 10, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithLines(
		styledLine("  Item 1  ", onBlue),
		catatui.NewLine(
			catatui.NewStyledSpan("  ", onBlue),
			catatui.NewStyledSpan("Item 2  ", bold.Patch(onBlue)),
		),
		styledLine("  Item 3  ", italic.Patch(onBlue)),
		catatui.NewLine(
			catatui.NewStyledSpan("  ", italic.Patch(onBlue)),
			catatui.NewStyledSpan("Item 4  ", bold.Patch(italic).Patch(onBlue)),
		),
		catatui.NewLine(
			catatui.NewStyledSpan(">>", italic.Patch(red).Patch(onBlue)),
			catatui.NewStyledSpan("Item 5  ", bold.Patch(italic).Patch(red).Patch(onBlue)),
		),
	))
}

func TestListRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	var state ListState
	list := NewList(NewListItem("Item 1"), NewListItem("Item 2"), NewListItem("Item 3"))
	// This should not panic, even if the buffer is too small to render the list.
	list.RenderStateful(buf.Area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("I"))
}

func TestListRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	var state ListState
	list := NewList(NewListItem("Item 1"), NewListItem("Item 2"), NewListItem("Item 3"))
	// This should not panic, even if the buffer has zero size.
	list.RenderStateful(buf.Area, buf, &state)
}

// --- list/item.rs --------------------------------------------------------

func assertListItem(t *testing.T, item ListItem, content catatui.Text, style catatui.Style) {
	t.Helper()
	if !reflect.DeepEqual(item.GetContent(), content) {
		t.Errorf("content = %+v, want %+v", item.GetContent(), content)
	}
	if item.GetStyle() != style {
		t.Errorf("style = %+v, want %+v", item.GetStyle(), style)
	}
}

func TestListItemNewFromString(t *testing.T) {
	item := NewListItem("Test item")
	assertListItem(t, item, catatui.TextFromString("Test item"), catatui.NewStyle())
}

func TestListItemNewFromSpan(t *testing.T) {
	span := catatui.NewStyledSpan("Test item", catatui.NewStyle().Fg(catatui.ColorBlue))
	item := NewListItemFromLine(catatui.NewLine(span))
	assertListItem(t, item, catatui.NewText(catatui.NewLine(span)), catatui.NewStyle())
}

func TestListItemNewFromSpans(t *testing.T) {
	line := catatui.NewLine(
		catatui.NewStyledSpan("Test ", catatui.NewStyle().Fg(catatui.ColorBlue)),
		catatui.NewStyledSpan("item", catatui.NewStyle().Fg(catatui.ColorRed)),
	)
	item := NewListItemFromLine(line)
	assertListItem(t, item, catatui.NewText(line), catatui.NewStyle())
}

func TestListItemNewFromLines(t *testing.T) {
	lines := []catatui.Line{
		catatui.NewLine(
			catatui.NewStyledSpan("Test ", catatui.NewStyle().Fg(catatui.ColorBlue)),
			catatui.NewStyledSpan("item", catatui.NewStyle().Fg(catatui.ColorRed)),
		),
		catatui.NewLine(
			catatui.NewStyledSpan("Second ", catatui.NewStyle().Fg(catatui.ColorGreen)),
			catatui.NewStyledSpan("line", catatui.NewStyle().Fg(catatui.ColorYellow)),
		),
	}
	item := NewListItemFromText(catatui.NewText(lines...))
	assertListItem(t, item, catatui.NewText(lines...), catatui.NewStyle())
}

func TestListItemStyle(t *testing.T) {
	item := NewListItem("Test item").Style(catatui.NewStyle().Bg(catatui.ColorRed))
	assertListItem(t, item, catatui.TextFromString("Test item"), catatui.NewStyle().Bg(catatui.ColorRed))
}

func TestListItemHeight(t *testing.T) {
	if got := NewListItem("Test item").Height(); got != 1 {
		t.Errorf("Height() = %d, want 1", got)
	}
	if got := NewListItem("Test item\nSecond line").Height(); got != 2 {
		t.Errorf("Height() = %d, want 2", got)
	}
}

func TestListItemWidth(t *testing.T) {
	if got := NewListItem("Test item").Width(); got != 9 {
		t.Errorf("Width() = %d, want 9", got)
	}
	if got := NewListItem("12345\n1234567").Width(); got != 7 {
		t.Errorf("Width() = %d, want 7", got)
	}
}

// --- list/state.rs -------------------------------------------------------

func TestListStateSelected(t *testing.T) {
	var state ListState
	assertSelected(t, state, 0, false)

	state.Select(1)
	assertSelected(t, state, 1, true)

	state.SelectNone()
	assertSelected(t, state, 0, false)
}

func TestListStateSelect(t *testing.T) {
	var state ListState
	assertSelected(t, state, 0, false)
	if state.Offset() != 0 {
		t.Errorf("Offset() = %d, want 0", state.Offset())
	}

	state.Select(2)
	assertSelected(t, state, 2, true)
	if state.Offset() != 0 {
		t.Errorf("Offset() = %d, want 0", state.Offset())
	}

	state.SelectNone()
	assertSelected(t, state, 0, false)
	if state.Offset() != 0 {
		t.Errorf("Offset() = %d, want 0", state.Offset())
	}
}

func TestListStateNavigation(t *testing.T) {
	var state ListState
	state.SelectFirst()
	assertSelected(t, state, 0, true)

	state.SelectPrevious() // should not go below 0
	assertSelected(t, state, 0, true)

	state.SelectNext()
	assertSelected(t, state, 1, true)

	state.SelectPrevious()
	assertSelected(t, state, 0, true)

	state.SelectLast()
	assertSelected(t, state, math.MaxInt, true)

	state.SelectNext() // should not go above MaxInt
	assertSelected(t, state, math.MaxInt, true)

	state.SelectPrevious()
	assertSelected(t, state, math.MaxInt-1, true)

	state.SelectNext()
	assertSelected(t, state, math.MaxInt, true)

	state = ListState{}
	state.SelectNext()
	assertSelected(t, state, 0, true)

	state = ListState{}
	state.SelectPrevious()
	assertSelected(t, state, math.MaxInt, true)

	state = ListState{}
	state.Select(2)
	state.ScrollDownBy(4)
	assertSelected(t, state, 6, true)

	state = ListState{}
	state.ScrollUpBy(3)
	assertSelected(t, state, 0, true)

	state.Select(6)
	state.ScrollUpBy(4)
	assertSelected(t, state, 2, true)

	state.ScrollUpBy(4)
	assertSelected(t, state, 0, true)
}

// --- list/rendering.rs ---------------------------------------------------

func TestListEmptyList(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 1))
	var state ListState
	list := NewList()
	state.SelectFirst()
	list.RenderStateful(buf.Area, buf, &state)
	assertSelected(t, state, 0, false)
}

func TestListSingleItem(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 1))
	var state ListState
	list := NewList(NewListItem("Item 1"))

	state.SelectFirst()
	list.RenderStateful(buf.Area, buf, &state)
	assertSelected(t, state, 0, true)

	state.SelectLast()
	list.RenderStateful(buf.Area, buf, &state)
	assertSelected(t, state, 0, true)

	state.SelectPrevious()
	list.RenderStateful(buf.Area, buf, &state)
	assertSelected(t, state, 0, true)

	state.SelectNext()
	list.RenderStateful(buf.Area, buf, &state)
	assertSelected(t, state, 0, true)
}

func TestListDoesNotRenderInSmallSpace(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2").HighlightSymbol(">>")
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 15, 3))

	// attempt to render into an area of the buffer with 0 width
	list.Render(catatui.NewRect(0, 0, 0, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBuffer(buf.Area))

	// attempt to render into an area of the buffer with 0 height
	list.Render(catatui.NewRect(0, 0, 15, 0), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBuffer(buf.Area))

	// attempt to render into an area of the buffer with zero height after
	// setting the block borders
	list = list.Block(Bordered())
	list.Render(catatui.NewRect(0, 0, 15, 2), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌─────────────┐",
		"└─────────────┘",
		"               ",
	))
}

func TestListCombinations(t *testing.T) {
	var emptyItems []ListItem
	singleItem := []ListItem{NewListItem("Item 0")}
	multipleItems := []ListItem{NewListItem("Item 0"), NewListItem("Item 1"), NewListItem("Item 2")}
	multiLineItems := []ListItem{NewListItem("Item 0\nLine 2"), NewListItem("Item 1"), NewListItem("Item 2")}

	blank := []string{
		"          ",
		"          ",
		"          ",
		"          ",
		"          ",
	}

	stateless := []struct {
		name  string
		items []ListItem
		want  []string
	}{
		{"empty", emptyItems, blank},
		{"single", singleItem, []string{
			"Item 0    ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"multiple", multipleItems, []string{
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
			"          ",
		}},
		{"multi line", multiLineItems, []string{
			"Item 0    ",
			"Line 2    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
		}},
	}
	for _, c := range stateless {
		t.Run("render "+c.name, func(t *testing.T) {
			list := NewList(c.items...).HighlightSymbol(">>")
			catatui.AssertBuffer(t, renderToBuffer(list, 10, 5), catatui.NewBufferWithStrings(c.want...))
		})
	}

	stateful := []struct {
		name        string
		items       []ListItem
		selected    int
		hasSelected bool
		want        []string
	}{
		{"empty none", emptyItems, 0, false, blank},
		{"empty 0", emptyItems, 0, true, blank},
		{"single none", singleItem, 0, false, []string{
			"Item 0    ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"single 0", singleItem, 0, true, []string{
			">>Item 0  ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"single 1", singleItem, 1, true, []string{
			">>Item 0  ",
			"          ",
			"          ",
			"          ",
			"          ",
		}},
		{"multiple none", multipleItems, 0, false, []string{
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
			"          ",
		}},
		{"multiple 0", multipleItems, 0, true, []string{
			">>Item 0  ",
			"  Item 1  ",
			"  Item 2  ",
			"          ",
			"          ",
		}},
		{"multiple 1", multipleItems, 1, true, []string{
			"  Item 0  ",
			">>Item 1  ",
			"  Item 2  ",
			"          ",
			"          ",
		}},
		{"multiple 3", multipleItems, 3, true, []string{
			"  Item 0  ",
			"  Item 1  ",
			">>Item 2  ",
			"          ",
			"          ",
		}},
		{"multi line none", multiLineItems, 0, false, []string{
			"Item 0    ",
			"Line 2    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
		}},
		{"multi line 0", multiLineItems, 0, true, []string{
			">>Item 0  ",
			"  Line 2  ",
			"  Item 1  ",
			"  Item 2  ",
			"          ",
		}},
		{"multi line 1", multiLineItems, 1, true, []string{
			"  Item 0  ",
			"  Line 2  ",
			">>Item 1  ",
			"  Item 2  ",
			"          ",
		}},
	}
	for _, c := range stateful {
		t.Run("stateful "+c.name, func(t *testing.T) {
			list := NewList(c.items...).HighlightSymbol(">>")
			var state ListState
			if c.hasSelected {
				state = state.WithSelected(c.selected)
			}
			catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestListItems(t *testing.T) {
	list := NewList().Items(NewListItem("Item 0"), NewListItem("Item 1"), NewListItem("Item 2"))
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 5), catatui.NewBufferWithStrings(
		"Item 0    ",
		"Item 1    ",
		"Item 2    ",
		"          ",
		"          ",
	))
}

func TestListEmptyStrings(t *testing.T) {
	list := NewListFromStrings("Item 0", "", "", "Item 1", "Item 2").
		Block(Bordered().Title("List"))
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 7), catatui.NewBufferWithStrings(
		"┌List────┐",
		"│Item 0  │",
		"│        │",
		"│        │",
		"│Item 1  │",
		"│Item 2  │",
		"└────────┘",
	))
}

func TestListBlock(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2").Block(Bordered().Title("List"))
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 7), catatui.NewBufferWithStrings(
		"┌List────┐",
		"│Item 0  │",
		"│Item 1  │",
		"│Item 2  │",
		"│        │",
		"│        │",
		"└────────┘",
	))
}

func TestListStyle(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2").Style(red)
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 5), catatui.NewBufferWithLines(
		styledLine("Item 0    ", red),
		styledLine("Item 1    ", red),
		styledLine("Item 2    ", red),
		styledLine("          ", red),
		styledLine("          ", red),
	))
}

func TestListHighlightSymbolAndStyle(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2").
		HighlightSymbol(">>").
		HighlightStyle(yellow)
	var state ListState
	state.Select(1)
	catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithLines(
		catatui.LineFromString("  Item 0  "),
		styledLine(">>Item 1  ", yellow),
		catatui.LineFromString("  Item 2  "),
		catatui.LineFromString("          "),
		catatui.LineFromString("          "),
	))
}

func TestListHighlightSymbolStyleAndStyle(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2").
		HighlightSymbolLine(catatui.LineFromString(">>").Style(red.Patch(bold))).
		HighlightStyle(yellow)
	var state ListState
	state.Select(1)
	expected := catatui.NewBufferWithLines(
		catatui.LineFromString("  Item 0  "),
		styledLine(">>Item 1  ", yellow),
		catatui.LineFromString("  Item 2  "),
		catatui.LineFromString("          "),
		catatui.LineFromString("          "),
	)
	expected.SetStyle(catatui.NewRect(0, 1, 2, 1), red.Patch(bold))
	catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), expected)
}

func TestListHighlightSpacingDefaultWhenSelected(t *testing.T) {
	t.Run("when not selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").HighlightSymbol(">>")
		var state ListState
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
			"          ",
		))
	})
	t.Run("when selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").HighlightSymbol(">>")
		var state ListState
		state.Select(1)
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"  Item 0  ",
			">>Item 1  ",
			"  Item 2  ",
			"          ",
			"          ",
		))
	})
}

func TestListHighlightSpacingDefaultAlways(t *testing.T) {
	t.Run("when not selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").
			HighlightSymbol(">>").
			HighlightSpacing(HighlightSpacingAlways)
		var state ListState
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"  Item 0  ",
			"  Item 1  ",
			"  Item 2  ",
			"          ",
			"          ",
		))
	})
	t.Run("when selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").
			HighlightSymbol(">>").
			HighlightSpacing(HighlightSpacingAlways)
		var state ListState
		state.Select(1)
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"  Item 0  ",
			">>Item 1  ",
			"  Item 2  ",
			"          ",
			"          ",
		))
	})
}

func TestListHighlightSpacingDefaultNever(t *testing.T) {
	t.Run("when not selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").
			HighlightSymbol(">>").
			HighlightSpacing(HighlightSpacingNever)
		var state ListState
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
			"          ",
		))
	})
	t.Run("when selected", func(t *testing.T) {
		list := NewListFromStrings("Item 0", "Item 1", "Item 2").
			HighlightSymbol(">>").
			HighlightSpacing(HighlightSpacingNever)
		var state ListState
		state.Select(1)
		catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), catatui.NewBufferWithStrings(
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
			"          ",
		))
	})
}

func TestListRepeatHighlightSymbol(t *testing.T) {
	list := NewListFromStrings("Item 0\nLine 2", "Item 1", "Item 2").
		HighlightSymbolLine(catatui.LineFromString(">>").Style(red.Patch(bold))).
		HighlightStyle(yellow).
		RepeatHighlightSymbol(true)
	var state ListState
	state.Select(0)
	expected := catatui.NewBufferWithLines(
		styledLine(">>Item 0  ", yellow),
		styledLine(">>Line 2  ", yellow),
		catatui.LineFromString("  Item 1  "),
		catatui.LineFromString("  Item 2  "),
		catatui.LineFromString("          "),
	)
	expected.SetStyle(catatui.NewRect(0, 0, 2, 2), red.Patch(bold))
	catatui.AssertBuffer(t, renderListStateful(list, &state, 10, 5), expected)
}

func TestListDirection(t *testing.T) {
	cases := []struct {
		name      string
		direction ListDirection
		want      []string
	}{
		{"top to bottom", ListDirectionTopToBottom, []string{
			"Item 0    ",
			"Item 1    ",
			"Item 2    ",
			"          ",
		}},
		{"bottom to top", ListDirectionBottomToTop, []string{
			"          ",
			"Item 2    ",
			"Item 1    ",
			"Item 0    ",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list := NewListFromStrings("Item 0", "Item 1", "Item 2").Direction(c.direction)
			catatui.AssertBuffer(t, renderToBuffer(list, 10, 4), catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestListTruncateItems(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4")
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 3), catatui.NewBufferWithStrings(
		"Item 0    ",
		"Item 1    ",
		"Item 2    ",
	))
}

func TestListOffsetRendersShifted(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6")
	state := NewListState().WithOffset(3)
	catatui.AssertBuffer(t, renderListStateful(list, &state, 6, 3),
		catatui.NewBufferWithStrings("Item 3", "Item 4", "Item 5"))
}

func TestListLongLines(t *testing.T) {
	cases := []struct {
		name        string
		selected    int
		hasSelected bool
		want        []string
	}{
		{"none", 0, false, []string{
			"Item 0 with a v",
			"Item 1         ",
			"Item 2         ",
		}},
		{"selected 0", 0, true, []string{
			">>Item 0 with a",
			"  Item 1       ",
			"  Item 2       ",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list := NewListFromStrings(
				"Item 0 with a very long line that will be truncated",
				"Item 1",
				"Item 2",
			).HighlightSymbol(">>")
			var state ListState
			if c.hasSelected {
				state = state.WithSelected(c.selected)
			}
			catatui.AssertBuffer(t, renderListStateful(list, &state, 15, 3), catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestListSelectedItemEnsuresSelectedItemIsVisibleWhenOffsetIsBeforeVisibleRange(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6").
		HighlightSymbol(">>")
	// Set the initial visible range to items 3, 4, and 5
	state := NewListState().WithSelected(1).WithOffset(3)
	buf := renderListStateful(list, &state, 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		">>Item 1  ",
		"  Item 2  ",
		"  Item 3  ",
	))
	assertSelected(t, state, 1, true)
	if state.Offset() != 1 {
		t.Errorf("did not scroll the selected item into view: offset = %d, want 1", state.Offset())
	}
}

func TestListSelectedItemEnsuresSelectedItemIsVisibleWhenOffsetIsAfterVisibleRange(t *testing.T) {
	list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6").
		HighlightSymbol(">>")
	// Set the initial visible range to items 3, 4, and 5
	state := NewListState().WithSelected(6).WithOffset(3)
	buf := renderListStateful(list, &state, 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  Item 4  ",
		"  Item 5  ",
		">>Item 6  ",
	))
	assertSelected(t, state, 6, true)
	if state.Offset() != 4 {
		t.Errorf("did not scroll the selected item into view: offset = %d, want 4", state.Offset())
	}
}

func alignedList(left, center, right string) List {
	return NewList(
		NewListItemFromLine(catatui.LineFromString(left).Left()),
		NewListItemFromLine(catatui.LineFromString(center).Centered()),
		NewListItemFromLine(catatui.LineFromString(right).Right()),
	)
}

func TestListWithAlignment(t *testing.T) {
	buf := renderToBuffer(alignedList("Left", "Center", "Right"), 10, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Left      ", "  Center  ", "     Right", ""))
}

func TestListAlignmentOddLineOddArea(t *testing.T) {
	buf := renderToBuffer(alignedList("Odd", "Even", "Width"), 7, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Odd    ", " Even  ", "  Width", ""))
}

func TestListAlignmentEvenLineEvenArea(t *testing.T) {
	buf := renderToBuffer(alignedList("Odd", "Even", "Width"), 6, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Odd   ", " Even ", " Width", ""))
}

func TestListAlignmentOddLineEvenArea(t *testing.T) {
	buf := renderToBuffer(alignedList("Odd", "Even", "Width"), 8, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Odd     ", "  Even  ", "   Width", ""))
}

func TestListAlignmentEvenLineOddArea(t *testing.T) {
	buf := renderToBuffer(alignedList("Odd", "Even", "Width"), 6, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("Odd   ", " Even ", " Width", ""))
}

func TestListAlignmentZeroLineWidth(t *testing.T) {
	list := NewList(NewListItemFromLine(catatui.LineFromString("This line has zero width").Centered()))
	buf := renderToBuffer(list, 0, 2)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("", ""))
}

func TestListAlignmentZeroAreaWidth(t *testing.T) {
	list := NewList(NewListItemFromLine(catatui.LineFromString("Text").Left()))
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 4, 1))
	list.Render(catatui.NewRect(0, 0, 4, 0), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("    "))
}

func TestListAlignmentLineLessThanWidth(t *testing.T) {
	list := NewList(NewListItemFromLine(catatui.LineFromString("Small").Centered()))
	catatui.AssertBuffer(t, renderToBuffer(list, 10, 2), catatui.NewBufferWithStrings("  Small   ", ""))
}

func TestListAlignmentLineEqualToWidth(t *testing.T) {
	list := NewList(NewListItemFromLine(catatui.LineFromString("Exact").Left()))
	catatui.AssertBuffer(t, renderToBuffer(list, 5, 2), catatui.NewBufferWithStrings("Exact", ""))
}

func TestListAlignmentLineGreaterThanWidth(t *testing.T) {
	list := NewList(NewListItemFromLine(catatui.LineFromString("Large line").Left()))
	catatui.AssertBuffer(t, renderToBuffer(list, 5, 2), catatui.NewBufferWithStrings("Large", ""))
}

func TestListWithPadding(t *testing.T) {
	cases := []struct {
		name         string
		renderHeight uint16
		offset       int
		padding      int
		selected     int
		want         []string
	}{
		{"no padding", 4, 2, 0, 2, []string{
			">> Item 2 ",
			"   Item 3 ",
			"   Item 4 ",
			"   Item 5 ",
		}},
		{"one before", 4, 2, 1, 2, []string{
			"   Item 1 ",
			">> Item 2 ",
			"   Item 3 ",
			"   Item 4 ",
		}},
		{"one after", 4, 1, 1, 4, []string{
			"   Item 2 ",
			"   Item 3 ",
			">> Item 4 ",
			"   Item 5 ",
		}},
		{"check padding overflow", 4, 1, 2, 4, []string{
			"   Item 2 ",
			"   Item 3 ",
			">> Item 4 ",
			"   Item 5 ",
		}},
		{"no padding offset behavior", 5, 2, 0, 3, []string{
			"   Item 2 ",
			">> Item 3 ",
			"   Item 4 ",
			"   Item 5 ",
			"          ",
		}},
		{"two before", 5, 2, 2, 3, []string{
			"   Item 1 ",
			"   Item 2 ",
			">> Item 3 ",
			"   Item 4 ",
			"   Item 5 ",
		}},
		{"keep selected visible", 4, 0, 4, 1, []string{
			"   Item 0 ",
			">> Item 1 ",
			"   Item 2 ",
			"   Item 3 ",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, c.renderHeight))
			var state ListState
			state.SetOffset(c.offset)
			state.Select(c.selected)

			list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "Item 5").
				ScrollPadding(c.padding).
				HighlightSymbol(">> ")
			list.RenderStateful(buf.Area, buf, &state)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

// TestListPaddingFlicker: if there is not enough room for the selected item
// and the requested padding, the list can jump up and down every frame unless
// something is done about it. This checks that it does not.
func TestListPaddingFlicker(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 5))
	var state ListState
	state.SetOffset(2)
	state.Select(4)

	list := NewListFromStrings("Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6", "Item 7").
		ScrollPadding(3).
		HighlightSymbol(">> ")

	list.RenderStateful(buf.Area, buf, &state)
	offsetAfterRender := state.Offset()

	list.RenderStateful(buf.Area, buf, &state)

	// Offset after rendering twice should remain the same as after once
	if state.Offset() != offsetAfterRender {
		t.Errorf("offset changed between renders: %d then %d", offsetAfterRender, state.Offset())
	}
}

func TestListPaddingInconsistentItemSizes(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 3))
	state := NewListState().WithOffset(0).WithSelected(3)

	list := NewList(
		NewListItem("Item 0"),
		NewListItem("Item 1"),
		NewListItem("Item 2"),
		NewListItem("Item 3"),
		NewListItem("Item 4\nTest\nTest"),
		NewListItem("Item 5"),
	).ScrollPadding(1).HighlightSymbol(">> ")

	list.RenderStateful(buf.Area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   Item 1 ",
		"   Item 2 ",
		">> Item 3 ",
	))
}

// TestListPaddingOffsetPushbackBreak makes sure that when the first visible
// index is pushed back, an item that is too large is not included.
func TestListPaddingOffsetPushbackBreak(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 4))
	var state ListState
	state.SetOffset(1)
	state.Select(2)

	list := NewList(
		NewListItem("Item 0\nTest\nTest"),
		NewListItem("Item 1"),
		NewListItem("Item 2"),
		NewListItem("Item 3"),
	).ScrollPadding(2).HighlightSymbol(">> ")

	list.RenderStateful(buf.Area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   Item 1 ",
		">> Item 2 ",
		"   Item 3 ",
		"          ",
	))
}

// TestListHighlightSymbolOverflow is a regression test for a highlight symbol
// wider than the area causing an underflow (ratatui #949).
func TestListHighlightSymbolOverflow(t *testing.T) {
	cases := []struct {
		name, symbol, item, want string
	}{
		{"under", ">>>>", "Item1", ">>>>Item1 "},      // enough space to render the highlight symbol
		{"exact", ">>>>>", "Item1", ">>>>>Item1"},     // exact space to render the highlight symbol
		{"overflow", ">>>>>>", "Item1", ">>>>>>Item"}, // not enough space
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list := NewListFromStrings(c.item).HighlightSymbol(c.symbol)
			var state ListState
			state.Select(0)
			buf := renderListStateful(list, &state, 10, 1)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want))
		})
	}
}

// TestListImplementsWidgetInterfaces pins the two interfaces a List satisfies.
func TestListImplementsWidgetInterfaces(t *testing.T) {
	var _ catatui.Widget = List{}
	var _ catatui.StatefulWidget[ListState] = List{}
	// A stateful render through the generic helper must work the same way.
	list := NewListFromStrings("a", "b").HighlightSymbol(">")
	state := NewListState().WithSelected(1)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 3, 2))
	catatui.RenderStatefulWidget(list, buf.Area, buf, &state)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" a ", ">b "))
}
