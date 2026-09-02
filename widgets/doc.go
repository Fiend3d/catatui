// Package widgets holds catatui's widget library: the ready-made things you
// render into a Frame.
//
// Port of ratatui-widgets @ ratatui-v0.30.2.
//
// # The catalogue
//
// Frames and text:
//
//   - Block — borders on any combination of sides, titles top and bottom,
//     padding, and an optional shadow. Mostly used for its Inner method: draw
//     the block, then draw the content into the area it leaves.
//   - Paragraph — styled text with wrapping, scrolling and alignment.
//   - Clear — blanks an area, which is how a popup stops showing what is under
//     it.
//   - Fill — fills an area with one symbol.
//
// Lists and tables:
//
//   - List, ListItem, ListState — a scrolling list with a selection, top to
//     bottom or bottom to top.
//   - Table, Row, Cell, TableState — columns sized by Constraint, with row,
//     column and cell selection.
//   - Tabs — a row of titles with one highlighted.
//   - Scrollbar, ScrollbarState — a track and thumb along any edge.
//
// Numbers:
//
//   - Gauge, LineGauge — a progress bar, in a block or on one row.
//   - BarChart, Bar, BarGroup — bars in either direction, grouped or not.
//   - Sparkline — one row of history, no axes.
//   - Chart, Dataset, Axis — scatter, line, bar and area plots on shared axes.
//
// Drawing:
//
//   - Canvas, Context, Painter — a coordinate space drawn with braille, block
//     or half-block markers, with the CanvasLine, Rectangle, Circle, Points
//     and Map shapes.
//   - Monthly, CalendarEventStore — a month, with styles per date.
//   - CatatuiLogo, RatatuiMascot — the logo and the rat.
//
// # State
//
// Most widgets are values you build and throw away each frame. The ones that
// have to remember something between frames — which row is selected, how far a
// view has scrolled — keep that in a separate state value the caller owns:
//
//	state := widgets.NewListState().WithSelected(0)
//
//	// each frame:
//	list := widgets.NewListFromStrings("one", "two", "three").
//		HighlightSymbol("> ")
//	catatui.RenderStatefulWidget(list, area, f.Buffer(), &state)
//
// The state outlives the widget, so the selection survives the list being
// rebuilt. RenderStatefulWidget is a free function because Go methods cannot
// have type parameters of their own.
//
// # Composition
//
// Widgets that can be framed take a Block, and use whatever it leaves inside:
//
//	widgets.NewParagraph("hello").Block(widgets.Bordered().Title("catatui"))
//
// To draw something else inside a block, render the block and ask it for the
// inner area:
//
//	block := widgets.Bordered().Title("Files")
//	inner := block.Inner(area)
//	f.RenderWidget(block, area)
//	f.RenderWidget(list, inner)
//
// Blocks that touch can share one border rather than drawing two, by merging
// their border characters — see Block.MergeBorders and the collapsed-borders
// example.
//
// # Conventions
//
// Builders return modified copies, so a partly built widget stays usable and
// can be shared. Fields are unexported behind ratatui's builder names, and the
// getters are prefixed with Get: Style sets, GetStyle reads.
//
// Every widget is safe to render into an area too small for it, including a
// zero-sized one. Nothing draws outside the area it is given.
package widgets
