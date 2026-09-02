package catatui_test

import (
	"fmt"
	"strings"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/widgets"
)

// printBuffer writes a buffer one row per line, with the padding on the right
// trimmed so that the expected output of an example stays readable.
func printBuffer(buf *catatui.Buffer) {
	for row := range strings.SplitSeq(buf.String(), "\n") {
		fmt.Println(strings.TrimRight(row, " "))
	}
}

// A whole frame: split the area, then render a widget into each part. This is
// what the body of a Terminal.Draw callback looks like, and TestBackend is how
// it is exercised without a terminal.
func Example() {
	backend := catatui.NewTestBackend(24, 5)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		panic(err)
	}

	terminal.Draw(func(f *catatui.Frame) {
		rows := catatui.VerticalLayout(
			catatui.Length(3),
			catatui.Fill(1),
		).Split(f.Area())

		f.RenderWidget(
			widgets.NewParagraph("hello").
				Block(widgets.Bordered().Title("catatui")),
			rows[0])
		f.RenderWidget(widgets.NewParagraph("q to quit"), rows[1])
	})

	printBuffer(backend.Buffer())
	// Output:
	// ┌catatui───────────────┐
	// │hello                 │
	// └──────────────────────┘
	// q to quit
	//
}

// Constraints decide how a rect is divided. Length takes a fixed number of
// cells, Fill shares out what is left.
func ExampleLayout_Split() {
	area := catatui.NewRect(0, 0, 20, 10)

	columns := catatui.HorizontalLayout(
		catatui.Length(4),
		catatui.Fill(1),
		catatui.Percentage(25),
	).Split(area)

	for _, column := range columns {
		fmt.Printf("x=%d width=%d\n", column.X, column.Width)
	}
	// Output:
	// x=0 width=4
	// x=4 width=11
	// x=15 width=5
}

// Spacing puts a gap between the parts; Overlap makes them share cells, which
// is how neighbouring blocks come to share a border.
func ExampleLayout_Spacing() {
	area := catatui.NewRect(0, 0, 20, 3)

	spaced := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Space(2)).Split(area)
	overlapped := catatui.HorizontalLayout(catatui.Fill(1), catatui.Fill(1)).
		Spacing(catatui.Overlap(1)).Split(area)

	fmt.Println("spaced:    ", spaced[0], spaced[1])
	fmt.Println("overlapped:", overlapped[0], overlapped[1])
	// Output:
	// spaced:     {0 0 9 3} {11 0 9 3}
	// overlapped: {0 0 11 3} {10 0 10 3}
}

// Styles layer rather than replace: an unset color inherits whatever is
// underneath, so patching a foreground over a background keeps both.
func ExampleStyle_Patch() {
	base := catatui.NewStyle().
		Fg(catatui.ColorWhite).
		Bg(catatui.ColorBlue)

	// Only the foreground is set here, so the blue background survives.
	patched := base.Patch(catatui.NewStyle().
		Fg(catatui.ColorYellow).
		AddModifier(catatui.ModifierBold))

	fmt.Println(patched.GetFg(), patched.GetBg(), patched.GetAddModifier())
	// Output:
	// Yellow Blue BOLD
}

// A grapheme cluster that does not fit in the space left is dropped rather than
// clipped, and drawing stops there. StringWidth measures exactly what
// SetStringn fills.
func ExampleBuffer_SetStringn() {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 5, 1))

	// "日" is two columns wide, so three of them need six.
	buf.SetStringn(0, 0, "日本語", 5, catatui.NewStyle())

	fmt.Printf("%q width=%d\n", buf.String(), catatui.StringWidth("日本語"))
	// Output:
	// "日本 " width=6
}
