// Command custom-widget is a button drawn from scratch, and three of them
// wired to the mouse.
//
//	go run ./examples/apps/custom-widget
//
// Left and right or h/l move the selection, space toggles the one selected,
// q quits. The mouse works too: moving over a button selects it, clicking
// toggles it.
//
// A widget is anything with a Render method, so this one is a struct with a
// label, a theme and a state, drawing itself straight into the buffer: a
// highlight along the top, a shadow along the bottom, and the label centred
// between them.
//
// Port of examples/apps/custom-widget @ ratatui-v0.30.2
package main

import (
	"fmt"
	"os"
	"strings"

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

	// Without this the mouse events never arrive, and the buttons only answer
	// to the keyboard.
	terminal, restore, err := term.Init(term.WithMouse())
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	a := &app{states: [3]buttonState{stateSelected, stateNormal, stateNormal}}

	for !a.quit {
		if err := terminal.Draw(a.render); err != nil {
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

// buttonWidth is how wide each button is drawn, and so where one ends and the
// next begins as far as the mouse is concerned.
const buttonWidth = 15

// app is the state of the three buttons and which one is selected.
type app struct {
	states   [3]buttonState
	selected int
	quit     bool
}

// handle applies one event, from the keyboard or the mouse.
func (a *app) handle(ev term.Event) {
	switch ev.Kind {
	case term.EventKey:
		a.handleKey(ev)
	case term.EventMouse:
		a.handleMouse(ev)
	}
}

func (a *app) handleKey(ev term.Event) {
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape), ev.IsCtrl('c'):
		a.quit = true
	case ev.IsRune('h'), ev.IsKey(term.KeyLeft):
		a.selectButton(max(a.selected-1, 0))
	case ev.IsRune('l'), ev.IsKey(term.KeyRight):
		a.selectButton(min(a.selected+1, len(a.states)-1))
	case ev.IsRune(' '):
		a.toggle()
	}
}

func (a *app) handleMouse(ev term.Event) {
	switch ev.MouseKind {
	case term.MouseMove:
		// The buttons are laid out from the left edge, one after another, so
		// which one the pointer is over is a division.
		a.selectButton(min(int(ev.X)/buttonWidth, len(a.states)-1))
	case term.MouseDown:
		if ev.Button == term.MouseButtonLeft {
			a.toggle()
		}
	}
}

// selectButton moves the selection, leaving an active button active: pressing
// a button and then moving away should not un-press it.
func (a *app) selectButton(index int) {
	if index == a.selected {
		return
	}
	if a.states[a.selected] != stateActive {
		a.states[a.selected] = stateNormal
	}
	a.selected = index
	if a.states[a.selected] != stateActive {
		a.states[a.selected] = stateSelected
	}
}

// toggle presses the selected button, or lets it back up.
func (a *app) toggle() {
	if a.states[a.selected] == stateActive {
		a.states[a.selected] = stateNormal
	} else {
		a.states[a.selected] = stateActive
	}
}

// render draws the title, the buttons and the help line.
func (a *app) render(f *catatui.Frame) {
	rows := catatui.VerticalLayout(
		catatui.Length(1),
		catatui.Max(3),
		catatui.Length(1),
		catatui.Min(0), // whatever is left over is not used
	).Split(f.Area())

	f.RenderWidget(widgets.NewParagraph("Custom Widget Example (mouse enabled)"), rows[0])
	a.renderButtons(f, rows[1])
	f.RenderWidget(widgets.NewParagraph("←/→: select, Space: toggle, q: quit"), rows[2])
}

func (a *app) renderButtons(f *catatui.Frame, area catatui.Rect) {
	widths := make([]catatui.Constraint, len(a.states))
	for i := range widths {
		widths[i] = catatui.Length(buttonWidth)
	}
	areas := catatui.HorizontalLayout(widths...).Flex(catatui.FlexStart).Split(area)

	for i, theme := range []theme{redTheme, greenTheme, blueTheme} {
		f.RenderWidget(button{
			label: buttonLabels[i],
			theme: theme,
			state: a.states[i],
		}, areas[i])
	}
}

var buttonLabels = [3]string{"Red", "Green", "Blue"}

// buttonState is how a button is drawn: idle, under the pointer, or pressed.
type buttonState uint8

const (
	stateNormal buttonState = iota
	stateSelected
	stateActive
)

// theme is the four colours a button is drawn from.
type theme struct {
	text       catatui.Color
	background catatui.Color
	highlight  catatui.Color
	shadow     catatui.Color
}

var (
	blueTheme = theme{
		text:       catatui.Rgb(16, 24, 48),
		background: catatui.Rgb(48, 72, 144),
		highlight:  catatui.Rgb(64, 96, 192),
		shadow:     catatui.Rgb(32, 48, 96),
	}
	redTheme = theme{
		text:       catatui.Rgb(48, 16, 16),
		background: catatui.Rgb(144, 48, 48),
		highlight:  catatui.Rgb(192, 64, 64),
		shadow:     catatui.Rgb(96, 32, 32),
	}
	greenTheme = theme{
		text:       catatui.Rgb(16, 48, 16),
		background: catatui.Rgb(48, 144, 48),
		highlight:  catatui.Rgb(64, 192, 64),
		shadow:     catatui.Rgb(32, 96, 32),
	}
)

// button is the custom widget: a label on a coloured slab, lit along one edge
// and shadowed along the other.
type button struct {
	label string
	theme theme
	state buttonState
}

// colors picks the four the current state calls for. Pressing a button swaps
// the highlight and the shadow over, which is what makes it look pushed in.
func (b button) colors() (background, text, shadow, highlight catatui.Color) {
	switch b.state {
	case stateSelected:
		return b.theme.highlight, b.theme.text, b.theme.shadow, b.theme.highlight
	case stateActive:
		return b.theme.background, b.theme.text, b.theme.highlight, b.theme.shadow
	default:
		return b.theme.background, b.theme.text, b.theme.shadow, b.theme.highlight
	}
}

// Render is all a widget needs. It draws straight into the buffer rather than
// composing other widgets, which is the point of the example.
func (b button) Render(area catatui.Rect, buf *catatui.Buffer) {
	if area.IsEmpty() {
		return
	}
	background, text, shadow, highlight := b.colors()
	buf.SetStyle(area, catatui.NewStyle().Bg(background).Fg(text))

	// The top and bottom edges are only drawn where there is room for them
	// and for the label between.
	if area.Height > 2 {
		buf.SetStringn(area.X, area.Y, strings.Repeat("▔", int(area.Width)), area.Width,
			catatui.NewStyle().Fg(highlight).Bg(background))
	}
	if area.Height > 1 {
		buf.SetStringn(area.X, area.Bottom()-1, strings.Repeat("▁", int(area.Width)), area.Width,
			catatui.NewStyle().Fg(shadow).Bg(background))
	}

	label := catatui.LineFromString(b.label)
	x := area.X + catatui.SatSub(area.Width, uint16(label.Width()))/2
	y := area.Y + catatui.SatSub(area.Height, 1)/2
	// The room left from x, not the whole width: a label that starts halfway
	// along must still stop at the button's right edge.
	buf.SetLine(x, y, label, catatui.SatSub(area.Right(), x))
}
