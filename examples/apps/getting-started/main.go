// Command getting-started is Purrfect Day, the app built in docs/getting-started.md.
//
//	go run ./examples/apps/getting-started
//
// Up/Down or j/k select, Space/Enter toggle, r resets, q/Esc/Ctrl+C quit.
package main

import (
	"fmt"
	"os"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/palette/tailwind"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

type task struct {
	name string
	done bool
}

// Keep data between frames; rebuild the widgets when drawing each frame.
type app struct {
	tasks []task
	list  widgets.ListState
	quit  bool
}

func newApp() *app {
	return &app{
		tasks: []task{{name: "Serve breakfast"}, {name: "Play together"}, {name: "Prepare a cozy nap"}},
		list:  widgets.NewListState().WithSelected(0),
	}
}

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
	a := newApp()
	for !a.quit {
		if err := terminal.Draw(a.draw); err != nil {
			return err
		}
		// Wait for input or a resize instead of repeatedly drawing an idle app.
		ev, ok := <-events.Events()
		if !ok {
			return events.Err()
		}
		a.handle(ev)
	}
	return nil
}

func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return // The loop still redraws after resize events.
	}
	i, _ := a.list.Selected()
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsKey(term.KeyUp), ev.IsRune('k'):
		a.list.Select(max(0, i-1))
	case ev.IsKey(term.KeyDown), ev.IsRune('j'):
		a.list.Select(min(len(a.tasks)-1, i+1))
	case ev.IsRune(' '), ev.IsKey(term.KeyEnter):
		a.tasks[i].done = !a.tasks[i].done
	case ev.IsRune('r'):
		*a = *newApp()
	}
}

func (a *app) completed() int {
	n := 0
	for _, task := range a.tasks {
		if task.done {
			n++
		}
	}
	return n
}

var (
	baseStyle = catatui.NewStyle().Fg(tailwind.Amber.C100).Bg(tailwind.Slate.C950)
	cyanStyle = catatui.NewStyle().Fg(tailwind.Cyan.C300)
	pinkStyle = catatui.NewStyle().Fg(tailwind.Pink.C300)
	mintStyle = catatui.NewStyle().Fg(tailwind.Emerald.C300)
)

func panel(title string, style catatui.Style) widgets.Block {
	return widgets.Bordered().Title(" " + title + " ").
		BorderType(widgets.BorderRounded).BorderStyle(style).
		Padding(widgets.HorizontalPadding(1))
}

func (a *app) draw(f *catatui.Frame) {
	area := f.Area()
	f.Buffer().SetStyle(area, baseStyle)
	if area.Width < 40 || area.Height < 18 {
		f.RenderWidget(widgets.NewParagraph("Purrfect Day\nResize to at least 40 x 18.\nq / Esc / Ctrl+C: quit").
			Wrap(widgets.Wrap{Trim: true}), area)
		return
	}

	rows := catatui.VerticalLayout(
		catatui.Length(2), catatui.Fill(1), catatui.Length(3), catatui.Length(2),
	).Split(area)
	f.RenderWidget(widgets.NewParagraph("PURRFECT DAY\nLittle acts of care. One very happy cat.").
		Style(cyanStyle.AddModifier(catatui.ModifierBold)).Centered(), rows[0])

	panes := catatui.HorizontalLayout(catatui.Percentage(50), catatui.Fill(1)).
		Spacing(catatui.Space(1)).Split(rows[1])
	if area.Width < 72 {
		panes = catatui.VerticalLayout(catatui.Length(5), catatui.Fill(1)).Split(rows[1])
	}
	a.renderTasks(f, panes[0])
	a.renderCat(f, panes[1])
	a.renderHappiness(f, rows[2])
	f.RenderWidget(widgets.NewParagraph("Up/Down or j/k: move  Space/Enter: tick\nr: reset  q/Esc/Ctrl+C: quit").
		Centered(), rows[3])
}

func (a *app) renderTasks(f *catatui.Frame, area catatui.Rect) {
	items := make([]widgets.ListItem, len(a.tasks))
	for i, task := range a.tasks {
		mark, style := "[ ] ", baseStyle
		if task.done {
			mark, style = "[x] ", mintStyle
		}
		items[i] = widgets.NewListItem(mark + task.name).Style(style)
	}
	list := widgets.NewList(items...).Block(panel("TODAY'S LITTLE WINS", cyanStyle)).
		HighlightSymbol("> ").HighlightSpacing(widgets.HighlightSpacingAlways).
		HighlightStyle(catatui.NewStyle().Bg(tailwind.Slate.C800).AddModifier(catatui.ModifierBold))
	catatui.RenderStatefulWidget(list, area, f.Buffer(), &a.list)
}

func (a *app) renderCat(f *catatui.Frame, area catatui.Rect) {
	face, message := "( o.o )", "A little care goes a long way."
	if a.completed() == len(a.tasks) {
		face, message = "( ^.^ )", "One happy cat. Nicely done!"
	}
	cat := ` /\_/\` + "\n" + face + "\n > ^ <\n" + message
	f.RenderWidget(widgets.NewParagraph(cat).Centered().
		Block(panel("YOUR TINY COMPANION", pinkStyle)).Style(pinkStyle), area)
}

func (a *app) renderHappiness(f *catatui.Frame, area catatui.Rect) {
	n := a.completed()
	f.RenderWidget(widgets.NewGauge().
		Block(panel("HAPPINESS", mintStyle)).
		Ratio(float64(n)/float64(len(a.tasks))).
		Label(fmt.Sprintf("%d / %d little wins", n, len(a.tasks))).
		GaugeStyle(mintStyle.Bg(tailwind.Slate.C800)), area)
}
