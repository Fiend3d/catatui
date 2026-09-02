// Tests ported from ratatui-widgets/src/block.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// renderToBuffer draws a widget into a fresh buffer of the given size.
func renderToBuffer(w catatui.Widget, width, height uint16) *catatui.Buffer {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, width, height))
	w.Render(buf.Area, buf)
	return buf
}

func TestBlockRendersBorders(t *testing.T) {
	buf := renderToBuffer(Bordered(), 5, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌───┐",
		"│   │",
		"└───┘",
	))
}

func TestBlockRendersPartialBorders(t *testing.T) {
	cases := []struct {
		name    string
		borders Borders
		want    []string
	}{
		{"none", BordersNone, []string{"     ", "     ", "     "}},
		{"top", BordersTop, []string{"─────", "     ", "     "}},
		{"bottom", BordersBottom, []string{"     ", "     ", "─────"}},
		{"left", BordersLeft, []string{"│    ", "│    ", "│    "}},
		{"right", BordersRight, []string{"    │", "    │", "    │"}},
		{"top and left", BordersTop | BordersLeft, []string{"┌────", "│    ", "│    "}},
		{"all", BordersAll, []string{"┌───┐", "│   │", "└───┘"}},
	}
	for _, c := range cases {
		buf := renderToBuffer(NewBlock().Borders(c.borders), 5, 3)
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestBlockBorderTypes(t *testing.T) {
	cases := []struct {
		name string
		t    BorderType
		want []string
	}{
		{"plain", BorderPlain, []string{"┌───┐", "│   │", "└───┘"}},
		{"rounded", BorderRounded, []string{"╭───╮", "│   │", "╰───╯"}},
		{"double", BorderDouble, []string{"╔═══╗", "║   ║", "╚═══╝"}},
		{"thick", BorderThick, []string{"┏━━━┓", "┃   ┃", "┗━━━┛"}},
	}
	for _, c := range cases {
		buf := renderToBuffer(Bordered().BorderType(c.t), 5, 3)
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestBlockTitle(t *testing.T) {
	buf := renderToBuffer(Bordered().Title("Hi"), 8, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌Hi────┐",
		"│      │",
		"└──────┘",
	))
}

func TestBlockTitleAlignment(t *testing.T) {
	cases := []struct {
		name string
		a    catatui.Alignment
		want string
	}{
		{"left", catatui.AlignmentLeft, "┌Hi──────┐"},
		{"center", catatui.AlignmentCenter, "┌───Hi───┐"},
		{"right", catatui.AlignmentRight, "┌──────Hi┐"},
	}
	for _, c := range cases {
		buf := renderToBuffer(Bordered().Title("Hi").TitleAlignment(c.a), 10, 3)
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, buf,
				catatui.NewBufferWithStrings(c.want, "│        │", "└────────┘"))
		})
	}
}

func TestBlockTitleOnBottom(t *testing.T) {
	buf := renderToBuffer(Bordered().TitleBottom(catatui.LineFromString("Hi")), 8, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌──────┐",
		"│      │",
		"└Hi────┘",
	))
}

// TestBlockMultipleTitles is ratatui's left/center/right title cases. Titles on
// one edge are laid out in order with a one-cell gap, and when they do not fit
// the later ones are truncated rather than pushed off.
//
// Note that the gap cell is not cleared: it keeps whatever is underneath, which
// is why a bordered block shows the border character between two titles.
func TestBlockMultipleTitles(t *testing.T) {
	line := catatui.LineFromString
	cases := []struct {
		name  string
		block Block
		want  string
	}{
		{"left", NewBlock().Title("L12").Title("L34"), "L12 L34   "},
		{"left truncated", NewBlock().Title("L12345").Title("L67890"), "L12345 L67"},
		{
			"center",
			NewBlock().TitleLine(line("C12").Centered()).TitleLine(line("C34").Centered()),
			" C12 C34  ",
		},
		{
			"center truncated",
			NewBlock().TitleLine(line("C12345").Centered()).TitleLine(line("C67890").Centered()),
			"12345 C678",
		},
		{
			"right",
			NewBlock().TitleLine(line("R12").Right()).TitleLine(line("R34").Right()),
			"   R12 R34",
		},
		{
			"right truncated",
			NewBlock().TitleLine(line("R12345").Right()).TitleLine(line("R67890").Right()),
			"345 R67890",
		},
	}
	for _, c := range cases {
		buf := renderToBuffer(c.block, 10, 1)
		t.Run(c.name, func(t *testing.T) {
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want))
		})
	}
}

// TestBlockTitleGapKeepsTheBorder pins the consequence of the rule above: the
// separator between titles is not blanked, so on a bordered block the border
// shows through.
func TestBlockTitleGapKeepsTheBorder(t *testing.T) {
	buf := renderToBuffer(Bordered().Title("A").Title("B"), 10, 3)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌A─B─────┐",
		"│        │",
		"└────────┘",
	))
}

func TestBlockInner(t *testing.T) {
	area := catatui.NewRect(0, 0, 10, 10)
	cases := []struct {
		name  string
		block Block
		want  catatui.Rect
	}{
		{"no borders", NewBlock(), catatui.NewRect(0, 0, 10, 10)},
		{"all borders", Bordered(), catatui.NewRect(1, 1, 8, 8)},
		{"top only", NewBlock().Borders(BordersTop), catatui.NewRect(0, 1, 10, 9)},
		{"left only", NewBlock().Borders(BordersLeft), catatui.NewRect(1, 0, 9, 10)},
		{"right only", NewBlock().Borders(BordersRight), catatui.NewRect(0, 0, 9, 10)},
		{"bottom only", NewBlock().Borders(BordersBottom), catatui.NewRect(0, 0, 10, 9)},
		{"padding", NewBlock().Padding(UniformPadding(2)), catatui.NewRect(2, 2, 6, 6)},
		{"borders and padding", Bordered().Padding(UniformPadding(1)), catatui.NewRect(2, 2, 6, 6)},
	}
	for _, c := range cases {
		if got := c.block.Inner(area); got != c.want {
			t.Errorf("%s: Inner() = %+v, want %+v", c.name, got, c.want)
		}
	}
}

// TestBlockInnerAccountsForTitleWithoutBorder is the subtle one: a title needs a
// row even when there is no border for it to sit on, or it would overwrite the
// first line of content.
func TestBlockInnerAccountsForTitleWithoutBorder(t *testing.T) {
	area := catatui.NewRect(0, 0, 10, 10)
	if got, want := NewBlock().Title("T").Inner(area), catatui.NewRect(0, 1, 10, 9); got != want {
		t.Errorf("a top title should reserve a row: got %+v, want %+v", got, want)
	}
	got := NewBlock().TitleBottom(catatui.LineFromString("T")).Inner(area)
	if want := catatui.NewRect(0, 0, 10, 9); got != want {
		t.Errorf("a bottom title should reserve a row: got %+v, want %+v", got, want)
	}
}

// TestBlockInnerSaturatesInTinyAreas checks the widget cannot produce a rect
// with a negative-turned-huge size, which is what unchecked subtraction gives.
func TestBlockInnerSaturatesInTinyAreas(t *testing.T) {
	for _, size := range []uint16{0, 1, 2} {
		area := catatui.NewRect(0, 0, size, size)
		inner := Bordered().Padding(UniformPadding(3)).Inner(area)
		if inner.Width > size || inner.Height > size {
			t.Errorf("area %dx%d gave an inner rect larger than itself: %+v", size, size, inner)
		}
	}
}

func TestBlockStyle(t *testing.T) {
	block := Bordered().
		Style(catatui.NewStyle().Bg(catatui.ColorBlue)).
		BorderStyle(catatui.NewStyle().Fg(catatui.ColorRed))
	buf := renderToBuffer(block, 4, 3)

	if got := buf.Get(0, 0).Fg; got != catatui.ColorRed {
		t.Errorf("border fg = %v, want red", got)
	}
	if got := buf.Get(0, 0).Bg; got != catatui.ColorBlue {
		t.Errorf("the block style should show through on the border: bg = %v, want blue", got)
	}
	if got := buf.Get(1, 1).Bg; got != catatui.ColorBlue {
		t.Errorf("interior bg = %v, want blue", got)
	}
}

// TestBlockRendersNothingInAnEmptyArea guards against a panic or stray write
// when a layout collapses a pane to nothing, which happens constantly on resize.
func TestBlockRendersNothingInAnEmptyArea(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 5))
	Bordered().Title("x").Render(catatui.NewRect(0, 0, 0, 0), buf)
	Bordered().Title("x").Render(catatui.NewRect(2, 2, 0, 3), buf)
	Bordered().Title("x").Render(catatui.NewRect(9, 9, 3, 3), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"     ", "     ", "     ", "     ", "     ",
	))
}

// TestBlockRendersInOneByOneArea checks the degenerate case where every corner
// lands on the same cell.
func TestBlockRendersInOneByOneArea(t *testing.T) {
	buf := renderToBuffer(Bordered(), 1, 1)
	// All four corners collapse onto one cell; the last one drawn wins.
	if got := buf.Get(0, 0).GetSymbol(); got == " " {
		t.Errorf("a 1x1 bordered block should draw something, got %q", got)
	}
}

func TestBlockTitleTruncatedInANarrowArea(t *testing.T) {
	buf := renderToBuffer(Bordered().Title("LongTitle"), 6, 3)
	// The title is clipped to the space between the corners.
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"┌Long┐",
		"│    │",
		"└────┘",
	))
}

func TestPaddingConstructors(t *testing.T) {
	cases := []struct {
		name string
		got  Padding
		want Padding
	}{
		{"horizontal", HorizontalPadding(1), NewPadding(1, 1, 0, 0)},
		{"vertical", VerticalPadding(1), NewPadding(0, 0, 1, 1)},
		{"uniform", UniformPadding(1), NewPadding(1, 1, 1, 1)},
		{"proportional", ProportionalPadding(1), NewPadding(2, 2, 1, 1)},
		{"symmetric", SymmetricPadding(1, 2), NewPadding(1, 1, 2, 2)},
		{"left", LeftPadding(1), NewPadding(1, 0, 0, 0)},
		{"right", RightPadding(1), NewPadding(0, 1, 0, 0)},
		{"top", TopPadding(1), NewPadding(0, 0, 1, 0)},
		{"bottom", BottomPadding(1), NewPadding(0, 0, 0, 1)},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, c.got, c.want)
		}
	}
}
