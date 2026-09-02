// Tests ported from ratatui-widgets/src/clear.rs @ ratatui-v0.30.2

package widgets

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// repeatLines returns n copies of s, which is ratatui's ["..."; n].
func repeatLines(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func TestClearRender(t *testing.T) {
	buf := catatui.NewBufferWithStrings(repeatLines("xxxxxxxxxxxxxxx", 7)...)
	Clear{}.Render(catatui.NewRect(1, 2, 3, 4), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(
		"xxxxxxxxxxxxxxx",
		"xxxxxxxxxxxxxxx",
		"x   xxxxxxxxxxx",
		"x   xxxxxxxxxxx",
		"x   xxxxxxxxxxx",
		"x   xxxxxxxxxxx",
		"xxxxxxxxxxxxxxx",
	))
}

func TestClearRenderPartiallyOutOfBounds(t *testing.T) {
	buf := catatui.NewBufferWithStrings(repeatLines("xxxxxxxxxxxxxxx", 7)...)
	Clear{}.Render(catatui.NewRect(2, 0, 100, 100), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(repeatLines("xx             ", 7)...))
}

func TestClearRenderFullyOutOfBounds(t *testing.T) {
	buf := catatui.NewBufferWithStrings(repeatLines("xxxxxxxxxxxxxxx", 7)...)
	Clear{}.Render(catatui.NewRect(100, 0, 100, 100), buf)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(repeatLines("xxxxxxxxxxxxxxx", 7)...))
}
