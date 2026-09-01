//go:build !windows

package term

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fiend3d/catatui"
	xterm "golang.org/x/term"
)

// platformState holds the terminal attributes to put back on restore.
type platformState struct {
	fd    int
	state *xterm.State
}

// enterRawMode puts the terminal into raw mode: no line buffering, no echo, no
// signal generation from control characters.
func enterRawMode(in, _ *os.File) (*platformState, error) {
	fd := int(in.Fd())
	if !xterm.IsTerminal(fd) {
		return nil, fmt.Errorf("catatui/term: not attached to a terminal")
	}
	state, err := xterm.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("catatui/term: entering raw mode: %w", err)
	}
	return &platformState{fd: fd, state: state}, nil
}

// exitRawMode puts the terminal attributes back.
func exitRawMode(st *platformState) error {
	if st == nil {
		return nil
	}
	return xterm.Restore(st.fd, st.state)
}

func terminalSize(f *os.File) (catatui.Size, error) {
	w, h, err := xterm.GetSize(int(f.Fd()))
	if err != nil {
		return catatui.Size{}, fmt.Errorf("catatui/term: reading the terminal size: %w", err)
	}
	return catatui.Size{Width: uint16(w), Height: uint16(h)}, nil
}

func terminalWindowSize(f *os.File) (catatui.WindowSize, error) {
	size, err := terminalSize(f)
	if err != nil {
		return catatui.WindowSize{}, err
	}
	return catatui.WindowSize{Columns: size}, nil
}

// watchResize reports terminal size changes, driven by SIGWINCH.
func watchResize(f *os.File, onResize func(catatui.Size), stop <-chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	defer signal.Stop(ch)

	for {
		select {
		case <-stop:
			return
		case <-ch:
			if size, err := terminalSize(f); err == nil {
				onResize(size)
			}
		}
	}
}
