// Port of ratatui-core/src/text/span.rs, line.rs and text.rs, and the Alignment
// part of ratatui-core/src/layout/alignment.rs @ ratatui-v0.30.2
//
// Span, Line and Text form a three-level hierarchy: a Span is a run of text in
// one style, a Line is a row of Spans, and a Text is a block of Lines. Each
// level carries its own Style, and the styles are layered outermost-first with
// Style.Patch when the text is drawn.
//
// Deviation from ratatui: Rust exposes public fields (`span.style`) alongside
// same-named builder methods (`span.style(s)`). Go allows only one, so fields
// are unexported, builders keep ratatui's names, and readers use Get*.

package catatui

import "strings"

// Alignment is the horizontal alignment of a line within its area.
//
// The zero value, AlignmentNone, means "not specified" and stands in for
// ratatui's Option<Alignment>::None: the containing widget decides. Note that
// ratatui's standalone Alignment defaults to Left instead, so write
// AlignmentLeft when you mean left rather than relying on the zero value.
type Alignment uint8

const (
	// AlignmentNone leaves alignment up to the containing widget.
	AlignmentNone Alignment = iota
	// AlignmentLeft aligns to the left edge.
	AlignmentLeft
	// AlignmentCenter centers within the area.
	AlignmentCenter
	// AlignmentRight aligns to the right edge.
	AlignmentRight
)

// String returns the alignment's name.
func (a Alignment) String() string {
	switch a {
	case AlignmentLeft:
		return "Left"
	case AlignmentCenter:
		return "Center"
	case AlignmentRight:
		return "Right"
	default:
		return "None"
	}
}

// --- Span ----------------------------------------------------------------

// Span is a run of text drawn in a single style. It is the smallest unit of
// styled text, and never contains a line break.
type Span struct {
	content string
	style   Style
}

// NewSpan returns an unstyled span holding the given content. It is the
// equivalent of ratatui's Span::raw.
func NewSpan(content string) Span { return Span{content: content} }

// NewStyledSpan returns a span holding the given content in the given style.
func NewStyledSpan(content string, style Style) Span {
	return Span{content: content, style: style}
}

// Content returns a copy of s with the content replaced.
func (s Span) Content(content string) Span { s.content = content; return s }

// Style returns a copy of s with the style replaced. The style is not patched
// onto the existing one; use Patch for that.
func (s Span) Style(style Style) Span { s.style = style; return s }

// Patch returns a copy of s with the given style layered on top of its own.
func (s Span) Patch(style Style) Span { s.style = s.style.Patch(style); return s }

// GetContent returns the span's text.
func (s Span) GetContent() string { return s.content }

// GetStyle returns the span's style.
func (s Span) GetStyle() Style { return s.style }

// Width is the number of columns the span occupies when drawn.
func (s Span) Width() int { return displayWidth(s.content) }

// String returns the span's text, so a Span prints as its content.
func (s Span) String() string { return s.content }

// --- Line ----------------------------------------------------------------

// Line is a single row of styled text, made of one or more Spans.
//
// A Line's own Style sits underneath its spans' styles, and its Alignment
// overrides whatever the containing widget would otherwise use.
type Line struct {
	spans     []Span
	style     Style
	alignment Alignment
}

// NewLine returns a line made of the given spans.
func NewLine(spans ...Span) Line { return Line{spans: spans} }

// LineFromString returns an unstyled single-span line.
func LineFromString(s string) Line { return Line{spans: []Span{NewSpan(s)}} }

// LineFromStyledString returns a single-span line in the given style. The style
// is applied to the span, matching ratatui's Line::styled.
func LineFromStyledString(s string, style Style) Line {
	return Line{spans: []Span{NewStyledSpan(s, style)}}
}

// Spans returns a copy of l with the spans replaced.
func (l Line) Spans(spans ...Span) Line { l.spans = spans; return l }

// Style returns a copy of l with the line style replaced. The line style sits
// beneath the spans' own styles.
func (l Line) Style(style Style) Line { l.style = style; return l }

// Patch returns a copy of l with the given style layered on top of its own.
func (l Line) Patch(style Style) Line { l.style = l.style.Patch(style); return l }

// Alignment returns a copy of l with the alignment replaced.
func (l Line) Alignment(a Alignment) Line { l.alignment = a; return l }

// Left returns a copy of l aligned to the left.
func (l Line) Left() Line { l.alignment = AlignmentLeft; return l }

// Centered returns a copy of l centered in its area.
func (l Line) Centered() Line { l.alignment = AlignmentCenter; return l }

// Right returns a copy of l aligned to the right.
func (l Line) Right() Line { l.alignment = AlignmentRight; return l }

// GetSpans returns the line's spans. The slice is the line's own; do not modify it.
func (l Line) GetSpans() []Span { return l.spans }

// GetStyle returns the line's style.
func (l Line) GetStyle() Style { return l.style }

// GetAlignment returns the line's alignment, which may be AlignmentNone.
func (l Line) GetAlignment() Alignment { return l.alignment }

// Width is the number of columns the line occupies when drawn.
func (l Line) Width() int {
	w := 0
	for _, s := range l.spans {
		w += s.Width()
	}
	return w
}

// String returns the line's text with all styling dropped.
func (l Line) String() string {
	var b strings.Builder
	for _, s := range l.spans {
		b.WriteString(s.content)
	}
	return b.String()
}

// --- Text ----------------------------------------------------------------

// Text is a block of styled Lines.
//
// Its Style sits underneath every line's style, and its Alignment applies to
// any line that does not set its own.
type Text struct {
	lines     []Line
	style     Style
	alignment Alignment
}

// NewText returns a Text made of the given lines.
func NewText(lines ...Line) Text { return Text{lines: lines} }

// TextFromString returns a Text by splitting s on newlines, so that multi-line
// strings become multiple Lines. This matches ratatui's From<&str> for Text.
func TextFromString(s string) Text {
	parts := strings.Split(s, "\n")
	lines := make([]Line, len(parts))
	for i, p := range parts {
		lines[i] = LineFromString(strings.TrimSuffix(p, "\r"))
	}
	return Text{lines: lines}
}

// TextFromStyledString is TextFromString with a style applied to the whole block.
func TextFromStyledString(s string, style Style) Text {
	t := TextFromString(s)
	t.style = style
	return t
}

// Lines returns a copy of t with the lines replaced.
func (t Text) Lines(lines ...Line) Text { t.lines = lines; return t }

// Style returns a copy of t with the style replaced.
func (t Text) Style(style Style) Text { t.style = style; return t }

// Patch returns a copy of t with the given style layered on top of its own.
func (t Text) Patch(style Style) Text { t.style = t.style.Patch(style); return t }

// Alignment returns a copy of t with the alignment replaced.
func (t Text) Alignment(a Alignment) Text { t.alignment = a; return t }

// Left returns a copy of t aligned to the left.
func (t Text) Left() Text { t.alignment = AlignmentLeft; return t }

// Centered returns a copy of t centered in its area.
func (t Text) Centered() Text { t.alignment = AlignmentCenter; return t }

// Right returns a copy of t aligned to the right.
func (t Text) Right() Text { t.alignment = AlignmentRight; return t }

// GetLines returns the text's lines. The slice is the text's own; do not modify it.
func (t Text) GetLines() []Line { return t.lines }

// GetStyle returns the text's style.
func (t Text) GetStyle() Style { return t.style }

// GetAlignment returns the text's alignment, which may be AlignmentNone.
func (t Text) GetAlignment() Alignment { return t.alignment }

// Width is the number of columns of the widest line.
func (t Text) Width() int {
	w := 0
	for _, l := range t.lines {
		w = max(w, l.Width())
	}
	return w
}

// Height is the number of lines.
func (t Text) Height() int { return len(t.lines) }

// String returns the text with all styling dropped, lines joined by newlines.
func (t Text) String() string {
	parts := make([]string, len(t.lines))
	for i, l := range t.lines {
		parts[i] = l.String()
	}
	return strings.Join(parts, "\n")
}
