// Tests ported from ratatui-core/src/buffer/buffer.rs @ ratatui-v0.30.2

package catatui

import "testing"

func TestBufferTranslatesToAndFromCoordinates(t *testing.T) {
	area := NewRect(200, 100, 50, 80)
	buf := NewBuffer(area)

	// Top-left, top-right, bottom-left and bottom-right corners. The buffer is
	// 50 wide and 80 tall, so a row step is 50.
	for _, c := range []struct {
		x, y  uint16
		index int
	}{
		{200, 100, 0},
		{249, 100, 49},
		{200, 179, 79 * 50},
		{249, 179, 79*50 + 49},
	} {
		if got := buf.IndexOf(c.x, c.y); got != c.index {
			t.Errorf("IndexOf(%d, %d) = %d, want %d", c.x, c.y, got, c.index)
		}
		gx, gy := buf.PosOf(c.index)
		if gx != c.x || gy != c.y {
			t.Errorf("PosOf(%d) = (%d, %d), want (%d, %d)", c.index, gx, gy, c.x, c.y)
		}
	}
}

func TestBufferIndexOfPanicsOutOfBounds(t *testing.T) {
	buf := NewBuffer(NewRect(10, 10, 10, 10))
	cases := []struct {
		name string
		x, y uint16
	}{
		{"left", 9, 10},
		{"top", 10, 9},
		{"right", 20, 10},
		{"bottom", 10, 20},
	}
	for _, c := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: IndexOf(%d, %d) should panic", c.name, c.x, c.y)
				}
			}()
			buf.IndexOf(c.x, c.y)
		}()
	}
}

func TestBufferCellReturnsNilOutOfBounds(t *testing.T) {
	buf := NewBuffer(NewRect(10, 10, 10, 10))
	if buf.CellAt(15, 15) == nil {
		t.Error("in-bounds cell should not be nil")
	}
	for _, p := range []Position{{9, 10}, {10, 9}, {20, 10}, {10, 20}} {
		if buf.Cell(p) != nil {
			t.Errorf("Cell(%+v) should be nil, outside the buffer", p)
		}
	}
}

func TestBufferSetString(t *testing.T) {
	area := NewRect(0, 0, 5, 1)
	buf := NewBuffer(area)

	// A zero max width draws nothing.
	buf.SetStringn(0, 0, "aaa", 0, NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("     "))

	buf.SetString(0, 0, "aaa", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("aaa  "))

	// The max width limits the draw.
	buf.SetStringn(0, 0, "bbbbbbbbbbbbbb", 4, NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("bbbb "))

	buf.SetString(0, 0, "12345", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("12345"))

	// The buffer edge truncates the draw.
	buf.SetString(0, 0, "123456", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("12345"))

	// Multiple lines.
	buf = NewBuffer(NewRect(0, 0, 5, 2))
	buf.SetString(0, 0, "12345", NewStyle())
	buf.SetString(0, 1, "67890", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("12345", "67890"))
}

// TestBufferSetStringMultiWidthOverwrite covers a wide character landing on top
// of narrow ones: it takes two columns and the third keeps its old content.
func TestBufferSetStringMultiWidthOverwrite(t *testing.T) {
	buf := NewBuffer(NewRect(0, 0, 5, 1))
	buf.SetString(0, 0, "aaaaa", NewStyle())
	buf.SetString(0, 0, "称号", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("称号a"))
}

func TestBufferSetStringZeroWidth(t *testing.T) {
	if got := cellWidth("​"); got != 0 {
		t.Fatalf("zero-width space should measure 0, got %d", got)
	}
	buf := NewBuffer(NewRect(0, 0, 1, 1))

	// A leading zero-width grapheme is skipped rather than consuming the cell.
	buf.SetStringn(0, 0, "​a", 1, NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("a"))

	// So is a trailing one.
	buf.SetStringn(0, 0, "a​", 1, NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("a"))
}

// TestBufferSetStringDoubleWidth is the rule the whole port hinges on: a wide
// grapheme that does not fit is dropped, not clipped, and drawing stops there.
func TestBufferSetStringDoubleWidth(t *testing.T) {
	buf := NewBuffer(NewRect(0, 0, 5, 1))
	buf.SetString(0, 0, "コン", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("コン "))

	// Only one column is left, so ピ is dropped entirely rather than drawn
	// as half a glyph.
	buf.SetString(0, 0, "コンピ", NewStyle())
	AssertBuffer(t, buf, NewBufferWithStrings("コン "))
}

func TestBufferSetStringHalfwidthKatakanaWithDakuten(t *testing.T) {
	area := NewRect(0, 0, 5, 1)

	// Halfwidth katakana alone occupies one column.
	buf := NewBuffer(area)
	buf.SetString(0, 0, "ｶ", NewStyle())
	if got := buf.Get(0, 0).GetSymbol(); got != "ｶ" {
		t.Errorf("cell 0 = %q, want %q", got, "ｶ")
	}
	if got := buf.Get(1, 0).GetSymbol(); got != " " {
		t.Errorf("cell 1 = %q, want a space", got)
	}

	// With a non-combining dakuten the cluster occupies two columns: the whole
	// cluster goes in the first cell and the second is blanked.
	buf = NewBuffer(area)
	buf.SetString(0, 0, "ｶﾞ", NewStyle())
	if got := buf.Get(0, 0).GetSymbol(); got != "ｶﾞ" {
		t.Errorf("cell 0 = %q, want %q", got, "ｶﾞ")
	}
	if got := buf.Get(1, 0).GetSymbol(); got != " " {
		t.Errorf("cell 1 = %q, want a blanked space", got)
	}
	if got := buf.Get(2, 0).GetSymbol(); got != " " {
		t.Errorf("cell 2 = %q, want a space", got)
	}
}

// TestBufferContinuationCellsAreBlankNotStyled pins the detail nezumi's row
// renderer has to compensate for: the columns after a wide grapheme are reset,
// so they do not carry the style the grapheme was drawn in.
func TestBufferContinuationCellsAreBlankNotStyled(t *testing.T) {
	buf := NewBuffer(NewRect(0, 0, 4, 1))
	style := NewStyle().Bg(ColorRed).Fg(ColorBlue)
	buf.SetString(0, 0, "あ", style)

	if got := buf.Get(0, 0).GetStyle(); got.GetBg() != ColorRed {
		t.Errorf("the grapheme's own cell should be styled, got %+v", got)
	}
	cont := buf.Get(1, 0)
	if cont.GetSymbol() != " " {
		t.Errorf("continuation cell symbol = %q, want a space", cont.GetSymbol())
	}
	if cont.Bg != ColorReset {
		t.Errorf("continuation cell bg = %v, want ColorReset (it is reset, not styled)", cont.Bg)
	}
}

func TestBufferSetLine(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"empty", "", "     "},
		{"one", "1", "1    "},
		{"full", "12345", "12345"},
		{"overflow", "123456", "12345"},
	}
	for _, c := range cases {
		buf := NewBuffer(NewRect(0, 0, 5, 1))
		buf.SetLine(0, 0, LineFromString(c.content), 5)
		if got := buf.String(); got != c.want {
			t.Errorf("%s: SetLine(%q) drew %q, want %q", c.name, c.content, got, c.want)
		}
	}
}

// TestBufferSetLinePatchesSpanStyleOntoLineStyle checks the layering order:
// the line's style sits underneath each span's own.
func TestBufferSetLinePatchesSpanStyleOntoLineStyle(t *testing.T) {
	buf := NewBuffer(NewRect(0, 0, 4, 1))
	line := NewLine(
		NewSpan("ab"),
		NewStyledSpan("cd", NewStyle().Fg(ColorGreen)),
	).Style(NewStyle().Fg(ColorRed).Bg(ColorBlack))
	buf.SetLine(0, 0, line, 4)

	if got := buf.Get(0, 0).Fg; got != ColorRed {
		t.Errorf("unstyled span should take the line's fg, got %v", got)
	}
	if got := buf.Get(2, 0).Fg; got != ColorGreen {
		t.Errorf("styled span's own fg should win, got %v", got)
	}
	if got := buf.Get(2, 0).Bg; got != ColorBlack {
		t.Errorf("the line's bg should show through, got %v", got)
	}
}

func TestBufferSetStyle(t *testing.T) {
	buf := NewBufferWithStrings("aaaaa", "bbbbb", "ccccc")
	buf.SetStyle(NewRect(0, 1, 5, 1), NewStyle().Fg(ColorRed))

	for x := range uint16(5) {
		if got := buf.Get(x, 1).Fg; got != ColorRed {
			t.Errorf("cell (%d, 1) fg = %v, want ColorRed", x, got)
		}
		if got := buf.Get(x, 0).Fg; got != ColorReset {
			t.Errorf("cell (%d, 0) fg = %v, want it left alone", x, got)
		}
	}
	// Styling must not change the symbols.
	if got, want := buf.String(), "aaaaa\nbbbbb\nccccc"; got != want {
		t.Errorf("SetStyle changed the content: %q", got)
	}
}

func TestBufferSetStyleDoesNotPanicWhenOutOfArea(t *testing.T) {
	buf := NewBufferWithStrings("aaaaa", "bbbbb", "ccccc")
	// Entirely outside, and partially overlapping: both must be clipped, not panic.
	buf.SetStyle(NewRect(15, 15, 5, 5), NewStyle().Fg(ColorRed))
	buf.SetStyle(NewRect(3, 2, 10, 10), NewStyle().Fg(ColorBlue))
	if got := buf.Get(4, 2).Fg; got != ColorBlue {
		t.Errorf("the overlapping part should be styled, got %v", got)
	}
}

func TestBufferWithLines(t *testing.T) {
	buf := NewBufferWithStrings("┌────────┐", "│ Hello  │", "└────────┘")
	if got, want := buf.Area, NewRect(0, 0, 10, 3); got != want {
		t.Errorf("area = %+v, want %+v", got, want)
	}
	if got, want := buf.String(), "┌────────┐\n│ Hello  │\n└────────┘"; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

// TestBufferWithLinesUsesWidestLine checks that shorter lines are padded rather
// than leaving the buffer ragged.
func TestBufferWithLinesUsesWidestLine(t *testing.T) {
	buf := NewBufferWithStrings("ab", "abcd", "a")
	if got, want := buf.Area, NewRect(0, 0, 4, 3); got != want {
		t.Errorf("area = %+v, want %+v", got, want)
	}
	if got, want := buf.String(), "ab  \nabcd\na   "; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestBufferResize(t *testing.T) {
	buf := NewBufferWithStrings("aaa", "bbb")
	buf.Resize(NewRect(0, 0, 5, 5))
	if got, want := len(buf.Content), 25; got != want {
		t.Errorf("content length = %d, want %d", got, want)
	}
	buf.Resize(NewRect(0, 0, 2, 2))
	if got, want := len(buf.Content), 4; got != want {
		t.Errorf("content length after shrink = %d, want %d", got, want)
	}
	if got, want := buf.Area, NewRect(0, 0, 2, 2); got != want {
		t.Errorf("area = %+v, want %+v", got, want)
	}
}

func TestBufferReset(t *testing.T) {
	buf := NewBufferWithStrings("aaa")
	buf.SetStyle(buf.Area, NewStyle().Fg(ColorRed))
	buf.Reset()
	for i, c := range buf.Content {
		if !c.Equal(EmptyCell()) {
			t.Errorf("cell %d should be empty after Reset, got %s", i, describeCell(c))
		}
	}
}

func TestBufferMerge(t *testing.T) {
	a := NewBufferFilled(NewRect(0, 0, 2, 2), NewCell("1"))
	b := NewBufferFilled(NewRect(0, 2, 2, 2), NewCell("2"))
	a.Merge(b)
	if got, want := a.Area, NewRect(0, 0, 2, 4); got != want {
		t.Errorf("merged area = %+v, want %+v", got, want)
	}
	if got, want := a.String(), "11\n11\n22\n22"; got != want {
		t.Errorf("merged content = %q, want %q", got, want)
	}
}

// TestBufferMergeOffsetAreas covers the reindexing path, where cells have to
// move because the union is wider than the original.
func TestBufferMergeOffsetAreas(t *testing.T) {
	a := NewBufferFilled(NewRect(2, 2, 2, 2), NewCell("1"))
	b := NewBufferFilled(NewRect(0, 0, 2, 2), NewCell("2"))
	a.Merge(b)
	if got, want := a.Area, NewRect(0, 0, 4, 4); got != want {
		t.Errorf("merged area = %+v, want %+v", got, want)
	}
	if got, want := a.String(), "22  \n22  \n  11\n  11"; got != want {
		t.Errorf("merged content = %q, want %q", got, want)
	}
}

// TestAssertBufferDetectsStyleDifference guards the harness itself: two buffers
// with identical text but different styling must not compare equal, or every
// widget style test would silently pass.
func TestAssertBufferDetectsStyleDifference(t *testing.T) {
	a := NewBufferWithStrings("ab")
	b := NewBufferWithStrings("ab")
	if err := BufferEqual(a, b); err != nil {
		t.Fatalf("identical buffers should be equal: %v", err)
	}
	b.SetStyle(b.Area, NewStyle().Fg(ColorRed))
	if err := BufferEqual(a, b); err == nil {
		t.Error("buffers differing only in style should not be equal")
	}
}

// TestDiffDoesNotBlankBehindAFreshEmojiSequence is the rainbow-flag case.
//
// ratatui-v0.30.2 clears the trailing column of an emoji presentation sequence
// on every frame that draws one. That write lands on the column the terminal
// has just shaped the second half of the glyph into, which breaks the ligature:
// 🏳️‍🌈 falls back to a bare 🏳 until some later frame rewrites the leading
// cell alone and it repairs itself. A wide cluster covers its own columns, so
// there is nothing to write there.
func TestDiffDoesNotBlankBehindAFreshEmojiSequence(t *testing.T) {
	const rainbowFlag = "\U0001F3F3\uFE0F\u200D\U0001F308"

	prev := NewBuffer(NewRect(0, 0, 4, 1))
	prev.SetString(0, 0, "ab", NewStyle())
	next := NewBuffer(NewRect(0, 0, 4, 1))
	next.SetString(0, 0, rainbowFlag, NewStyle())

	for _, pc := range prev.Diff(next) {
		if pc.X == 1 {
			t.Errorf("column 1 was written (%q); it is covered by the glyph at column 0",
				pc.Cell.GetSymbol())
		}
	}
}

// TestDiffClearsBehindAReplacedEmojiSequence is what the clearing was for:
// terminals do not reliably drop the trailing column of an emoji presentation
// sequence when it is replaced, and unlike plain CJK that holds even on a
// default background.
func TestDiffClearsBehindAReplacedEmojiSequence(t *testing.T) {
	const rainbowFlag = "\U0001F3F3\uFE0F\u200D\U0001F308"

	prev := NewBuffer(NewRect(0, 0, 4, 1))
	prev.SetString(0, 0, rainbowFlag, NewStyle())
	next := NewBuffer(NewRect(0, 0, 4, 1))
	next.SetString(0, 0, "a", NewStyle())

	// Column 1 holds a blank in both buffers, so only the forced rewrite puts
	// it on the wire.
	var cleared bool
	for _, pc := range prev.Diff(next) {
		if pc.X == 1 {
			cleared = true
			if got := pc.Cell.GetSymbol(); got != " " {
				t.Errorf("column 1 = %q, want a blank", got)
			}
		}
	}
	if !cleared {
		t.Error("the column the flag used to cover was never rewritten")
	}
}
