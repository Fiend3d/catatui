// Command hyperlink draws a line of text that is a clickable link, using the
// OSC 8 escape sequence.
//
//	go run ./examples/apps/hyperlink
//
// Any key quits. Ctrl-click or click the underlined word, depending on the
// terminal; one that does not know OSC 8 shows the text and nothing else.
//
// OSC 8 wraps text in an escape sequence carrying a URL:
//
//	ESC ] 8 ; ; https://example.com BEL  the text  ESC ] 8 ; ; BEL
//
// The escape is not a character on screen, so a buffer that measured the cell
// by its symbol would think it fifty columns wide and blank the rest of the
// line. CellForcedWidth is the answer: it tells the diff how many columns the
// cell really occupies, and the columns it covers are left alone.
//
// This is where the port departs from ratatui, whose example predates
// ForcedWidth and works around the same problem by cutting the link into
// two-character pieces, one per cell, because two is what its width calculation
// came to. One cell per span is both simpler and correct for any text.
//
// Port of examples/apps/hyperlink @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

func render(f *catatui.Frame) {
	link := hyperlink{
		text: catatui.NewLine(
			catatui.NewSpan("Example "),
			catatui.NewStyledSpan("hyperlink", catatui.NewStyle().Fg(catatui.ColorBlue)),
		),
		url: "https://example.com",
	}
	f.RenderWidget(link, f.Area())
}

// hyperlink is a line of text that links to a URL.
type hyperlink struct {
	text catatui.Line
	url  string
}

// Render draws the text one span per cell, each span wrapped in its own OSC 8
// sequence and given the width it draws.
//
// Keeping the spans apart is what keeps their styles: the escape carries no
// colour of its own, so everything inside one cell comes out in that cell's
// style, and merging the whole line into a single cell would flatten the blue
// word into the colour of the first one.
func (h hyperlink) Render(area catatui.Rect, buf *catatui.Buffer) {
	if area.IsEmpty() {
		return
	}

	x := area.X
	for _, span := range h.text.GetSpans() {
		if x >= area.Right() {
			break
		}
		// The line's own style sits under each span's, as Line.Render layers
		// them.
		span = span.Style(h.text.GetStyle().Patch(span.GetStyle()))

		// Draw the span normally first. That clips it to the area and settles
		// what actually fits, which is what the escape has to agree with.
		start := x
		x, _ = buf.SetSpan(start, area.Y, span, area.Right()-start)
		if x == start {
			continue
		}
		buf.Get(start, area.Y).
			SetSymbol(osc8(h.url, drawnText(buf, start, x, area.Y))).
			SetDiffOption(catatui.CellForcedWidth(x - start))
	}
}

// drawnText reads back the characters written between two columns, so that the
// escape sequence carries exactly what the buffer would otherwise have shown.
// Reading them back rather than reusing the span is what accounts for a
// grapheme too wide for the space left, which the buffer drops.
func drawnText(buf *catatui.Buffer, from, to, y uint16) string {
	var b strings.Builder
	for x := from; x < to; {
		cell := buf.Get(x, y)
		b.WriteString(cell.GetSymbol())
		// Step over the columns a wide grapheme covers. The buffer blanks them,
		// and they are not part of the text.
		x += max(cell.Width(), 1)
	}
	return b.String()
}

// osc8 wraps text in a hyperlink to url. BEL (\a) terminates the sequence;
// ST (ESC \) is the other spelling, and terminals accept both.
func osc8(url, text string) string {
	return "\x1b]8;;" + url + "\a" + text + "\x1b]8;;\a"
}
