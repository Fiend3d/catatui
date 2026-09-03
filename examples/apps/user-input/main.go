// Command user-input shows how to handle keyboard input: an input box that
// records what is typed, and a list of the messages entered so far.
//
//	go run ./examples/apps/user-input
//
// Press e to start editing, Esc to stop, Enter to record a message, q to quit.
// While editing, the left and right arrows move the cursor and Backspace
// deletes the character before it.
//
// catatui does not handle text input for you, any more than ratatui does. This
// is what the smallest reasonable version of it looks like.
//
// Port of examples/apps/user-input @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// inputMode is whether keys type into the input box or drive the app.
type inputMode uint8

const (
	modeNormal inputMode = iota
	modeEditing
)

// app holds the state of the application.
//
// The input is kept as a []rune rather than ratatui's String, which is what
// lets this port handle multi-byte characters that ratatui's example gives up
// on: an index into runes cannot land in the middle of one.
type app struct {
	// input is the current value of the input box.
	input []rune
	// cursor is the position of the cursor in the input box, counted in
	// characters rather than bytes or columns.
	cursor int
	// mode is the current input mode.
	mode inputMode
	// messages is the history of recorded messages.
	messages []string
	// quit is set once the app is done.
	quit bool
}

func (a *app) moveCursorLeft()  { a.cursor = a.clampCursor(a.cursor - 1) }
func (a *app) moveCursorRight() { a.cursor = a.clampCursor(a.cursor + 1) }

// enterChar inserts a character at the cursor and steps over it.
func (a *app) enterChar(r rune) {
	a.input = append(a.input, 0)
	copy(a.input[a.cursor+1:], a.input[a.cursor:])
	a.input[a.cursor] = r
	a.moveCursorRight()
}

// deleteChar removes the character to the left of the cursor, if there is one.
func (a *app) deleteChar() {
	if a.cursor == 0 {
		return
	}
	a.input = append(a.input[:a.cursor-1], a.input[a.cursor:]...)
	a.moveCursorLeft()
}

func (a *app) clampCursor(pos int) int { return min(max(pos, 0), len(a.input)) }

// submitMessage files the input away and empties the box.
func (a *app) submitMessage() {
	a.messages = append(a.messages, string(a.input))
	a.input = a.input[:0]
	a.cursor = 0
}

// run draws the app and waits for a key between frames.
func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{}

	for !a.quit {
		if err := terminal.Draw(func(f *catatui.Frame) { a.render(f) }); err != nil {
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

// handle applies one event, which key it is depending on the mode.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	if ev.IsCtrl('c') {
		a.quit = true
		return
	}
	switch a.mode {
	case modeNormal:
		switch {
		case ev.IsRune('e'):
			a.mode = modeEditing
		case ev.IsRune('q'):
			a.quit = true
		}
	case modeEditing:
		switch {
		case ev.IsKey(term.KeyEnter):
			a.submitMessage()
		case ev.IsKey(term.KeyBackspace):
			a.deleteChar()
		case ev.IsKey(term.KeyLeft):
			a.moveCursorLeft()
		case ev.IsKey(term.KeyRight):
			a.moveCursorRight()
		case ev.IsKey(term.KeyEscape):
			a.mode = modeNormal
		case ev.Key == term.KeyRune:
			a.enterChar(ev.Rune)
		}
	}
}

// render draws the help line, the input box and the messages.
func (a *app) render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Length(3),
		catatui.Min(1),
	).Split(f.Area())
	helpArea, inputArea, messagesArea := rows[0], rows[1], rows[2]

	f.RenderWidget(widgets.NewParagraphFromText(a.helpMessage()), helpArea)
	f.RenderWidget(a.inputBox(), inputArea)
	a.placeCursor(f, inputArea)
	f.RenderWidget(a.messageList(), messagesArea)
}

// helpMessage says which keys do what in the current mode.
func (a *app) helpMessage() catatui.Text {
	bold := catatui.NewStyle().AddModifier(catatui.ModifierBold)

	if a.mode == modeNormal {
		return catatui.NewText(catatui.NewLine(
			catatui.NewSpan("Press "),
			catatui.NewStyledSpan("q", bold),
			catatui.NewSpan(" to exit, "),
			catatui.NewStyledSpan("e", bold),
			catatui.NewStyledSpan(" to start editing.", bold),
		)).Patch(catatui.NewStyle().AddModifier(catatui.ModifierRapidBlink))
	}
	return catatui.NewText(catatui.NewLine(
		catatui.NewSpan("Press "),
		catatui.NewStyledSpan("Esc", bold),
		catatui.NewSpan(" to stop editing, "),
		catatui.NewStyledSpan("Enter", bold),
		catatui.NewSpan(" to record the message"),
	))
}

// inputBox is the bordered box holding what has been typed.
func (a *app) inputBox() widgets.Paragraph {
	style := catatui.NewStyle()
	if a.mode == modeEditing {
		style = style.Fg(catatui.ColorYellow)
	}
	return widgets.NewParagraph(string(a.input)).
		Style(style).
		Block(widgets.Bordered().Title("Input"))
}

// placeCursor asks the terminal to leave the cursor in the input box after the
// frame is drawn. In normal mode nothing is asked for, and a Frame hides the
// cursor by default.
//
// The column is the width of the text before the cursor rather than the number
// of characters in it, so that a wide character moves the cursor two columns.
func (a *app) placeCursor(f *catatui.Frame, inputArea catatui.Rect) {
	if a.mode != modeEditing {
		return
	}
	width := uint16(catatui.StringWidth(string(a.input[:a.cursor])))
	f.SetCursorPosition(catatui.Position{
		// One column in from the border, plus whatever is typed already.
		X: catatui.SatAdd(inputArea.X, width+1),
		// One row down, from the border to the input line.
		Y: catatui.SatAdd(inputArea.Y, 1),
	})
}

// messageList is everything recorded so far, numbered.
func (a *app) messageList() widgets.List {
	items := make([]widgets.ListItem, len(a.messages))
	for i, m := range a.messages {
		items[i] = widgets.NewListItemFromLine(
			catatui.LineFromString(fmt.Sprintf("%d: %s", i, m)))
	}
	return widgets.NewList(items...).Block(widgets.Bordered().Title("Messages"))
}
