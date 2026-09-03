// Command chart draws four charts at once: a scrolling pair of sine waves, a
// bell curve as bars, a line through two points, and a scatter plot.
//
//	go run ./examples/apps/chart
//
// Press q to quit. Only the first chart moves; the rest are there to show what
// the other graph types and legend positions look like beside it.
//
// Port of examples/apps/chart @ ratatui-v0.30.2
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run redraws four times a second, which is what scrolls the first chart.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	a := newApp()

	for {
		if err := terminal.Draw(a.render); err != nil {
			return err
		}
		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			if ev.Kind == term.EventKey && (ev.IsRune('q') || ev.IsCtrl('c')) {
				return nil
			}
		case <-ticker.C:
			a.onTick()
		}
	}
}

// sinSignal is a sine wave sampled a point at a time, keeping its place
// between calls so the wave carries on where it left off.
type sinSignal struct {
	x        float64
	interval float64
	period   float64
	scale    float64
}

// next is the next point on the wave.
func (s *sinSignal) next() [2]float64 {
	point := [2]float64{s.x, math.Sin(s.x/s.period) * s.scale}
	s.x += s.interval
	return point
}

// take is the next n points.
func (s *sinSignal) take(n int) [][2]float64 {
	points := make([][2]float64, n)
	for i := range points {
		points[i] = s.next()
	}
	return points
}

// app is the two waves, the points of each one currently on screen, and the
// stretch of the x axis they are shown over.
type app struct {
	signal1 sinSignal
	data1   [][2]float64
	signal2 sinSignal
	data2   [][2]float64
	window  [2]float64
}

func newApp() *app {
	a := &app{
		signal1: sinSignal{interval: 0.2, period: 3.0, scale: 18.0},
		signal2: sinSignal{interval: 0.1, period: 2.0, scale: 10.0},
		window:  [2]float64{0.0, 20.0},
	}
	a.data1 = a.signal1.take(200)
	a.data2 = a.signal2.take(200)
	return a
}

// onTick drops the oldest points off each wave, adds as many new ones, and
// slides the window along to follow them.
func (a *app) onTick() {
	a.data1 = append(a.data1[5:], a.signal1.take(5)...)
	a.data2 = append(a.data2[10:], a.signal2.take(10)...)
	a.window[0]++
	a.window[1]++
}

// render lays the four charts out two by two.
func (a *app) render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Fill(1), catatui.Fill(1)).Split(f.Area())

	top := catatui.HorizontalLayout(catatui.Fill(1), catatui.Length(29)).Split(rows[0])
	bottom := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).Split(rows[1])

	f.RenderWidget(a.animatedChart(), top[0])
	f.RenderWidget(barChart(), top[1])
	f.RenderWidget(lineChart(), bottom[0])
	f.RenderWidget(scatterChart(), bottom[1])
}

// animatedChart is the two waves scrolling past, with the ends of the window
// labelled on the x axis as it moves.
func (a *app) animatedChart() widgets.Chart {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	gray := catatui.NewStyle().Fg(catatui.ColorGray)

	xLabels := []catatui.Line{
		catatui.LineFromStyledString(number(a.window[0]), bold),
		catatui.LineFromString(number((a.window[0] + a.window[1]) / 2)),
		catatui.LineFromStyledString(number(a.window[1]), bold),
	}

	return widgets.NewChart(
		widgets.NewDataset().
			Name("data2").
			Marker(symbols.Dot).
			Style(catatui.NewStyle().Fg(catatui.ColorCyan)).
			Data(a.data1),
		widgets.NewDataset().
			Name("data3").
			Marker(symbols.Braille).
			Style(catatui.NewStyle().Fg(catatui.ColorYellow)).
			Data(a.data2),
	).
		Block(widgets.Bordered()).
		XAxis(widgets.NewAxis().
			Title("X Axis").
			Style(gray).
			LabelLines(xLabels...).
			Bounds(a.window)).
		YAxis(widgets.NewAxis().
			Title("Y Axis").
			Style(gray).
			LabelLines(
				catatui.LineFromStyledString("-20", bold),
				catatui.LineFromString("0"),
				catatui.LineFromStyledString("20", bold),
			).
			Bounds([2]float64{-20.0, 20.0}))
}

// barChart is a bell curve drawn as bars up from the axis, in half blocks so
// the tops of the bars land between rows.
func barChart() widgets.Chart {
	dataset := widgets.NewDataset().
		Marker(symbols.HalfBlock).
		Style(catatui.NewStyle().Fg(catatui.ColorBlue)).
		GraphType(widgets.GraphTypeBar).
		Data([][2]float64{
			{0, 0.4}, {10, 2.9}, {20, 13.5}, {30, 41.1}, {40, 80.1}, {50, 100.0},
			{60, 80.1}, {70, 41.1}, {80, 13.5}, {90, 2.9}, {100, 0.4},
		})

	return widgets.NewChart(dataset).
		Block(widgets.Bordered().TitleTop(title("Bar chart"))).
		XAxis(axis([2]float64{0.0, 100.0}, "0", "50", "100.0")).
		YAxis(axis([2]float64{0.0, 100.0}, "0", "50", "100.0")).
		HiddenLegendConstraints(catatui.Ratio(1, 2), catatui.Ratio(1, 2))
}

// lineChart is a line through two points, with its legend moved out of the way
// into the top-left corner.
func lineChart() widgets.Chart {
	dataset := widgets.NewDataset().
		NameLine(catatui.LineFromStyledString("Line from only 2 points",
			catatui.NewStyle().AddModifier(catatui.ModifierItalic))).
		Marker(symbols.Braille).
		Style(catatui.NewStyle().Fg(catatui.ColorYellow)).
		GraphType(widgets.GraphTypeLine).
		Data([][2]float64{{1, 1}, {4, 4}})

	return widgets.NewChart(dataset).
		Block(widgets.Bordered().TitleLine(title("Line chart"))).
		XAxis(axis([2]float64{0.0, 5.0}, "0", "2.5", "5.0").Title("X Axis")).
		YAxis(axis([2]float64{0.0, 5.0}, "0", "2.5", "5.0").Title("Y Axis")).
		LegendPosition(widgets.LegendPositionTopLeft).
		HiddenLegendConstraints(catatui.Ratio(1, 2), catatui.Ratio(1, 2))
}

// scatterChart is three sets of points, each with its own marker, which is how
// they are told apart where they overlap.
func scatterChart() widgets.Chart {
	gray := catatui.NewStyle().Fg(catatui.ColorGray)

	return widgets.NewChart(
		widgets.NewDataset().
			Name("Heavy").
			Marker(symbols.Dot).
			GraphType(widgets.GraphTypeScatter).
			Style(catatui.NewStyle().Fg(catatui.ColorYellow)).
			Data(heavyPayloadData),
		widgets.NewDataset().
			NameLine(catatui.LineFromStyledString("Medium",
				catatui.NewStyle().AddModifier(catatui.ModifierUnderlined))).
			Marker(symbols.Custom('×')).
			GraphType(widgets.GraphTypeScatter).
			Style(catatui.NewStyle().Fg(catatui.ColorMagenta)).
			Data(mediumPayloadData),
		widgets.NewDataset().
			Name("Small").
			Marker(symbols.Custom('+')).
			GraphType(widgets.GraphTypeScatter).
			Style(catatui.NewStyle().Fg(catatui.ColorCyan)).
			Data(smallPayloadData),
	).
		Block(widgets.Bordered().TitleLine(title("Scatter chart"))).
		XAxis(widgets.NewAxis().
			Title("Year").
			Bounds([2]float64{1960, 2020}).
			Style(gray).
			Labels("1960", "1990", "2020")).
		YAxis(widgets.NewAxis().
			Title("Cost").
			Bounds([2]float64{0, 75000}).
			Style(gray).
			Labels("0", "37 500", "75 000")).
		HiddenLegendConstraints(catatui.Ratio(1, 2), catatui.Ratio(1, 2))
}

// title is the centred cyan heading each of the three static charts carries.
func title(text string) catatui.Line {
	return catatui.LineFromStyledString(text, catatui.NewStyle().
		Fg(catatui.ColorCyan).
		AddModifier(catatui.ModifierBold)).Centered()
}

// axis is a grey axis over the given bounds, with the ends of it in bold and
// the middle label plain.
func axis(bounds [2]float64, start, middle, end string) widgets.Axis {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	return widgets.NewAxis().
		Style(catatui.NewStyle().Fg(catatui.ColorGray)).
		Bounds(bounds).
		LabelLines(
			catatui.LineFromStyledString(start, bold),
			catatui.LineFromString(middle),
			catatui.LineFromStyledString(end, bold),
		)
}

// number formats an axis label the way Rust prints an f64: no trailing zeros,
// and no decimal point when there is nothing after it.
func number(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%g", v)
}

// The payload data, from https://ourworldindata.org/space-exploration-satellites
var (
	heavyPayloadData = [][2]float64{
		{1965, 8200}, {1967, 5400}, {1981, 65400}, {1989, 30800}, {1997, 10200},
		{2004, 11600}, {2014, 4500}, {2016, 7900}, {2018, 1500},
	}
	mediumPayloadData = [][2]float64{
		{1963, 29500}, {1964, 30600}, {1965, 177900}, {1965, 21000}, {1966, 17900},
		{1966, 8400}, {1975, 17500}, {1982, 8300}, {1985, 5100}, {1988, 18300},
		{1990, 38800}, {1990, 9900}, {1991, 18700}, {1992, 9100}, {1994, 10500},
		{1994, 8500}, {1994, 8700}, {1997, 6200}, {1999, 18000}, {1999, 7600},
		{1999, 8900}, {1999, 9600}, {2000, 16000}, {2001, 10000}, {2002, 10400},
		{2002, 8100}, {2010, 2600}, {2013, 13600}, {2017, 8000},
	}
	smallPayloadData = [][2]float64{
		{1961, 118500}, {1962, 14900}, {1975, 21400}, {1980, 32800}, {1988, 31100},
		{1990, 41100}, {1993, 23600}, {1994, 20600}, {1994, 34600}, {1996, 50600},
		{1997, 19200}, {1997, 45800}, {1998, 19100}, {2000, 73100}, {2003, 11200},
		{2008, 12600}, {2010, 30500}, {2012, 20000}, {2013, 10600}, {2013, 34500},
		{2015, 10600}, {2018, 23100}, {2019, 17300},
	}
)
