// Widget implementations for the text primitives.
//
// Port of the Widget impls in ratatui-core/src/text/{span,line,text}.rs
// @ ratatui-v0.30.2

package catatui

// Render draws the span at the top left of the area, clipped to its width.
func (s Span) Render(area Rect, buf *Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStringn(area.X, area.Y, s.content, area.Width, s.style)
}

// Render draws the line on the first row of the area, honouring its own
// alignment.
func (l Line) Render(area Rect, buf *Buffer) {
	l.RenderWithFallbackAlignment(area, buf, AlignmentNone)
}

// RenderWithFallbackAlignment draws the line, falling back to the given
// alignment when the line does not set one of its own. Container widgets use it
// to impose their alignment on the lines they hold.
//
// A line that is too wide is truncated on the right, so the alignment decides
// how much is skipped on the left: a right-aligned line drops its beginning, a
// centered one drops half from each side.
func (l Line) RenderWithFallbackAlignment(area Rect, buf *Buffer, parent Alignment) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	area.Height = 1

	width := l.Width()
	if width == 0 {
		return
	}
	buf.SetStyle(area, l.style)

	alignment := l.alignment
	if alignment == AlignmentNone {
		alignment = parent
	}

	areaWidth := int(area.Width)
	if width <= areaWidth {
		var indent int
		switch alignment {
		case AlignmentCenter:
			indent = (areaWidth - width) / 2
		case AlignmentRight:
			indent = areaWidth - width
		}
		renderSpans(l.spans, indentX(area, uint16(min(indent, maxU16))), buf, 0)
		return
	}

	var skip int
	switch alignment {
	case AlignmentCenter:
		skip = (width - areaWidth) / 2
	case AlignmentRight:
		skip = width - areaWidth
	}
	renderSpans(l.spans, area, buf, skip)
}

// indentX moves a rect right by n, shrinking its width to match, so that the
// right edge stays put.
func indentX(area Rect, n uint16) Rect {
	area.X = SatAdd(area.X, n)
	area.Width = SatSub(area.Width, n)
	return area
}

// renderSpans draws a run of spans left to right, skipping the first skipWidth
// columns.
func renderSpans(spans []Span, area Rect, buf *Buffer, skipWidth int) {
	for _, s := range spans {
		w := s.Width()
		if skipWidth >= w {
			// Entirely before the visible region.
			skipWidth -= w
			continue
		}
		if skipWidth > 0 {
			// Partially visible: drop the leading columns of this span. The
			// span is re-cut by grapheme so that a wide cluster straddling the
			// boundary is dropped rather than split.
			s = s.Content(TrimLeftColumns(s.content, skipWidth))
			w -= skipWidth
			skipWidth = 0
		}
		if area.IsEmpty() {
			return
		}
		s.Render(area, buf)
		area = indentX(area, uint16(min(w, maxU16)))
	}
}

// TrimLeftColumns drops the first n columns of s, working in grapheme clusters
// so that a wide cluster is never cut in half. A cluster straddling the cut is
// dropped whole, which is the same rule Buffer.SetStringn applies at the right
// edge.
func TrimLeftColumns(s string, n int) string {
	if n <= 0 {
		return s
	}
	for g := range AllGraphemes(s) {
		if n <= 0 {
			break
		}
		n -= int(g.Width)
		s = s[len(g.Symbol):]
	}
	return s
}

// Render draws the text, one line per row, honouring the text's alignment for
// any line that does not set its own.
func (t Text) Render(area Rect, buf *Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, t.style)
	for i, line := range t.lines {
		if uint16(i) >= area.Height {
			break
		}
		row := Rect{X: area.X, Y: area.Y + uint16(i), Width: area.Width, Height: 1}
		line.RenderWithFallbackAlignment(row, buf, t.alignment)
	}
}

// Compile-time checks that the text primitives really are widgets.
var (
	_ Widget = Span{}
	_ Widget = Line{}
	_ Widget = Text{}
)
