// Tests ported from ratatui-widgets/src/sparkline.rs @ ratatui-v0.30.2

package widgets

import (
	"math"
	"reflect"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// renderSparkline draws a sparkline into a one-row buffer of the given width
// pre-filled with "x", so that untouched cells are visible in the result.
func renderSparkline(w Sparkline, width uint16) *catatui.Buffer {
	area := catatui.NewRect(0, 0, width, 1)
	buf := catatui.NewBufferFilled(area, catatui.NewCell("x"))
	w.Render(area, buf)
	return buf
}

func TestSparklineRenderDirectionToString(t *testing.T) {
	if got := RenderLeftToRight.String(); got != "LeftToRight" {
		t.Errorf("got %q, want LeftToRight", got)
	}
	if got := RenderRightToLeft.String(); got != "RightToLeft" {
		t.Errorf("got %q, want RightToLeft", got)
	}
}

func TestSparklineItCanBeCreatedFromU64(t *testing.T) {
	got := NewSparkline().Data(1, 2, 3).data
	want := []SparklineBar{NewSparklineBar(1), NewSparklineBar(2), NewSparklineBar(3)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestSparklineItCanBeCreatedFromOptionU64(t *testing.T) {
	got := NewSparkline().DataBars(NewSparklineBar(1), AbsentSparklineBar(), NewSparklineBar(3)).data
	want := []SparklineBar{NewSparklineBar(1), AbsentSparklineBar(), NewSparklineBar(3)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestSparklineItDoesNotPanicIfMaxIsZero(t *testing.T) {
	buf := renderSparkline(NewSparkline().Data(0, 0, 0), 6)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("   xxx"))
}

func TestSparklineItDoesNotPanicIfMaxIsSetToZero(t *testing.T) {
	buf := renderSparkline(NewSparkline().Data(0, 1, 2).Max(0), 6)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("   xxx"))
}

func TestSparklineItDraws(t *testing.T) {
	buf := renderSparkline(NewSparkline().Data(0, 1, 2, 3, 4, 5, 6, 7, 8), 12)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" ▁▂▃▄▅▆▇█xxx"))
}

func TestSparklineItDrawsDoubleHeight(t *testing.T) {
	w := NewSparkline().Data(0, 1, 2, 3, 4, 5, 6, 7, 8)
	area := catatui.NewRect(0, 0, 12, 2)
	buf := catatui.NewBufferFilled(area, catatui.NewCell("x"))
	w.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"     ▂▄▆█xxx",
		" ▂▄▆█████xxx",
	))
}

func TestSparklineRenderHandlesU64MaxValue(t *testing.T) {
	w := NewSparkline().Data(math.MaxUint64).Max(math.MaxUint64)
	area := catatui.NewRect(0, 0, 1, 3)
	buf := catatui.NewBuffer(area)
	w.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("█", "█", "█"))
}

func TestSparklineRenderKeepsIntegerPrecisionForLargeValues(t *testing.T) {
	w := NewSparkline().Data(math.MaxUint64 - 1).Max(math.MaxUint64)
	area := catatui.NewRect(0, 0, 1, 1)
	buf := catatui.NewBuffer(area)
	w.Render(area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("▇"))
}

func TestSparklineItRendersLeftToRight(t *testing.T) {
	w := NewSparkline().Data(0, 1, 2, 3, 4, 5, 6, 7, 8).Direction(RenderLeftToRight)
	buf := renderSparkline(w, 12)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" ▁▂▃▄▅▆▇█xxx"))
}

func TestSparklineItRendersRightToLeft(t *testing.T) {
	w := NewSparkline().Data(0, 1, 2, 3, 4, 5, 6, 7, 8).Direction(RenderRightToLeft)
	buf := renderSparkline(w, 12)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("xxx█▇▆▅▄▃▂▁ "))
}

// absentThenOneToEight is the dataset [None, 1..=8] used by the absent-value
// tests.
func absentThenOneToEight() []SparklineBar {
	bars := []SparklineBar{AbsentSparklineBar()}
	for v := uint64(1); v <= 8; v++ {
		bars = append(bars, NewSparklineBar(v))
	}
	return bars
}

func TestSparklineItRendersWithAbsentValueStyle(t *testing.T) {
	w := NewSparkline().
		AbsentValueStyle(catatui.NewStyle().Fg(catatui.ColorRed)).
		AbsentValueSymbol(symbols.ShadeFull).
		DataBars(absentThenOneToEight()...)
	buf := renderSparkline(w, 12)
	want := catatui.NewBufferWithStrings("█▁▂▃▄▅▆▇█xxx")
	want.SetStyle(catatui.NewRect(0, 0, 1, 1), catatui.NewStyle().Fg(catatui.ColorRed))
	catatui.AssertBuffer(t, buf, want)
}

func TestSparklineItRendersWithAbsentValueStyleDoubleHeight(t *testing.T) {
	w := NewSparkline().
		AbsentValueStyle(catatui.NewStyle().Fg(catatui.ColorRed)).
		AbsentValueSymbol(symbols.ShadeFull).
		DataBars(absentThenOneToEight()...)
	area := catatui.NewRect(0, 0, 12, 2)
	buf := catatui.NewBufferFilled(area, catatui.NewCell("x"))
	w.Render(area, buf)
	want := catatui.NewBufferWithStrings(
		"█    ▂▄▆█xxx",
		"█▂▄▆█████xxx",
	)
	want.SetStyle(catatui.NewRect(0, 0, 1, 2), catatui.NewStyle().Fg(catatui.ColorRed))
	catatui.AssertBuffer(t, buf, want)
}

func TestSparklineItRendersWithCustomAbsentValueStyle(t *testing.T) {
	w := NewSparkline().AbsentValueSymbol("*").DataBars(absentThenOneToEight()...)
	buf := renderSparkline(w, 12)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("*▁▂▃▄▅▆▇█xxx"))
}

func TestSparklineItRendersWithCustomBarStyles(t *testing.T) {
	red := catatui.NewStyle().Fg(catatui.ColorRed)
	green := catatui.NewStyle().Fg(catatui.ColorGreen)
	blue := catatui.NewStyle().Fg(catatui.ColorBlue)
	w := NewSparkline().DataBars(
		NewSparklineBar(0).Style(red),
		NewSparklineBar(1).Style(red),
		NewSparklineBar(2).Style(red),
		NewSparklineBar(3).Style(green),
		NewSparklineBar(4).Style(green),
		NewSparklineBar(5).Style(green),
		NewSparklineBar(6).Style(blue),
		NewSparklineBar(7).Style(blue),
		NewSparklineBar(8).Style(blue),
	)
	buf := renderSparkline(w, 12)
	want := catatui.NewBufferWithStrings(" ▁▂▃▄▅▆▇█xxx")
	want.SetStyle(catatui.NewRect(0, 0, 3, 1), red)
	want.SetStyle(catatui.NewRect(3, 0, 3, 1), green)
	want.SetStyle(catatui.NewRect(6, 0, 3, 1), blue)
	catatui.AssertBuffer(t, buf, want)
}

func TestSparklineRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	w := NewSparkline().Data(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Max(10)
	// This must not panic, even though the buffer is too small for the data.
	w.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestSparklineRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	w := NewSparkline().Data(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Max(10)
	// This must not panic, even though the buffer has zero size.
	w.Render(buf.Area, buf)
}

// TestSparklineDataBarsCopiesInput pins the builder rule that the caller's
// slice stays independent of the widget.
func TestSparklineDataBarsCopiesInput(t *testing.T) {
	bars := []SparklineBar{NewSparklineBar(1), NewSparklineBar(2)}
	w := NewSparkline().DataBars(bars...)
	bars[0] = NewSparklineBar(9)
	if v, _ := w.data[0].Value(); v != 1 {
		t.Errorf("widget data changed with caller's slice: got %d, want 1", v)
	}
}
