// Command tracing logs to a file while the terminal is given over to the UI,
// and shows the last few events it handled.
//
//	go run ./examples/apps/tracing
//	CATATUI_LOG=DEBUG-4 go run ./examples/apps/tracing   # per-frame logs too
//
// Press q to quit, then read tracing.log.
//
// A program that owns the screen cannot print its logs: anything written to
// stdout lands in the middle of a frame and is wiped by the next redraw. So
// they go to a file instead. ratatui's example does that with the tracing
// crate; this one uses log/slog from the standard library, which is why it
// needs no dependency to do it.
//
// Port of examples/apps/tracing @ ratatui-v0.30.2
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// logPath is where the log goes, in the working directory, as in ratatui's.
const logPath = "tracing.log"

// levelTrace is finer than slog's own lowest level, for the things that happen
// every frame and would bury everything else. slog has no name for it, so it
// is asked for by arithmetic: CATATUI_LOG=DEBUG-4.
const levelTrace = slog.LevelDebug - 4

// recentEvents is how many events the UI keeps on screen.
const recentEvents = 10

// redrawInterval is how long the loop waits for an event before drawing
// anyway, so the frame counter keeps moving while nothing happens.
const redrawInterval = 100 * time.Millisecond

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("See the " + logPath + " file for the logs")
}

func run() error {
	defer term.RecoverAndRestore()

	closeLog, err := initLogging()
	if err != nil {
		return err
	}
	defer closeLog()

	slog.Info("starting the tracing example", "log_level", logLevel())
	defer slog.Info("exiting the tracing example")

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	var recent []term.Event
	for !shouldExit(recent) {
		next, ok := handleEvents(events, recent)
		if !ok {
			return events.Err()
		}
		recent = next
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, recent) }); err != nil {
			return err
		}
	}
	return nil
}

// shouldExit reports whether q has been pressed, by looking at what has been
// recorded rather than at the key directly, which is ratatui's arrangement:
// the event is on screen for a frame before it takes effect.
func shouldExit(recent []term.Event) bool {
	for _, ev := range recent {
		if ev.IsRune('q') {
			return true
		}
	}
	return false
}

// handleEvents waits up to redrawInterval for one event and records it,
// keeping the most recent few. It reports false once the reader is done.
func handleEvents(events *term.EventReader, recent []term.Event) ([]term.Event, bool) {
	timeout := time.NewTimer(redrawInterval)
	defer timeout.Stop()

	select {
	case ev, ok := <-events.Events():
		if !ok {
			return recent, false
		}
		slog.Debug("event", "event", describe(ev))
		recent = append([]term.Event{ev}, recent...)
	case <-timeout.C:
	}
	return recent[:min(len(recent), recentEvents)], true
}

// render draws the recorded events in a block.
func render(f *catatui.Frame, recent []term.Event) {
	// Only visible with CATATUI_LOG=DEBUG-4: one line per frame would drown
	// out everything else at the default level.
	slog.Log(context.Background(), levelTrace, "render",
		"frame_count", f.Count(), "event_count", len(recent))

	lines := make([]string, len(recent))
	for i, ev := range recent {
		lines[i] = describe(ev)
	}

	f.RenderWidget(
		widgets.NewParagraph(strings.Join(lines, "\n")).
			Block(widgets.Bordered().Title("Tracing example. Press 'q' to quit.")),
		f.Area())
}

// initLogging points the default logger at the file and returns a function
// that closes it. Nothing is buffered, so a log line survives a panic.
func initLogging() (func(), error) {
	file, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", logPath, err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: logLevel(),
	})))
	return func() { _ = file.Close() }, nil
}

// logLevel reads CATATUI_LOG, which takes the names slog prints — DEBUG, INFO,
// WARN, ERROR — and the offsets from them, so DEBUG-4 reaches levelTrace.
// Anything unreadable, the variable being unset included, means debug.
func logLevel() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(os.Getenv("CATATUI_LOG"))); err != nil {
		return slog.LevelDebug
	}
	return level
}

// describe is one short line for an event. catatui's Event is a struct with a
// field per kind rather than an enum, so this says which fields matter.
func describe(ev term.Event) string {
	switch ev.Kind {
	case term.EventKey:
		return "key " + keyName(ev) + modifiers(ev.Mods)
	case term.EventMouse:
		return fmt.Sprintf("mouse %s%s at %d,%d%s",
			mouseKindNames[ev.MouseKind], mouseButtonNames[ev.Button],
			ev.X, ev.Y, modifiers(ev.Mods))
	case term.EventResize:
		return fmt.Sprintf("resize to %dx%d", ev.Size.Width, ev.Size.Height)
	case term.EventPaste:
		return fmt.Sprintf("paste %q", ev.Text)
	case term.EventFocus:
		if ev.Focused {
			return "focus gained"
		}
		return "focus lost"
	}
	return "unknown event"
}

// keyName names the key, quoting it when it is a character.
func keyName(ev term.Event) string {
	if ev.Key == term.KeyRune {
		return fmt.Sprintf("%q", ev.Rune)
	}
	if name, ok := keyNames[ev.Key]; ok {
		return name
	}
	return fmt.Sprintf("key(%d)", ev.Key)
}

// modifiers is the held modifiers in brackets, or nothing when none are.
func modifiers(mods term.Modifiers) string {
	if mods == 0 {
		return ""
	}
	return " [" + mods.String() + "]"
}

var keyNames = map[term.KeyCode]string{
	term.KeyEnter: "Enter", term.KeyEscape: "Esc", term.KeyBackspace: "Backspace",
	term.KeyTab: "Tab", term.KeyBackTab: "BackTab", term.KeyDelete: "Delete",
	term.KeyInsert: "Insert", term.KeyLeft: "Left", term.KeyRight: "Right",
	term.KeyUp: "Up", term.KeyDown: "Down", term.KeyHome: "Home", term.KeyEnd: "End",
	term.KeyPageUp: "PageUp", term.KeyPageDown: "PageDown",
	term.KeyF1: "F1", term.KeyF2: "F2", term.KeyF3: "F3", term.KeyF4: "F4",
	term.KeyF5: "F5", term.KeyF6: "F6", term.KeyF7: "F7", term.KeyF8: "F8",
	term.KeyF9: "F9", term.KeyF10: "F10", term.KeyF11: "F11", term.KeyF12: "F12",
}

var mouseKindNames = map[term.MouseKind]string{
	term.MouseDown: "down", term.MouseUp: "up", term.MouseDrag: "drag",
	term.MouseMove: "move", term.MouseScrollUp: "scroll up",
	term.MouseScrollDown: "scroll down",
}

var mouseButtonNames = map[term.MouseButton]string{
	term.MouseButtonNone: "", term.MouseButtonLeft: " left",
	term.MouseButtonMiddle: " middle", term.MouseButtonRight: " right",
}
