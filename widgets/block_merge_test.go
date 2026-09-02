// Tests ported from ratatui-widgets/src/block.rs @ ratatui-v0.30.2, including
// the golden files under ratatui-widgets/tests/block, copied verbatim into
// testdata/block.

package widgets

import (
	"os"
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// mergeStrategies is every strategy, used by the tests that must come out the
// same whichever one is in force.
var mergeStrategies = []symbols.MergeStrategy{
	symbols.MergeReplace,
	symbols.MergeExact,
	symbols.MergeFuzzy,
}

func TestBlockRenderPartialBorders(t *testing.T) {
	cases := []struct {
		name    string
		borders Borders
		want    []string
	}{
		{"all", BordersTop | BordersLeft | BordersRight | BordersBottom, []string{
			"┌────────┐",
			"│        │",
			"└────────┘",
		}},
		{"top left", BordersTop | BordersLeft, []string{
			"┌─────────",
			"│         ",
			"│         ",
		}},
		{"top right", BordersTop | BordersRight, []string{
			"─────────┐",
			"         │",
			"         │",
		}},
		{"bottom left", BordersBottom | BordersLeft, []string{
			"│         ",
			"│         ",
			"└─────────",
		}},
		{"bottom right", BordersBottom | BordersRight, []string{
			"         │",
			"         │",
			"─────────┘",
		}},
		{"top bottom", BordersTop | BordersBottom, []string{
			"──────────",
			"          ",
			"──────────",
		}},
		{"left right", BordersLeft | BordersRight, []string{
			"│        │",
			"│        │",
			"│        │",
		}},
	}

	for _, strategy := range mergeStrategies {
		for _, c := range cases {
			t.Run(strategy.String()+"/"+c.name, func(t *testing.T) {
				area := catatui.NewRect(0, 0, 10, 3)
				buf := catatui.NewBuffer(area)
				NewBlock().Borders(c.borders).MergeBorders(strategy).Render(area, buf)
				catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
			})
		}
	}
}

func TestBlockMergedTitlesBottomFirst(t *testing.T) {
	cases := []struct {
		strategy symbols.MergeStrategy
		want     []string
	}{
		{symbols.MergeReplace, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┗━━━━━━━━━━━┛",
			"│           │",
			"└───────────┘",
		}},
		{symbols.MergeExact, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┡block btm━━┩",
			"│           │",
			"└───────────┘",
		}},
		{symbols.MergeFuzzy, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┡block btm━━┩",
			"│           │",
			"└───────────┘",
		}},
	}

	for _, c := range cases {
		t.Run(c.strategy.String(), func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 13, 5))
			Bordered().Title("block btm").
				Render(catatui.NewRect(0, 2, 13, 3), buf)
			Bordered().Title("block top").
				BorderType(BorderThick).
				MergeBorders(c.strategy).
				Render(catatui.NewRect(0, 0, 13, 3), buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

func TestBlockMergedTitlesTopFirst(t *testing.T) {
	cases := []struct {
		strategy symbols.MergeStrategy
		want     []string
	}{
		{symbols.MergeReplace, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┌block btm──┐",
			"│           │",
			"└───────────┘",
		}},
		{symbols.MergeExact, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┞block btm──┦",
			"│           │",
			"└───────────┘",
		}},
		{symbols.MergeFuzzy, []string{
			"┏block top━━┓",
			"┃           ┃",
			"┞block btm──┦",
			"│           │",
			"└───────────┘",
		}},
	}

	for _, c := range cases {
		t.Run(c.strategy.String(), func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 13, 5))
			Bordered().Title("block top").
				BorderType(BorderThick).
				Render(catatui.NewRect(0, 0, 13, 3), buf)
			Bordered().Title("block btm").
				MergeBorders(c.strategy).
				Render(catatui.NewRect(0, 2, 13, 3), buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}

// TestBlockRenderMergedBorders draws every pair of border types in the four
// ways two blocks can meet — at a corner, overlapping, along a vertical edge
// and along a horizontal edge — and compares the whole 43x1000 buffer against
// the file ratatui's own test produced.
func TestBlockRenderMergedBorders(t *testing.T) {
	borderTypes := []BorderType{
		BorderPlain,
		BorderRounded,
		BorderThick,
		BorderDouble,
		BorderLightDoubleDashed,
		BorderHeavyDoubleDashed,
		BorderLightTripleDashed,
		BorderHeavyTripleDashed,
		BorderLightQuadrupleDashed,
		BorderHeavyQuadrupleDashed,
	}
	rects := [][2]catatui.Rect{
		// touching at corners
		{catatui.NewRect(0, 0, 5, 5), catatui.NewRect(4, 4, 5, 5)},
		// overlapping
		{catatui.NewRect(10, 0, 5, 5), catatui.NewRect(12, 2, 5, 5)},
		// touching vertical edges
		{catatui.NewRect(18, 0, 5, 5), catatui.NewRect(22, 0, 5, 5)},
		// touching horizontal edges
		{catatui.NewRect(28, 0, 5, 5), catatui.NewRect(28, 4, 5, 5)},
	}

	cases := []struct {
		strategy symbols.MergeStrategy
		file     string
	}{
		{symbols.MergeReplace, "testdata/block/merge_replace.txt"},
		{symbols.MergeExact, "testdata/block/merge_exact.txt"},
		{symbols.MergeFuzzy, "testdata/block/merge_fuzzy.txt"},
	}

	for _, c := range cases {
		t.Run(c.strategy.String(), func(t *testing.T) {
			buf := catatui.NewBuffer(catatui.NewRect(0, 0, 43, 1000))

			var y int32
			for _, first := range borderTypes {
				for _, second := range borderTypes {
					title := catatui.LineFromString(first.String() + " + " + second.String())
					title.Render(catatui.NewRect(0, 0, 43, 1).Offset(catatui.Offset{Y: y}), buf)
					y++

					for _, pair := range rects {
						Bordered().BorderType(first).MergeBorders(c.strategy).
							Render(pair[0].Offset(catatui.Offset{Y: y}), buf)
						Bordered().BorderType(second).MergeBorders(c.strategy).
							Render(pair[1].Offset(catatui.Offset{Y: y}), buf)
					}
					y += 9
				}
			}

			assertBufferLines(t, buf, readGoldenLines(t, c.file))
		})
	}
}

// readGoldenLines reads a golden file as one string per buffer row, dropping
// the trailing newline. The files are ratatui's, so they may carry CRLF.
func readGoldenLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	return strings.Split(text, "\n")
}

// assertBufferLines compares a buffer against the text it should render as,
// reporting the first few rows that differ rather than dumping the whole
// buffer, which for the merge goldens is a thousand rows.
func assertBufferLines(t *testing.T, buf *catatui.Buffer, want []string) {
	t.Helper()

	got := bufferRows(buf)
	if len(got) != len(want) {
		t.Fatalf("buffer has %d rows, want %d", len(got), len(want))
	}

	const maxReported = 5
	reported := 0
	for y := range want {
		// The goldens have their trailing spaces trimmed, the buffer does
		// not, so compare with both trimmed.
		if strings.TrimRight(got[y], " ") == strings.TrimRight(want[y], " ") {
			continue
		}
		reported++
		if reported > maxReported {
			continue
		}
		t.Errorf("row %d (%s):\n got: %q\nwant: %q", y, goldenSection(want, y), got[y], want[y])
	}
	if reported > maxReported {
		t.Errorf("... and %d more rows differ", reported-maxReported)
	}
}

// bufferRows renders a buffer as one string per row.
func bufferRows(buf *catatui.Buffer) []string {
	rows := make([]string, 0, buf.Area.Height)
	for y := buf.Area.Top(); y < buf.Area.Bottom(); y++ {
		var row strings.Builder
		for x := buf.Area.Left(); x < buf.Area.Right(); x++ {
			row.WriteString(buf.Get(x, y).GetSymbol())
		}
		rows = append(rows, row.String())
	}
	return rows
}

// goldenSection names the pair of border types a row belongs to, which is the
// title row that starts every ten-row block of the golden file.
func goldenSection(want []string, y int) string {
	return strings.TrimSpace(want[y-y%10])
}

func TestBorderTypeNames(t *testing.T) {
	cases := []struct {
		borderType BorderType
		want       string
	}{
		{BorderPlain, "Plain"},
		{BorderRounded, "Rounded"},
		{BorderDouble, "Double"},
		{BorderThick, "Thick"},
		{BorderQuadrantOutside, "QuadrantOutside"},
		{BorderQuadrantInside, "QuadrantInside"},
		{BorderLightDoubleDashed, "LightDoubleDashed"},
		{BorderHeavyDoubleDashed, "HeavyDoubleDashed"},
		{BorderLightTripleDashed, "LightTripleDashed"},
		{BorderHeavyTripleDashed, "HeavyTripleDashed"},
		{BorderLightQuadrupleDashed, "LightQuadrupleDashed"},
		{BorderHeavyQuadrupleDashed, "HeavyQuadrupleDashed"},
	}
	for _, c := range cases {
		if got := c.borderType.String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
		got, err := ParseBorderType(c.want)
		if err != nil || got != c.borderType {
			t.Errorf("ParseBorderType(%q) = %v, %v; want %v, nil", c.want, got, err, c.borderType)
		}
	}
	if _, err := ParseBorderType(""); err == nil {
		t.Error("ParseBorderType(\"\") = nil error, want an error")
	}
}

func TestBlockDashedBorderTypes(t *testing.T) {
	cases := []struct {
		borderType BorderType
		want       []string
	}{
		{BorderLightDoubleDashed, []string{"┌╌╌╌┐", "╎   ╎", "└╌╌╌┘"}},
		{BorderHeavyDoubleDashed, []string{"┏╍╍╍┓", "╏   ╏", "┗╍╍╍┛"}},
		{BorderLightTripleDashed, []string{"┌┄┄┄┐", "┆   ┆", "└┄┄┄┘"}},
		{BorderHeavyTripleDashed, []string{"┏┅┅┅┓", "┇   ┇", "┗┅┅┅┛"}},
		{BorderLightQuadrupleDashed, []string{"┌┈┈┈┐", "┊   ┊", "└┈┈┈┘"}},
		{BorderHeavyQuadrupleDashed, []string{"┏┉┉┉┓", "┋   ┋", "┗┉┉┉┛"}},
	}
	for _, c := range cases {
		t.Run(c.borderType.String(), func(t *testing.T) {
			area := catatui.NewRect(0, 0, 5, 3)
			buf := catatui.NewBuffer(area)
			Bordered().BorderType(c.borderType).Render(area, buf)
			catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(c.want...))
		})
	}
}
