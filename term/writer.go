// Package term is catatui's terminal driver: the equivalent of the crossterm
// crate that ratatui uses.
//
// It provides raw mode, the alternate screen, mouse and paste reporting, a
// VT/ANSI writer, an input parser, and a Backend implementation that catatui's
// Terminal can drive.
//
//	terminal, restore, err := term.Init(term.WithMouse())
//	if err != nil {
//		return err
//	}
//	defer restore()
//
//	events := term.NewEventReader(os.Stdin, os.Stdout)
//	defer events.Close()
//	for {
//		terminal.Draw(func(f *catatui.Frame) { ... })
//		ev, ok := <-events.Events()
//		if !ok || ev.IsRune('q') {
//			return nil
//		}
//	}
package term

import (
	"bufio"
	"io"
	"strconv"

	"github.com/Fiend3d/catatui"
)

// writer emits VT escape sequences, tracking the state it has already put the
// terminal into so that it only emits what actually changes.
//
// This matters more than it looks: a full-screen redraw at 200x50 is 10,000
// cells, and naively emitting a colour sequence per cell makes the output tens
// of times larger than it needs to be, which is visible as tearing over ssh.
type writer struct {
	w *bufio.Writer

	// The style currently in effect on the terminal, so that SGR sequences are
	// only emitted on change. Both colors start unset, meaning "unknown", which
	// forces the first cell to emit a full reset.
	curFg, curBg, curUnderline catatui.Color
	curModifier                catatui.Modifier
	styleKnown                 bool

	// The cursor position the terminal is at, so that a run of adjacent cells
	// needs no cursor moves at all.
	curX, curY  uint16
	cursorKnown bool
}

func newWriter(w io.Writer) *writer {
	return &writer{w: bufio.NewWriterSize(w, 16<<10)}
}

func (w *writer) flush() error { return w.w.Flush() }

func (w *writer) esc(s string) { w.w.WriteString("\x1b" + s) }

// csi writes a Control Sequence Introducer followed by s.
func (w *writer) csi(s string) {
	w.w.WriteString("\x1b[")
	w.w.WriteString(s)
}

// moveTo positions the cursor, emitting nothing when it is already there and a
// single-character step when the next cell is directly to the right.
func (w *writer) moveTo(x, y uint16) {
	if w.cursorKnown && w.curX == x && w.curY == y {
		return
	}
	w.csi(strconv.Itoa(int(y)+1) + ";" + strconv.Itoa(int(x)+1) + "H")
	w.curX, w.curY, w.cursorKnown = x, y, true
}

// advance records that printing a symbol moved the cursor right by n columns.
func (w *writer) advance(n uint16) {
	w.curX += n
}

// invalidateCursor forgets where the cursor is, forcing the next move to be
// emitted. Anything that can move the cursor behind our back must call this.
func (w *writer) invalidateCursor() { w.cursorKnown = false }

// setStyle emits the SGR sequences needed to get from the current style to the
// requested one.
func (w *writer) setStyle(fg, bg, underline catatui.Color, mod catatui.Modifier) {
	if w.styleKnown && w.curFg == fg && w.curBg == bg &&
		w.curUnderline == underline && w.curModifier == mod {
		return
	}

	// Modifiers can only be turned off individually with codes some terminals
	// ignore, so when any modifier is being removed the whole style is reset
	// and rebuilt. This is what ratatui's backends do too.
	removed := w.curModifier &^ mod
	if !w.styleKnown || removed != 0 {
		w.csi("0m")
		w.curFg, w.curBg, w.curUnderline = catatui.ColorReset, catatui.ColorReset, catatui.ColorReset
		w.curModifier = 0
	}

	for _, m := range modifierCodes {
		if mod&m.bit != 0 && w.curModifier&m.bit == 0 {
			w.csi(m.on)
		}
	}
	if fg != w.curFg {
		w.writeColor(fg, false)
	}
	if bg != w.curBg {
		w.writeColor(bg, true)
	}
	if underline != w.curUnderline {
		w.writeUnderlineColor(underline)
	}

	w.curFg, w.curBg, w.curUnderline, w.curModifier = fg, bg, underline, mod
	w.styleKnown = true
}

// resetStyle puts the terminal back to its default attributes and records
// that, so the next cell re-emits whatever style it needs.
//
// This is not only tidiness. Erasing and scrolling fill the cells they touch
// with the background colour that is currently in effect — "background colour
// erase", which every common terminal does — so a clear issued while the last
// cell's blue background is still set paints the whole screen blue. The diff
// then only rewrites the cells that differ from a blank buffer, and every cell
// the frame leaves blank keeps that blue. Anything that erases or scrolls must
// reset first.
func (w *writer) resetStyle() {
	if w.styleKnown && w.curFg == catatui.ColorReset && w.curBg == catatui.ColorReset &&
		w.curUnderline == catatui.ColorReset && w.curModifier == 0 {
		return
	}
	w.csi("0m")
	w.curFg, w.curBg, w.curUnderline = catatui.ColorReset, catatui.ColorReset, catatui.ColorReset
	w.curModifier = 0
	w.styleKnown = true
}

var modifierCodes = []struct {
	bit catatui.Modifier
	on  string
}{
	{catatui.ModifierBold, "1m"},
	{catatui.ModifierDim, "2m"},
	{catatui.ModifierItalic, "3m"},
	{catatui.ModifierUnderlined, "4m"},
	{catatui.ModifierSlowBlink, "5m"},
	{catatui.ModifierRapidBlink, "6m"},
	{catatui.ModifierReversed, "7m"},
	{catatui.ModifierHidden, "8m"},
	{catatui.ModifierCrossedOut, "9m"},
}

// writeColor emits the SGR code for a foreground or background color.
func (w *writer) writeColor(c catatui.Color, background bool) {
	base := 30
	if background {
		base = 40
	}
	switch {
	case !c.IsSet() || c.IsReset():
		w.csi(strconv.Itoa(base+9) + "m")
	default:
		if n, ok := c.Named(); ok {
			if n < 8 {
				w.csi(strconv.Itoa(base+int(n)) + "m")
			} else {
				// The bright half of the palette lives at 90..97 / 100..107.
				w.csi(strconv.Itoa(base+60+int(n)-8) + "m")
			}
			return
		}
		if i, ok := c.Index(); ok {
			w.csi(strconv.Itoa(base+8) + ";5;" + strconv.Itoa(int(i)) + "m")
			return
		}
		if r, g, b, ok := c.RGB(); ok {
			w.csi(strconv.Itoa(base+8) + ";2;" +
				strconv.Itoa(int(r)) + ";" + strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)) + "m")
		}
	}
}

// writeUnderlineColor emits the SGR 58/59 sequence, which only some terminals
// understand; the rest ignore it harmlessly.
func (w *writer) writeUnderlineColor(c catatui.Color) {
	switch {
	case !c.IsSet() || c.IsReset():
		w.csi("59m")
	default:
		if i, ok := c.Index(); ok {
			w.csi("58;5;" + strconv.Itoa(int(i)) + "m")
			return
		}
		if r, g, b, ok := c.RGB(); ok {
			w.csi("58;2;" + strconv.Itoa(int(r)) + ";" +
				strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)) + "m")
			return
		}
		if n, ok := c.Named(); ok {
			w.csi("58;5;" + strconv.Itoa(int(n)) + "m")
		}
	}
}

// CursorShape is how the terminal draws the cursor.
//
// Setting it is DECSCUSR, which most modern terminals support — Windows
// Terminal, xterm, kitty, alacritty, iTerm2 — and which the rest ignore
// harmlessly. There is no way to query the current shape, so a program that
// changes it should set it back; Init's restore function does that for you.
type CursorShape uint8

const (
	// CursorDefault restores the shape the user configured in their terminal.
	CursorDefault CursorShape = iota
	// CursorBlinkingBlock is a blinking filled rectangle.
	CursorBlinkingBlock
	// CursorSteadyBlock is a filled rectangle that does not blink.
	CursorSteadyBlock
	// CursorBlinkingUnderline is a blinking underscore.
	CursorBlinkingUnderline
	// CursorSteadyUnderline is an underscore that does not blink.
	CursorSteadyUnderline
	// CursorBlinkingBar is a blinking vertical bar.
	CursorBlinkingBar
	// CursorSteadyBar is a vertical bar that does not blink.
	CursorSteadyBar
)

// String returns the shape's name.
func (s CursorShape) String() string {
	switch s {
	case CursorBlinkingBlock:
		return "BlinkingBlock"
	case CursorSteadyBlock:
		return "SteadyBlock"
	case CursorBlinkingUnderline:
		return "BlinkingUnderline"
	case CursorSteadyUnderline:
		return "SteadyUnderline"
	case CursorBlinkingBar:
		return "BlinkingBar"
	case CursorSteadyBar:
		return "SteadyBar"
	default:
		return "Default"
	}
}

// decscusr is the DECSCUSR parameter for a shape. The numbering is the
// standard's: 0 resets to the terminal's default, then the shapes run in
// blinking/steady pairs.
func (s CursorShape) decscusr() int {
	if s > CursorSteadyBar {
		return 0
	}
	return int(s)
}

// setCursorShape emits DECSCUSR. It does not move the cursor, so the writer's
// idea of the cursor position stays valid.
func (w *writer) setCursorShape(s CursorShape) {
	w.csi(strconv.Itoa(s.decscusr()) + " q")
}

// Terminal mode sequences. These are the private DEC modes catatui toggles.
const (
	seqAltScreenOn  = "[?1049h"
	seqAltScreenOff = "[?1049l"
	seqCursorHide   = "[?25l"
	seqCursorShow   = "[?25h"
	seqMouseOn      = "[?1000h\x1b[?1002h\x1b[?1003h\x1b[?1006h"
	seqMouseOff     = "[?1006l\x1b[?1003l\x1b[?1002l\x1b[?1000l"
	seqPasteOn      = "[?2004h"
	seqPasteOff     = "[?2004l"
	seqFocusOn      = "[?1004h"
	seqFocusOff     = "[?1004l"
)
