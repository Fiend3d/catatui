// Port of ratatui-widgets/src/paragraph.rs and reflow.rs @ ratatui-v0.30.2

package widgets

import (
	"unicode"

	"github.com/Fiend3d/catatui"
)

// Wrap configures word wrapping for a Paragraph.
type Wrap struct {
	// Trim removes leading whitespace from each wrapped line. Leave it false to
	// keep indentation, which matters for code and pre-formatted text.
	Trim bool
}

// Paragraph draws styled text in an area, optionally inside a Block, with
// optional word wrapping and scrolling.
//
//	p := widgets.NewParagraph("some long text").
//		Block(widgets.Bordered().Title("Notes")).
//		Wrap(widgets.Wrap{Trim: true})
//	f.RenderWidget(p, area)
//
// Without Wrap, lines are truncated at the right edge and Scroll's X offset
// slides them horizontally. With Wrap, lines are reflowed and the X offset is
// ignored, because there is nothing left to scroll horizontally.
type Paragraph struct {
	block     Block
	hasBlock  bool
	style     catatui.Style
	wrap      Wrap
	hasWrap   bool
	text      catatui.Text
	scroll    catatui.Position
	alignment catatui.Alignment
}

// NewParagraph returns a paragraph holding the given text, split on newlines.
func NewParagraph(text string) Paragraph {
	return Paragraph{text: catatui.TextFromString(text)}
}

// NewParagraphFromText returns a paragraph holding already-styled text.
func NewParagraphFromText(text catatui.Text) Paragraph {
	return Paragraph{text: text}
}

// Block returns a copy of p drawn inside the given block.
func (p Paragraph) Block(b Block) Paragraph { p.block, p.hasBlock = b, true; return p }

// Style returns a copy of p with a style applied beneath the text's own.
func (p Paragraph) Style(s catatui.Style) Paragraph { p.style = s; return p }

// Wrap returns a copy of p with word wrapping enabled.
func (p Paragraph) Wrap(w Wrap) Paragraph { p.wrap, p.hasWrap = w, true; return p }

// Scroll returns a copy of p scrolled by the given offset. The X offset only
// applies when wrapping is off.
func (p Paragraph) Scroll(pos catatui.Position) Paragraph { p.scroll = pos; return p }

// Alignment returns a copy of p with the default alignment for lines that do
// not set their own.
func (p Paragraph) Alignment(a catatui.Alignment) Paragraph { p.alignment = a; return p }

// Left returns a copy of p aligned to the left.
func (p Paragraph) Left() Paragraph { p.alignment = catatui.AlignmentLeft; return p }

// Centered returns a copy of p centered in its area.
func (p Paragraph) Centered() Paragraph { p.alignment = catatui.AlignmentCenter; return p }

// Right returns a copy of p aligned to the right.
func (p Paragraph) Right() Paragraph { p.alignment = catatui.AlignmentRight; return p }

// LineCount reports how many lines the paragraph needs at the given width.
//
// It is what you use to size a scrollbar or to decide how tall to make the
// paragraph's area, and it accounts for wrapping.
func (p Paragraph) LineCount(width uint16) int {
	inner := width
	if p.hasBlock {
		inner = p.block.Inner(catatui.NewRect(0, 0, width, 1)).Width
	}
	if inner == 0 {
		return 0
	}
	if !p.hasWrap {
		return len(p.text.GetLines())
	}
	n := 0
	for _, line := range p.text.GetLines() {
		n += len(wrapLine(line, inner, p.wrap.Trim))
	}
	return n
}

// Render draws the paragraph.
func (p Paragraph) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, p.style)

	textArea := area
	if p.hasBlock {
		p.block.Render(area, buf)
		textArea = p.block.Inner(area)
	}
	if textArea.IsEmpty() {
		return
	}
	buf.SetStyle(textArea, p.style)

	// Each source line is turned into one or more display lines, then drawn
	// from the scroll offset down.
	var display []catatui.Line
	for _, line := range p.text.GetLines() {
		// The text's style sits beneath the line's, and both beneath each span's.
		line = line.Style(p.text.GetStyle().Patch(line.GetStyle()))
		if p.hasWrap {
			display = append(display, wrapLine(line, textArea.Width, p.wrap.Trim)...)
			continue
		}
		display = append(display, line)
	}

	for i := int(p.scroll.Y); i < len(display); i++ {
		y := textArea.Y + uint16(i) - p.scroll.Y
		if y >= textArea.Bottom() {
			break
		}
		line := display[i]
		if !p.hasWrap && p.scroll.X > 0 {
			line = scrollLineHorizontally(line, p.scroll.X)
		}
		row := catatui.Rect{X: textArea.X, Y: y, Width: textArea.Width, Height: 1}
		line.RenderWithFallbackAlignment(row, buf, p.alignment)
	}
}

// scrollLineHorizontally drops the first n columns of a line, cutting on
// grapheme boundaries so a wide cluster is never split.
func scrollLineHorizontally(line catatui.Line, n uint16) catatui.Line {
	remaining := int(n)
	var out []catatui.Span
	for _, s := range line.GetSpans() {
		w := s.Width()
		if remaining >= w {
			remaining -= w
			continue
		}
		if remaining > 0 {
			s = s.Content(catatui.TrimLeftColumns(s.GetContent(), remaining))
			remaining = 0
		}
		out = append(out, s)
	}
	return line.Spans(out...)
}

// wrapLine breaks a line into as many display lines as it takes to fit within
// width.
//
// This is a port of ratatui's WordWrapper. The shape is unusual because it has
// to hold a pending word and the whitespace before it separately: whitespace at
// a wrap point is dropped, but whitespace in the middle of a line is kept, and
// which one you have is only known once the following word is complete.
func wrapLine(line catatui.Line, width uint16, trim bool) []catatui.Line {
	if width == 0 {
		return nil
	}

	var (
		out             []catatui.Line
		pendingLine     []catatui.Span
		pendingWord     []catatui.Span
		pendingSpace    []catatui.Span
		lineWidth       uint16
		wordWidth       uint16
		whitespaceWidth uint16
		prevNonSpace    bool
	)

	flushLine := func() {
		out = append(out, catatui.NewLine(pendingLine...).
			Style(line.GetStyle()).Alignment(line.GetAlignment()))
		pendingLine = nil
	}

	for _, g := range styledGraphemes(line) {
		isSpace := isWhitespaceSymbol(g.symbol)
		w := g.width

		// A grapheme wider than the whole area can never be placed.
		if w > width {
			continue
		}

		wordFound := prevNonSpace && isSpace
		empty := len(pendingLine) == 0
		trimmedOverflow := empty && trim && wordWidth+w > width
		whitespaceOverflow := empty && trim && whitespaceWidth+w > width
		untrimmedOverflow := empty && !trim && wordWidth+whitespaceWidth+w > width

		if wordFound || trimmedOverflow || whitespaceOverflow || untrimmedOverflow {
			// The pending word is complete: commit it, along with the
			// whitespace before it unless we are trimming at a line start.
			if len(pendingLine) > 0 || !trim {
				pendingLine = append(pendingLine, pendingSpace...)
				lineWidth += whitespaceWidth
			}
			pendingSpace = nil
			pendingLine = append(pendingLine, pendingWord...)
			pendingWord = nil
			lineWidth += wordWidth
			whitespaceWidth, wordWidth = 0, 0
		}

		lineFull := lineWidth >= width
		wordOverflow := w > 0 && lineWidth+whitespaceWidth+wordWidth >= width

		if lineFull || wordOverflow {
			remaining := catatui.SatSub(width, lineWidth)
			flushLine()
			lineWidth = 0

			// Trailing whitespace that still fits is kept on the line just
			// emitted; the rest is dropped at the wrap point.
			for len(pendingSpace) > 0 {
				sw := uint16(min(pendingSpace[0].Width(), 0xFFFF))
				if sw > remaining {
					break
				}
				whitespaceWidth -= sw
				remaining -= sw
				pendingSpace = pendingSpace[1:]
			}

			// The whitespace that caused the wrap does not start the next line.
			if isSpace && len(pendingSpace) == 0 {
				prevNonSpace = !isSpace
				continue
			}
		}

		span := catatui.NewStyledSpan(g.symbol, g.style)
		if isSpace {
			whitespaceWidth += w
			pendingSpace = append(pendingSpace, span)
		} else {
			wordWidth += w
			pendingWord = append(pendingWord, span)
		}
		prevNonSpace = !isSpace
	}

	// A line of nothing but whitespace becomes an empty line when trimming.
	if len(pendingLine) == 0 && len(pendingWord) == 0 && len(pendingSpace) > 0 && trim {
		out = append(out, catatui.NewLine().Style(line.GetStyle()).Alignment(line.GetAlignment()))
	}
	if len(pendingLine) > 0 || !trim {
		pendingLine = append(pendingLine, pendingSpace...)
	}
	pendingLine = append(pendingLine, pendingWord...)
	if len(pendingLine) > 0 {
		flushLine()
	}

	// An empty source line still occupies a row.
	if len(out) == 0 {
		out = append(out, catatui.NewLine().Style(line.GetStyle()).Alignment(line.GetAlignment()))
	}
	return out
}

// styledGrapheme is one grapheme cluster with the style it inherited.
type styledGrapheme struct {
	symbol string
	width  uint16
	style  catatui.Style
}

// styledGraphemes flattens a line into grapheme clusters, each carrying the
// span style it came from. Wrapping has to work at this granularity because a
// word can span several differently styled spans.
func styledGraphemes(line catatui.Line) []styledGrapheme {
	var out []styledGrapheme
	for _, s := range line.GetSpans() {
		style := s.GetStyle()
		for _, g := range catatui.Graphemes(s.GetContent()) {
			out = append(out, styledGrapheme{symbol: g.Symbol, width: g.Width, style: style})
		}
	}
	return out
}

// isWhitespaceSymbol reports whether a grapheme counts as a wrap point.
//
// A non-breaking space deliberately does not, which is the whole point of it; a
// zero-width space does, which is the whole point of that.
func isWhitespaceSymbol(s string) bool {
	const (
		zeroWidthSpace   = "​"
		nonBreakingSpace = " "
	)
	if s == zeroWidthSpace {
		return true
	}
	if s == nonBreakingSpace {
		return false
	}
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return s != ""
}

var _ catatui.Widget = Paragraph{}
