// Tests ported from ratatui-widgets/src/paragraph.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// paragraphCase renders a paragraph into a buffer the size of the expected
// output and compares them, which is ratatui's test_case helper.
func paragraphCase(t *testing.T, p Paragraph, want ...string) {
	t.Helper()
	expected := catatui.NewBufferWithStrings(want...)
	buf := catatui.NewBuffer(expected.Area)
	p.Render(buf.Area, buf)
	catatui.AssertBuffer(t, buf, expected)
}

func TestParagraphRendersSingleLine(t *testing.T) {
	text := "Hello, world!"
	for _, p := range []Paragraph{
		NewParagraph(text),
		NewParagraph(text).Wrap(Wrap{Trim: false}),
		NewParagraph(text).Wrap(Wrap{Trim: true}),
	} {
		paragraphCase(t, p, "Hello, world!  ")
		paragraphCase(t, p, "Hello, world!")
		paragraphCase(t, p, "Hello, world!  ", "               ")
	}
}

func TestParagraphRendersEmpty(t *testing.T) {
	for _, p := range []Paragraph{
		NewParagraph(""),
		NewParagraph("").Wrap(Wrap{Trim: false}),
		NewParagraph("").Wrap(Wrap{Trim: true}),
	} {
		paragraphCase(t, p, " ")
		paragraphCase(t, p, "          ")
		paragraphCase(t, p, " ", " ")
	}
}

// TestParagraphZeroWidthCharAtEndOfLine covers a zero-width space, which
// occupies no columns and so must not push the line over the edge.
func TestParagraphZeroWidthCharAtEndOfLine(t *testing.T) {
	line := "foo​" // "foo" followed by U+200B
	for _, p := range []Paragraph{
		NewParagraph(line),
		NewParagraph(line).Wrap(Wrap{Trim: false}),
		NewParagraph(line).Wrap(Wrap{Trim: true}),
	} {
		paragraphCase(t, p, "foo")
		paragraphCase(t, p, "foo   ")
		paragraphCase(t, p, "foo   ", "      ")
		paragraphCase(t, p, "foo", "   ")
	}
}

func TestParagraphRendersMultipleLines(t *testing.T) {
	text := "This is a\nmultiline\nparagraph."
	for _, p := range []Paragraph{
		NewParagraph(text),
		NewParagraph(text).Wrap(Wrap{Trim: false}),
		NewParagraph(text).Wrap(Wrap{Trim: true}),
	} {
		paragraphCase(t, p, "This is a ", "multiline ", "paragraph.")
	}
}

// TestParagraphWordWrap is ratatui's word-wrap case, and the reason the
// wrapper is written the way it is: runs of whitespace are preserved untrimmed
// but collapsed at wrap points when trimming, and a word longer than the line
// is broken rather than dropped.
func TestParagraphWordWrap(t *testing.T) {
	text := "This is a long line of text that should wrap      and contains a superultramegagigalong word."
	wrapped := NewParagraph(text).Wrap(Wrap{Trim: false})
	trimmed := NewParagraph(text).Wrap(Wrap{Trim: true})

	paragraphCase(t, wrapped,
		"This is a long line",
		"of text that should",
		"wrap      and      ",
		"contains a         ",
		"superultramegagigal",
		"ong word.          ",
	)
	paragraphCase(t, wrapped,
		"This is a   ",
		"long line of",
		"text that   ",
		"should wrap ",
		"    and     ",
		"contains a  ",
		"superultrame",
		"gagigalong  ",
		"word.       ",
	)

	paragraphCase(t, trimmed,
		"This is a long line",
		"of text that should",
		"wrap      and      ",
		"contains a         ",
		"superultramegagigal",
		"ong word.          ",
	)
	paragraphCase(t, trimmed,
		"This is a   ",
		"long line of",
		"text that   ",
		"should wrap ",
		"and contains",
		"a           ",
		"superultrame",
		"gagigalong  ",
		"word.       ",
	)
}

// TestParagraphWrapWhitespaceOnlyLine checks that a line of nothing but spaces
// survives untrimmed and collapses to empty when trimming.
func TestParagraphWrapWhitespaceOnlyLine(t *testing.T) {
	lines := []catatui.Line{
		catatui.LineFromString("A"),
		catatui.LineFromString("  "),
		catatui.LineFromString("B"),
		catatui.LineFromString("  a"),
		catatui.LineFromString("C"),
	}
	text := catatui.NewText(lines...)

	paragraphCase(t, NewParagraphFromText(text).Wrap(Wrap{Trim: false}),
		"A  ", "   ", "B  ", "  a", "C  ")
	paragraphCase(t, NewParagraphFromText(text).Wrap(Wrap{Trim: true}),
		"A  ", "   ", "B  ", "a  ", "C  ")
}

func TestParagraphTruncatesWithoutWrap(t *testing.T) {
	p := NewParagraph("This is a long line of text that should be truncated.")
	paragraphCase(t, p, "This is a long line of")
}

func TestParagraphScrollVertically(t *testing.T) {
	p := NewParagraph("one\ntwo\nthree\nfour")
	paragraphCase(t, p.Scroll(catatui.Position{Y: 0}), "one  ", "two  ")
	paragraphCase(t, p.Scroll(catatui.Position{Y: 1}), "two  ", "three")
	paragraphCase(t, p.Scroll(catatui.Position{Y: 2}), "three", "four ")
	// Scrolling past the end leaves the area blank rather than panicking.
	paragraphCase(t, p.Scroll(catatui.Position{Y: 10}), "     ", "     ")
}

func TestParagraphScrollHorizontally(t *testing.T) {
	p := NewParagraph("abcdefgh")
	paragraphCase(t, p.Scroll(catatui.Position{X: 0}), "abcd")
	paragraphCase(t, p.Scroll(catatui.Position{X: 2}), "cdef")
	paragraphCase(t, p.Scroll(catatui.Position{X: 6}), "gh  ")
	paragraphCase(t, p.Scroll(catatui.Position{X: 20}), "    ")
}

func TestParagraphAlignment(t *testing.T) {
	p := NewParagraph("ab")
	paragraphCase(t, p.Left(), "ab    ")
	paragraphCase(t, p.Centered(), "  ab  ")
	paragraphCase(t, p.Right(), "    ab")
}

func TestParagraphInBlock(t *testing.T) {
	p := NewParagraph("hi").Block(Bordered().Title("T"))
	paragraphCase(t, p,
		"┌T──┐",
		"│hi │",
		"└───┘",
	)
}

func TestParagraphLineCount(t *testing.T) {
	p := NewParagraph("one\ntwo\nthree")
	if got := p.LineCount(10); got != 3 {
		t.Errorf("unwrapped LineCount = %d, want 3", got)
	}

	wrapped := NewParagraph("aaa bbb ccc").Wrap(Wrap{Trim: true})
	if got := wrapped.LineCount(4); got != 3 {
		t.Errorf("wrapped LineCount at width 4 = %d, want 3", got)
	}
	if got := wrapped.LineCount(11); got != 1 {
		t.Errorf("wrapped LineCount at width 11 = %d, want 1", got)
	}

	// A block eats into the width available for text.
	inBlock := NewParagraph("aaa bbb").Wrap(Wrap{Trim: true}).Block(Bordered())
	if got := inBlock.LineCount(5); got != 2 {
		t.Errorf("LineCount inside a border at width 5 = %d, want 2", got)
	}
}

// TestParagraphStylePropagates checks the layering: the paragraph's style sits
// beneath the text's, and both beneath each span's.
func TestParagraphStylePropagates(t *testing.T) {
	p := NewParagraphFromText(catatui.NewText(
		catatui.NewLine(
			catatui.NewSpan("ab"),
			catatui.NewStyledSpan("cd", catatui.NewStyle().Fg(catatui.ColorGreen)),
		),
	)).Style(catatui.NewStyle().Bg(catatui.ColorBlue).Fg(catatui.ColorRed))

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 4, 1))
	p.Render(buf.Area, buf)

	if got := buf.Get(0, 0).Fg; got != catatui.ColorRed {
		t.Errorf("unstyled span fg = %v, want red from the paragraph", got)
	}
	if got := buf.Get(2, 0).Fg; got != catatui.ColorGreen {
		t.Errorf("styled span fg = %v, want green from the span", got)
	}
	for x := range uint16(4) {
		if got := buf.Get(x, 0).Bg; got != catatui.ColorBlue {
			t.Errorf("cell %d bg = %v, want blue everywhere", x, got)
		}
	}
}

// TestParagraphWrapsWideCharacters checks that CJK text wraps on cell width
// rather than character count.
//
// Two wide glyphs fill four of five columns, so each row must hold exactly two
// characters and the text must be spread over four rows with nothing lost. If
// wrapping counted characters instead of columns it would fit five per row and
// overrun the area.
func TestParagraphWrapsWideCharacters(t *testing.T) {
	const text = "日本語のテキスト" // eight wide glyphs, 16 columns
	p := NewParagraph(text).Wrap(Wrap{Trim: false})
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 4))
	p.Render(buf.Area, buf)

	var got string
	for y := range uint16(4) {
		var row string
		for x := range uint16(5) {
			// A wide glyph sits in its own cell and blanks the next, so
			// collecting symbols and dropping the blanks rebuilds the text.
			if s := buf.Get(x, y).Symbol; s != "" {
				row += s
			}
		}
		if w := catatui.Graphemes(row); len(w) > 2 {
			t.Errorf("row %d holds %d wide glyphs, which cannot fit in 5 columns: %q", y, len(w), row)
		}
		got += row
	}
	if got != text {
		t.Errorf("wrapping lost or reordered text:\n  got  %q\n  want %q", got, text)
	}
}

func TestParagraphRendersNothingInAnEmptyArea(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 4, 2))
	NewParagraph("hello").Render(catatui.NewRect(0, 0, 0, 0), buf)
	NewParagraph("hello").Wrap(Wrap{Trim: true}).Render(catatui.NewRect(0, 0, 0, 2), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings("    ", "    "))
}
