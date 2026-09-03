// Command todo-list is a todo list whose items can be selected and ticked off.
//
//	go run ./examples/todo-list
//
// Arrow keys or h/j/k/l move the selection, left unselects, right or Enter
// toggles the status, g/G jump to the top and bottom, q quits.
//
// Port of examples/apps/todo-list @ ratatui-v0.30.2
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

// status is whether an item has been done.
type status uint8

const (
	statusTodo status = iota
	statusCompleted
)

// todoItem is one line of the list, with the longer text shown underneath it
// while it is selected.
type todoItem struct {
	status status
	todo   string
	info   string
}

// todoList is the items and the ListState that remembers which one is
// selected. Keeping the state around between frames is what a stateful widget
// is for: the widget is rebuilt every frame, the selection and scroll are not.
type todoList struct {
	items []todoItem
	state widgets.ListState
}

// app is the whole program.
type app struct {
	shouldExit bool
	todoList   todoList
}

// newApp builds the app with the list ratatui's example ships with, with its
// two Rust jokes turned into Go ones.
func newApp() *app {
	return &app{todoList: todoList{items: []todoItem{
		{statusTodo, "Rewrite everything with Go!",
			"I can't hold my inner voice. He tells me to rewrite the complete universe with Go"},
		{statusCompleted, "Rewrite all of your tui apps with catatui",
			"Yes, you heard that right. Go and replace your tui with catatui."},
		{statusTodo, "Pet your cat",
			"Minnak loves to be pet by you! Don't forget to pet and give some treats!"},
		{statusTodo, "Walk with your dog",
			"Max is bored, go walk with him!"},
		{statusCompleted, "Pay the bills",
			"Pay the train subscription!!!"},
		{statusCompleted, "Refactor list example",
			"If you see this info that means I completed this task!"},
	}}}
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

	a := newApp()

	for !a.shouldExit {
		// The app is a widget in its own right, as in ratatui, so drawing a
		// frame is one RenderWidget call.
		if err := terminal.Draw(func(f *catatui.Frame) {
			f.RenderWidget(a, f.Area())
		}); err != nil {
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

// handle applies one event.
func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.shouldExit = true
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.todoList.state.SelectNone()
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.todoList.state.SelectNext()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.todoList.state.SelectPrevious()
	case ev.IsRune('g'), ev.IsKey(term.KeyHome):
		a.todoList.state.SelectFirst()
	case ev.IsRune('G'), ev.IsKey(term.KeyEnd):
		a.todoList.state.SelectLast()
	case ev.IsRune('l'), ev.IsKey(term.KeyRight), ev.IsKey(term.KeyEnter):
		a.toggleStatus()
	}
}

// toggleStatus ticks the selected item off, or puts it back.
func (a *app) toggleStatus() {
	i, ok := a.todoList.state.Selected()
	if !ok || i >= len(a.todoList.items) {
		return
	}
	if a.todoList.items[i].status == statusCompleted {
		a.todoList.items[i].status = statusTodo
	} else {
		a.todoList.items[i].status = statusCompleted
	}
}
