// Package catatui draws terminal user interfaces: a cell-addressed buffer, a
// constraint-solved layout, and a terminal that writes only what changed.
//
// It is a near-literal Go port of [ratatui], following the structure and
// semantics of ratatui 0.30.x. Names, behaviour and edge cases match, and every
// ported file names its Rust source in a header comment.
//
// # The model
//
// catatui is immediate mode. There is no widget tree that lives between frames:
// every frame your code builds widget values, renders them into a buffer, and
// throws them away. Whatever the UI has to remember lives in your own structs.
//
// One frame is one call to [Terminal.Draw], which:
//
//  1. resizes its buffers if the terminal changed size,
//  2. hands your callback a blank [Frame],
//  3. diffs what you drew against the previous frame,
//  4. writes only the cells that differ, then places or hides the cursor.
//
// Nothing is written to the terminal until the callback returns, so a frame is
// atomic: the screen never shows a half-drawn UI.
//
// # A whole program
//
//	package main
//
//	import (
//		"os"
//
//		"github.com/Fiend3d/catatui"
//		"github.com/Fiend3d/catatui/term"
//		"github.com/Fiend3d/catatui/widgets"
//	)
//
//	func main() {
//		defer term.RecoverAndRestore()
//
//		terminal, restore, err := term.Init()
//		if err != nil {
//			panic(err)
//		}
//		defer restore()
//
//		events := term.NewEventReader(os.Stdin, os.Stdout)
//		defer events.Close()
//
//		for {
//			terminal.Draw(func(f *catatui.Frame) {
//				rows := catatui.VerticalLayout(
//					catatui.Length(3),
//					catatui.Fill(1),
//				).Split(f.Area())
//
//				f.RenderWidget(
//					widgets.NewParagraph("hello").
//						Block(widgets.Bordered().Title("catatui")),
//					rows[0])
//			})
//
//			ev, ok := <-events.Events()
//			if !ok || ev.IsRune('q') {
//				return
//			}
//		}
//	}
//
// # The packages
//
//   - catatui, this package: [Rect] and layout, [Buffer] and [Cell], [Style]
//     and styled text, [Widget], [Frame] and [Terminal].
//   - [github.com/Fiend3d/catatui/term]: the terminal driver — raw mode, the
//     alternate screen, input events, and the [Backend] that Terminal writes
//     through. This is what crossterm is to ratatui.
//   - [github.com/Fiend3d/catatui/widgets]: the widget library — Block,
//     Paragraph, List, Table, Chart, Canvas and the rest.
//   - [github.com/Fiend3d/catatui/symbols]: the characters widgets draw with.
//   - [github.com/Fiend3d/catatui/palette/tailwind] and
//     [github.com/Fiend3d/catatui/palette/material]: named colour ramps, for
//     when the sixteen ANSI colours are not enough.
//
// # Things worth knowing early
//
// Coordinates are uint16 with saturating arithmetic ([SatAdd], [SatSub]), as in
// ratatui, so a rect can never wrap around into a huge one. [Rect.Right] and
// [Rect.Bottom] are exclusive: they name the first column and row past the
// rect.
//
// There is exactly one width function, [StringWidth], and what it measures is
// exactly the number of columns [Buffer.SetString] fills. A grapheme cluster
// that does not fit in the space left is dropped rather than clipped, and
// drawing stops there.
//
// The zero [Color] means "unset" — inherit whatever is underneath — while
// [ColorReset] is an explicit reset to the terminal's default. This is the one
// place where catatui's model differs from ratatui's Option[Color], and it is
// what keeps [Style] comparable and allocation-free.
//
// Builders return modified copies rather than consuming the receiver, so a
// half-built widget stays usable:
//
//	base := widgets.Bordered().Title("Files")
//	left := base.BorderStyle(cyan)  // base is unchanged
//
// Widgets hold no state. A widget that has to remember something — which row of
// a list is selected, how far a view is scrolled — takes that state from the
// caller through [RenderStatefulWidget], which is a free function because Go
// methods cannot have type parameters of their own.
//
// # Where to read next
//
// The guides under docs/concepts cover each area in depth: rendering, layout,
// text and style, events, the terminal, and testing. The examples directory has
// a runnable program for every widget.
//
// [ratatui]: https://ratatui.rs
package catatui
