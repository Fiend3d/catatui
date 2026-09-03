// Command paragraph shows catatui's Paragraph widget: alignment and wrapping.
//
//	go run ./examples/widgets/paragraph
//
// Press any key to quit.
//
// Port of ratatui-widgets/examples/paragraph.rs @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run draws the UI and waits for a key.
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

// render draws a centred paragraph and a wrapped one.
func render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(f.Area())
	cols := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Percentage(50)).
		Spacing(catatui.Space(1)).Split(rows[1])

	f.RenderWidget(title("Paragraph Widget"), rows[0])
	renderCenteredParagraph(f, cols[0])
	renderWrappedParagraph(f, cols[1])
}

// renderCenteredParagraph draws plain text centred in its area.
func renderCenteredParagraph(f *catatui.Frame, area catatui.Rect) {
	text := "Centered text\nwith multiple lines.\nCheck out the recipe!"
	paragraph := widgets.NewParagraph(text).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		Alignment(catatui.AlignmentCenter)
	f.RenderWidget(paragraph, area)
}

// renderWrappedParagraph draws styled lines wrapped to the area's width.
func renderWrappedParagraph(f *catatui.Frame, area catatui.Rect) {
	paragraph := widgets.NewParagraphFromText(recipe(area)).
		Style(catatui.NewStyle().Fg(catatui.ColorWhite)).
		Scroll(catatui.Position{X: 0, Y: 0}).
		Wrap(widgets.Wrap{Trim: true})
	f.RenderWidget(paragraph, area)
}

// recipe returns the styled text of the paragraph, long enough to wrap.
func recipe(area catatui.Rect) catatui.Text {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)
	italic := catatui.NewStyle().AddModifier(catatui.ModifierItalic)

	shortLine := "Slice, layer, and bake the vegetables. "
	longLine := strings.Repeat(shortLine, int(area.Width)/len(shortLine)+2)

	return catatui.NewText(
		catatui.LineFromString("Recipe: Ratatouille"),
		catatui.LineFromStyledString("Ingredients:", bold),
		catatui.NewLine(
			catatui.NewSpan("Bell Peppers"),
			catatui.NewStyledSpan(", Eggplant", italic),
			catatui.NewStyledSpan(", Tomatoes", bold),
			catatui.NewSpan(", Onion"),
		),
		catatui.NewLine(
			catatui.NewStyledSpan("Secret Ingredient: ",
				catatui.NewStyle().AddModifier(catatui.ModifierUnderlined)),
			catatui.NewStyledSpan(masked("herbs de Provence", '*'),
				catatui.NewStyle().Fg(catatui.ColorRed)),
		),
		catatui.LineFromStyledString("Instructions:",
			bold.Fg(catatui.ColorYellow)),
		catatui.LineFromStyledString(longLine,
			italic.Fg(catatui.ColorGreen)),
	)
}

// masked hides a string behind a mask character, one per grapheme cluster, the
// way ratatui's Masked does.
func masked(s string, mask rune) string {
	width := 0
	for range catatui.AllGraphemes(s) {
		width++
	}
	return strings.Repeat(string(mask), width)
}

// title returns the heading every example draws along its top row.
func title(name string) catatui.Line {
	return catatui.NewLine(
		catatui.NewStyledSpan(name, catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(" (press any key to quit)"),
	).Centered()
}
