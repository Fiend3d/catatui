// Port of examples/apps/demo/src/app.rs @ ratatui-v0.30.2

package main

import (
	"math"
	"math/rand/v2"

	"github.com/Fiend3d/catatui/symbols"
	"github.com/Fiend3d/catatui/widgets"
)

// tasks fills the list on the first tab.
var tasks = []string{
	"Item1", "Item2", "Item3", "Item4", "Item5", "Item6", "Item7", "Item8",
	"Item9", "Item10", "Item11", "Item12", "Item13", "Item14", "Item15",
	"Item16", "Item17", "Item18", "Item19", "Item20", "Item21", "Item22",
	"Item23", "Item24",
}

// logEntry is one line of the log list: what happened, and how bad it was.
type logEntry struct {
	event string
	level string
}

var logs = []logEntry{
	{"Event1", "INFO"}, {"Event2", "INFO"}, {"Event3", "CRITICAL"},
	{"Event4", "ERROR"}, {"Event5", "INFO"}, {"Event6", "INFO"},
	{"Event7", "WARNING"}, {"Event8", "INFO"}, {"Event9", "INFO"},
	{"Event10", "INFO"}, {"Event11", "CRITICAL"}, {"Event12", "INFO"},
	{"Event13", "INFO"}, {"Event14", "INFO"}, {"Event15", "INFO"},
	{"Event16", "INFO"}, {"Event17", "ERROR"}, {"Event18", "ERROR"},
	{"Event19", "INFO"}, {"Event20", "INFO"}, {"Event21", "WARNING"},
	{"Event22", "INFO"}, {"Event23", "INFO"}, {"Event24", "WARNING"},
	{"Event25", "INFO"}, {"Event26", "INFO"},
}

// events fills the bar chart.
var events = []widgets.BarPair{
	{Label: "B1", Value: 9}, {Label: "B2", Value: 12}, {Label: "B3", Value: 5},
	{Label: "B4", Value: 8}, {Label: "B5", Value: 2}, {Label: "B6", Value: 4},
	{Label: "B7", Value: 5}, {Label: "B8", Value: 9}, {Label: "B9", Value: 14},
	{Label: "B10", Value: 15}, {Label: "B11", Value: 1}, {Label: "B12", Value: 0},
	{Label: "B13", Value: 4}, {Label: "B14", Value: 6}, {Label: "B15", Value: 4},
	{Label: "B16", Value: 6}, {Label: "B17", Value: 4}, {Label: "B18", Value: 7},
	{Label: "B19", Value: 13}, {Label: "B20", Value: 8}, {Label: "B21", Value: 11},
	{Label: "B22", Value: 9}, {Label: "B23", Value: 3}, {Label: "B24", Value: 5},
}

// randomSignal returns values spread evenly over [lower, upper).
func randomSignal(lower, upper uint64) func() uint64 {
	return func() uint64 { return lower + rand.Uint64N(upper-lower) }
}

// sinSignal returns points along a sine wave, one interval further along the x
// axis each time it is called.
func sinSignal(interval, period, scale float64) func() [2]float64 {
	x := 0.0
	return func() [2]float64 {
		point := [2]float64{x, math.Sin(x*1.0/period) * scale}
		x += interval
		return point
	}
}

// signal is a window onto an endless source: it holds the points currently on
// screen, and each tick drops the oldest few and pulls in as many new ones.
type signal[T any] struct {
	next     func() T
	points   []T
	tickRate int
}

// newSignal fills a window of count points from source.
func newSignal[T any](source func() T, count, tickRate int) signal[T] {
	s := signal[T]{next: source, tickRate: tickRate, points: make([]T, 0, count)}
	for range count {
		s.points = append(s.points, source())
	}
	return s
}

// onTick slides the window along by tickRate points.
func (s *signal[T]) onTick() {
	if s.tickRate > len(s.points) {
		return
	}
	s.points = append(s.points[:0], s.points[s.tickRate:]...)
	for range s.tickRate {
		s.points = append(s.points, s.next())
	}
}

// signals is the pair of sine waves the chart draws, and the window of the x
// axis they scroll through.
type signals struct {
	sin1   signal[[2]float64]
	sin2   signal[[2]float64]
	window [2]float64
}

func (s *signals) onTick() {
	s.sin1.onTick()
	s.sin2.onTick()
	s.window[0]++
	s.window[1]++
}

// tabsState is which of the tab titles is selected.
type tabsState struct {
	titles []string
	index  int
}

func (t *tabsState) next() { t.index = (t.index + 1) % len(t.titles) }

func (t *tabsState) previous() {
	if t.index > 0 {
		t.index--
	} else {
		t.index = len(t.titles) - 1
	}
}

// statefulList pairs items with the selection the List widget renders against.
// The widget is rebuilt every frame; this is what survives.
type statefulList[T any] struct {
	state widgets.ListState
	items []T
}

func newStatefulList[T any](items []T) statefulList[T] {
	return statefulList[T]{items: items}
}

// next moves the selection down, wrapping at the end.
func (l *statefulList[T]) next() {
	i, ok := l.state.Selected()
	switch {
	case !ok, i >= len(l.items)-1:
		i = 0
	default:
		i++
	}
	l.state.Select(i)
}

// previous moves the selection up, wrapping at the start.
func (l *statefulList[T]) previous() {
	i, ok := l.state.Selected()
	switch {
	case !ok:
		i = 0
	case i == 0:
		i = len(l.items) - 1
	default:
		i--
	}
	l.state.Select(i)
}

// server is one row of the table on the second tab, and one point on the map
// beside it.
type server struct {
	name     string
	location string
	// lat and lon are degrees, which is what the map's bounds are in.
	lat, lon float64
	status   string
}

// app is everything the demo draws, and the only thing that lives between
// frames.
type app struct {
	title            string
	shouldQuit       bool
	tabs             tabsState
	showChart        bool
	progress         float64
	sparkline        signal[uint64]
	tasks            statefulList[string]
	logs             statefulList[logEntry]
	signals          signals
	barchart         []widgets.BarPair
	servers          []server
	enhancedGraphics bool
}

// newApp returns the demo with its signals already filled, so the charts have
// something to show on the first frame.
func newApp(title string, enhancedGraphics bool) *app {
	return &app{
		title:     title,
		tabs:      tabsState{titles: []string{"Tab0", "Tab1", "Tab2"}},
		showChart: true,
		sparkline: newSignal(randomSignal(0, 100), 300, 1),
		tasks:     newStatefulList(tasks),
		logs:      newStatefulList(logs),
		signals: signals{
			sin1:   newSignal(sinSignal(0.2, 3.0, 18.0), 100, 5),
			sin2:   newSignal(sinSignal(0.1, 2.0, 10.0), 200, 10),
			window: [2]float64{0, 20},
		},
		barchart: append([]widgets.BarPair(nil), events...),
		servers: []server{
			{"NorthAmerica-1", "New York City", 40.71, -74.00, "Up"},
			{"Europe-1", "Paris", 48.85, 2.35, "Failure"},
			{"SouthAmerica-1", "São Paulo", -23.54, -46.62, "Up"},
			{"Asia-1", "Singapore", 1.35, 103.86, "Up"},
		},
		enhancedGraphics: enhancedGraphics,
	}
}

// barSet is the set of bar characters to draw with: the finer one when the
// terminal can be trusted with it, three levels otherwise.
func (a *app) barSet() symbols.LevelSet {
	if a.enhancedGraphics {
		return symbols.BarNineLevels
	}
	return symbols.BarThreeLevels
}

// marker is what the chart and the map plot points with.
func (a *app) marker() symbols.Marker {
	if a.enhancedGraphics {
		return symbols.Braille
	}
	return symbols.Dot
}

func (a *app) onUp()    { a.tasks.previous() }
func (a *app) onDown()  { a.tasks.next() }
func (a *app) onRight() { a.tabs.next() }
func (a *app) onLeft()  { a.tabs.previous() }

func (a *app) onKey(c rune) {
	switch c {
	case 'q':
		a.shouldQuit = true
	case 't':
		a.showChart = !a.showChart
	}
}

// onTick advances everything that moves on its own: the gauges creep forward,
// the signals scroll, and the two rotating lists rotate.
func (a *app) onTick() {
	a.progress += 0.001
	if a.progress > 1.0 {
		a.progress = 0
	}

	a.sparkline.onTick()
	a.signals.onTick()

	a.logs.items = rotate(a.logs.items)
	a.barchart = rotate(a.barchart)
}

// rotate moves the last item to the front, which is how the demo makes its
// lists look live.
func rotate[T any](items []T) []T {
	if len(items) == 0 {
		return items
	}
	last := items[len(items)-1]
	return append([]T{last}, items[:len(items)-1]...)
}
