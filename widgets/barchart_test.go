// Tests ported from ratatui-widgets/src/barchart.rs, barchart/bar.rs and
// barchart/bar_group.rs @ ratatui-v0.30.2

package widgets

import (
	"math"
	"reflect"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

func pairs(ps ...BarPair) BarGroup { return BarGroupFromPairs(ps...) }

func pair(label string, value uint64) BarPair { return BarPair{Label: label, Value: value} }

func line(s string) catatui.Line { return catatui.LineFromString(s) }

func TestBarChartDefault(t *testing.T) {
	buf := renderToBuffer(NewBarChart(), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"          ",
		"          ",
		"          ",
	))
}

func TestBarChartHorizontalEmptyBarchartDoesNotPanic(t *testing.T) {
	renderToBuffer(HorizontalBarChart(), 10, 3)
}

func TestBarChartConstructorsIgnoreEmptyGroups(t *testing.T) {
	if len(VerticalBarChart().data) != 0 {
		t.Errorf("VerticalBarChart() with no bars should have no groups")
	}
	if len(HorizontalBarChart().data) != 0 {
		t.Errorf("HorizontalBarChart() with no bars should have no groups")
	}
	chart := GroupedBarChart(NewBarGroup(), NewBarGroup(NewBar(1)))
	want := []BarGroup{NewBarGroup(NewBar(1))}
	if !reflect.DeepEqual(chart.data, want) {
		t.Errorf("data = %#v, want %#v", chart.data, want)
	}
}

func TestBarChartData(t *testing.T) {
	buf := renderToBuffer(NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2)), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  █       ",
		"1 2       ",
		"f b       ",
	))
}

func TestBarChartBlock(t *testing.T) {
	block := Bordered().BorderType(BorderDouble).Title("Block")
	buf := renderToBuffer(NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2)).Block(block), 10, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"╔Block═══╗",
		"║  █     ║",
		"║1 2     ║",
		"║f b     ║",
		"╚════════╝",
	))
}

func TestBarChartMax(t *testing.T) {
	withoutMax := NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2), pair("baz", 100))
	buf := renderToBuffer(withoutMax, 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    █     ",
		"    █     ",
		"f b b     ",
	))

	withMax := NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2), pair("baz", 100)).Max(2)
	withMax.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  █ █     ",
		"1 2 █     ",
		"f b b     ",
	))
}

func TestBarChartBarStyle(t *testing.T) {
	buf := renderToBuffer(NewBarChart().
		DataPairs(pair("foo", 1), pair("bar", 2)).
		BarStyle(catatui.NewStyle().Fg(catatui.ColorRed)), 10, 3)
	expected := catatui.NewBufferWithStrings(
		"  █       ",
		"1 2       ",
		"f b       ",
	)
	for _, x := range []uint16{0, 2} {
		for _, y := range []uint16{0, 1} {
			expected.Get(x, y).SetFg(catatui.ColorRed)
		}
	}
	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartBarWidth(t *testing.T) {
	buf := renderToBuffer(NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2)).BarWidth(3), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    ███   ",
		"█1█ █2█   ",
		"foo bar   ",
	))
}

func TestBarChartBarGap(t *testing.T) {
	buf := renderToBuffer(NewBarChart().DataPairs(pair("foo", 1), pair("bar", 2)).BarGap(2), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   █      ",
		"1  2      ",
		"f  b      ",
	))
}

func TestBarChartBarSet(t *testing.T) {
	buf := renderToBuffer(NewBarChart().
		DataPairs(pair("foo", 0), pair("bar", 1), pair("baz", 3)).
		BarSet(symbols.BarThreeLevels), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    █     ",
		"  ▄ 3     ",
		"f b b     ",
	))
}

func nineLevelPairs() []BarPair {
	return []BarPair{
		pair("a", 0), pair("b", 1), pair("c", 2), pair("d", 3), pair("e", 4),
		pair("f", 5), pair("g", 6), pair("h", 7), pair("i", 8),
	}
}

func TestBarChartBarSetNineLevels(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 18, 3))
	NewBarChart().DataPairs(nineLevelPairs()...).BarSet(symbols.BarNineLevels).
		Render(catatui.NewRect(0, 1, 18, 2), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                  ",
		"  ▁ ▂ ▃ ▄ ▅ ▆ ▇ 8 ",
		"a b c d e f g h i ",
	))
}

func TestBarChartValueStyle(t *testing.T) {
	buf := renderToBuffer(NewBarChart().
		DataPairs(pair("foo", 1), pair("bar", 2)).
		BarWidth(3).
		ValueStyle(catatui.NewStyle().Fg(catatui.ColorRed)), 10, 3)
	expected := catatui.NewBufferWithStrings(
		"    ███   ",
		"█1█ █2█   ",
		"foo bar   ",
	)
	expected.Get(1, 1).SetFg(catatui.ColorRed)
	expected.Get(5, 1).SetFg(catatui.ColorRed)
	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartLabelStyle(t *testing.T) {
	buf := renderToBuffer(NewBarChart().
		DataPairs(pair("foo", 1), pair("bar", 2)).
		LabelStyle(catatui.NewStyle().Fg(catatui.ColorRed)), 10, 3)
	expected := catatui.NewBufferWithStrings(
		"  █       ",
		"1 2       ",
		"f b       ",
	)
	expected.Get(0, 2).SetFg(catatui.ColorRed)
	expected.Get(2, 2).SetFg(catatui.ColorRed)
	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartStyle(t *testing.T) {
	buf := renderToBuffer(NewBarChart().
		DataPairs(pair("foo", 1), pair("bar", 2)).
		Style(catatui.NewStyle().Fg(catatui.ColorRed)), 10, 3)
	expected := catatui.NewBufferWithStrings(
		"  █       ",
		"1 2       ",
		"f b       ",
	)
	expected.SetStyle(expected.Area, catatui.NewStyle().Fg(catatui.ColorRed))
	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartEmptyGroup(t *testing.T) {
	chart := NewBarChart().
		Data(NewBarGroup().Label(line("invisible"))).
		Data(NewBarGroup().Label(line("G")).Bars(NewBar(0).Value(1), NewBar(0).Value(2)))
	buf := renderToBuffer(chart, 3, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  █",
		"1 2",
		"G  ",
	))
}

func buildTestBarchart() BarChart {
	return NewBarChart().
		Data(NewBarGroup().Label(line("G1")).Bars(NewBar(2), NewBar(3), NewBar(4))).
		Data(NewBarGroup().Label(line("G2")).Bars(NewBar(3), NewBar(4), NewBar(5))).
		GroupGap(1).
		Direction(catatui.Horizontal).
		BarGap(0)
}

func TestBarChartHorizontalBars(t *testing.T) {
	buf := renderToBuffer(buildTestBarchart(), 5, 8)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"2█   ",
		"3██  ",
		"4███ ",
		"G1   ",
		"3██  ",
		"4███ ",
		"5████",
		"G2   ",
	))
}

func TestBarChartHorizontalBarsNoSpaceForGroupLabel(t *testing.T) {
	buf := renderToBuffer(buildTestBarchart(), 5, 7)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"2█   ",
		"3██  ",
		"4███ ",
		"G1   ",
		"3██  ",
		"4███ ",
		"5████",
	))
}

func TestBarChartHorizontalBarsNoSpaceForAllBars(t *testing.T) {
	buf := renderToBuffer(buildTestBarchart(), 5, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"2█   ",
		"3██  ",
		"4███ ",
		"G1   ",
		"3██  ",
	))
}

func testHorizontalBarsLabelWidthGreaterThanBar(t *testing.T, barColor catatui.Color) {
	t.Helper()
	bar := NewBar(2).TextValue("label").ValueStyle(catatui.NewStyle().Fg(catatui.ColorRed))
	if barColor.IsSet() {
		bar = bar.Style(catatui.NewStyle().Fg(barColor))
	}

	chart := NewBarChart().
		Data(NewBarGroup().Bars(bar, NewBar(5))).
		Direction(catatui.Horizontal).
		BarStyle(catatui.NewStyle().Fg(catatui.ColorYellow)).
		ValueStyle(catatui.NewStyle().AddModifier(catatui.ModifierItalic)).
		BarGap(0)
	buf := renderToBuffer(chart, 5, 2)

	expected := catatui.NewBufferWithStrings("label", "5████")

	// The second line has a yellow foreground; its first cell holds an italic "5".
	expected.Get(0, 1).SetStyle(catatui.NewStyle().AddModifier(catatui.ModifierItalic))
	for x := uint16(0); x < 5; x++ {
		expected.Get(x, 1).SetFg(catatui.ColorYellow)
	}

	expectedColor := barColor
	if !expectedColor.IsSet() {
		expectedColor = catatui.ColorYellow
	}

	// The first line holds the word "label". Since the bar value is 2, the
	// first 2 characters are italic red; the rest use the bar's style.
	italic := catatui.NewStyle().AddModifier(catatui.ModifierItalic)
	expected.Get(0, 0).SetFg(catatui.ColorRed).SetStyle(italic)
	expected.Get(1, 0).SetFg(catatui.ColorRed).SetStyle(italic)
	expected.Get(2, 0).SetFg(expectedColor)
	expected.Get(3, 0).SetFg(expectedColor)
	expected.Get(4, 0).SetFg(expectedColor)

	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartHorizontalBarsLabelWidthGreaterThanBarWithoutStyle(t *testing.T) {
	testHorizontalBarsLabelWidthGreaterThanBar(t, catatui.Color{})
}

func TestBarChartHorizontalBarsLabelWidthGreaterThanBarWithStyle(t *testing.T) {
	testHorizontalBarsLabelWidthGreaterThanBar(t, catatui.ColorWhite)
}

// TestBarChartHorizontalLabel checks that horizontal bars show their labels.
func TestBarChartHorizontalLabel(t *testing.T) {
	chart := NewBarChart().
		Direction(catatui.Horizontal).
		BarGap(0).
		DataPairs(pair("Jan", 10), pair("Feb", 20), pair("Mar", 5))
	buf := renderToBuffer(chart, 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"Jan 10█   ",
		"Feb 20████",
		"Mar 5     ",
	))
}

func TestBarChartGroupLabelStyle(t *testing.T) {
	chart := NewBarChart().
		Data(NewBarGroup().
			Label(catatui.NewLine(catatui.NewStyledSpan("G1", catatui.NewStyle().Fg(catatui.ColorRed)))).
			Bars(NewBar(2))).
		GroupGap(1).
		Direction(catatui.Horizontal).
		LabelStyle(catatui.NewStyle().AddModifier(catatui.ModifierBold).Fg(catatui.ColorYellow))
	buf := renderToBuffer(chart, 5, 2)

	// G1 is bold from the chart's LabelStyle and red from the label itself.
	expected := catatui.NewBufferWithStrings("2████", "G1   ")
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	expected.Get(0, 1).SetFg(catatui.ColorRed).SetStyle(bold)
	expected.Get(1, 1).SetFg(catatui.ColorRed).SetStyle(bold)
	catatui.AssertBuffer(t, buf, expected)
}

func TestBarChartGroupLabelCenter(t *testing.T) {
	// The centered group label when one bar of the group is out of view.
	group := pairs(pair("a", 1), pair("b", 2), pair("c", 3), pair("c", 4))
	chart := NewBarChart().
		Data(group.Label(line("G1").Alignment(catatui.AlignmentCenter))).
		Data(group.Label(line("G2").Alignment(catatui.AlignmentCenter)))
	buf := renderToBuffer(chart, 13, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    ▂ █     ▂",
		"  ▄ █ █   ▄ █",
		"▆ 2 3 4 ▆ 2 3",
		"a b c c a b c",
		"  G1     G2  ",
	))
}

func TestBarChartGroupLabelRight(t *testing.T) {
	chart := NewBarChart().Data(NewBarGroup().
		Label(catatui.NewLine(catatui.NewSpan("G")).Alignment(catatui.AlignmentRight)).
		Bars(NewBar(2), NewBar(5)))
	buf := renderToBuffer(chart, 3, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  █",
		"▆ 5",
		"  G",
	))
}

func TestBarChartUnicodeAsValue(t *testing.T) {
	group := NewBarGroup().Bars(
		NewBar(123).Label(line("B1")).TextValue("写"),
		NewBar(321).Label(line("B2")).TextValue("写"),
		NewBar(333).Label(line("B2")).TextValue("写"),
	)
	chart := NewBarChart().Data(group).BarWidth(3).BarGap(1)
	buf := renderToBuffer(chart, 11, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    ▆▆▆ ███",
		"    ███ ███",
		"▃▃▃ ███ ███",
		"写█ 写█ 写█",
		"B1  B2  B2 ",
	))
}

// TestBarChartHandlesZeroWidth ensures a chart with zero bar and gap width
// does not panic.
func TestBarChartHandlesZeroWidth(t *testing.T) {
	chart := NewBarChart().DataPairs(pair("A", 1)).BarWidth(0).BarGap(0)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 0, 10))
	chart.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBuffer(catatui.NewRect(0, 0, 0, 10)))
}

func TestBarChartSingleLine(t *testing.T) {
	group := pairs(nineLevelPairs()...).Label(line("Group"))
	chart := NewBarChart().Data(group).BarSet(symbols.BarNineLevels)
	buf := renderToBuffer(chart, 17, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("  ▁ ▂ ▃ ▄ ▅ ▆ ▇ 8"))
}

func TestBarChartTwoLines(t *testing.T) {
	group := pairs(nineLevelPairs()...).Label(line("Group"))
	chart := NewBarChart().Data(group).BarSet(symbols.BarNineLevels)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 17, 3))
	chart.Render(catatui.NewRect(0, 1, buf.Area.Width, 2), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                 ",
		"  ▁ ▂ ▃ ▄ ▅ ▆ ▇ 8",
		"a b c d e f g h i",
	))
}

func TestBarChartThreeLines(t *testing.T) {
	group := pairs(nineLevelPairs()...).Label(line("Group").Alignment(catatui.AlignmentCenter))
	chart := NewBarChart().Data(group).BarSet(symbols.BarNineLevels)
	buf := renderToBuffer(chart, 17, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"  ▁ ▂ ▃ ▄ ▅ ▆ ▇ 8",
		"a b c d e f g h i",
		"      Group      ",
	))
}

func TestBarChartThreeLinesDoubleWidth(t *testing.T) {
	group := pairs(nineLevelPairs()...).Label(line("Group").Alignment(catatui.AlignmentCenter))
	chart := NewBarChart().Data(group).BarWidth(2).BarSet(symbols.BarNineLevels)
	buf := renderToBuffer(chart, 26, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   1▁ 2▂ 3▃ 4▄ 5▅ 6▆ 7▇ 8█",
		"a  b  c  d  e  f  g  h  i ",
		"          Group           ",
	))
}

func TestBarChartFourLines(t *testing.T) {
	group := pairs(nineLevelPairs()...).Label(line("Group").Alignment(catatui.AlignmentCenter))
	chart := NewBarChart().Data(group).BarSet(symbols.BarNineLevels)
	buf := renderToBuffer(chart, 17, 4)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"          ▂ ▄ ▆ █",
		"  ▂ ▄ ▆ 4 5 6 7 8",
		"a b c d e f g h i",
		"      Group      ",
	))
}

func TestBarChartTwoLinesWithoutBarLabels(t *testing.T) {
	group := NewBarGroup().
		Label(line("Group").Alignment(catatui.AlignmentCenter)).
		Bars(NewBar(0), NewBar(1), NewBar(2), NewBar(3), NewBar(4), NewBar(5), NewBar(6), NewBar(7), NewBar(8))
	chart := NewBarChart().Data(group)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 17, 3))
	chart.Render(catatui.NewRect(0, 1, buf.Area.Width, 2), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                 ",
		"  ▁ ▂ ▃ ▄ ▅ ▆ ▇ 8",
		"      Group      ",
	))
}

func TestBarChartOneLinesWithMoreBars(t *testing.T) {
	bars := make([]Bar, 0, 30)
	for i := range uint64(30) {
		bars = append(bars, NewBar(i))
	}
	chart := NewBarChart().Data(NewBarGroup().Bars(bars...))
	buf := renderToBuffer(chart, 59, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"        ▁ ▁ ▁ ▁ ▂ ▂ ▂ ▃ ▃ ▃ ▃ ▄ ▄ ▄ ▄ ▅ ▅ ▅ ▆ ▆ ▆ ▆ ▇ ▇ ▇ █",
	))
}

func TestBarChartFirstBarOfTheGroupIsHalfOutsideView(t *testing.T) {
	chart := NewBarChart().
		DataPairs(pair("a", 1), pair("b", 2)).
		DataPairs(pair("a", 1), pair("b", 2)).
		BarWidth(2)
	buf := renderToBuffer(chart, 7, 6)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"   ██  ",
		"   ██  ",
		"▄▄ ██  ",
		"██ ██  ",
		"1█ 2█  ",
		"a  b   ",
	))
}

func TestBarChartRenderHandlesU64MaxValue(t *testing.T) {
	chart := VerticalBarChart(NewBar(math.MaxUint64)).Max(math.MaxUint64)
	buf := renderToBuffer(chart, 5, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"█    ",
		"█    ",
		"█    ",
		"█    ",
		"█    ",
	))
}

func TestBarChartRenderKeepsIntegerPrecisionForLargeValues(t *testing.T) {
	chart := VerticalBarChart(NewBar(math.MaxUint64 - 1)).Max(math.MaxUint64)
	buf := renderToBuffer(chart, 1, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("▇"))
}

func TestBarChartNew(t *testing.T) {
	bars := []Bar{BarWithLabel(line("Red"), 1), BarWithLabel(line("Green"), 2)}

	chart := VerticalBarChart(bars...)
	if len(chart.data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(chart.data))
	}
	if !reflect.DeepEqual(chart.data[0].bars, bars) {
		t.Errorf("data[0].bars = %#v, want %#v", chart.data[0].bars, bars)
	}

	updated := chart.DataPairs(pair("Blue", 3))
	if len(updated.data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(updated.data))
	}
	want := []Bar{BarWithLabel(line("Blue"), 3)}
	if !reflect.DeepEqual(updated.data[1].bars, want) {
		t.Errorf("data[1].bars = %#v, want %#v", updated.data[1].bars, want)
	}
}

// TestBarChartRegression1928 is ratatui's regression test for
// https://github.com/ratatui/ratatui/issues/1928: a text value made of
// multi-byte characters must not panic when it is split at the bar's end.
func TestBarChartRegression1928(t *testing.T) {
	const textValue = " " // narrow no-break space
	bars := []Bar{
		NewBar(0).TextValue(textValue),
		NewBar(1).TextValue(textValue),
		NewBar(2).TextValue(textValue),
		NewBar(3).TextValue(textValue),
		NewBar(4).TextValue(textValue),
	}
	chart := NewBarChart().
		Data(NewBarGroup().Bars(bars...)).
		BarGap(0).
		Direction(catatui.Horizontal)
	buf := renderToBuffer(chart, 4, 5)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"    ",
		"    ",
		" █  ",
		" ██ ",
		" ███",
	))
}

func TestBarChartRenderInMinimalBuffer(t *testing.T) {
	for _, d := range []catatui.Direction{catatui.Horizontal, catatui.Vertical} {
		t.Run(d.String(), func(t *testing.T) {
			chart := NewBarChart().
				DataPairs(pair("A", 1), pair("B", 2)).
				BarWidth(3).
				BarGap(1).
				Direction(d)
			// This must not panic, even though the buffer is too small for
			// the chart.
			buf := renderToBuffer(chart, 1, 1)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
		})
	}
}

func TestBarChartRenderInZeroSizeBuffer(t *testing.T) {
	for _, d := range []catatui.Direction{catatui.Horizontal, catatui.Vertical} {
		t.Run(d.String(), func(t *testing.T) {
			chart := NewBarChart().
				DataPairs(pair("A", 1), pair("B", 2)).
				BarWidth(3).
				BarGap(1).
				Direction(d)
			buf := catatui.NewBuffer(catatui.ZeroRect)
			// This must not panic, even though the buffer has zero size.
			chart.Render(buf.Area, buf)
		})
	}
}

// --- bar.rs ---------------------------------------------------------------

func TestBarNew(t *testing.T) {
	bar := NewBar(42).Label(line("Label"))
	if !bar.hasLabel || !reflect.DeepEqual(bar.label, line("Label")) {
		t.Errorf("label = %#v (set: %v), want %q", bar.label, bar.hasLabel, "Label")
	}
	if bar.value != 42 {
		t.Errorf("value = %d, want 42", bar.value)
	}
}

func TestBarWithLabel(t *testing.T) {
	bar := BarWithLabel(line("Label"), 42)
	if !bar.hasLabel || !reflect.DeepEqual(bar.label, line("Label")) {
		t.Errorf("label = %#v (set: %v), want %q", bar.label, bar.hasLabel, "Label")
	}
	if bar.value != 42 {
		t.Errorf("value = %d, want 42", bar.value)
	}
}

// --- bar_group.rs ---------------------------------------------------------

func TestBarGroupNew(t *testing.T) {
	group := NewBarGroup(BarWithLabel(line("Label1"), 1), BarWithLabel(line("Label2"), 2)).
		Label(line("Group1"))
	if !group.hasLabel || !reflect.DeepEqual(group.label, line("Group1")) {
		t.Errorf("label = %#v (set: %v), want %q", group.label, group.hasLabel, "Group1")
	}
	if len(group.bars) != 2 {
		t.Errorf("len(bars) = %d, want 2", len(group.bars))
	}
}
