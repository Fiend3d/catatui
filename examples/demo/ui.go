// Port of examples/apps/demo/src/ui.rs @ ratatui-v0.30.2

package main

import (
	"fmt"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// render draws the tab bar and whichever tab is selected.
func render(f *catatui.Frame, a *app) {
	rows := catatui.VerticalLayout(catatui.Length(3), catatui.Min(0)).Split(f.Area())

	titles := make([]catatui.Line, len(a.tabs.titles))
	for i, t := range a.tabs.titles {
		titles[i] = catatui.LineFromStyledString(t, catatui.NewStyle().Fg(catatui.ColorGreen))
	}
	tabs := widgets.NewTabsFromLines(titles...).
		Block(widgets.Bordered().Title(a.title)).
		HighlightStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
		Select(a.tabs.index)
	f.RenderWidget(tabs, rows[0])

	switch a.tabs.index {
	case 0:
		drawFirstTab(f, a, rows[1])
	case 1:
		drawSecondTab(f, a, rows[1])
	case 2:
		drawThirdTab(f, rows[1])
	}
}

// drawFirstTab stacks the gauges, the charts and the footer text.
func drawFirstTab(f *catatui.Frame, a *app, area catatui.Rect) {
	rows := catatui.VerticalLayout(
		catatui.Length(9),
		catatui.Min(8),
		catatui.Length(7),
	).Split(area)

	drawGauges(f, a, rows[0])
	drawCharts(f, a, rows[1])
	drawText(f, rows[2])
}

// drawGauges shows the same progress value three ways, inside one block.
func drawGauges(f *catatui.Frame, a *app, area catatui.Rect) {
	rows := catatui.VerticalLayout(
		catatui.Length(2),
		catatui.Length(3),
		catatui.Length(2),
	).Margin(1).Split(area)

	f.RenderWidget(widgets.Bordered().Title("Graphs"), area)

	gauge := widgets.NewGauge().
		Block(widgets.NewBlock().Title("Gauge:")).
		GaugeStyle(catatui.NewStyle().
			Fg(catatui.ColorMagenta).
			Bg(catatui.ColorBlack).
			AddModifier(catatui.ModifierItalic | catatui.ModifierBold)).
		UseUnicode(a.enhancedGraphics).
		Label(fmt.Sprintf("%.2f%%", a.progress*100)).
		Ratio(a.progress)
	f.RenderWidget(gauge, rows[0])

	sparkline := widgets.NewSparkline().
		Block(widgets.NewBlock().Title("Sparkline:")).
		Style(catatui.NewStyle().Fg(catatui.ColorGreen)).
		Data(a.sparkline.points...).
		BarSet(a.barSet())
	f.RenderWidget(sparkline, rows[1])

	// The heavier line reads as a bar; the plain one is there for fonts that
	// do not have it.
	lineSymbol := symbols.Horizontal
	if a.enhancedGraphics {
		lineSymbol = symbols.ThickHorizontal
	}
	lineGauge := widgets.NewLineGauge().
		Block(widgets.NewBlock().Title("LineGauge:")).
		FilledStyle(catatui.NewStyle().Fg(catatui.ColorMagenta)).
		FilledSymbol(lineSymbol).
		UnfilledSymbol(lineSymbol).
		Ratio(a.progress)
	f.RenderWidget(lineGauge, rows[2])
}

// drawCharts fills the middle of the first tab: two lists over a bar chart on
// the left, and the sine chart on the right when it is turned on.
func drawCharts(f *catatui.Frame, a *app, area catatui.Rect) {
	constraints := []catatui.Constraint{catatui.Percentage(100)}
	if a.showChart {
		constraints = []catatui.Constraint{catatui.Percentage(50), catatui.Percentage(50)}
	}
	columns := catatui.HorizontalLayout(constraints...).Split(area)

	left := catatui.VerticalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Split(columns[0])
	lists := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Split(left[0])

	drawTasks(f, a, lists[0])
	drawLogs(f, a, lists[1])
	drawBarChart(f, a, left[1])

	if a.showChart {
		drawChart(f, a, columns[1])
	}
}

// drawTasks is the list whose selection the arrow keys move.
func drawTasks(f *catatui.Frame, a *app, area catatui.Rect) {
	items := make([]widgets.ListItem, len(a.tasks.items))
	for i, task := range a.tasks.items {
		items[i] = widgets.NewListItemFromLine(catatui.LineFromString(task))
	}
	list := widgets.NewList(items...).
		Block(widgets.Bordered().Title("List")).
		HighlightStyle(catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		HighlightSymbol("> ")
	catatui.RenderStatefulWidget(list, area, f.Buffer(), &a.tasks.state)
}

// drawLogs colours each line by its level.
func drawLogs(f *catatui.Frame, a *app, area catatui.Rect) {
	levelStyles := map[string]catatui.Style{
		"ERROR":    catatui.NewStyle().Fg(catatui.ColorMagenta),
		"CRITICAL": catatui.NewStyle().Fg(catatui.ColorRed),
		"WARNING":  catatui.NewStyle().Fg(catatui.ColorYellow),
	}
	info := catatui.NewStyle().Fg(catatui.ColorBlue)

	items := make([]widgets.ListItem, len(a.logs.items))
	for i, log := range a.logs.items {
		style, ok := levelStyles[log.level]
		if !ok {
			style = info
		}
		items[i] = widgets.NewListItemFromLine(catatui.NewLine(
			catatui.NewStyledSpan(fmt.Sprintf("%-9s", log.level), style),
			catatui.NewSpan(log.event),
		))
	}
	list := widgets.NewList(items...).Block(widgets.Bordered().Title("List"))
	catatui.RenderStatefulWidget(list, area, f.Buffer(), &a.logs.state)
}

func drawBarChart(f *catatui.Frame, a *app, area catatui.Rect) {
	barchart := widgets.NewBarChart().
		Block(widgets.Bordered().Title("Bar chart")).
		DataPairs(a.barchart...).
		BarWidth(3).
		BarGap(2).
		BarSet(a.barSet()).
		ValueStyle(catatui.NewStyle().
			Fg(catatui.ColorBlack).
			Bg(catatui.ColorGreen).
			AddModifier(catatui.ModifierItalic)).
		LabelStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
		BarStyle(catatui.NewStyle().Fg(catatui.ColorGreen))
	f.RenderWidget(barchart, area)
}

// drawChart plots the two sine waves against the scrolling window.
func drawChart(f *catatui.Frame, a *app, area catatui.Rect) {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	window := a.signals.window

	xLabels := []catatui.Line{
		catatui.LineFromStyledString(fmt.Sprintf("%g", window[0]), bold),
		catatui.LineFromString(fmt.Sprintf("%g", (window[0]+window[1])/2)),
		catatui.LineFromStyledString(fmt.Sprintf("%g", window[1]), bold),
	}

	datasets := []widgets.Dataset{
		widgets.NewDataset().
			Name("data2").
			Marker(symbols.Dot).
			Style(catatui.NewStyle().Fg(catatui.ColorCyan)).
			Data(a.signals.sin1.points),
		widgets.NewDataset().
			Name("data3").
			Marker(a.marker()).
			Style(catatui.NewStyle().Fg(catatui.ColorYellow)).
			Data(a.signals.sin2.points),
	}

	gray := catatui.NewStyle().Fg(catatui.ColorGray)
	chart := widgets.NewChart(datasets...).
		Block(widgets.Bordered().TitleLine(catatui.LineFromStyledString("Chart",
			catatui.NewStyle().Fg(catatui.ColorCyan).AddModifier(catatui.ModifierBold)))).
		XAxis(widgets.NewAxis().
			Title("X Axis").
			Style(gray).
			Bounds(window).
			LabelLines(xLabels...)).
		YAxis(widgets.NewAxis().
			Title("Y Axis").
			Style(gray).
			Bounds([2]float64{-20, 20}).
			LabelLines(
				catatui.LineFromStyledString("-20", bold),
				catatui.LineFromString("0"),
				catatui.LineFromStyledString("20", bold),
			))
	f.RenderWidget(chart, area)
}

// drawText is the footer: a paragraph showing what styling looks like.
func drawText(f *catatui.Frame, area catatui.Rect) {
	mod := func(m catatui.Modifier) catatui.Style {
		return catatui.NewStyle().AddModifier(m)
	}
	fg := func(c catatui.Color) catatui.Style { return catatui.NewStyle().Fg(c) }

	text := catatui.NewText(
		catatui.LineFromString(
			"This is a paragraph with several lines. You can change style your text the way you want"),
		catatui.LineFromString(""),
		catatui.NewLine(
			catatui.NewSpan("For example: "),
			catatui.NewStyledSpan("under", fg(catatui.ColorRed)),
			catatui.NewSpan(" "),
			catatui.NewStyledSpan("the", fg(catatui.ColorGreen)),
			catatui.NewSpan(" "),
			catatui.NewStyledSpan("rainbow", fg(catatui.ColorBlue)),
			catatui.NewSpan("."),
		),
		catatui.NewLine(
			catatui.NewSpan("Oh and if you did not "),
			catatui.NewStyledSpan("notice", mod(catatui.ModifierItalic)),
			catatui.NewSpan(" you can "),
			catatui.NewStyledSpan("automatically", mod(catatui.ModifierBold)),
			catatui.NewSpan(" "),
			catatui.NewStyledSpan("wrap", mod(catatui.ModifierReversed)),
			catatui.NewSpan(" your "),
			catatui.NewStyledSpan("text", mod(catatui.ModifierUnderlined)),
			catatui.NewSpan("."),
		),
		catatui.LineFromString(
			"One more thing is that it should display unicode characters: 10€"),
	)

	block := widgets.Bordered().TitleLine(catatui.LineFromStyledString("Footer",
		catatui.NewStyle().Fg(catatui.ColorMagenta).AddModifier(catatui.ModifierBold)))
	f.RenderWidget(
		widgets.NewParagraphFromText(text).Block(block).Wrap(widgets.Wrap{Trim: true}),
		area)
}

// drawSecondTab is the server table beside a world map with the servers on it.
func drawSecondTab(f *catatui.Frame, a *app, area catatui.Rect) {
	columns := catatui.HorizontalLayout(catatui.Percentage(30), catatui.Percentage(70)).
		Split(area)

	up := catatui.NewStyle().Fg(catatui.ColorGreen)
	failure := catatui.NewStyle().Fg(catatui.ColorRed).
		AddModifier(catatui.ModifierRapidBlink | catatui.ModifierCrossedOut)

	rows := make([]widgets.Row, len(a.servers))
	for i, s := range a.servers {
		style := failure
		if s.status == "Up" {
			style = up
		}
		rows[i] = widgets.NewRowFromStrings(s.name, s.location, s.status).Style(style)
	}

	table := widgets.NewTable(rows,
		catatui.Length(15), catatui.Length(15), catatui.Length(10)).
		Header(widgets.NewRowFromStrings("Server", "Location", "Status").
			Style(catatui.NewStyle().Fg(catatui.ColorYellow)).
			BottomMargin(1)).
		Block(widgets.Bordered().Title("Servers"))
	f.RenderWidget(table, columns[0])

	drawMap(f, a, columns[1])
}

// drawMap draws the world, the links between the servers, and a marker on each
// of them coloured by status.
func drawMap(f *catatui.Frame, a *app, area catatui.Rect) {
	canvas := widgets.NewCanvas().
		Block(widgets.Bordered().Title("World")).
		Marker(a.marker()).
		XBounds([2]float64{-180, 180}).
		YBounds([2]float64{-90, 90}).
		Paint(func(ctx *widgets.Context) {
			ctx.Draw(widgets.Map{
				Resolution: widgets.MapResolutionHigh,
				Color:      catatui.ColorWhite,
			})
			ctx.Layer()
			ctx.Draw(widgets.Rectangle{
				X: 0, Y: 30, Width: 10, Height: 10,
				Color: catatui.ColorYellow,
			})
			ctx.Draw(widgets.Circle{
				X: a.servers[2].lon, Y: a.servers[2].lat, Radius: 10,
				Color: catatui.ColorGreen,
			})

			// A line from every server to every other one.
			for i, s1 := range a.servers {
				for _, s2 := range a.servers[i+1:] {
					ctx.Draw(widgets.NewCanvasLine(
						s1.lon, s1.lat, s2.lon, s2.lat, catatui.ColorYellow))
				}
			}

			for _, s := range a.servers {
				color := catatui.ColorRed
				if s.status == "Up" {
					color = catatui.ColorGreen
				}
				ctx.Print(s.lon, s.lat,
					catatui.LineFromStyledString("X", catatui.NewStyle().Fg(color)))
			}
		})
	f.RenderWidget(canvas, area)
}

// drawThirdTab shows every named colour as a foreground and as a background.
func drawThirdTab(f *catatui.Frame, area catatui.Rect) {
	columns := catatui.HorizontalLayout(catatui.Ratio(1, 2), catatui.Ratio(1, 2)).
		Split(area)

	colors := []catatui.Color{
		catatui.ColorReset,
		catatui.ColorBlack, catatui.ColorRed, catatui.ColorGreen, catatui.ColorYellow,
		catatui.ColorBlue, catatui.ColorMagenta, catatui.ColorCyan, catatui.ColorGray,
		catatui.ColorDarkGray, catatui.ColorLightRed, catatui.ColorLightGreen,
		catatui.ColorLightYellow, catatui.ColorLightBlue, catatui.ColorLightMagenta,
		catatui.ColorLightCyan, catatui.ColorWhite,
	}

	rows := make([]widgets.Row, len(colors))
	for i, c := range colors {
		rows[i] = widgets.NewRow(
			widgets.NewCell(c.String()+": "),
			widgets.NewCellFromLine(catatui.LineFromStyledString("Foreground",
				catatui.NewStyle().Fg(c))),
			widgets.NewCellFromLine(catatui.LineFromStyledString("Background",
				catatui.NewStyle().Bg(c))),
		)
	}

	table := widgets.NewTable(rows,
		catatui.Ratio(1, 3), catatui.Ratio(1, 3), catatui.Ratio(1, 3)).
		Block(widgets.Bordered().Title("Colors"))
	f.RenderWidget(table, columns[0])
}
