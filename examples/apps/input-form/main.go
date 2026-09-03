// Command input-form is a form with three fields and a focus that moves
// between them.
//
//	go run ./examples/apps/input-form
//
// Tab moves to the next field, Enter submits, Esc cancels. On submit the form
// is printed as JSON; on cancel it says so.
//
// The point of it is where the cursor goes: each field says where its cursor
// belongs, and the form asks whichever one has the focus. Nothing here handles
// cursor movement within a field — see the user-input example for that.
//
// Port of examples/apps/input-form @ ratatui-v0.30.2
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func main() {
	form, submitted, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if !submitted {
		fmt.Println("Canceled")
		return
	}
	encoded, err := json.MarshalIndent(form, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}

// run draws the form until it is submitted or cancelled, and reports which.
func run() (inputForm, bool, error) {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return inputForm{}, false, err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	form := newInputForm()

	for {
		if err := terminal.Draw(form.render); err != nil {
			return inputForm{}, false, err
		}
		ev, ok := <-events.Events()
		if !ok {
			return inputForm{}, false, events.Err()
		}
		if ev.Kind != term.EventKey {
			continue
		}
		switch {
		case ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
			return inputForm{}, false, nil
		case ev.IsKey(term.KeyEnter):
			return form, true, nil
		default:
			form.onKeyPress(ev)
		}
	}
}

// focus is which field the keys go to.
type focus int

const (
	focusFirstName focus = iota
	focusLastName
	focusAge
)

// next moves the focus on, round to the first from the last.
func (f focus) next() focus { return (f + 1) % 3 }

// inputForm is the three fields and which of them has the focus.
//
// The labels and the focus stay out of the JSON, as ratatui's example marks
// them serde(skip): they are how the form is drawn, not what it collected.
// Being unexported is all that takes in Go.
type inputForm struct {
	focus     focus
	FirstName stringField `json:"first_name"`
	LastName  stringField `json:"last_name"`
	Age       ageField    `json:"age"`
}

func newInputForm() inputForm {
	return inputForm{
		focus:     focusFirstName,
		FirstName: stringField{label: "First Name"},
		LastName:  stringField{label: "Last Name"},
		Age:       ageField{label: "Age"},
	}
}

// onKeyPress moves the focus, or hands the key to the field that has it.
func (f *inputForm) onKeyPress(ev term.Event) {
	if ev.IsKey(term.KeyTab) {
		f.focus = f.focus.next()
		return
	}
	switch f.focus {
	case focusFirstName:
		f.FirstName.onKeyPress(ev)
	case focusLastName:
		f.LastName.onKeyPress(ev)
	case focusAge:
		f.Age.onKeyPress(ev)
	}
}

// render draws the three fields and puts the cursor at the end of whichever
// one has the focus.
func (f inputForm) render(frame *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(1), catatui.Length(1), catatui.Length(1),
	).Split(frame.Area())

	frame.RenderWidget(f.FirstName, rows[0])
	frame.RenderWidget(f.LastName, rows[1])
	frame.RenderWidget(f.Age, rows[2])

	var cursor catatui.Rect
	switch f.focus {
	case focusFirstName:
		cursor = rows[0].Offset(f.FirstName.cursorOffset())
	case focusLastName:
		cursor = rows[1].Offset(f.LastName.cursorOffset())
	case focusAge:
		cursor = rows[2].Offset(f.Age.cursorOffset())
	}
	frame.SetCursorPosition(cursor.AsPosition())
}

// stringField is a labelled line of text.
type stringField struct {
	label string
	Value string `json:"value"`
}

// onKeyPress appends a character or takes the last one back.
func (f *stringField) onKeyPress(ev term.Event) {
	switch {
	case ev.Key == term.KeyRune:
		f.Value += string(ev.Rune)
	case ev.IsKey(term.KeyBackspace):
		if runes := []rune(f.Value); len(runes) > 0 {
			f.Value = string(runes[:len(runes)-1])
		}
	}
}

// cursorOffset is where the cursor sits: past the label, its colon and space,
// and whatever has been typed.
func (f stringField) cursorOffset() catatui.Offset {
	return catatui.Offset{X: int32(labelWidth(f.label)) + int32(catatui.StringWidth(f.Value))}
}

func (f stringField) Render(area catatui.Rect, buf *catatui.Buffer) {
	renderField(area, buf, f.label, f.Value)
}

// ageField is a number between zero and 130, typed a digit at a time or nudged
// with the arrows.
type ageField struct {
	label string
	Value uint8 `json:"value"`
}

// maxAge is as old as the field will go.
const maxAge = 130

// onKeyPress takes digits, backspace, and the up and down keys.
func (f *ageField) onKeyPress(ev term.Event) {
	switch {
	case ev.Key == term.KeyRune && ev.Rune >= '0' && ev.Rune <= '9':
		// Typing a digit shifts the number along, unless that would take it
		// past the maximum, in which case the digit is dropped.
		value := int(f.Value)*10 + int(ev.Rune-'0')
		if value <= maxAge {
			f.Value = uint8(value)
		}
	case ev.IsKey(term.KeyBackspace):
		f.Value /= 10
	case ev.IsKey(term.KeyUp), ev.IsRune('k'):
		if f.Value < maxAge {
			f.Value++
		}
	case ev.IsKey(term.KeyDown), ev.IsRune('j'):
		if f.Value > 0 {
			f.Value--
		}
	}
}

func (f ageField) cursorOffset() catatui.Offset {
	return catatui.Offset{X: int32(labelWidth(f.label)) + int32(len(f.text()))}
}

func (f ageField) text() string { return strconv.Itoa(int(f.Value)) }

func (f ageField) Render(area catatui.Rect, buf *catatui.Buffer) {
	renderField(area, buf, f.label, f.text())
}

// renderField draws one row: a bold label, then the value beside it.
func renderField(area catatui.Rect, buf *catatui.Buffer, label, value string) {
	columns := catatui.HorizontalLayout(
		catatui.Length(labelWidth(label)),
		catatui.Fill(1),
	).Split(area)

	catatui.LineFromStyledString(label+": ",
		catatui.NewStyle().AddModifier(catatui.ModifierBold)).
		Render(columns[0], buf)
	catatui.LineFromString(value).Render(columns[1], buf)
}

// labelWidth is the room a label takes, colon and space included.
func labelWidth(label string) uint16 {
	return uint16(catatui.StringWidth(label)) + 2
}
