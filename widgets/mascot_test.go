// Tests ported from ratatui-widgets/src/mascot.rs @ ratatui-v0.30.2

package widgets

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

func TestRatatuiMascotNew(t *testing.T) {
	if got := NewRatatuiMascot().GetEye(); got != MascotEyeDefault {
		t.Errorf("eye = %v, want default", got)
	}
}

func TestRatatuiMascotSetEyeColor(t *testing.T) {
	mascot := NewRatatuiMascot().SetEye(MascotEyeRed)
	buf := renderToBuffer(mascot, 32, 16)
	if got := mascot.GetEye(); got != MascotEyeRed {
		t.Errorf("eye = %v, want red", got)
	}
	if got := buf.Get(21, 5).Bg; got != catatui.Indexed(196) {
		t.Errorf("(21, 5) bg = %v, want indexed(196)", got)
	}
}

// bufferSymbols concatenates every cell's symbol, which is how ratatui's test
// compares the mascot: by shape only, ignoring color.
func bufferSymbols(buf *catatui.Buffer) string {
	var sb strings.Builder
	for _, c := range buf.Content {
		sb.WriteString(c.GetSymbol())
	}
	return sb.String()
}

func TestRatatuiMascotRender(t *testing.T) {
	buf := renderToBuffer(NewRatatuiMascot(), 32, 16)
	if got, want := buf.Area.AsSize(), (catatui.Size{Width: 32, Height: 16}); got != want {
		t.Errorf("area size = %+v, want %+v", got, want)
	}
	if got := buf.Get(21, 5).Bg; got != catatui.Indexed(236) {
		t.Errorf("(21, 5) bg = %v, want indexed(236)", got)
	}
	want := catatui.NewBufferWithStrings(
		"             ▄▄███              ",
		"           ▄███████             ",
		"         ▄█████████             ",
		"        ████████████            ",
		"        ▀███████████▀   ▄▄██████",
		"              ▀███▀▄█▀▀████████ ",
		"            ▄▄▄▄▀▄████████████  ",
		"           ████████████████     ",
		"           ▀███▀██████████      ",
		"         ▄▀▀▄   █████████       ",
		"       ▄▀ ▄  ▀▄▀█████████       ",
		"     ▄▀  ▀▀    ▀▄▀███████       ",
		"   ▄▀      ▄▄    ▀▄▀█████████   ",
		" ▄▀         ▀▀     ▀▄▀██▀  ███  ",
		"█                    ▀▄▀  ▄██   ",
		" ▀▄                    ▀▄▀█     ",
	)
	if got, want := bufferSymbols(buf), bufferSymbols(want); got != want {
		t.Errorf("mascot shape differs:\ngot:\n%s\nwant:\n%s", buf, want)
	}
}

func TestRatatuiMascotRenderInMinimalBuffer(t *testing.T) {
	// This should not panic, even if the buffer is too small to render the
	// mascot.
	buf := renderToBuffer(NewRatatuiMascot(), 1, 1)
	catatui.AssertBuffer(t, buf, catatui.NewBufferWithStrings(" "))
}

func TestRatatuiMascotRenderInZeroSizeBuffer(t *testing.T) {
	// This should not panic, even if the buffer has zero size.
	renderToBuffer(NewRatatuiMascot(), 0, 0)
}
