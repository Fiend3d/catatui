package main

import (
	"encoding/json"
	"testing"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

func keyRune(r rune) term.Event {
	return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
}

func key(k term.KeyCode) term.Event {
	return term.Event{Kind: term.EventKey, Key: k}
}

// TestRender draws the form empty and filled, with the focus on each field, at
// sizes from nothing to bigger than a screen. Rendering outside the area given
// panics in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	forms := []inputForm{newInputForm(), filledForm()}
	for i, form := range forms {
		for _, where := range []focus{focusFirstName, focusLastName, focusAge} {
			form.focus = where
			for _, size := range sizes {
				terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
				if err != nil {
					t.Fatalf("form %d, focus %d, %dx%d: %v", i, where, size[0], size[1], err)
				}
				if err := terminal.Draw(form.render); err != nil {
					t.Fatalf("form %d, focus %d, %dx%d: %v", i, where, size[0], size[1], err)
				}
			}
		}
	}
}

func filledForm() inputForm {
	form := newInputForm()
	form.FirstName.Value = "Ada"
	form.LastName.Value = "Lovelace"
	form.Age.Value = 36
	return form
}

// TestTabCyclesTheFocus checks Tab reaches every field and comes back round.
func TestTabCyclesTheFocus(t *testing.T) {
	form := newInputForm()
	seen := map[focus]bool{form.focus: true}
	for range 3 {
		form.onKeyPress(key(term.KeyTab))
		seen[form.focus] = true
	}
	if len(seen) != 3 {
		t.Errorf("Tab reached %d of the 3 fields", len(seen))
	}
	if form.focus != focusFirstName {
		t.Errorf("a full cycle ended on field %d, want it back at the first", form.focus)
	}
}

// TestKeysGoToTheFocusedField checks typing lands in the field with the focus
// and nowhere else.
func TestKeysGoToTheFocusedField(t *testing.T) {
	form := newInputForm()
	for _, r := range "Ada" {
		form.onKeyPress(keyRune(r))
	}
	form.onKeyPress(key(term.KeyTab))
	for _, r := range "Lovelace" {
		form.onKeyPress(keyRune(r))
	}
	if form.FirstName.Value != "Ada" || form.LastName.Value != "Lovelace" {
		t.Errorf("typed %q and %q", form.FirstName.Value, form.LastName.Value)
	}

	form.onKeyPress(key(term.KeyBackspace))
	if form.LastName.Value != "Lovelac" {
		t.Errorf("backspace left %q", form.LastName.Value)
	}
}

// TestAgeFieldStaysInRange checks the age takes digits up to its maximum,
// refuses the ones that would take it past, and stops at both ends.
func TestAgeFieldStaysInRange(t *testing.T) {
	field := &ageField{label: "Age"}
	for _, r := range "36" {
		field.onKeyPress(keyRune(r))
	}
	if field.Value != 36 {
		t.Errorf("typing 36 gave %d", field.Value)
	}
	// 366 is past the maximum, so the last digit is dropped rather than
	// wrapping the uint8 round to 110.
	field.onKeyPress(keyRune('6'))
	if field.Value != 36 {
		t.Errorf("a digit past the maximum changed the value to %d", field.Value)
	}
	field.onKeyPress(key(term.KeyBackspace))
	if field.Value != 3 {
		t.Errorf("backspace left %d", field.Value)
	}

	field.Value = maxAge
	field.onKeyPress(key(term.KeyUp))
	if field.Value != maxAge {
		t.Errorf("incrementing past the maximum gave %d", field.Value)
	}
	field.Value = 0
	field.onKeyPress(key(term.KeyDown))
	if field.Value != 0 {
		t.Errorf("decrementing below zero gave %d", field.Value)
	}
}

// TestJSONHoldsOnlyTheValues checks the labels and the focus stay out of what
// is printed on submit.
func TestJSONHoldsOnlyTheValues(t *testing.T) {
	encoded, err := json.Marshal(filledForm())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"first_name":{"value":"Ada"},"last_name":{"value":"Lovelace"},"age":{"value":36}}`
	if got := string(encoded); got != want {
		t.Errorf("form encodes as\n%s\nwant\n%s", got, want)
	}
}
