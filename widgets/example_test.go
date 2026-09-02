package widgets_test

import (
	"fmt"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// draw renders a widget into a buffer of the given size and prints it, one row
// per line with the padding on the right trimmed.
func draw(w catatui.Widget, width, height uint16) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, height))
	w.Render(buf.Area, buf)
	for _, row := range strings.Split(buf.String(), "\n") {
		fmt.Println(strings.TrimRight(row, " "))
	}
}

func ExampleBlock() {
	draw(widgets.Bordered().Title("Files"), 16, 3)
	// Output:
	// ┌Files─────────┐
	// │              │
	// └──────────────┘
}

// Inner is the area a block leaves for its content: what is left after the
// borders that are drawn, the padding, and any title row.
func ExampleBlock_Inner() {
	area := catatui.NewRect(0, 0, 20, 10)

	fmt.Println(widgets.Bordered().Inner(area))
	fmt.Println(widgets.Bordered().Padding(widgets.UniformPadding(2)).Inner(area))
	fmt.Println(widgets.NewBlock().Borders(widgets.BordersLeft).Inner(area))
	// Output:
	// {1 1 18 8}
	// {3 3 14 4}
	// {1 0 19 10}
}

func ExampleParagraph() {
	text := "Slice, layer, and bake the vegetables."
	draw(widgets.NewParagraph(text).Wrap(widgets.Wrap{Trim: true}), 14, 4)
	// Output:
	// Slice, layer,
	// and bake the
	// vegetables.
}

// A list keeps its selection in a ListState the caller owns, so the selection
// survives the list being rebuilt each frame.
func ExampleList() {
	state := widgets.NewListState().WithSelected(1)

	list := widgets.NewListFromStrings("Eggplant", "Tomato", "Zucchini").
		HighlightSymbol("> ")

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 12, 3))
	catatui.RenderStatefulWidget(list, buf.Area, buf, &state)
	for _, row := range strings.Split(buf.String(), "\n") {
		fmt.Println(strings.TrimRight(row, " "))
	}
	// Output:
	//   Eggplant
	// > Tomato
	//   Zucchini
}

func ExampleGauge() {
	draw(widgets.NewGauge().Percent(40).Label("40%"), 20, 1)
	// Output:
	// ████████40%
}

// Merging turns the two borders where blocks meet into the one character that
// shows both lines. The block drawn last reads what is underneath, so it has to
// be rendered after its neighbour.
func ExampleBlock_MergeBorders() {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 9, 3))

	widgets.Bordered().Render(catatui.NewRect(0, 0, 5, 3), buf)
	widgets.Bordered().
		MergeBorders(symbols.MergeExact).
		Render(catatui.NewRect(4, 0, 5, 3), buf)

	for _, row := range strings.Split(buf.String(), "\n") {
		fmt.Println(strings.TrimRight(row, " "))
	}
	// Output:
	// ┌───┬───┐
	// │   │   │
	// └───┴───┘
}
