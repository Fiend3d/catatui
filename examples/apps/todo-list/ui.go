// Port of the rendering half of examples/apps/todo-list @ ratatui-v0.30.2

package main

import (
	"fmt"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// The colours, straight from the tailwind palette as ratatui's example takes
// them. The rows alternate between two backgrounds so that a long list is
// easier to follow across.
var (
	todoHeaderStyle      = catatui.NewStyle().Fg(tailwind.Slate.C100).Bg(tailwind.Blue.C800)
	normalRowBg          = tailwind.Slate.C950
	altRowBgColor        = tailwind.Slate.C900
	selectedStyle        = catatui.NewStyle().Bg(tailwind.Slate.C800).AddModifier(catatui.ModifierBold)
	textFgColor          = tailwind.Slate.C200
	completedTextFgColor = tailwind.Green.C500
)

// Render draws the header, the list, the selected item's info and the footer.
// The app is a widget itself, which is how ratatui's example is written.
func (a *app) Render(area catatui.Rect, buf *catatui.Buffer) {
	rows := catatui.VerticalLayout(
		catatui.Length(2),
		catatui.Fill(1),
		catatui.Length(1),
	).Split(area)
	headerArea, contentArea, footerArea := rows[0], rows[1], rows[2]

	content := catatui.VerticalLayout(catatui.Fill(1), catatui.Fill(1)).Split(contentArea)
	listArea, itemArea := content[0], content[1]

	renderHeader(headerArea, buf)
	renderFooter(footerArea, buf)
	a.renderList(listArea, buf)
	a.renderSelectedItem(itemArea, buf)
}

func renderHeader(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Catatui Todo List Example").
		Style(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		Centered().
		Render(area, buf)
}

func renderFooter(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Use ↓↑ to move, ← to unselect, → to change status, g/G to go top/bottom.").
		Centered().
		Render(area, buf)
}

// renderList draws the items, highlighting whichever one is selected.
func (a *app) renderList(area catatui.Rect, buf *catatui.Buffer) {
	// An empty border set draws nothing but still reserves the row, which is
	// what turns the top border into a coloured header bar.
	block := widgets.NewBlock().
		TitleLine(catatui.LineFromString("TODO List").Centered()).
		Borders(widgets.BordersTop).
		BorderSet(symbols.BorderEmpty).
		BorderStyle(todoHeaderStyle).
		Style(catatui.NewStyle().Bg(normalRowBg))

	items := make([]widgets.ListItem, len(a.todoList.items))
	for i, item := range a.todoList.items {
		items[i] = item.listItem().Style(catatui.NewStyle().Bg(alternateColors(i)))
	}

	list := widgets.NewList(items...).
		Block(block).
		HighlightStyle(selectedStyle).
		HighlightSymbol(">").
		HighlightSpacing(widgets.HighlightSpacingAlways)

	// The list is a stateful widget, so it renders through the free function
	// rather than as a plain Widget. See the note on StatefulWidget.
	catatui.RenderStatefulWidget(list, area, buf, &a.todoList.state)
}

// renderSelectedItem draws the longer text of whichever item is selected.
func (a *app) renderSelectedItem(area catatui.Rect, buf *catatui.Buffer) {
	info := "Nothing selected..."
	if i, ok := a.todoList.state.Selected(); ok && i < len(a.todoList.items) {
		item := a.todoList.items[i]
		if item.status == statusCompleted {
			info = fmt.Sprintf("✓ DONE: %s", item.info)
		} else {
			info = fmt.Sprintf("☐ TODO: %s", item.info)
		}
	}

	block := widgets.NewBlock().
		TitleLine(catatui.LineFromString("TODO Info").Centered()).
		Borders(widgets.BordersTop).
		BorderSet(symbols.BorderEmpty).
		BorderStyle(todoHeaderStyle).
		Style(catatui.NewStyle().Bg(normalRowBg)).
		Padding(widgets.HorizontalPadding(1))

	widgets.NewParagraph(info).
		Block(block).
		Style(catatui.NewStyle().Fg(textFgColor)).
		Wrap(widgets.Wrap{Trim: false}).
		Render(area, buf)
}

// listItem is the one-line form of an item: a tick box, the text, and a colour
// that says whether it is done.
func (i todoItem) listItem() widgets.ListItem {
	if i.status == statusCompleted {
		return widgets.NewListItemFromLine(catatui.LineFromStyledString(
			" ✓ "+i.todo, catatui.NewStyle().Fg(completedTextFgColor)))
	}
	return widgets.NewListItemFromLine(catatui.LineFromStyledString(
		" ☐ "+i.todo, catatui.NewStyle().Fg(textFgColor)))
}

// alternateColors gives every other row the lighter background.
func alternateColors(i int) catatui.Color {
	if i%2 == 0 {
		return normalRowBg
	}
	return altRowBgColor
}
