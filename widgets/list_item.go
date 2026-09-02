// Port of ratatui-widgets/src/list/item.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// ListItem is a single entry in a List.
//
// An item is a Text plus a style for the whole row. Its height is the number
// of lines in the text, and its width is that of the widest line. The item's
// style sits beneath the text's own styles, so a style on the Text or on one of
// its spans wins over the item's.
//
//	item := widgets.NewListItem("Multi-line\nitem").
//		Style(catatui.NewStyle().Fg(catatui.ColorRed))
//
// An item is aligned by aligning its Text or its Lines: a right-aligned Text
// gives a right-aligned item, and a single Line can override that.
type ListItem struct {
	content catatui.Text
	style   catatui.Style
}

// NewListItem returns an unstyled item holding the given text, split on
// newlines so that a multi-line string becomes a multi-line item.
func NewListItem(content string) ListItem {
	return ListItem{content: catatui.TextFromString(content)}
}

// NewListItemFromText returns an item holding already-styled text.
func NewListItemFromText(content catatui.Text) ListItem {
	return ListItem{content: content}
}

// NewListItemFromLine returns a one-line item holding a styled line.
func NewListItemFromLine(line catatui.Line) ListItem {
	return ListItem{content: catatui.NewText(line)}
}

// Style returns a copy of i with the row style replaced. The style can be
// overridden by the style of the item's Text.
func (i ListItem) Style(s catatui.Style) ListItem { i.style = s; return i }

// GetContent returns the item's text.
func (i ListItem) GetContent() catatui.Text { return i.content }

// GetStyle returns the item's row style.
func (i ListItem) GetStyle() catatui.Style { return i.style }

// Height is the number of lines the item occupies.
func (i ListItem) Height() int { return i.content.Height() }

// Width is the number of columns of the item's widest line.
func (i ListItem) Width() int { return i.content.Width() }
