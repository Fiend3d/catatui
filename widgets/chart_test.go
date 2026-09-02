// Tests ported from ratatui-widgets/src/chart.rs and
// ratatui/tests/widgets_chart.rs @ ratatui-v0.30.2

package widgets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// renderChart draws a chart into a buffer of the given size.
func renderChart(c Chart, width, height uint16) *catatui.Buffer {
	area := catatui.NewRect(0, 0, width, height)
	buf := catatui.NewBuffer(area)
	c.Render(area, buf)
	return buf
}

func TestChartItShouldHideTheLegend(t *testing.T) {
	data := [][2]float64{{0, 5}, {1, 6}, {3, 7}}
	cases := []struct {
		chartArea  catatui.Rect
		hiddenW    catatui.Constraint
		hiddenH    catatui.Constraint
		legendArea catatui.Rect
		hasLegend  bool
	}{
		{
			chartArea:  catatui.NewRect(0, 0, 100, 100),
			hiddenW:    catatui.Ratio(1, 4),
			hiddenH:    catatui.Ratio(1, 4),
			legendArea: catatui.NewRect(88, 0, 12, 12),
			hasLegend:  true,
		},
		{
			chartArea: catatui.NewRect(0, 0, 100, 100),
			hiddenW:   catatui.Ratio(1, 10),
			hiddenH:   catatui.Ratio(1, 4),
		},
	}

	for i, tc := range cases {
		datasets := make([]Dataset, 10)
		for j := range datasets {
			datasets[j] = NewDataset().Name(fmt.Sprintf("Dataset #%d", j)).Data(data)
		}
		chart := NewChart(datasets...).
			XAxis(NewAxis().Title("X axis")).
			YAxis(NewAxis().Title("Y axis")).
			HiddenLegendConstraints(tc.hiddenW, tc.hiddenH)

		layout, ok := chart.layout(tc.chartArea)
		if !ok {
			t.Fatalf("case %d: layout failed", i)
		}
		if layout.hasLegend != tc.hasLegend || layout.legendArea != tc.legendArea {
			t.Errorf("case %d: got legend %v %v, want %v %v",
				i, layout.legendArea, layout.hasLegend, tc.legendArea, tc.hasLegend)
		}
	}
}

func TestChartAxisCanBeStylized(t *testing.T) {
	want := catatui.NewStyle().
		Fg(catatui.ColorBlack).
		Bg(catatui.ColorWhite).
		AddModifier(catatui.ModifierBold).
		RemoveModifier(catatui.ModifierDim)
	if got := NewAxis().Style(want).GetStyle(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestChartDatasetCanBeStylized(t *testing.T) {
	want := catatui.NewStyle().
		Fg(catatui.ColorBlack).
		Bg(catatui.ColorWhite).
		AddModifier(catatui.ModifierBold).
		RemoveModifier(catatui.ModifierDim)
	if got := NewDataset().Style(want).GetStyle(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestChartCanBeStylized(t *testing.T) {
	want := catatui.NewStyle().
		Fg(catatui.ColorBlack).
		Bg(catatui.ColorWhite).
		AddModifier(catatui.ModifierBold).
		RemoveModifier(catatui.ModifierDim)
	if got := NewChart().Style(want).GetStyle(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestChartGraphTypeToString(t *testing.T) {
	cases := []struct {
		graphType GraphType
		want      string
	}{
		{GraphTypeScatter, "Scatter"},
		{GraphTypeLine, "Line"},
		{GraphTypeBar, "Bar"},
		{GraphTypeArea, "Area"},
	}
	for _, tc := range cases {
		if got := tc.graphType.String(); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestChartGraphTypeFromString(t *testing.T) {
	cases := []struct {
		s    string
		want GraphType
	}{
		{"Scatter", GraphTypeScatter},
		{"Line", GraphTypeLine},
		{"Bar", GraphTypeBar},
		{"Area", GraphTypeArea},
	}
	for _, tc := range cases {
		got, err := ParseGraphType(tc.s)
		if err != nil || got != tc.want {
			t.Errorf("ParseGraphType(%q) = %v, %v; want %v, nil", tc.s, got, err, tc.want)
		}
	}
	if _, err := ParseGraphType(""); err == nil {
		t.Error("ParseGraphType(\"\") = nil error, want an error")
	}
}

func TestChartLegendPositionToString(t *testing.T) {
	cases := []struct {
		position LegendPosition
		want     string
	}{
		{LegendPositionTop, "Top"},
		{LegendPositionTopRight, "TopRight"},
		{LegendPositionTopLeft, "TopLeft"},
		{LegendPositionLeft, "Left"},
		{LegendPositionRight, "Right"},
		{LegendPositionBottom, "Bottom"},
		{LegendPositionBottomRight, "BottomRight"},
		{LegendPositionBottomLeft, "BottomLeft"},
	}
	for _, tc := range cases {
		if got := tc.position.String(); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestChartItDoesNotPanicIfTitleIsWiderThanBuffer(t *testing.T) {
	chart := Chart{}.
		YAxis(NewAxis().Title("xxxxxxxxxxxxxxxx")).
		XAxis(NewAxis().Title("xxxxxxxxxxxxxxxx"))
	buf := renderChart(chart, 8, 4)
	blank := strings.Repeat(" ", 8)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(blank, blank, blank, blank))
}

func TestChartItDoesNotPanicIfYAxisHasOneLabel(t *testing.T) {
	chart := NewChart().YAxis(NewAxis().Bounds([2]float64{0, 1}).Labels("only"))
	renderChart(chart, 20, 5)
}

func TestChartDatasetsWithoutNameDoNotContributeToLegendHeight(t *testing.T) {
	chart := NewChart(
		NewDataset().Name("data1"), // occupies a row in the legend
		NewDataset(),               // must not occupy a row
		NewDataset().Name(""),      // occupies a row, even with an empty name
	)
	layout, ok := chart.layout(catatui.NewRect(0, 0, 50, 25))
	if !ok {
		t.Fatal("layout failed")
	}
	if !layout.hasLegend {
		t.Fatal("got no legend, want one")
	}
	if got := layout.legendArea.Height; got != 4 { // 2 for the borders, 2 for the rows
		t.Errorf("got legend height %d, want 4", got)
	}
}

func TestChartNoLegendIfNoNamedDatasets(t *testing.T) {
	chart := NewChart(NewDataset(), NewDataset(), NewDataset())
	layout, ok := chart.layout(catatui.NewRect(0, 0, 50, 25))
	if !ok {
		t.Fatal("layout failed")
	}
	if layout.hasLegend {
		t.Errorf("got legend %v, want none", layout.legendArea)
	}
}

func TestChartDatasetLegendStyleIsPatched(t *testing.T) {
	longName := NewDataset().Name("Very long name")
	shortName := NewDataset().NameLine(catatui.LineFromString("Short name").Right())
	chart := NewChart(longName, shortName).
		HiddenLegendConstraints(catatui.Length(100), catatui.Length(100))

	buf := renderChart(chart, 20, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    ┌──────────────┐",
		"    │Very long name│",
		"    │    Short name│",
		"    └──────────────┘",
		"                    ",
	))
}

func TestChartHaveATopLeftLegend(t *testing.T) {
	chart := NewChart(NewDataset().Name("Ds1")).LegendPosition(LegendPositionTopLeft)
	buf := renderChart(chart, 30, 20)

	want := make([]string, 20)
	for i := range want {
		want[i] = strings.Repeat(" ", 30)
	}
	want[0] = "┌───┐                         "
	want[1] = "│Ds1│                         "
	want[2] = "└───┘                         "
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(want...))
}

func TestChartHaveALongYAxisTitleOverlappingLegend(t *testing.T) {
	chart := NewChart(NewDataset().Name("Ds1")).
		YAxis(NewAxis().Title("The title overlap a legend."))
	buf := renderChart(chart, 30, 20)

	want := make([]string, 20)
	for i := range want {
		want[i] = strings.Repeat(" ", 30)
	}
	want[0] = "The title overlap a legend.   "
	want[1] = "                         ┌───┐"
	want[2] = "                         │Ds1│"
	want[3] = "                         └───┘"
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(want...))
}

func TestChartHaveOverflowedYAxis(t *testing.T) {
	chart := NewChart(NewDataset().Name("Ds1")).
		YAxis(NewAxis().Title("The title overlap a legend."))
	buf := renderChart(chart, 10, 10)

	want := make([]string, 10)
	for i := range want {
		want[i] = strings.Repeat(" ", 10)
	}
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(want...))
}

func TestChartLegendAreaCanFitSameChartArea(t *testing.T) {
	positions := []LegendPosition{
		LegendPositionTopLeft,
		LegendPositionTop,
		LegendPositionTopRight,
		LegendPositionLeft,
		LegendPositionRight,
		LegendPositionBottom,
		LegendPositionBottomLeft,
		LegendPositionBottomRight,
	}
	for _, position := range positions {
		chart := NewChart(NewDataset().Name("Data")).
			HiddenLegendConstraints(catatui.Percentage(100), catatui.Percentage(100)).
			LegendPosition(position)
		buf := renderChart(chart, 6, 3)
		want := catatui.NewBufferWithStrings(
			"┌────┐",
			"│Data│",
			"└────┘",
		)
		t.Run(position.String(), func(t *testing.T) {
			catatui.AssertBuffer(t, buf, want)
		})
	}
}

func TestChartLegendOfChartHaveOddMarginSize(t *testing.T) {
	cases := []struct {
		name     string
		position LegendPosition
		hidden   bool
		want     []string
	}{
		{name: "TopLeft", position: LegendPositionTopLeft, want: []string{
			"┌────┐   ",
			"│Data│   ",
			"└────┘   ",
			"         ",
			"         ",
			"         ",
		}},
		{name: "Top", position: LegendPositionTop, want: []string{
			" ┌────┐  ",
			" │Data│  ",
			" └────┘  ",
			"         ",
			"         ",
			"         ",
		}},
		{name: "TopRight", position: LegendPositionTopRight, want: []string{
			"   ┌────┐",
			"   │Data│",
			"   └────┘",
			"         ",
			"         ",
			"         ",
		}},
		{name: "Left", position: LegendPositionLeft, want: []string{
			"         ",
			"┌────┐   ",
			"│Data│   ",
			"└────┘   ",
			"         ",
			"         ",
		}},
		{name: "Right", position: LegendPositionRight, want: []string{
			"         ",
			"   ┌────┐",
			"   │Data│",
			"   └────┘",
			"         ",
			"         ",
		}},
		{name: "BottomLeft", position: LegendPositionBottomLeft, want: []string{
			"         ",
			"         ",
			"         ",
			"┌────┐   ",
			"│Data│   ",
			"└────┘   ",
		}},
		{name: "Bottom", position: LegendPositionBottom, want: []string{
			"         ",
			"         ",
			"         ",
			" ┌────┐  ",
			" │Data│  ",
			" └────┘  ",
		}},
		{name: "BottomRight", position: LegendPositionBottomRight, want: []string{
			"         ",
			"         ",
			"         ",
			"   ┌────┐",
			"   │Data│",
			"   └────┘",
		}},
		{name: "None", hidden: true, want: []string{
			"         ",
			"         ",
			"         ",
			"         ",
			"         ",
			"         ",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart := NewChart(NewDataset().Name("Data")).
				HiddenLegendConstraints(catatui.Percentage(100), catatui.Percentage(100))
			if tc.hidden {
				chart = chart.LegendPositionNone()
			} else {
				chart = chart.LegendPosition(tc.position)
			}
			buf := renderChart(chart, 9, 6)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(tc.want...))
		})
	}
}

func TestChartBarChart(t *testing.T) {
	data := [][2]float64{{0, 0}, {2, 1}, {4, 4}, {6, 8}, {8, 9}, {10, 10}}
	chart := NewChart(
		NewDataset().Data(data).Marker(symbols.Dot).GraphType(GraphTypeBar),
	).
		XAxis(NewAxis().Bounds([2]float64{0, 10})).
		YAxis(NewAxis().Bounds([2]float64{0, 10}))

	buf := renderChart(chart, 11, 11)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"          •",
		"        • •",
		"      • • •",
		"      • • •",
		"      • • •",
		"      • • •",
		"    • • • •",
		"    • • • •",
		"    • • • •",
		"  • • • • •",
		"• • • • • •",
	))
}

func TestChartOverlappingLines(t *testing.T) {
	cases := []struct {
		name   string
		marker symbols.Marker
		symbol string
	}{
		{name: "Dot", marker: symbols.Dot, symbol: "•"},
		{name: "Braille", marker: symbols.Braille, symbol: "⢣"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diagonalUp := [][2]float64{{0, 0}, {5, 5}}
			diagonalDown := [][2]float64{{0, 5}, {5, 0}}
			chart := NewChart(
				NewDataset().
					Data(diagonalUp).
					Marker(symbols.Block).
					GraphType(GraphTypeLine).
					Style(catatui.NewStyle().Fg(catatui.ColorBlue)),
				NewDataset().
					Data(diagonalDown).
					Marker(tc.marker).
					GraphType(GraphTypeLine).
					Style(catatui.NewStyle().Fg(catatui.ColorRed)),
			).
				XAxis(NewAxis().Bounds([2]float64{0, 5})).
				YAxis(NewAxis().Bounds([2]float64{0, 5}))

			buf := renderChart(chart, 5, 5)

			want := catatui.NewBufferWithStrings(
				tc.symbol+"   █",
				" "+tc.symbol+" █ ",
				"  "+tc.symbol+"  ",
				" █ "+tc.symbol+" ",
				"█   "+tc.symbol,
			)
			for i := uint16(0); i < 5; i++ {
				// The dot and braille markers only set the foreground.
				want.SetStyle(catatui.NewRect(i, i, 1, 1),
					catatui.NewStyle().Fg(catatui.ColorRed))
				// The block marker sets both foreground and background.
				want.SetStyle(catatui.NewRect(i, 4-i, 1, 1),
					catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorBlue))
			}
			// Where the dot or braille overlaps the block, the background
			// stays blue from the block and the foreground is red from the
			// dot, which is what lets two line plots overlap as long as one
			// of them is drawn with blocks.
			want.SetStyle(catatui.NewRect(2, 2, 1, 1),
				catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorBlue))

			catatui.AssertBuffer(t, buf, want)
		})
	}
}

func TestChartFilledLine(t *testing.T) {
	data := [][2]float64{{0, 0}, {5, 5}, {10, 5}}
	chart := NewChart(
		NewDataset().
			Data(data).
			Marker(symbols.Dot).
			FillToY(0).
			GraphType(GraphTypeArea),
	).
		XAxis(NewAxis().Bounds([2]float64{0, 10})).
		YAxis(NewAxis().Bounds([2]float64{0, 10}))

	buf := renderChart(chart, 11, 11)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"           ",
		"           ",
		"           ",
		"           ",
		"           ",
		"     ••••••",
		"    •••••••",
		"   ••••••••",
		"  •••••••••",
		" ••••••••••",
		"•••••••••••",
	))
}

func TestChartRenderInMinimalBuffer(t *testing.T) {
	chart := NewChart(NewDataset().Data([][2]float64{{0, 0}, {1, 1}})).
		XAxis(NewAxis().Bounds([2]float64{0, 1})).
		YAxis(NewAxis().Bounds([2]float64{0, 1}))
	// This should not panic, even though the buffer is too small for the chart.
	buf := renderChart(chart, 1, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("•"))
}

func TestChartRenderInZeroSizeBuffer(t *testing.T) {
	chart := NewChart(NewDataset().Data([][2]float64{{0, 0}, {1, 1}})).
		XAxis(NewAxis().Bounds([2]float64{0, 1})).
		YAxis(NewAxis().Bounds([2]float64{0, 1}))
	// This should not panic on a buffer with no size at all.
	renderChart(chart, 0, 0)
}

// --- Tests ported from ratatui/tests/widgets_chart.rs ---------------------

// renderChartAxes draws a chart with the given axes and no datasets, which is
// how ratatui's integration tests check label and axis placement.
func renderChartAxes(width, height uint16, xAxis, yAxis Axis) *catatui.Buffer {
	return renderChart(NewChart().XAxis(xAxis).YAxis(yAxis), width, height)
}

func TestChartCanRenderOnSmallAreas(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 2}}
	for _, size := range sizes {
		chart := NewChart(
			NewDataset().
				Marker(symbols.Braille).
				Style(catatui.NewStyle().Fg(catatui.ColorMagenta)).
				Data([][2]float64{{0, 0}}),
		).
			Block(Bordered().Title("Plot")).
			XAxis(NewAxis().Bounds([2]float64{0, 0}).Labels("0.0", "1.0")).
			YAxis(NewAxis().Bounds([2]float64{0, 0}).Labels("0.0", "1.0"))
		// This should not panic at any size.
		renderChart(chart, size[0], size[1])
	}
}

func TestChartHandlesLongLabels(t *testing.T) {
	cases := []struct {
		name       string
		xLabels    []string
		yLabels    []string
		xAlignment catatui.Alignment
		want       []string
	}{
		{
			name:       "long first x label",
			xLabels:    []string{"AAAA", "B"},
			xAlignment: catatui.AlignmentLeft,
			want: []string{
				"          ",
				"          ",
				"          ",
				"   ───────",
				"AAA      B",
			},
		},
		{
			name:       "long last x label",
			xLabels:    []string{"A", "BBBB"},
			xAlignment: catatui.AlignmentLeft,
			want: []string{
				"          ",
				"          ",
				"          ",
				" ─────────",
				"A     BBBB",
			},
		},
		{
			name:       "very long first x label",
			xLabels:    []string{"AAAAAAAAAAA", "B"},
			xAlignment: catatui.AlignmentLeft,
			want: []string{
				"          ",
				"          ",
				"          ",
				"   ───────",
				"AAA      B",
			},
		},
		{
			name:       "long first y label",
			xLabels:    []string{"A", "B"},
			yLabels:    []string{"CCCCCCC", "D"},
			xAlignment: catatui.AlignmentLeft,
			want: []string{
				"D  │      ",
				"   │      ",
				"CCC│      ",
				"   └──────",
				"   A     B",
			},
		},
		{
			name:       "centered x labels",
			xLabels:    []string{"AAAAAAAAAA", "B"},
			yLabels:    []string{"C", "D"},
			xAlignment: catatui.AlignmentCenter,
			want: []string{
				"D  │      ",
				"   │      ",
				"C  │      ",
				"   └──────",
				"AAAAAAA  B",
			},
		},
		{
			name:       "right aligned x labels",
			xLabels:    []string{"AAAAAAA", "B"},
			yLabels:    []string{"C", "D"},
			xAlignment: catatui.AlignmentRight,
			want: []string{
				"D│        ",
				" │        ",
				"C│        ",
				" └────────",
				" AAAAA   B",
			},
		},
		{
			name:       "right aligned long x labels",
			xLabels:    []string{"AAAAAAA", "BBBBBBB"},
			yLabels:    []string{"C", "D"},
			xAlignment: catatui.AlignmentRight,
			want: []string{
				"D│        ",
				" │        ",
				"C│        ",
				" └────────",
				" AAAAABBBB",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xAxis := NewAxis().Bounds([2]float64{0, 1})
			if len(tc.xLabels) > 0 {
				xAxis = xAxis.Labels(tc.xLabels...).LabelsAlignment(tc.xAlignment)
			}
			yAxis := NewAxis().Bounds([2]float64{0, 1})
			if len(tc.yLabels) > 0 {
				yAxis = yAxis.Labels(tc.yLabels...)
			}
			buf := renderChartAxes(10, 5, xAxis, yAxis)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(tc.want...))
		})
	}
}

func TestChartHandlesXAxisLabelsAlignments(t *testing.T) {
	cases := []struct {
		name      string
		alignment catatui.Alignment
		want      []string
	}{
		{name: "Left", alignment: catatui.AlignmentLeft, want: []string{
			"          ",
			"          ",
			"          ",
			"   ───────",
			"AAA   B  C",
		}},
		{name: "Center", alignment: catatui.AlignmentCenter, want: []string{
			"          ",
			"          ",
			"          ",
			"  ────────",
			"AAAA B   C",
		}},
		{name: "Right", alignment: catatui.AlignmentRight, want: []string{
			"          ",
			"          ",
			"          ",
			"──────────",
			"AAA B    C",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xAxis := NewAxis().Labels("AAAA", "B", "C").LabelsAlignment(tc.alignment)
			buf := renderChartAxes(10, 5, xAxis, NewAxis())
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(tc.want...))
		})
	}
}

func TestChartHandlesYAxisLabelsAlignments(t *testing.T) {
	cases := []struct {
		name      string
		alignment catatui.Alignment
		want      []string
	}{
		{name: "Left", alignment: catatui.AlignmentLeft, want: []string{
			"D   │               ",
			"    │               ",
			"C   │               ",
			"    └───────────────",
			"AAAAA              B",
		}},
		{name: "Center", alignment: catatui.AlignmentCenter, want: []string{
			" D  │               ",
			"    │               ",
			" C  │               ",
			"    └───────────────",
			"AAAAA              B",
		}},
		{name: "Right", alignment: catatui.AlignmentRight, want: []string{
			"   D│               ",
			"    │               ",
			"   C│               ",
			"    └───────────────",
			"AAAAA              B",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xAxis := NewAxis().Labels("AAAAA", "B")
			yAxis := NewAxis().Labels("C", "D").LabelsAlignment(tc.alignment)
			buf := renderChartAxes(20, 5, xAxis, yAxis)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(tc.want...))
		})
	}
}

func TestChartCanHaveAxisWithZeroLengthBounds(t *testing.T) {
	chart := NewChart(
		NewDataset().
			Marker(symbols.Braille).
			Style(catatui.NewStyle().Fg(catatui.ColorMagenta)).
			Data([][2]float64{{0, 0}}),
	).
		Block(Bordered().Title("Plot")).
		XAxis(NewAxis().Bounds([2]float64{0, 0}).Labels("0.0", "1.0")).
		YAxis(NewAxis().Bounds([2]float64{0, 0}).Labels("0.0", "1.0"))
	// Bounds of zero length must not divide by zero.
	renderChart(chart, 100, 100)
}

func TestChartHandlesOverflows(t *testing.T) {
	chart := NewChart(
		NewDataset().
			Marker(symbols.Braille).
			Style(catatui.NewStyle().Fg(catatui.ColorMagenta)).
			Data([][2]float64{
				{1_588_298_471, 1},
				{1_588_298_473, 0},
				{1_588_298_496, 1},
			}),
	).
		Block(Bordered().Title("Plot")).
		XAxis(NewAxis().
			Bounds([2]float64{1_588_298_471, 1_588_992_600}).
			Labels("1588298471.0", "1588992600.0")).
		YAxis(NewAxis().Bounds([2]float64{0, 1}).Labels("0.0", "1.0"))
	// Coordinates far outside uint16 must not wrap into the buffer.
	renderChart(chart, 80, 30)
}

func TestChartCanHaveEmptyDatasets(t *testing.T) {
	chart := NewChart(NewDataset().Data(nil).GraphType(GraphTypeLine)).
		Block(Bordered().Title("Empty Dataset With Line")).
		XAxis(NewAxis().Bounds([2]float64{0, 0}).Labels("0.0", "1.0")).
		YAxis(NewAxis().Bounds([2]float64{0, 1}).Labels("0.0", "1.0"))
	// A line plot with no points has no pairs to draw.
	renderChart(chart, 100, 100)
}

func TestChartTopLineStylingIsCorrect(t *testing.T) {
	titleStyle := catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorLightRed)
	dataStyle := catatui.NewStyle().Fg(catatui.ColorBlue)

	chart := NewChart(
		NewDataset().
			Data([][2]float64{{0, 1}, {1, 1}}).
			GraphType(GraphTypeLine).
			Style(dataStyle),
	).
		YAxis(NewAxis().
			TitleLine(catatui.NewLine(catatui.NewStyledSpan("abc", titleStyle))).
			Bounds([2]float64{0, 1}).
			Labels("a", "b")).
		XAxis(NewAxis().Bounds([2]float64{0, 1}))

	buf := renderChart(chart, 9, 5)

	want := catatui.NewBufferWithStrings(
		"b│abc••••",
		" │       ",
		" │       ",
		" │       ",
		"a│       ",
	)
	want.SetStyle(catatui.NewRect(2, 0, 3, 1), titleStyle)
	want.SetStyle(catatui.NewRect(5, 0, 4, 1), dataStyle)
	catatui.AssertBuffer(t, buf, want)
}
