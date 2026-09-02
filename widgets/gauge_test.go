// Tests ported from ratatui-widgets/src/gauge.rs and ratatui/tests/widgets_gauge.rs
// @ ratatui-v0.30.2

package widgets

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// mustPanic runs f and fails unless it panics with a message containing want,
// which is ratatui's #[should_panic = "..."].
func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected a panic containing %q, got none", want)
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic = %v, want one containing %q", r, want)
		}
	}()
	f()
}

func TestGaugeInvalidPercentage(t *testing.T) {
	mustPanic(t, "Percentage should be between 0 and 100 inclusively", func() {
		_ = NewGauge().Percent(110)
	})
}

func TestGaugeInvalidRatioUpperBound(t *testing.T) {
	mustPanic(t, "Ratio should be between 0 and 1 inclusively", func() {
		_ = NewGauge().Ratio(1.1)
	})
}

func TestGaugeInvalidRatioLowerBound(t *testing.T) {
	mustPanic(t, "Ratio should be between 0 and 1 inclusively", func() {
		_ = NewGauge().Ratio(-0.5)
	})
}

func TestLineGaugeInvalidRatio(t *testing.T) {
	mustPanic(t, "Ratio should be between 0 and 1 inclusively", func() {
		_ = NewLineGauge().Ratio(1.1)
	})
}

func TestLineGaugeCanBeStylizedWithDeprecatedGaugeStyle(t *testing.T) {
	gauge := NewLineGauge().GaugeStyle(catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorBlue))
	if want := catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorReset); gauge.filledStyle != want {
		t.Errorf("filledStyle = %+v, want %+v", gauge.filledStyle, want)
	}
	if want := catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorReset); gauge.unfilledStyle != want {
		t.Errorf("unfilledStyle = %+v, want %+v", gauge.unfilledStyle, want)
	}
}

func TestLineGaugeSetFilledSymbol(t *testing.T) {
	if got := NewLineGauge().FilledSymbol("▰").filledSymbol; got != "▰" {
		t.Errorf("filledSymbol = %q, want ▰", got)
	}
}

func TestLineGaugeSetUnfilledSymbol(t *testing.T) {
	if got := NewLineGauge().UnfilledSymbol("▱").unfilledSymbol; got != "▱" {
		t.Errorf("unfilledSymbol = %q, want ▱", got)
	}
}

func TestLineGaugeDeprecatedLineSet(t *testing.T) {
	gauge := NewLineGauge().LineSet(symbols.LineDouble)
	if gauge.filledSymbol != symbols.LineDouble.Horizontal {
		t.Errorf("filledSymbol = %q, want %q", gauge.filledSymbol, symbols.LineDouble.Horizontal)
	}
	if gauge.unfilledSymbol != symbols.LineDouble.Horizontal {
		t.Errorf("unfilledSymbol = %q, want %q", gauge.unfilledSymbol, symbols.LineDouble.Horizontal)
	}
}

func TestLineGaugeDefault(t *testing.T) {
	got := NewLineGauge()
	want := LineGauge{
		filledSymbol:   symbols.Horizontal,
		unfilledSymbol: symbols.Horizontal,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewLineGauge() = %+v, want %+v", got, want)
	}
}

func TestGaugeRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	gauge := NewGauge().Percent(50)
	// This must not panic, even though the buffer is too small for the gauge.
	gauge.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("5"))
}

func TestLineGaugeRenderInMinimalBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 1, 1))
	gauge := NewLineGauge().Ratio(0.5)
	// This must not panic, even though the buffer is too small for the gauge.
	gauge.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestGaugeRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	// This must not panic, even though the buffer has zero size.
	NewGauge().Percent(50).Render(buf.Area, buf)
}

func TestLineGaugeRenderInZeroSizeBuffer(t *testing.T) {
	buf := catatui.NewBuffer(catatui.ZeroRect)
	// This must not panic, even though the buffer has zero size.
	NewLineGauge().Ratio(0.5).Render(buf.Area, buf)
}

// The cases below come from ratatui/tests/widgets_gauge.rs. The Rust tests
// draw through a Terminal with a vertical layout; here the two chunks that
// layout produces are given directly.

func TestGaugeRenders(t *testing.T) {
	redOnBlue := catatui.NewStyle().Fg(catatui.ColorRed).Bg(catatui.ColorBlue)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 40, 10))
	NewGauge().
		Block(Bordered().Title("Percentage")).
		GaugeStyle(catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorRed)).
		UseUnicode(true).
		Percent(43).
		Render(catatui.NewRect(2, 2, 36, 3), buf)
	NewGauge().
		Block(Bordered().Title("Ratio")).
		GaugeStyle(catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorRed)).
		UseUnicode(true).
		Ratio(0.511_313_934_313_1).
		Render(catatui.NewRect(2, 5, 36, 3), buf)

	expected := catatui.NewBufferWithStrings(
		"                                        ",
		"                                        ",
		"  ┌Percentage────────────────────────┐  ",
		"  │██████████████▋43%                │  ",
		"  └──────────────────────────────────┘  ",
		"  ┌Ratio─────────────────────────────┐  ",
		"  │███████████████51%                │  ",
		"  └──────────────────────────────────┘  ",
		"                                        ",
		"                                        ",
	)
	expected.SetStyle(catatui.NewRect(3, 3, 34, 1), redOnBlue)

	expected.SetStyle(catatui.NewRect(3, 6, 15, 1), redOnBlue)
	// The filled part of the gauge only covers the 5 and the 1, not the %.
	expected.SetStyle(catatui.NewRect(18, 6, 2, 1), catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorRed))
	expected.SetStyle(catatui.NewRect(20, 6, 17, 1), redOnBlue)

	catatui.AssertBuffer(t, buf, expected)
}

func TestGaugeRendersNoUnicode(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 40, 10))
	NewGauge().
		Block(Bordered().Title("Percentage")).
		Percent(43).
		UseUnicode(false).
		Render(catatui.NewRect(2, 2, 36, 3), buf)
	NewGauge().
		Block(Bordered().Title("Ratio")).
		Ratio(0.211_313_934_313_1).
		UseUnicode(false).
		Render(catatui.NewRect(2, 5, 36, 3), buf)

	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"                                        ",
		"                                        ",
		"  ┌Percentage────────────────────────┐  ",
		"  │███████████████43%                │  ",
		"  └──────────────────────────────────┘  ",
		"  ┌Ratio─────────────────────────────┐  ",
		"  │███████        21%                │  ",
		"  └──────────────────────────────────┘  ",
		"                                        ",
		"                                        ",
	))
}

func TestGaugeAppliesStyles(t *testing.T) {
	red := catatui.NewStyle().Fg(catatui.ColorRed)
	blueOnRed := catatui.NewStyle().Fg(catatui.ColorBlue).Bg(catatui.ColorRed)
	labelStyle := catatui.NewStyle().Fg(catatui.ColorGreen).AddModifier(catatui.ModifierBold)

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 12, 5))
	NewGauge().
		Block(Bordered().TitleLine(catatui.NewLine(catatui.NewStyledSpan("Test", red)))).
		GaugeStyle(blueOnRed).
		Percent(43).
		LabelSpan(catatui.NewStyledSpan("43%", labelStyle)).
		Render(buf.Area, buf)

	expected := catatui.NewBufferWithStrings(
		"┌Test──────┐",
		"│████      │",
		"│███43%    │",
		"│████      │",
		"└──────────┘",
	)
	// title
	expected.SetStyle(catatui.NewRect(1, 0, 4, 1), red)
	// gauge area
	expected.SetStyle(catatui.NewRect(1, 1, 10, 3), blueOnRed)
	// filled area
	expected.SetStyle(catatui.NewRect(1, 1, 4, 3), blueOnRed)
	// label: foreground and modifier from the label style. The "4" is in the
	// filled area, so its background is the gauge style's foreground.
	expected.SetStyle(catatui.NewRect(4, 2, 1, 1), labelStyle.Bg(catatui.ColorBlue))
	// "3%" is not in the filled area, so its background is the gauge style's
	// background.
	expected.SetStyle(catatui.NewRect(5, 2, 2, 1), labelStyle.Bg(catatui.ColorRed))
	catatui.AssertBuffer(t, buf, expected)
}

func TestGaugeSupportsLargeLabels(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 1))
	NewGauge().
		Percent(43).
		Label("43333333333333333333333333333%").
		Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("4333333333"))
}

func TestLineGaugeRenders(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 20, 6))
	NewLineGauge().
		FilledStyle(catatui.NewStyle().Fg(catatui.ColorGreen)).
		UnfilledStyle(catatui.NewStyle().Fg(catatui.ColorWhite)).
		Ratio(0.43).
		Render(catatui.NewRect(0, 0, 20, 1), buf)
	// Custom (same) symbols for the filled and unfilled parts.
	NewLineGauge().
		Block(Bordered().Title("Gauge 2")).
		FilledStyle(catatui.NewStyle().Fg(catatui.ColorGreen)).
		FilledSymbol(symbols.ThickHorizontal).
		UnfilledSymbol(symbols.ThickHorizontal).
		Ratio(0.211_313_934_313_1).
		Render(catatui.NewRect(0, 1, 20, 3), buf)
	// The default symbol for the filled part, but empty for the unfilled part.
	NewLineGauge().UnfilledSymbol(" ").Ratio(0.50).
		Render(catatui.NewRect(0, 4, 20, 1), buf)
	// Different custom symbols for the filled and unfilled parts.
	NewLineGauge().
		FilledSymbol("█").
		UnfilledSymbol("░").
		Ratio(0.80).
		Render(catatui.NewRect(0, 5, 20, 1), buf)

	expected := catatui.NewBufferWithStrings(
		" 43% ───────────────",
		"┌Gauge 2───────────┐",
		"│ 21% ━━━━━━━━━━━━━│",
		"└──────────────────┘",
		" 50% ───────        ",
		" 80% ████████████░░░",
	)
	for col := uint16(5); col < 11; col++ {
		expected.Get(col, 0).SetFg(catatui.ColorGreen)
	}
	for col := uint16(11); col < 20; col++ {
		expected.Get(col, 0).SetFg(catatui.ColorWhite)
	}
	for col := uint16(6); col < 8; col++ {
		expected.Get(col, 2).SetFg(catatui.ColorGreen)
	}
	catatui.AssertBuffer(t, buf, expected)
}
