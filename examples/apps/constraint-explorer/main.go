// Command constraint-explorer lets you build a set of constraints and watch
// every Flex mode lay them out at once.
//
//	go run ./examples/apps/constraint-explorer
//
// h/l or the left and right arrows select a block, j/k or up and down change
// its number, 1 to 6 change its kind, a adds a block and x deletes one, +/-
// change the spacing between them, and q quits.
//
// Where the constraints example stacks up fixed cases to read, this one is the
// one to poke at: the six panels below are the same constraints under Start,
// Center, End, SpaceBetween, SpaceAround and SpaceEvenly, so the difference
// between them is what moves when you edit.
//
// Deviation from ratatui: a catatui Constraint does not hand back the number it
// was built with, so the app keeps each block as a kind and a number of its own
// and builds the Constraint from the pair. ratatui edits the constraint in
// place.
//
// Port of examples/apps/constraint-explorer @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run draws the app and waits for a key between frames. Nothing animates, so a
// blocking read is all it takes.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := newApp()

	for !a.quit {
		if err := terminal.Draw(func(f *catatui.Frame) { f.RenderWidget(a, f.Area()) }); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		a.handle(ev)
	}
	return nil
}

// constraintName is a kind of constraint, in the order the number keys pick
// them.
type constraintName int

const (
	nameMin constraintName = iota
	nameMax
	nameLength
	namePercentage
	nameRatio
	nameFill
)

// swapOrder is the order the number keys and the legend run in, which is the
// order the solver gives way in rather than the order above.
var swapOrder = []constraintName{nameMin, nameMax, nameLength, namePercentage, nameRatio, nameFill}

func (n constraintName) String() string {
	switch n {
	case nameMin:
		return "Min"
	case nameMax:
		return "Max"
	case nameLength:
		return "Length"
	case namePercentage:
		return "Percentage"
	case nameRatio:
		return "Ratio"
	default:
		return "Fill"
	}
}

// block is one constraint the app is showing: which kind it is and the number
// it carries. For every kind but Ratio the number is the constraint's own; for
// Ratio it is the denominator, with the numerator fixed at one, which is what
// ratatui edits there too.
type block struct {
	name  constraintName
	value uint16
}

// constraint builds the catatui Constraint this block stands for.
func (b block) constraint() catatui.Constraint {
	switch b.name {
	case nameMin:
		return catatui.Min(b.value)
	case nameMax:
		return catatui.Max(b.value)
	case nameLength:
		return catatui.Length(b.value)
	case namePercentage:
		return catatui.Percentage(b.value)
	case nameRatio:
		return catatui.Ratio(1, uint32(b.value))
	default:
		return catatui.Fill(b.value)
	}
}

// app is the list of blocks, which one is selected, and the spacing between
// them. The spacing is signed: a negative one overlaps the blocks instead of
// separating them.
type app struct {
	blocks   []block
	selected int
	spacing  int16
	quit     bool
}

// defaultValue is the number a new block starts with, and the one a block takes
// when its kind is swapped. ratatui calls this the app's `value` and never
// changes it.
const defaultValue uint16 = 20

// newApp starts with three fixed-width blocks, as ratatui does.
func newApp() *app {
	return &app{blocks: []block{
		{nameLength, defaultValue},
		{nameLength, defaultValue},
		{nameLength, defaultValue},
	}}
}

// constraints is the blocks as the layout wants them.
func (a *app) constraints() []catatui.Constraint {
	constraints := make([]catatui.Constraint, len(a.blocks))
	for i, b := range a.blocks {
		constraints[i] = b.constraint()
	}
	return constraints
}

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('1'):
		a.swapConstraint(nameMin)
	case ev.IsRune('2'):
		a.swapConstraint(nameMax)
	case ev.IsRune('3'):
		a.swapConstraint(nameLength)
	case ev.IsRune('4'):
		a.swapConstraint(namePercentage)
	case ev.IsRune('5'):
		a.swapConstraint(nameRatio)
	case ev.IsRune('6'):
		a.swapConstraint(nameFill)
	case ev.IsRune('+'), ev.IsRune('='):
		a.spacing = satAddI16(a.spacing, 1)
	case ev.IsRune('-'):
		a.spacing = satAddI16(a.spacing, -1)
	case ev.IsRune('a'):
		a.insertBlock()
	case ev.IsRune('x'):
		a.deleteBlock()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.changeValue(1)
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.changeValue(-1)
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.previousBlock()
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.nextBlock()
	}
}

// changeValue moves the selected block's number by one, saturating at both
// ends.
func (a *app) changeValue(delta int) {
	if len(a.blocks) == 0 {
		return
	}
	b := &a.blocks[a.selected]
	if delta > 0 {
		b.value = catatui.SatAdd(b.value, 1)
	} else {
		b.value = catatui.SatSub(b.value, 1)
	}
}

// nextBlock and previousBlock move the selection, wrapping around.
func (a *app) nextBlock() {
	if len(a.blocks) == 0 {
		return
	}
	a.selected = (a.selected + 1) % len(a.blocks)
}

func (a *app) previousBlock() {
	if len(a.blocks) == 0 {
		return
	}
	a.selected = (a.selected + len(a.blocks) - 1) % len(a.blocks)
}

// deleteBlock removes the selected block and selects the one before it.
func (a *app) deleteBlock() {
	if len(a.blocks) == 0 {
		return
	}
	a.blocks = append(a.blocks[:a.selected], a.blocks[a.selected+1:]...)
	a.selected = max(a.selected-1, 0)
}

// insertBlock adds a block after the selected one and selects it.
func (a *app) insertBlock() {
	index := min(a.selected+1, len(a.blocks))
	a.blocks = append(a.blocks, block{})
	copy(a.blocks[index+1:], a.blocks[index:])
	a.blocks[index] = block{nameLength, defaultValue}
	a.selected = index
}

// swapConstraint changes the selected block's kind, resetting its number to the
// one new blocks start with, as ratatui does. A Ratio takes a quarter of it
// instead, so that swapping to one leaves the layout looking much as it did
// rather than collapsing it to a sliver.
func (a *app) swapConstraint(name constraintName) {
	if len(a.blocks) == 0 {
		return
	}
	value := defaultValue
	if name == nameRatio {
		value = defaultValue / 4
	}
	a.blocks[a.selected] = block{name, value}
}

// satAddI16 adds without wrapping around, which is what an int16 spacing does
// at the ends of its range.
func satAddI16(a, b int16) int16 {
	sum := int32(a) + int32(b)
	switch {
	case sum > 32767:
		return 32767
	case sum < -32768:
		return -32768
	}
	return int16(sum)
}
