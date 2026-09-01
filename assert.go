package catatui

import (
	"fmt"
	"strings"
)

// TestingT is the slice of *testing.T that the assertion helpers need.
//
// It is declared here rather than importing "testing" so that the assertions
// can live in the main package, where widgets and backends can reach them,
// without pulling the testing package's flags into a production binary.
type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
	FailNow()
}

// AssertBuffer fails the test unless got and want hold the same cells, and
// reports what differed as a side-by-side rendering plus a per-cell list.
//
// This is the Go counterpart of ratatui's assert_eq!(buf, Buffer::with_lines(..)),
// and the main way widget behaviour is pinned down:
//
//	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 2))
//	widget.Render(buf.Area, buf)
//	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
//		"┌────────┐",
//		"└────────┘",
//	))
func AssertBuffer(t TestingT, got, want *Buffer) {
	t.Helper()
	if err := BufferEqual(got, want); err != nil {
		t.Errorf("%v", err)
	}
}

// BufferEqual reports whether two buffers hold the same cells, returning nil if
// they do and an error describing every difference if they do not.
//
// Symbols are compared as ratatui compares them, so a cell with no symbol
// equals a cell holding a single space. Styles are compared in full: two cells
// showing the same character in different colors are not equal.
func BufferEqual(got, want *Buffer) error {
	if got.Area != want.Area {
		return fmt.Errorf("buffer areas differ:\n  got:  %+v\n  want: %+v\n\ngot:\n%s\n\nwant:\n%s",
			got.Area, want.Area, indent(got.String()), indent(want.String()))
	}

	var diffs []string
	for y := got.Area.Top(); y < got.Area.Bottom(); y++ {
		for x := got.Area.Left(); x < got.Area.Right(); x++ {
			g, w := *got.Get(x, y), *want.Get(x, y)
			if g.Equal(w) {
				continue
			}
			diffs = append(diffs, fmt.Sprintf("  (%d, %d): got %s, want %s",
				x, y, describeCell(g), describeCell(w)))
		}
	}
	if len(diffs) == 0 {
		return nil
	}

	const maxListed = 40
	listed := diffs
	truncated := ""
	if len(listed) > maxListed {
		listed = listed[:maxListed]
		truncated = fmt.Sprintf("\n  ... and %d more", len(diffs)-maxListed)
	}
	return fmt.Errorf("buffers differ in %d cell(s):\n\ngot:\n%s\n\nwant:\n%s\n\ndifferences:\n%s%s",
		len(diffs), indent(got.String()), indent(want.String()),
		strings.Join(listed, "\n"), truncated)
}

// describeCell renders a cell as its symbol plus only the styling that is set,
// so that failure output stays readable when most cells are unstyled.
func describeCell(c Cell) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%q", c.GetSymbol())
	var attrs []string
	if fg := orReset(c.Fg); !fg.IsReset() {
		attrs = append(attrs, "fg="+fg.String())
	}
	if bg := orReset(c.Bg); !bg.IsReset() {
		attrs = append(attrs, "bg="+bg.String())
	}
	if uc := orReset(c.UnderlineColor); !uc.IsReset() {
		attrs = append(attrs, "underline="+uc.String())
	}
	if !c.Modifier.IsEmpty() {
		attrs = append(attrs, "mod="+c.Modifier.String())
	}
	if c.DiffOption != CellDiffNone {
		attrs = append(attrs, fmt.Sprintf("diff=%d", c.DiffOption.kind))
	}
	if len(attrs) > 0 {
		b.WriteString(" [" + strings.Join(attrs, " ") + "]")
	}
	return b.String()
}

// indent frames each row so that leading and trailing spaces stay visible in
// test output, which matters because most buffer bugs are off-by-one padding.
func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  |" + l + "|"
	}
	return strings.Join(lines, "\n")
}
