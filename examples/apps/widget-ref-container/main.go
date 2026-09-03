// Command widget-ref-container keeps widgets of different types in one
// container and renders them down the screen.
//
//	go run ./examples/apps/widget-ref-container
//
// Any key quits.
//
// ratatui needs a second trait for this. Widget::render consumes the widget, so
// a Box<dyn Widget> cannot be rendered — the value cannot be moved out of the
// box — and WidgetRef exists to render through a shared reference instead.
//
// Go has no such split: Render takes the receiver like any other method, so a
// catatui.Widget in an interface value renders as it stands and a slice of them
// is all the container needs.
//
// Port of examples/apps/widget-ref-container @ ratatui-v0.30.2
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

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	for {
		if err := terminal.Draw(render); err != nil {
			return err
		}
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		if ev.Kind == term.EventKey {
			return nil
		}
	}
}

func render(f *catatui.Frame) {
	container := stackContainer{
		direction: catatui.Vertical,
		children: []child{
			{greeting{}, catatui.Percentage(50)},
			{farewell{}, catatui.Percentage(50)},
		},
	}
	f.RenderWidget(container, f.Area())
}

// child is one widget in a stackContainer, with the constraint that decides how
// much room it gets.
type child struct {
	widget     catatui.Widget
	constraint catatui.Constraint
}

// stackContainer lays its children out along one axis and draws each in the
// area the solver gives it. The children have nothing in common but the Widget
// interface, which is the point of the example.
type stackContainer struct {
	direction catatui.Direction
	children  []child
}

func (s stackContainer) Render(area catatui.Rect, buf *catatui.Buffer) {
	constraints := make([]catatui.Constraint, len(s.children))
	for i, c := range s.children {
		constraints[i] = c.constraint
	}
	areas := catatui.NewLayout().
		Direction(s.direction).
		Constraints(constraints...).
		Split(area)

	for i, c := range s.children {
		c.widget.Render(areas[i], buf)
	}
}

// greeting and farewell are two widgets of different types, each drawing itself
// however it likes.
type greeting struct{}

func (greeting) Render(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Hello").Block(widgets.Bordered()).Render(area, buf)
}

type farewell struct{}

func (farewell) Render(area catatui.Rect, buf *catatui.Buffer) {
	widgets.NewParagraph("Goodbye").Block(widgets.Bordered()).Render(area, buf)
}
