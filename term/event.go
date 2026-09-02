package term

import "github.com/Fiend3d/catatui"

// EventKind distinguishes the kinds of terminal event.
type EventKind uint8

const (
	// EventKey is a key press.
	EventKey EventKind = iota
	// EventMouse is a mouse press, release, drag, move or wheel movement.
	EventMouse
	// EventResize is a change in the terminal's size.
	EventResize
	// EventPaste is a block of bracketed-paste text.
	EventPaste
	// EventFocus is the terminal window gaining or losing focus.
	EventFocus
)

// Modifiers are the modifier keys held during an event.
type Modifiers uint8

const (
	// ModShift is the shift key.
	ModShift Modifiers = 1 << iota
	// ModAlt is the alt or meta key.
	ModAlt
	// ModCtrl is the control key.
	ModCtrl
)

// Contains reports whether every modifier in other is held.
func (m Modifiers) Contains(other Modifiers) bool { return m&other == other }

// String lists the held modifiers, e.g. "ctrl+shift".
func (m Modifiers) String() string {
	var s string
	for _, n := range []struct {
		bit  Modifiers
		name string
	}{{ModCtrl, "ctrl"}, {ModAlt, "alt"}, {ModShift, "shift"}} {
		if m&n.bit != 0 {
			if s != "" {
				s += "+"
			}
			s += n.name
		}
	}
	return s
}

// KeyCode identifies a key. Printable keys use KeyRune and carry the character
// in Event.Rune; everything else has its own constant.
type KeyCode uint16

const (
	// KeyRune is a printable character, carried in Event.Rune.
	KeyRune KeyCode = iota
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyTab
	KeyBackTab
	KeyDelete
	KeyInsert
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// MouseKind is what the mouse did.
type MouseKind uint8

const (
	// MouseDown is a button press.
	MouseDown MouseKind = iota
	// MouseUp is a button release.
	MouseUp
	// MouseDrag is movement with a button held.
	MouseDrag
	// MouseMove is movement with no button held.
	MouseMove
	// MouseScrollUp is a wheel scroll away from the user.
	MouseScrollUp
	// MouseScrollDown is a wheel scroll toward the user.
	MouseScrollDown
)

// MouseButton identifies which button an event concerns.
type MouseButton uint8

const (
	// MouseButtonNone means no button, as for a bare move or a wheel event.
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

// Event is a single thing that happened at the terminal.
//
// Which fields carry meaning depends on Kind:
//
//	EventKey     Key, Rune, Mods
//	EventMouse   MouseKind, Button, X, Y, Mods
//	EventResize  Size
//	EventPaste   Text
//	EventFocus   Focused
type Event struct {
	Kind EventKind

	// Key events.
	Key  KeyCode
	Rune rune

	// Mouse events.
	MouseKind MouseKind
	Button    MouseButton
	X, Y      uint16

	// Modifier keys held, for key and mouse events.
	Mods Modifiers

	// Resize events.
	Size catatui.Size

	// Paste events.
	Text string

	// Focus events.
	Focused bool
}

// IsKey reports whether e is a press of the given key with no modifiers.
func (e Event) IsKey(k KeyCode) bool {
	return e.Kind == EventKey && e.Key == k && e.Mods == 0
}

// IsRune reports whether e is a press of the given character with no modifiers.
func (e Event) IsRune(r rune) bool {
	return e.Kind == EventKey && e.Key == KeyRune && e.Rune == r && e.Mods == 0
}

// IsCtrl reports whether e is the given character pressed with control held.
func (e Event) IsCtrl(r rune) bool {
	return e.Kind == EventKey && e.Key == KeyRune && e.Rune == r && e.Mods.Contains(ModCtrl)
}

// cellColumns is the number of columns a symbol occupies on screen.
//
// It measures with catatui.StringWidth rather than uniseg directly: the Buffer
// decides how many columns a cell covers with that function, and a second,
// disagreeing width implementation here is exactly what the library exists to
// avoid. The two differ on halfwidth katakana sound marks and control
// characters.
func cellColumns(s string) uint16 {
	w := catatui.StringWidth(s)
	if w < 0 {
		return 0
	}
	if w > 0xFFFF {
		return 0xFFFF
	}
	return uint16(w)
}
