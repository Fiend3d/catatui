package main

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// TestRender draws the app at sizes from nothing to bigger than a screen, in
// each of the states the keys can put it in. Rendering outside the area given
// panics in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	apps := map[string]*app{
		"default":   newApp(),
		"no blocks": {},
		"one of each": {blocks: []block{
			{nameMin, 20}, {nameMax, 20}, {nameLength, 20},
			{namePercentage, 20}, {nameRatio, 5}, {nameFill, 1},
		}},
		"wide spacing":  {blocks: newApp().blocks, spacing: 40},
		"overlapping":   {blocks: newApp().blocks, spacing: -10},
		"huge values":   {blocks: []block{{nameLength, 65535}, {namePercentage, 65535}}},
		"zero values":   {blocks: []block{{nameLength, 0}, {nameRatio, 0}, {nameFill, 0}}},
		"last selected": {blocks: newApp().blocks, selected: 2},
		"many blocks":   manyBlocks(20),
	}
	for name, a := range apps {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
			err = terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) })
			if err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
		}
	}
}

func manyBlocks(n int) *app {
	a := &app{}
	for range n {
		a.blocks = append(a.blocks, block{nameLength, 20})
	}
	return a
}

// key builds a key event for the tests below.
func key(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

// TestAddingAndDeletingBlocks checks a is an insert after the selection, x
// removes the selected block, and the selection stays in range through both —
// an index past the end would panic on the next frame.
func TestAddingAndDeletingBlocks(t *testing.T) {
	a := newApp()
	a.handle(key('a'))
	if len(a.blocks) != 4 || a.selected != 1 {
		t.Errorf("after a: %d blocks, selection %d; want 4 and 1", len(a.blocks), a.selected)
	}

	for range 10 {
		a.handle(key('x'))
	}
	if len(a.blocks) != 0 {
		t.Errorf("x left %d blocks after deleting more times than there were", len(a.blocks))
	}
	if a.selected != 0 {
		t.Errorf("the selection is %d with nothing to select", a.selected)
	}

	// Every key has to survive an empty list, since x can empty it.
	for _, r := range []rune{'1', '2', '3', '4', '5', '6', 'j', 'k', 'h', 'l', 'x', '+', '-'} {
		a.handle(key(r))
		if a.selected < 0 || a.selected > len(a.blocks) {
			t.Fatalf("%q left the selection at %d with %d blocks", r, a.selected, len(a.blocks))
		}
	}

	a.handle(key('a'))
	if len(a.blocks) != 1 || a.selected != 0 {
		t.Errorf("a on an empty list gave %d blocks and selection %d; want 1 and 0",
			len(a.blocks), a.selected)
	}
}

// TestTheSelectionWrapsAround checks h and l cycle rather than stopping, which
// is what the app promises.
func TestTheSelectionWrapsAround(t *testing.T) {
	a := newApp()
	for range len(a.blocks) {
		a.handle(key('l'))
	}
	if a.selected != 0 {
		t.Errorf("l stopped at %d after a full cycle, want back at 0", a.selected)
	}
	a.handle(key('h'))
	if a.selected != len(a.blocks)-1 {
		t.Errorf("h from the first block went to %d, want the last", a.selected)
	}
}

// TestValuesSaturate checks the number on a block stops at both ends of uint16
// rather than wrapping, which would jump a block from empty to the full width.
func TestValuesSaturate(t *testing.T) {
	a := &app{blocks: []block{{nameLength, 1}}}
	for range 5 {
		a.handle(key('j'))
	}
	if a.blocks[0].value != 0 {
		t.Errorf("the value went to %d below zero, want 0", a.blocks[0].value)
	}

	a.blocks[0].value = 65534
	for range 5 {
		a.handle(key('k'))
	}
	if a.blocks[0].value != 65535 {
		t.Errorf("the value went to %d past the top, want 65535", a.blocks[0].value)
	}
}

// TestSpacingGoesBothWays checks - takes the spacing negative, which is what
// turns it into an overlap, and that both directions saturate.
func TestSpacingGoesBothWays(t *testing.T) {
	a := newApp()
	for range 3 {
		a.handle(key('-'))
	}
	if a.spacing != -3 {
		t.Fatalf("three presses of - gave a spacing of %d, want -3", a.spacing)
	}
	if _, isOverlap := strings.CutPrefix(a.spacingValue().String(), "Overlap"); !isOverlap {
		t.Errorf("a negative spacing came out as %v, want an overlap", a.spacingValue())
	}

	a.spacing = 32767
	a.handle(key('+'))
	if a.spacing != 32767 {
		t.Errorf("+ wrapped the spacing round to %d", a.spacing)
	}
	a.spacing = -32768
	a.handle(key('-'))
	if a.spacing != -32768 {
		t.Errorf("- wrapped the spacing round to %d", a.spacing)
	}
}

// TestSwappingKinds checks each number key swaps the selected block and leaves
// the rest alone, and that a Ratio comes out with a denominator rather than the
// raw value, which would make it a sliver of the width.
func TestSwappingKinds(t *testing.T) {
	for i, name := range swapOrder {
		a := newApp()
		a.handle(key('l')) // select the middle block
		a.handle(key(rune('1' + i)))

		if got := a.blocks[1].name; got != name {
			t.Errorf("key %d swapped to %v, want %v", i+1, got, name)
		}
		if a.blocks[0].name != nameLength || a.blocks[2].name != nameLength {
			t.Errorf("key %d changed a block that was not selected", i+1)
		}

		want := "Ratio(1, 5)"
		if name == nameRatio {
			if got := a.blocks[1].constraint().String(); got != want {
				t.Errorf("swapping to a ratio gave %s, want %s", got, want)
			}
		}
	}
}

// TestTheAxisNamesTheSpacing checks the bar over each panel says what the
// spacing is doing, since a gap and an overlap are otherwise hard to tell apart
// from the blocks alone.
func TestTheAxisNamesTheSpacing(t *testing.T) {
	for _, tc := range []struct {
		spacing int16
		want    string
	}{
		{0, "40 px"},
		{2, "40 px (gap: 2 px)"},
		{-2, "40 px (overlap: 2 px)"},
	} {
		buf := catatui.NewBuffer(catatui.NewRect(0, 0, 40, 1))
		(&app{spacing: tc.spacing}).axis(40).Render(buf.Area, buf)

		line := strings.TrimSpace(buf.String())
		if !strings.Contains(line, tc.want) {
			t.Errorf("the axis for a spacing of %d reads %q, want it to contain %q",
				tc.spacing, line, tc.want)
		}
		if !strings.HasPrefix(line, "<") || !strings.HasSuffix(line, ">") {
			t.Errorf("the axis for a spacing of %d reads %q, want it between < and >",
				tc.spacing, line)
		}
	}
}

// TestABlockNamesItsConstraintAndWidth checks the four-row block, which is the
// one the app is read from: the constraint as it was written, and the width it
// actually came to.
func TestABlockNamesItsConstraintAndWidth(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 14, 4))
	constraintBlock{block: block{nameLength, 20}}.Render(buf.Area, buf)

	text := buf.String()
	for _, want := range []string{"Length(20)", "14 px"} {
		if !strings.Contains(text, want) {
			t.Errorf("the block reads\n%s\nwant it to contain %q", text, want)
		}
	}
}

// TestANarrowBlockDropsTheLabels checks the width label gives way before the
// constraint does, and that neither is drawn past the border.
func TestANarrowBlockDropsTheLabels(t *testing.T) {
	for _, width := range []uint16{0, 1, 2, 3, 4, 5, 6, 7} {
		got := constraintBlock{block: block{nameLength, 20}}.label(width)
		lines := strings.SplitN(got, "\n", 2)
		if len(lines) != 2 {
			t.Fatalf("width %d gave %q, want two lines", width, got)
		}
		if lines[1] != "" && len(lines[1]) >= int(catatui.SatSub(width, 2)) {
			t.Errorf("width %d labels itself %q, which does not fit inside the border",
				width, lines[1])
		}
	}
}

// TestASpacerNamesItself checks the gap between blocks is labelled when there
// is room, and silently is not when there is not.
func TestASpacerNamesItself(t *testing.T) {
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 8, 4))
	spacerBlock{}.Render(buf.Area, buf)
	if text := buf.String(); !strings.Contains(text, "Spacer") || !strings.Contains(text, "8 px") {
		t.Errorf("a wide spacer reads\n%s\nwant it to name itself and its width", text)
	}

	narrow := catatui.NewBuffer(catatui.NewRect(0, 0, 3, 4))
	spacerBlock{}.Render(narrow.Area, narrow)
	if text := narrow.String(); strings.Contains(text, "Spacer") {
		t.Errorf("a three-column spacer still labels itself:\n%s", text)
	}
}
