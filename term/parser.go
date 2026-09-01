package term

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// parse consumes one event from the front of buf.
//
// It returns the event, how many bytes it consumed, and whether an event was
// produced. When ok is false and n is zero, buf holds the start of a sequence
// that is not yet complete and the caller should read more; when ok is false and
// n is positive, that many bytes were consumed without producing an event (an
// unrecognised sequence, which is discarded rather than shown to the user).
//
// Terminal input is ambiguous by design: a lone ESC is both the escape key and
// the start of dozens of sequences, so a caller has to decide by timeout whether
// a trailing ESC is a key press. See EventReader.
func parse(buf []byte) (ev Event, n int, ok bool) {
	if len(buf) == 0 {
		return Event{}, 0, false
	}

	if buf[0] != 0x1b {
		return parseChar(buf)
	}
	if len(buf) == 1 {
		// Could be the escape key, or the start of a sequence. Undecidable here.
		return Event{}, 0, false
	}

	switch buf[1] {
	case '[':
		return parseCSI(buf)
	case 'O':
		return parseSS3(buf)
	case 0x1b:
		// A doubled escape: report the first as the escape key.
		return Event{Kind: EventKey, Key: KeyEscape}, 1, true
	default:
		// ESC followed by a character is that character with alt held.
		ev, n, ok := parseChar(buf[1:])
		if !ok {
			return Event{}, 0, false
		}
		ev.Mods |= ModAlt
		return ev, n + 1, true
	}
}

// parseChar decodes a single character or C0 control byte.
func parseChar(buf []byte) (Event, int, bool) {
	b := buf[0]
	switch {
	case b == 0x0d || b == 0x0a:
		return Event{Kind: EventKey, Key: KeyEnter}, 1, true
	case b == 0x09:
		return Event{Kind: EventKey, Key: KeyTab}, 1, true
	case b == 0x7f || b == 0x08:
		return Event{Kind: EventKey, Key: KeyBackspace}, 1, true
	case b == 0x1b:
		return Event{Kind: EventKey, Key: KeyEscape}, 1, true
	case b == 0x00:
		// NUL arrives from ctrl+space on many terminals.
		return Event{Kind: EventKey, Key: KeyRune, Rune: ' ', Mods: ModCtrl}, 1, true
	case b < 0x20:
		// The remaining C0 controls are ctrl+letter: 0x01 is ctrl+a.
		return Event{Kind: EventKey, Key: KeyRune, Rune: rune('a' + b - 1), Mods: ModCtrl}, 1, true
	}

	r, size := utf8.DecodeRune(buf)
	if r == utf8.RuneError && size <= 1 {
		if !utf8.FullRune(buf) {
			// A partial multi-byte character; wait for the rest.
			return Event{}, 0, false
		}
		// Genuinely invalid input: drop the byte rather than emitting garbage.
		return Event{}, 1, false
	}
	return Event{Kind: EventKey, Key: KeyRune, Rune: r}, size, true
}

// parseSS3 handles the ESC O sequences, which are how some terminals report the
// arrow and function keys in application cursor mode.
func parseSS3(buf []byte) (Event, int, bool) {
	if len(buf) < 3 {
		return Event{}, 0, false
	}
	key, ok := ss3Keys[buf[2]]
	if !ok {
		return Event{}, 3, false
	}
	return Event{Kind: EventKey, Key: key}, 3, true
}

var ss3Keys = map[byte]KeyCode{
	'A': KeyUp, 'B': KeyDown, 'C': KeyRight, 'D': KeyLeft,
	'H': KeyHome, 'F': KeyEnd,
	'P': KeyF1, 'Q': KeyF2, 'R': KeyF3, 'S': KeyF4,
}

// parseCSI handles ESC [ sequences: cursor keys, function keys, mouse reports,
// bracketed paste and focus events.
func parseCSI(buf []byte) (Event, int, bool) {
	// Find the final byte, which is the first in the range 0x40..0x7e after the
	// parameter and intermediate bytes.
	end := -1
	for i := 2; i < len(buf); i++ {
		if buf[i] >= 0x40 && buf[i] <= 0x7e {
			end = i
			break
		}
	}
	if end < 0 {
		if len(buf) > 64 {
			// Nothing legitimate is this long; resynchronise rather than
			// stalling forever on a malformed sequence.
			return Event{}, len(buf), false
		}
		return Event{}, 0, false
	}

	body := string(buf[2:end])
	final := buf[end]
	n := end + 1

	// SGR mouse reports: ESC [ < b ; x ; y (M|m)
	if strings.HasPrefix(body, "<") && (final == 'M' || final == 'm') {
		ev, ok := parseSGRMouse(body[1:], final == 'M')
		return ev, n, ok
	}

	// Bracketed paste: ESC [ 200 ~ ... ESC [ 201 ~
	if body == "200" && final == '~' {
		return parsePaste(buf, n)
	}

	// Focus in and out.
	if final == 'I' {
		return Event{Kind: EventFocus, Focused: true}, n, true
	}
	if final == 'O' {
		return Event{Kind: EventFocus, Focused: false}, n, true
	}

	params := parseParams(body)
	mods := csiModifiers(params)

	// The tilde-terminated keys are identified by their first parameter.
	if final == '~' {
		if len(params) == 0 {
			return Event{}, n, false
		}
		key, ok := tildeKeys[params[0]]
		if !ok {
			return Event{}, n, false
		}
		return Event{Kind: EventKey, Key: key, Mods: mods}, n, true
	}

	// CSI Z is shift+tab on every terminal that emits it.
	if final == 'Z' {
		return Event{Kind: EventKey, Key: KeyBackTab, Mods: mods}, n, true
	}

	if key, ok := csiLetterKeys[final]; ok {
		return Event{Kind: EventKey, Key: key, Mods: mods}, n, true
	}
	return Event{}, n, false
}

var tildeKeys = map[int]KeyCode{
	1: KeyHome, 2: KeyInsert, 3: KeyDelete, 4: KeyEnd, 5: KeyPageUp, 6: KeyPageDown,
	7: KeyHome, 8: KeyEnd,
	11: KeyF1, 12: KeyF2, 13: KeyF3, 14: KeyF4, 15: KeyF5,
	17: KeyF6, 18: KeyF7, 19: KeyF8, 20: KeyF9, 21: KeyF10,
	23: KeyF11, 24: KeyF12,
}

var csiLetterKeys = map[byte]KeyCode{
	'A': KeyUp, 'B': KeyDown, 'C': KeyRight, 'D': KeyLeft,
	'H': KeyHome, 'F': KeyEnd,
	'P': KeyF1, 'Q': KeyF2, 'S': KeyF4,
}

// parseParams splits a CSI parameter string into numbers, treating an empty
// parameter as zero the way terminals do.
func parseParams(body string) []int {
	if body == "" {
		return nil
	}
	// Sub-parameters after a colon are not used here; keep the leading number.
	parts := strings.Split(body, ";")
	out := make([]int, len(parts))
	for i, p := range parts {
		if j := strings.IndexByte(p, ':'); j >= 0 {
			p = p[:j]
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			v = 0
		}
		out[i] = v
	}
	return out
}

// csiModifiers decodes the modifier parameter, which terminals encode as a
// bitmask plus one in the second parameter.
func csiModifiers(params []int) Modifiers {
	if len(params) < 2 || params[1] < 1 {
		return 0
	}
	return decodeModifierMask(params[1] - 1)
}

func decodeModifierMask(mask int) Modifiers {
	var m Modifiers
	if mask&1 != 0 {
		m |= ModShift
	}
	if mask&2 != 0 {
		m |= ModAlt
	}
	if mask&4 != 0 {
		m |= ModCtrl
	}
	return m
}

// parseSGRMouse decodes an SGR (1006) mouse report body, "b;x;y".
//
// SGR is the only mouse encoding catatui enables, because the older X10 one
// cannot report coordinates past column 223 and gives no way to tell a press
// from a release.
func parseSGRMouse(body string, press bool) (Event, bool) {
	params := parseParams(body)
	if len(params) < 3 {
		return Event{}, false
	}
	b, x, y := params[0], params[1], params[2]
	if x < 1 || y < 1 {
		return Event{}, false
	}

	ev := Event{
		Kind: EventMouse,
		// Terminals report 1-based coordinates; catatui is 0-based throughout.
		X:    uint16(x - 1),
		Y:    uint16(y - 1),
		Mods: sgrMouseModifiers(b),
	}

	switch {
	case b&64 != 0:
		// Wheel events. Bit 0 selects the direction.
		if b&1 != 0 {
			ev.MouseKind = MouseScrollDown
		} else {
			ev.MouseKind = MouseScrollUp
		}
	case b&32 != 0:
		// Motion. With a button held it is a drag, otherwise a bare move.
		ev.Button = sgrMouseButton(b)
		if ev.Button == MouseButtonNone {
			ev.MouseKind = MouseMove
		} else {
			ev.MouseKind = MouseDrag
		}
	default:
		ev.Button = sgrMouseButton(b)
		if press {
			ev.MouseKind = MouseDown
		} else {
			ev.MouseKind = MouseUp
		}
	}
	return ev, true
}

func sgrMouseButton(b int) MouseButton {
	switch b & 3 {
	case 0:
		return MouseButtonLeft
	case 1:
		return MouseButtonMiddle
	case 2:
		return MouseButtonRight
	default:
		// 3 means "no button", which is what a bare move reports.
		return MouseButtonNone
	}
}

func sgrMouseModifiers(b int) Modifiers {
	var m Modifiers
	if b&4 != 0 {
		m |= ModShift
	}
	if b&8 != 0 {
		m |= ModAlt
	}
	if b&16 != 0 {
		m |= ModCtrl
	}
	return m
}

// pasteEnd terminates a bracketed paste.
var pasteEnd = []byte("\x1b[201~")

// parsePaste collects everything up to the paste terminator into one event, so
// that pasted text never looks like a burst of key presses.
func parsePaste(buf []byte, bodyStart int) (Event, int, bool) {
	i := indexBytes(buf[bodyStart:], pasteEnd)
	if i < 0 {
		// The paste is still arriving.
		return Event{}, 0, false
	}
	text := string(buf[bodyStart : bodyStart+i])
	return Event{Kind: EventPaste, Text: text}, bodyStart + i + len(pasteEnd), true
}

func indexBytes(haystack, needle []byte) int {
	return strings.Index(string(haystack), string(needle))
}
