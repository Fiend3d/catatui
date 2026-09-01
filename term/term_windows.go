//go:build windows

package term

import (
	"fmt"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"golang.org/x/sys/windows"
)

// Console mode flags. golang.org/x/sys/windows does not export all of them.
const (
	enableProcessedInput         = 0x0001
	enableLineInput              = 0x0002
	enableEchoInput              = 0x0004
	enableWindowInput            = 0x0008
	enableMouseInput             = 0x0010
	enableVirtualTerminalInput   = 0x0200
	enableProcessedOutput        = 0x0001
	enableVirtualTerminalProcess = 0x0004
	disableNewlineAutoReturn     = 0x0008
)

// platformState holds the console modes to put back on restore.
type platformState struct {
	inHandle   windows.Handle
	outHandle  windows.Handle
	inMode     uint32
	outMode    uint32
	restoreIn  bool
	restoreOut bool
}

// enterRawMode puts the console into raw VT mode.
//
// Windows needs both halves configured explicitly: the output handle has to opt
// in to interpreting escape sequences at all, and the input handle has to be
// switched to VT input so that keys arrive as escape sequences rather than
// console records. Without the output flag every sequence catatui writes would
// be printed literally.
func enterRawMode(in, out *os.File) (*platformState, error) {
	st := &platformState{
		inHandle:  windows.Handle(in.Fd()),
		outHandle: windows.Handle(out.Fd()),
	}

	if err := windows.GetConsoleMode(st.outHandle, &st.outMode); err == nil {
		mode := st.outMode | enableProcessedOutput | enableVirtualTerminalProcess | disableNewlineAutoReturn
		if err := windows.SetConsoleMode(st.outHandle, mode); err != nil {
			return nil, fmt.Errorf("catatui/term: enabling virtual terminal output: %w", err)
		}
		st.restoreOut = true
	}

	if err := windows.GetConsoleMode(st.inHandle, &st.inMode); err == nil {
		mode := st.inMode
		mode &^= enableLineInput | enableEchoInput | enableProcessedInput | enableWindowInput | enableMouseInput
		mode |= enableVirtualTerminalInput
		if err := windows.SetConsoleMode(st.inHandle, mode); err != nil {
			return nil, fmt.Errorf("catatui/term: enabling virtual terminal input: %w", err)
		}
		st.restoreIn = true
	}

	if !st.restoreIn && !st.restoreOut {
		return nil, fmt.Errorf("catatui/term: not attached to a console")
	}
	return st, nil
}

// exitRawMode puts the console modes back.
func exitRawMode(st *platformState) error {
	if st == nil {
		return nil
	}
	var firstErr error
	note := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if st.restoreIn {
		note(windows.SetConsoleMode(st.inHandle, st.inMode))
	}
	if st.restoreOut {
		note(windows.SetConsoleMode(st.outHandle, st.outMode))
	}
	return firstErr
}

func terminalSize(f *os.File) (catatui.Size, error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(f.Fd()), &info); err != nil {
		return catatui.Size{}, fmt.Errorf("catatui/term: reading the console size: %w", err)
	}
	// The window rectangle is inclusive on both ends, so a 80-column window
	// reports Right-Left == 79.
	return catatui.Size{
		Width:  uint16(info.Window.Right - info.Window.Left + 1),
		Height: uint16(info.Window.Bottom - info.Window.Top + 1),
	}, nil
}

func terminalWindowSize(f *os.File) (catatui.WindowSize, error) {
	size, err := terminalSize(f)
	if err != nil {
		return catatui.WindowSize{}, err
	}
	// The Windows console API reports no pixel dimensions.
	return catatui.WindowSize{Columns: size}, nil
}

// watchResize reports terminal size changes.
//
// Windows has no SIGWINCH, and with the input handle in VT mode there are no
// WINDOW_BUFFER_SIZE_EVENT records to read either, so the size is polled. The
// interval is short enough to feel immediate while a window is being dragged
// and cheap enough to ignore: one console API call per tick.
func watchResize(f *os.File, onResize func(catatui.Size), stop <-chan struct{}) {
	pollResize(f, onResize, stop)
}

// resizePollInterval is how often the terminal size is checked on platforms
// with no resize signal. It is short enough to track a window being dragged and
// cheap enough to ignore.
const resizePollInterval = 50 * time.Millisecond

// pollResize reports size changes by checking the size on a ticker. It is the
// fallback for platforms without SIGWINCH.
func pollResize(f *os.File, onResize func(catatui.Size), stop <-chan struct{}) {
	last, err := terminalSize(f)
	if err != nil {
		return
	}
	ticker := time.NewTicker(resizePollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			size, err := terminalSize(f)
			if err != nil || size == last {
				continue
			}
			last = size
			onResize(size)
		}
	}
}
