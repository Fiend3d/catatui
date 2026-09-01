package term

import (
	"strings"
	"testing"
)

// parseAll runs the parser over a whole input the way EventReader does, so that
// the tests exercise the same path a real session takes.
func parseAll(t *testing.T, in string) []Event {
	t.Helper()
	var out []Event
	pending := []byte(in)
	for len(pending) > 0 {
		ev, n, ok := parse(pending)
		if n == 0 && !ok {
			t.Fatalf("parser stalled on %q with %q left", in, pending)
		}
		pending = pending[n:]
		if ok {
			out = append(out, ev)
		}
	}
	return out
}

func parseOne(t *testing.T, in string) Event {
	t.Helper()
	evs := parseAll(t, in)
	if len(evs) != 1 {
		t.Fatalf("parsing %q gave %d events, want 1: %+v", in, len(evs), evs)
	}
	return evs[0]
}

func TestParseKeys(t *testing.T) {
	cases := []struct {
		in   string
		key  KeyCode
		mods Modifiers
	}{
		{"\r", KeyEnter, 0},
		{"\n", KeyEnter, 0},
		{"\t", KeyTab, 0},
		{"\x7f", KeyBackspace, 0},
		{"\x1b[A", KeyUp, 0},
		{"\x1b[B", KeyDown, 0},
		{"\x1b[C", KeyRight, 0},
		{"\x1b[D", KeyLeft, 0},
		{"\x1b[H", KeyHome, 0},
		{"\x1b[F", KeyEnd, 0},
		{"\x1bOA", KeyUp, 0},
		{"\x1bOH", KeyHome, 0},
		{"\x1b[2~", KeyInsert, 0},
		{"\x1b[3~", KeyDelete, 0},
		{"\x1b[5~", KeyPageUp, 0},
		{"\x1b[6~", KeyPageDown, 0},
		{"\x1b[15~", KeyF5, 0},
		{"\x1b[24~", KeyF12, 0},
		{"\x1bOP", KeyF1, 0},
		{"\x1b[Z", KeyBackTab, 0},

		// Modified keys use the ";mask+1" parameter.
		{"\x1b[1;2A", KeyUp, ModShift},
		{"\x1b[1;3A", KeyUp, ModAlt},
		{"\x1b[1;5A", KeyUp, ModCtrl},
		{"\x1b[1;6A", KeyUp, ModShift | ModCtrl},
		{"\x1b[1;8A", KeyUp, ModShift | ModAlt | ModCtrl},
		{"\x1b[3;5~", KeyDelete, ModCtrl},
	}
	for _, c := range cases {
		ev := parseOne(t, c.in)
		if ev.Kind != EventKey {
			t.Errorf("%q: kind = %v, want EventKey", c.in, ev.Kind)
			continue
		}
		if ev.Key != c.key {
			t.Errorf("%q: key = %v, want %v", c.in, ev.Key, c.key)
		}
		if ev.Mods != c.mods {
			t.Errorf("%q: mods = %v, want %v", c.in, ev.Mods, c.mods)
		}
	}
}

func TestParseRunes(t *testing.T) {
	cases := []struct {
		in   string
		r    rune
		mods Modifiers
	}{
		{"a", 'a', 0},
		{"Z", 'Z', 0},
		{"1", '1', 0},
		{" ", ' ', 0},
		{"é", 'é', 0},
		{"あ", 'あ', 0},
		{"\x01", 'a', ModCtrl},
		{"\x03", 'c', ModCtrl},
		{"\x1a", 'z', ModCtrl},
		{"\x00", ' ', ModCtrl},
		{"\x1ba", 'a', ModAlt},
	}
	for _, c := range cases {
		ev := parseOne(t, c.in)
		if ev.Kind != EventKey || ev.Key != KeyRune {
			t.Errorf("%q: got %+v, want a rune key", c.in, ev)
			continue
		}
		if ev.Rune != c.r || ev.Mods != c.mods {
			t.Errorf("%q: got rune %q mods %v, want %q %v", c.in, ev.Rune, ev.Mods, c.r, c.mods)
		}
	}
}

func TestParseSGRMouse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Event
	}{
		{
			"left press", "\x1b[<0;10;5M",
			Event{Kind: EventMouse, MouseKind: MouseDown, Button: MouseButtonLeft, X: 9, Y: 4},
		},
		{
			"left release", "\x1b[<0;10;5m",
			Event{Kind: EventMouse, MouseKind: MouseUp, Button: MouseButtonLeft, X: 9, Y: 4},
		},
		{
			"right press", "\x1b[<2;1;1M",
			Event{Kind: EventMouse, MouseKind: MouseDown, Button: MouseButtonRight, X: 0, Y: 0},
		},
		{
			"middle press", "\x1b[<1;3;4M",
			Event{Kind: EventMouse, MouseKind: MouseDown, Button: MouseButtonMiddle, X: 2, Y: 3},
		},
		{
			"left drag", "\x1b[<32;7;8M",
			Event{Kind: EventMouse, MouseKind: MouseDrag, Button: MouseButtonLeft, X: 6, Y: 7},
		},
		{
			"bare move", "\x1b[<35;7;8M",
			Event{Kind: EventMouse, MouseKind: MouseMove, Button: MouseButtonNone, X: 6, Y: 7},
		},
		{
			"scroll up", "\x1b[<64;2;3M",
			Event{Kind: EventMouse, MouseKind: MouseScrollUp, X: 1, Y: 2},
		},
		{
			"scroll down", "\x1b[<65;2;3M",
			Event{Kind: EventMouse, MouseKind: MouseScrollDown, X: 1, Y: 2},
		},
		{
			"ctrl+left press", "\x1b[<16;10;5M",
			Event{Kind: EventMouse, MouseKind: MouseDown, Button: MouseButtonLeft, X: 9, Y: 4, Mods: ModCtrl},
		},
		{
			"shift+left press", "\x1b[<4;10;5M",
			Event{Kind: EventMouse, MouseKind: MouseDown, Button: MouseButtonLeft, X: 9, Y: 4, Mods: ModShift},
		},
	}
	for _, c := range cases {
		got := parseOne(t, c.in)
		if got != c.want {
			t.Errorf("%s (%q):\n  got  %+v\n  want %+v", c.name, c.in, got, c.want)
		}
	}
}

// TestParseMouseCoordinatesAreZeroBased pins the off-by-one that makes mouse
// hit-testing line up with the buffer: terminals report 1-based columns.
func TestParseMouseCoordinatesAreZeroBased(t *testing.T) {
	ev := parseOne(t, "\x1b[<0;1;1M")
	if ev.X != 0 || ev.Y != 0 {
		t.Errorf("the top left cell reported as (%d, %d), want (0, 0)", ev.X, ev.Y)
	}
}

func TestParseBracketedPaste(t *testing.T) {
	ev := parseOne(t, "\x1b[200~hello world\x1b[201~")
	if ev.Kind != EventPaste {
		t.Fatalf("kind = %v, want EventPaste", ev.Kind)
	}
	if ev.Text != "hello world" {
		t.Errorf("text = %q, want %q", ev.Text, "hello world")
	}
}

// TestParsePasteKeepsControlCharacters checks that a paste containing newlines
// stays one event, which is the whole point of bracketed paste: a pasted block
// must not look like the user pressing enter.
func TestParsePasteKeepsControlCharacters(t *testing.T) {
	ev := parseOne(t, "\x1b[200~line1\nline2\r\n\x1b[201~")
	if ev.Kind != EventPaste {
		t.Fatalf("kind = %v, want EventPaste", ev.Kind)
	}
	if want := "line1\nline2\r\n"; ev.Text != want {
		t.Errorf("text = %q, want %q", ev.Text, want)
	}
}

func TestParseFocus(t *testing.T) {
	if ev := parseOne(t, "\x1b[I"); ev.Kind != EventFocus || !ev.Focused {
		t.Errorf("focus in gave %+v", ev)
	}
	if ev := parseOne(t, "\x1b[O"); ev.Kind != EventFocus || ev.Focused {
		t.Errorf("focus out gave %+v", ev)
	}
}

// TestParseIncompleteSequences checks that a partial sequence asks for more
// bytes rather than being misread. Input arrives in arbitrary chunks, so this
// happens constantly in practice.
func TestParseIncompleteSequences(t *testing.T) {
	for _, in := range []string{"\x1b", "\x1b[", "\x1b[1", "\x1b[1;", "\x1b[<0;10", "\x1b[200~partial"} {
		ev, n, ok := parse([]byte(in))
		if ok || n != 0 {
			t.Errorf("%q: got (%+v, %d, %v), want the parser to ask for more bytes", in, ev, n, ok)
		}
	}
}

// TestParseSequenceSplitAcrossReads checks that feeding a sequence one byte at a
// time still yields exactly one event, which is what the read loop does when
// input is slow.
func TestParseSequenceSplitAcrossReads(t *testing.T) {
	full := "\x1b[<0;10;5M"
	for split := 1; split < len(full); split++ {
		head := []byte(full[:split])
		if _, n, ok := parse(head); ok || n != 0 {
			t.Errorf("split at %d: the parser consumed a partial sequence", split)
			continue
		}
		ev := parseOne(t, full)
		if ev.X != 9 || ev.Y != 4 {
			t.Errorf("split at %d: reassembled event = %+v", split, ev)
		}
	}
}

// TestParseResynchronisesOnGarbage checks that malformed input is dropped rather
// than stalling the reader forever.
func TestParseResynchronisesOnGarbage(t *testing.T) {
	// An unrecognised final byte should be consumed, and the following key read.
	evs := parseAll(t, "\x1b[99999999xa")
	if len(evs) != 1 || !evs[0].IsRune('a') {
		t.Errorf("after garbage, got %+v, want just the 'a' key", evs)
	}

	// A runaway sequence with no final byte must not be held indefinitely.
	long := "\x1b[" + strings.Repeat("1;", 100)
	if _, n, ok := parse([]byte(long)); ok || n == 0 {
		t.Errorf("a runaway sequence should be discarded, got n=%d ok=%v", n, ok)
	}
}

func TestParseMultipleEventsInOneRead(t *testing.T) {
	evs := parseAll(t, "ab\x1b[Ac")
	if len(evs) != 4 {
		t.Fatalf("got %d events, want 4: %+v", len(evs), evs)
	}
	if !evs[0].IsRune('a') || !evs[1].IsRune('b') || !evs[2].IsKey(KeyUp) || !evs[3].IsRune('c') {
		t.Errorf("events = %+v", evs)
	}
}

func TestEventHelpers(t *testing.T) {
	if !parseOne(t, "q").IsRune('q') {
		t.Error("IsRune should match a plain letter")
	}
	if !parseOne(t, "\x03").IsCtrl('c') {
		t.Error("IsCtrl should match ctrl+c")
	}
	if parseOne(t, "\x03").IsRune('c') {
		t.Error("IsRune should not match when ctrl is held")
	}
	if !parseOne(t, "\x1b[A").IsKey(KeyUp) {
		t.Error("IsKey should match an unmodified arrow key")
	}
	if parseOne(t, "\x1b[1;5A").IsKey(KeyUp) {
		t.Error("IsKey should not match when ctrl is held")
	}
}

func TestModifiersString(t *testing.T) {
	cases := []struct {
		m    Modifiers
		want string
	}{
		{0, ""},
		{ModCtrl, "ctrl"},
		{ModAlt, "alt"},
		{ModShift, "shift"},
		{ModCtrl | ModShift, "ctrl+shift"},
		{ModCtrl | ModAlt | ModShift, "ctrl+alt+shift"},
	}
	for _, c := range cases {
		if got := c.m.String(); got != c.want {
			t.Errorf("Modifiers(%d).String() = %q, want %q", c.m, got, c.want)
		}
	}
}
