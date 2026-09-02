// Port of ratatui-widgets/src/barchart/bar.rs @ ratatui-v0.30.2

package widgets

import (
	"strconv"
	"unicode/utf8"

	"github.com/Fiend3d/catatui"
)

// Bar is one bar of a BarChart: a value, an optional label, and the styles
// each part is drawn in.
//
//	███                          ┐
//	█2█  <- text value or value  │ bar
//	foo  <- label                ┘
//
// Every part can be styled on its own. The chart's own BarStyle, ValueStyle and
// LabelStyle sit beneath the bar's, so a bar only needs to set what differs.
//
//	bar := widgets.BarWithLabel(catatui.LineFromString("Bar 1"), 10).
//		Style(catatui.NewStyle().Fg(catatui.ColorRed)).
//		ValueStyle(catatui.NewStyle().Fg(catatui.ColorWhite)).
//		TextValue("10°C")
type Bar struct {
	value        uint64
	label        catatui.Line
	hasLabel     bool
	style        catatui.Style
	valueStyle   catatui.Style
	textValue    string
	hasTextValue bool
}

// NewBar returns a bar with the given value and no label.
func NewBar(value uint64) Bar { return Bar{value: value} }

// BarWithLabel returns a bar with the given label and value.
func BarWithLabel(label catatui.Line, value uint64) Bar {
	return Bar{value: value, label: label, hasLabel: true}
}

// Value returns a copy of b with the given value. The value is drawn inside
// the bar unless TextValue replaces it.
func (b Bar) Value(value uint64) Bar { b.value = value; return b }

// Label returns a copy of b with a label. In a vertical chart the label is
// drawn under the bar; in a horizontal chart it is drawn to the left of it.
func (b Bar) Label(label catatui.Line) Bar { b.label, b.hasLabel = label, true; return b }

// Style returns a copy of b with a style applied to the bar itself. The
// chart's BarStyle sits beneath it.
func (b Bar) Style(s catatui.Style) Bar { b.style = s; return b }

// ValueStyle returns a copy of b with a style applied to the value drawn in
// the bar. The chart's ValueStyle sits beneath it.
func (b Bar) ValueStyle(s catatui.Style) Bar { b.valueStyle = s; return b }

// TextValue returns a copy of b that shows the given text in the bar instead
// of the numeric value.
func (b Bar) TextValue(text string) Bar { b.textValue, b.hasTextValue = text, true; return b }

// valueText is the text drawn in the bar: the text value if set, otherwise the
// value in decimal.
func (b Bar) valueText() string {
	if b.hasTextValue {
		return b.textValue
	}
	return strconv.FormatUint(b.value, 10)
}

// renderValueWithDifferentStyles draws the value of a horizontal bar.
//
// The value may be longer than the bar, so it is drawn in two parts: the part
// inside the bar in the value style, and the rest, outside the bar, in the bar
// style. As in ratatui, barLength and the split are measured in bytes, not
// columns; the split is moved back to a character boundary so a multi-byte
// character is never cut.
func (b Bar) renderValueWithDifferentStyles(buf *catatui.Buffer, area catatui.Rect, barLength int, defaultValueStyle, barStyle catatui.Style) {
	text := b.valueText()
	if text == "" {
		return
	}
	style := defaultValueStyle.Patch(b.valueStyle)
	buf.SetStringn(area.X, area.Y, text, uint16(min(barLength, 0xFFFF)), style)
	if len(text) > barLength {
		// Find the last character boundary at or before barLength.
		split := 0
		for i := 0; i < len(text) && i < barLength; {
			_, n := utf8.DecodeRuneInString(text[i:])
			split = i + n
			i = split
		}
		first, second := text[:split], text[split:]
		firstLen := uint16(min(len(first), 0xFFFF))
		style := barStyle.Patch(b.style)
		buf.SetStringn(catatui.SatAdd(area.X, firstLen), area.Y, second,
			catatui.SatSub(area.Width, firstLen), style)
	}
}

// renderValue draws the value of a vertical bar, centered in the bar's width,
// if there is room for it. A value as wide as the bar is only drawn when the
// bar is at least one full cell tall.
func (b Bar) renderValue(buf *catatui.Buffer, maxWidth, x, y uint16, defaultValueStyle catatui.Style, ticks uint64) {
	if b.value == 0 {
		return
	}
	const ticksPerLine = 8
	label := b.valueText()
	width := uint16(min(catatui.StringWidth(label), 0xFFFF))
	// Print the value if there is enough space, or if the bar is at least one
	// full cell (8 ticks) tall.
	if width < maxWidth || (width == maxWidth && ticks >= ticksPerLine) {
		// ratatui centers on the byte length of the label, not its width.
		labelLen := uint16(min(len(label), 0xFFFF))
		buf.SetString(catatui.SatAdd(x, catatui.SatSub(maxWidth, labelLen)>>1), y, label,
			defaultValueStyle.Patch(b.valueStyle))
	}
}

// renderLabel draws the label of a vertical bar, centered under the bar and
// truncated to the bar's width. Only the label's cells get the default label
// style, not the whole bar width.
func (b Bar) renderLabel(buf *catatui.Buffer, maxWidth, x, y uint16, defaultLabelStyle catatui.Style) {
	var width uint16
	if b.hasLabel {
		width = catatui.MinU16(lineWidth(b.label), maxWidth)
	}
	area := catatui.Rect{
		X:      catatui.SatAdd(x, catatui.SatSub(maxWidth, width)/2),
		Y:      y,
		Width:  width,
		Height: 1,
	}
	buf.SetStyle(area, defaultLabelStyle)
	if b.hasLabel {
		b.label.Render(area, buf)
	}
}
