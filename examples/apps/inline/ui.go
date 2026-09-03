// Port of the rendering half of examples/apps/inline @ ratatui-v0.30.2

package main

import (
	"fmt"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// render draws the eight rows of the viewport: the overall progress, the
// downloads in flight, and a gauge for each of them.
func render(f *catatui.Frame, d *downloads) {
	area := f.Area()

	f.RenderWidget(
		widgets.NewBlock().TitleLine(catatui.LineFromString("Progress").Centered()),
		area)

	rows := catatui.VerticalLayout(catatui.Length(2), catatui.Length(4)).Margin(1).Split(area)
	progressArea, main := rows[0], rows[1]

	columns := catatui.HorizontalLayout(catatui.Percentage(20), catatui.Percentage(80)).Split(main)
	listArea, gaugeArea := columns[0], columns[1]

	renderTotalProgress(f, progressArea, d)
	renderInProgressList(f, listArea, d)
	renderGauges(f, gaugeArea, d)
}

// renderTotalProgress is the "3/10" bar across the top.
func renderTotalProgress(f *catatui.Frame, area catatui.Rect, d *downloads) {
	done := numDownloads - len(d.pending) - d.running()
	f.RenderWidget(
		widgets.NewLineGauge().
			FilledStyle(catatui.NewStyle().Fg(catatui.ColorBlue)).
			Label(fmt.Sprintf("%d/%d", done, numDownloads)).
			Ratio(float64(done)/float64(numDownloads)),
		area)
}

// renderInProgressList names each download in flight and how long it has been
// running.
func renderInProgressList(f *catatui.Frame, area catatui.Rect, d *downloads) {
	var items []widgets.ListItem
	for _, p := range d.inProgress {
		if p == nil {
			continue
		}
		items = append(items, widgets.NewListItemFromLine(catatui.NewLine(
			catatui.NewSpan(symbols.DotFull),
			catatui.NewStyledSpan(fmt.Sprintf(" download %2d", p.id),
				catatui.NewStyle().
					Fg(catatui.ColorLightGreen).
					AddModifier(catatui.ModifierBold)),
			catatui.NewSpan(fmt.Sprintf(" (%dms)", time.Since(p.startedAt).Milliseconds())),
		)))
	}
	f.RenderWidget(widgets.NewList(items...), area)
}

// renderGauges draws one row per download in flight. The rows are taken one at
// a time out of the area given, and drawing stops when they run out: a widget
// that draws outside the area it was handed panics in catatui.
func renderGauges(f *catatui.Frame, area catatui.Rect, d *downloads) {
	var row uint16
	for _, p := range d.inProgress {
		if p == nil {
			continue
		}
		y := catatui.SatAdd(area.Top(), row)
		if y >= area.Bottom() {
			return
		}
		f.RenderWidget(
			widgets.NewGauge().
				GaugeStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
				Ratio(p.progress),
			catatui.NewRect(area.Left(), y, area.Width, 1))
		row++
	}
}
