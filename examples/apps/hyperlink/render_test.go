package main

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen.
// Rendering outside the area given panics in catatui, so this is what keeps the
// example honest when the library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for _, size := range sizes {
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		if err := terminal.Draw(render); err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
	}
}

// link is the widget the example draws, built here so the tests can render it
// on its own.
func link() hyperlink {
	return hyperlink{
		text: catatui.NewLine(
			catatui.NewSpan("Example "),
			catatui.NewStyledSpan("hyperlink", catatui.NewStyle().Fg(catatui.ColorBlue)),
		),
		url: "https://example.com",
	}
}

// TestEachSpanIsOneLinkedCell checks the escape sequence lands in the first
// cell of each span, wrapping that span's text and nothing else, and that the
// cell claims the columns the span draws.
func TestEachSpanIsOneLinkedCell(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 30, 1))
	link().Render(buf.Area, buf)

	for _, want := range []struct {
		x     uint16
		text  string
		width uint16
	}{
		{0, "Example ", 8},
		{8, "hyperlink", 9},
	} {
		cell := buf.Get(want.x, 0)
		if got := cell.GetSymbol(); got != osc8("https://example.com", want.text) {
			t.Errorf("cell %d holds %q", want.x, got)
		}
		if got, ok := cell.DiffOption.ForcedWidth(); !ok || got != want.width {
			t.Errorf("cell %d is %d columns wide (forced: %v), want %d", want.x, got, ok, want.width)
		}
	}

	if got := buf.Get(8, 0).Fg; got != catatui.ColorBlue {
		t.Errorf("the linked word is %v, want blue — one cell per span is what keeps its style", got)
	}
}

// TestTheLinkStopsAtTheEdge checks a narrow area truncates the text inside the
// escape too. A link carrying more text than the cell claims would overrun the
// line, since the terminal prints the payload and the diff does not.
func TestTheLinkStopsAtTheEdge(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 12, 1))
	link().Render(buf.Area, buf)

	cell := buf.Get(8, 0)
	width, ok := cell.DiffOption.ForcedWidth()
	if !ok || width != 4 {
		t.Fatalf("the clipped span claims %d columns (forced: %v), want 4", width, ok)
	}
	text := strings.TrimSuffix(strings.SplitN(cell.GetSymbol(), "\a", 2)[1], "\x1b]8;;\a")
	if text != "hype" {
		t.Errorf("the link carries %q, want %q", text, "hype")
	}
	if got := catatui.StringWidth(text); uint16(got) != width {
		t.Errorf("the link carries %d columns of text but claims %d", got, width)
	}
}

// TestWideGraphemesKeepTheirColumns checks the width claimed matches the
// columns drawn when the text is not one column per character. A cell holding a
// wide grapheme covers the column after it, and that column is not text.
func TestWideGraphemesKeepTheirColumns(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 10, 1))
	h := hyperlink{text: catatui.NewLine(catatui.NewSpan("日本語")), url: "https://example.com"}
	h.Render(buf.Area, buf)

	cell := buf.Get(0, 0)
	width, ok := cell.DiffOption.ForcedWidth()
	if !ok || width != 6 {
		t.Fatalf("the span claims %d columns (forced: %v), want 6", width, ok)
	}
	text := strings.TrimSuffix(strings.SplitN(cell.GetSymbol(), "\a", 2)[1], "\x1b]8;;\a")
	if text != "日本語" {
		t.Errorf("the link carries %q, want the three characters with no blanks between them", text)
	}
}
