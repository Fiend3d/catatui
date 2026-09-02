// Port of ratatui-widgets/src/list.rs and list/rendering.rs @ ratatui-v0.30.2

package widgets

import (
	"strings"

	"github.com/Fiend3d/catatui"
)

// ListDirection is the order in which a List lays out its items.
//
// If there are too few items to fill the area, the list sticks to the edge it
// starts from.
type ListDirection uint8

const (
	// ListDirectionTopToBottom puts the first item at the top, which is the
	// default.
	ListDirectionTopToBottom ListDirection = iota
	// ListDirectionBottomToTop puts the first item at the bottom.
	ListDirectionBottomToTop
)

// String returns the direction's ratatui name.
func (d ListDirection) String() string {
	if d == ListDirectionBottomToTop {
		return "BottomToTop"
	}
	return "TopToBottom"
}

// List draws a column of ListItems, optionally inside a Block, and lets one
// of them be selected.
//
// It differs from a Table in that it has no columns, headers or footers, each
// item's height follows from its text, and it can run bottom to top.
//
//	list := widgets.NewListFromStrings("Item 1", "Item 2", "Item 3").
//		Block(widgets.Bordered().Title("List")).
//		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierReversed)).
//		HighlightSymbol(">>")
//	f.RenderWidget(list, area)
//
// Rendered through catatui.RenderStatefulWidget with a ListState, the list
// highlights the selected item, draws the highlight symbol in front of it and
// scrolls so that it stays visible:
//
//	catatui.RenderStatefulWidget(list, area, buf, &state)
//
// Rendered as a plain Widget, it behaves as if nothing were selected.
type List struct {
	block                 Block
	hasBlock              bool
	items                 []ListItem
	style                 catatui.Style
	direction             ListDirection
	highlightStyle        catatui.Style
	highlightSymbol       catatui.Line
	hasHighlightSymbol    bool
	repeatHighlightSymbol bool
	highlightSpacing      HighlightSpacing
	scrollPadding         int
}

// NewList returns a list of the given items. With no items it is the same as
// the zero value, to which items can be added with Items.
func NewList(items ...ListItem) List {
	return List{items: items}
}

// NewListFromStrings returns a list with one unstyled item per string.
func NewListFromStrings(items ...string) List {
	list := List{items: make([]ListItem, len(items))}
	for i, s := range items {
		list.items[i] = NewListItem(s)
	}
	return list
}

// Items returns a copy of l with the items replaced.
func (l List) Items(items ...ListItem) List { l.items = items; return l }

// Block returns a copy of l drawn inside the given block.
func (l List) Block(b Block) List { l.block, l.hasBlock = b, true; return l }

// Style returns a copy of l with a base style applied to the whole area. It
// sits beneath the block's style, each item's style and the styles of the
// items' text.
func (l List) Style(s catatui.Style) List { l.style = s; return l }

// GetStyle returns the list's base style.
func (l List) GetStyle() catatui.Style { return l.style }

// HighlightSymbol returns a copy of l that draws the given symbol in front of
// the selected item. There is no symbol by default.
func (l List) HighlightSymbol(symbol string) List {
	return l.HighlightSymbolLine(catatui.LineFromString(symbol))
}

// HighlightSymbolLine is HighlightSymbol for a styled symbol.
func (l List) HighlightSymbolLine(symbol catatui.Line) List {
	l.highlightSymbol, l.hasHighlightSymbol = symbol, true
	return l
}

// HighlightStyle returns a copy of l with the style applied to the selected
// item. It covers the whole row, highlight symbol included, and is layered
// over the item's own styles.
func (l List) HighlightStyle(s catatui.Style) List { l.highlightStyle = s; return l }

// RepeatHighlightSymbol returns a copy of l that, when set, draws the
// highlight symbol on every line of a multi-line selected item rather than on
// its first line only. It is off by default.
func (l List) RepeatHighlightSymbol(repeat bool) List { l.repeatHighlightSymbol = repeat; return l }

// HighlightSpacing returns a copy of l with the rule for when the highlight
// symbol's column is reserved. See HighlightSpacing for the choices; the
// default reserves it only while an item is selected.
func (l List) HighlightSpacing(h HighlightSpacing) List { l.highlightSpacing = h; return l }

// Direction returns a copy of l laid out in the given direction.
func (l List) Direction(d ListDirection) List { l.direction = d; return l }

// ScrollPadding returns a copy of l that tries to keep the given number of
// items visible above and below the selected one when scrolling.
func (l List) ScrollPadding(padding int) List { l.scrollPadding = max(padding, 0); return l }

// Len is the number of items in the list.
func (l List) Len() int { return len(l.items) }

// IsEmpty reports whether the list has no items.
func (l List) IsEmpty() bool { return len(l.items) == 0 }

// Render draws the list as if nothing were selected.
func (l List) Render(area catatui.Rect, buf *catatui.Buffer) {
	var state ListState
	l.RenderStateful(area, buf, &state)
}

// RenderStateful draws the list, highlighting the selected item and updating
// the state's offset so that it is visible. A selection past the end of the
// list is moved to the last item, and the selection is cleared if the list is
// empty.
func (l List) RenderStateful(area catatui.Rect, buf *catatui.Buffer, state *ListState) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, l.style)
	listArea := area
	if l.hasBlock {
		l.block.Render(area, buf)
		listArea = l.block.Inner(area)
	}
	if listArea.IsEmpty() {
		return
	}

	if len(l.items) == 0 {
		state.SelectNone()
		return
	}

	// A selection past the end lands on the last item.
	if state.hasSelected && state.selected >= len(l.items) {
		state.Select(len(l.items) - 1)
	}

	listHeight := int(listArea.Height)
	firstVisible, lastVisible := l.getItemsBounds(state.selected, state.hasSelected, state.offset, listHeight)

	// This is where the state changes: the offset becomes the first of the
	// items now in view.
	state.offset = firstVisible

	var highlightSymbol catatui.Line
	if l.hasHighlightSymbol {
		highlightSymbol = l.highlightSymbol
	}
	highlightSymbolWidth := uint16(min(highlightSymbol.Width(), 0xFFFF))
	emptySymbol := catatui.LineFromString(strings.Repeat(" ", int(highlightSymbolWidth)))

	var currentHeight uint16
	selectionSpacing := l.highlightSpacing.ShouldAdd(state.hasSelected)
	for i := firstVisible; i < lastVisible; i++ {
		item := l.items[i]
		itemHeight := uint16(min(item.Height(), 0xFFFF))

		var x, y uint16
		if l.direction == ListDirectionBottomToTop {
			currentHeight = catatui.SatAdd(currentHeight, itemHeight)
			x, y = listArea.Left(), catatui.SatSub(listArea.Bottom(), currentHeight)
		} else {
			x, y = listArea.Left(), catatui.SatAdd(listArea.Top(), currentHeight)
			currentHeight = catatui.SatAdd(currentHeight, itemHeight)
		}

		rowArea := catatui.NewRect(x, y, listArea.Width, itemHeight)

		itemStyle := l.style.Patch(item.style)
		buf.SetStyle(rowArea, itemStyle)

		isSelected := state.hasSelected && state.selected == i

		itemArea := rowArea
		if selectionSpacing {
			itemArea.X = catatui.SatAdd(rowArea.X, highlightSymbolWidth)
			itemArea.Width = catatui.SatSub(rowArea.Width, highlightSymbolWidth)
		}
		item.content.Render(itemArea, buf)

		if isSelected {
			buf.SetStyle(rowArea, l.highlightStyle)
		}
		if selectionSpacing {
			for j := range item.content.Height() {
				// The selected item gets the symbol on its first line, or on
				// every line when RepeatHighlightSymbol is set; every other
				// line gets blanks of the same width.
				line := emptySymbol
				if isSelected && (j == 0 || l.repeatHighlightSymbol) {
					line = highlightSymbol
				}
				highlightArea := catatui.NewRect(x, catatui.SatAdd(y, uint16(min(j, 0xFFFF))), highlightSymbolWidth, 1)
				line.Render(highlightArea, buf)
			}
		}
	}
}

// getItemsBounds works out, from an offset, the half-open range of items that
// fit in maxHeight rows while keeping the selected item (or, with nothing
// selected, the offset item) in view. It must only be called on a non-empty
// list.
func (l List) getItemsBounds(selected int, hasSelected bool, offset, maxHeight int) (first, last int) {
	offset = min(offset, max(len(l.items)-1, 0))

	// "Visible" here means visible in the given area.
	first, last = offset, offset

	// Total height of the items rendered so far, starting at the offset.
	heightFromOffset := 0

	// Find the last visible index and the total height of the items that fit
	// in the available space.
	for _, item := range l.items[offset:] {
		if heightFromOffset+item.Height() > maxHeight {
			break
		}
		heightFromOffset += item.Height()
		last++
	}

	// Apply the scroll padding to the selection, but still honour the offset
	// when nothing is selected, so the list stays put after SelectNone.
	indexToDisplay := offset
	if idx, ok := l.applyScrollPaddingToSelectedIndex(selected, hasSelected, maxHeight, first, last); ok {
		indexToDisplay = idx
	}

	// If the item to display is past what fits after the offset, extend the
	// range to it and drop items from the front to make room.
	for indexToDisplay >= last {
		heightFromOffset += l.items[last].Height()
		last++
		for heightFromOffset > maxHeight {
			heightFromOffset = max(heightFromOffset-l.items[first].Height(), 0)
			first++
		}
	}

	// And the same the other way: if it is before the offset, extend the
	// range back to it and drop items from the end.
	for indexToDisplay < first {
		first--
		heightFromOffset += l.items[first].Height()
		for heightFromOffset > maxHeight {
			last--
			heightFromOffset = max(heightFromOffset-l.items[last].Height(), 0)
		}
	}

	return first, last
}

// applyScrollPaddingToSelectedIndex returns the index that has to be in view
// for the selected item to have its scroll padding around it, shrinking the
// padding as needed so that the selected item itself always fits, even when
// the items are of uneven height.
//
// This is sensitive to how getItemsBounds treats item heights.
func (l List) applyScrollPaddingToSelectedIndex(selected int, hasSelected bool, maxHeight, first, last int) (int, bool) {
	if !hasSelected {
		return 0, false
	}
	lastValid := max(len(l.items)-1, 0)
	selected = min(selected, lastValid)

	// Uneven item sizes can mean the padding excludes items that should be
	// shown, or makes the offset differ from one render to the next. The
	// padding is reduced until the padded range fits.
	padding := l.scrollPadding
	for padding > 0 {
		heightAroundSelected := 0
		lo := max(selected-padding, 0)
		hi := min(satAddInt(selected, padding), lastValid)
		for i := lo; i <= hi; i++ {
			heightAroundSelected += l.items[i].Height()
		}
		if heightAroundSelected <= maxHeight {
			break
		}
		padding--
	}

	var index int
	switch {
	case min(selected+padding, lastValid) >= last:
		index = selected + padding
	case max(selected-padding, 0) < first:
		index = max(selected-padding, 0)
	default:
		index = selected
	}
	return min(index, lastValid), true
}

var (
	_ catatui.Widget                    = List{}
	_ catatui.StatefulWidget[ListState] = List{}
)
